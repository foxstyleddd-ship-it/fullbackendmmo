package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Économie (valeurs "démo" — tunées pour que des drops arrivent pendant une
// session/un TikTok. En vrai, une vue de pub rapporte ~0,001–0,01 € via une
// régie rewarded/offerwall, et la cagnotte se remplit beaucoup plus lentement).
// Tout est en micro-euros (1 € = 1_000_000) pour éviter les flottants.
// ---------------------------------------------------------------------------
const (
	Micro       = 1_000_000
	GamePrice   = 69_990_000 // 69,99 € le GTA6
	FeeRate     = 0.05       // on récupère 5 % des bénéfices
	RewardMin   = 100_000    // 0,10 €  par pub (démo)
	RewardMax   = 800_000    // 0,80 €  par pub (démo)
	AdDuration  = 5          // secondes à "regarder" avant de pouvoir valider
	Cooldown    = 3 * time.Second
	RefBonusYou = 2 // tickets offerts au nouvel inscrit parrainé
	RefBonusRef = 5 // tickets offerts au parrain
)

type User struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Salt       string    `json:"-"`
	PassHash   string    `json:"-"`
	Weight     int       `json:"weight"`     // "poids" = tickets de tirage (actions vérifiées)
	AdsWatched int       `json:"adsWatched"`
	Earned     int64     `json:"earned"`     // micro-€ générés par ce compte
	RefCode    string    `json:"refCode"`
	ReferredBy string    `json:"-"`
	Referrals  int       `json:"referrals"`
	Wins       int       `json:"wins"`
	CreatedAt  time.Time `json:"createdAt"`
	LastWatch  time.Time `json:"-"`
	IP         string    `json:"-"`
}

type Winner struct {
	Username string    `json:"username"`
	GameNo   int       `json:"gameNo"`
	At       time.Time `json:"at"`
}

type pendingAd struct {
	UserID      string
	StartAt     time.Time
	Sponsor     string
	NeedCaptcha bool
	CaptchaAns  int
}

type snapshot struct {
	Users      map[string]*User `json:"users"`
	Pot        int64            `json:"pot"`
	Fee        int64            `json:"fee"`
	GamesGiven int              `json:"gamesGiven"`
	Winners    []Winner         `json:"winners"`
	Slots      []*AdSlot        `json:"slots"`
}

type Store struct {
	mu         sync.RWMutex
	users      map[string]*User
	byName     map[string]string // username(lower) -> id
	byRef      map[string]string // refcode -> id
	sessions   map[string]string // token -> userID
	pending    map[string]*pendingAd
	pot        int64
	fee        int64
	gamesGiven int
	winners    []Winner
	slots      []*AdSlot // emplacements pub du hub (loués par les marques)
	path       string
}

func NewStore(path string) *Store {
	s := &Store{
		users:    map[string]*User{},
		byName:   map[string]string{},
		byRef:    map[string]string{},
		sessions: map[string]string{},
		pending:  map[string]*pendingAd{},
		path:     path,
	}
	s.load()
	if len(s.slots) == 0 {
		s.slots = seedSlots()
	}
	return s
}

// --- helpers ---------------------------------------------------------------

func token(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func randInt(max int64) int64 {
	if max <= 0 {
		return 0
	}
	n, _ := rand.Int(rand.Reader, big.NewInt(max))
	return n.Int64()
}

func hashPw(salt, pw string) string {
	sum := sha256.Sum256([]byte(salt + ":" + pw))
	return hex.EncodeToString(sum[:])
}

func refFromName(name string) string {
	base := strings.ToUpper(strings.NewReplacer(" ", "", "-", "", "_", "").Replace(name))
	if len(base) > 5 {
		base = base[:5]
	}
	return base + strings.ToUpper(token(2))
}

// --- persistence -----------------------------------------------------------

func (s *Store) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var snap snapshot
	if json.Unmarshal(data, &snap) != nil {
		return
	}
	if snap.Users != nil {
		s.users = snap.Users
	}
	s.pot, s.fee, s.gamesGiven, s.winners, s.slots = snap.Pot, snap.Fee, snap.GamesGiven, snap.Winners, snap.Slots
	for id, u := range s.users {
		s.byName[strings.ToLower(u.Username)] = id
		if u.RefCode != "" {
			s.byRef[u.RefCode] = id
		}
	}
}

// persist must be called with the lock held.
func (s *Store) persist() {
	snap := snapshot{Users: s.users, Pot: s.pot, Fee: s.fee, GamesGiven: s.gamesGiven, Winners: s.winners, Slots: s.slots}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return
	}
	tmp := s.path + ".tmp"
	if os.WriteFile(tmp, data, 0o644) == nil {
		os.Rename(tmp, s.path)
	}
}

// --- auth ------------------------------------------------------------------

var (
	ErrTaken   = errors.New("ce pseudo est déjà pris")
	ErrBadCred = errors.New("pseudo ou mot de passe incorrect")
	ErrInput   = errors.New("pseudo : 3 à 20 caractères · mot de passe : 4 caractères minimum")
)

func (s *Store) Register(name, pw, refCode, ip string) (*User, string, error) {
	name = strings.TrimSpace(name)
	if len(name) < 3 || len(pw) < 4 || len(name) > 20 {
		return nil, "", ErrInput
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byName[strings.ToLower(name)]; ok {
		return nil, "", ErrTaken
	}
	salt := token(8)
	u := &User{
		ID:        token(8),
		Username:  name,
		Salt:      salt,
		PassHash:  hashPw(salt, pw),
		RefCode:   refFromName(name),
		CreatedAt: time.Now(),
		IP:        ip,
	}
	// Parrainage : actions vérifiées => bonus de poids des deux côtés.
	if refCode != "" {
		if rid, ok := s.byRef[strings.ToUpper(refCode)]; ok && rid != u.ID {
			if ref := s.users[rid]; ref != nil {
				ref.Weight += RefBonusRef
				ref.Referrals++
				u.ReferredBy = rid
				u.Weight += RefBonusYou
			}
		}
	}
	s.users[u.ID] = u
	s.byName[strings.ToLower(name)] = u.ID
	s.byRef[u.RefCode] = u.ID
	tok := token(16)
	s.sessions[tok] = u.ID
	s.persist()
	return u, tok, nil
}

func (s *Store) Login(name, pw string) (*User, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byName[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return nil, "", ErrBadCred
	}
	u := s.users[id]
	if u == nil || u.PassHash != hashPw(u.Salt, pw) {
		return nil, "", ErrBadCred
	}
	tok := token(16)
	s.sessions[tok] = u.ID
	return u, tok, nil
}

func (s *Store) Logout(tok string) {
	s.mu.Lock()
	delete(s.sessions, tok)
	s.mu.Unlock()
}

func (s *Store) UserByToken(tok string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if id, ok := s.sessions[tok]; ok {
		return s.users[id]
	}
	return nil
}

// --- ad flow (anti-fraude démo : durée mini + cooldown + captcha) ----------

var sponsors = []string{
	"Vinewood Premium Detailing", "Maze Bank — ouvre ton compte",
	"Sprunk — la boisson officielle", "Los Santos Customs",
	"eCola Zéro", "Ammu-Nation", "Lifeinvader Pro", "Bean Machine Café",
}

// pickSponsor privilégie une marque réellement présente dans le hub, sinon
// retombe sur la liste par défaut. Doit être appelé avec le lock tenu.
func (s *Store) pickSponsor() string {
	var filled []string
	for _, sl := range s.slots {
		if sl.Filled {
			filled = append(filled, sl.Brand)
		}
	}
	if len(filled) > 0 {
		return filled[randInt(int64(len(filled)))]
	}
	return sponsors[randInt(int64(len(sponsors)))]
}

type AdStart struct {
	AdID        string `json:"adId"`
	Duration    int    `json:"duration"`
	Sponsor     string `json:"sponsor"`
	NeedCaptcha bool   `json:"needCaptcha"`
	Captcha     string `json:"captcha,omitempty"`
}

func (s *Store) StartAd(userID string) (*AdStart, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := s.users[userID]
	if u == nil {
		return nil, ErrBadCred
	}
	if d := time.Since(u.LastWatch); d < Cooldown {
		return nil, errors.New("patiente quelques secondes avant la prochaine pub")
	}
	adID := token(8)
	sponsor := s.pickSponsor()
	pa := &pendingAd{UserID: userID, StartAt: time.Now(), Sponsor: sponsor}
	out := &AdStart{AdID: adID, Duration: AdDuration, Sponsor: sponsor}
	// captcha "humain" un coup sur trois pour ne pas être lourd
	if u.AdsWatched%3 == 2 {
		a, b := int(randInt(8))+1, int(randInt(8))+1
		pa.NeedCaptcha = true
		pa.CaptchaAns = a + b
		out.NeedCaptcha = true
		out.Captcha = formatCaptcha(a, b)
	}
	s.pending[adID] = pa
	return out, nil
}

func formatCaptcha(a, b int) string {
	return itoa(a) + " + " + itoa(b) + " = ?"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

type WatchResult struct {
	Reward   int64    `json:"reward"`
	Weight   int      `json:"weight"`
	Earned   int64    `json:"earned"`
	NewWins  []Winner `json:"newWins"`
}

// CompleteAd valide une pub regardée et crédite la cagnotte. Renvoie les
// éventuels drops déclenchés (cagnotte >= prix d'un GTA6).
func (s *Store) CompleteAd(userID, adID string, captcha int) (*WatchResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pa := s.pending[adID]
	if pa == nil || pa.UserID != userID {
		return nil, errors.New("pub introuvable")
	}
	delete(s.pending, adID)
	if time.Since(pa.StartAt) < time.Duration(AdDuration)*time.Second {
		return nil, errors.New("tu n'as pas regardé la pub jusqu'au bout 👀")
	}
	if pa.NeedCaptcha && captcha != pa.CaptchaAns {
		return nil, errors.New("vérification anti-bot échouée")
	}
	u := s.users[userID]
	if u == nil {
		return nil, ErrBadCred
	}
	if time.Since(u.LastWatch) < Cooldown {
		return nil, errors.New("trop rapide, patiente un peu")
	}

	reward := RewardMin + randInt(RewardMax-RewardMin)
	fee := int64(float64(reward) * FeeRate)
	u.LastWatch = time.Now()
	u.AdsWatched++
	u.Weight++ // chaque pub vérifiée = +1 ticket (le "poids")
	u.Earned += reward
	s.fee += fee
	s.pot += reward - fee

	res := &WatchResult{NewWins: s.runDrops()}
	s.persist()
	res.Reward, res.Weight, res.Earned = reward, u.Weight, u.Earned
	return res, nil
}

// runDrops vide la cagnotte tant qu'elle dépasse le prix d'un GTA6 et tire un
// gagnant pondéré à chaque fois. Doit être appelé avec le lock tenu.
func (s *Store) runDrops() []Winner {
	var wins []Winner
	for s.pot >= GamePrice {
		s.pot -= GamePrice
		s.gamesGiven++
		w := s.drawWinner()
		win := Winner{Username: w, GameNo: s.gamesGiven, At: time.Now()}
		s.winners = append([]Winner{win}, s.winners...)
		if len(s.winners) > 50 {
			s.winners = s.winners[:50]
		}
		wins = append(wins, win)
	}
	return wins
}

// drawWinner : tirage pondéré par le poids parmi ceux qui ont regardé.
// Doit être appelé avec le lock tenu.
func (s *Store) drawWinner() string {
	var total int
	for _, u := range s.users {
		if u.Weight > 0 {
			total += u.Weight
		}
	}
	if total == 0 {
		return "—"
	}
	r := int(randInt(int64(total)))
	// ordre déterministe pour le tirage
	ids := make([]string, 0, len(s.users))
	for id := range s.users {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		u := s.users[id]
		if u.Weight <= 0 {
			continue
		}
		if r < u.Weight {
			u.Wins++
			return u.Username
		}
		r -= u.Weight
	}
	return "—"
}

// --- vues ------------------------------------------------------------------

type LeaderRow struct {
	Username   string `json:"username"`
	Weight     int    `json:"weight"`
	AdsWatched int    `json:"adsWatched"`
	Wins       int    `json:"wins"`
}

type State struct {
	Pot         int64       `json:"pot"`
	Price       int64       `json:"price"`
	Progress    float64     `json:"progress"`
	Fee         int64       `json:"fee"`
	GamesGiven  int         `json:"gamesGiven"`
	TotalUsers  int         `json:"totalUsers"`
	TotalWeight int         `json:"totalWeight"`
	Winners     []Winner    `json:"winners"`
	Leaderboard []LeaderRow `json:"leaderboard"`
}

func (s *Store) State() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st := State{
		Pot:        s.pot,
		Price:      GamePrice,
		Fee:        s.fee,
		GamesGiven: s.gamesGiven,
		TotalUsers: len(s.users),
		Winners:    s.winners,
	}
	if len(st.Winners) > 12 {
		st.Winners = st.Winners[:12]
	}
	st.Progress = float64(s.pot) / float64(GamePrice) * 100

	rows := make([]LeaderRow, 0, len(s.users))
	for _, u := range s.users {
		st.TotalWeight += u.Weight
		rows = append(rows, LeaderRow{u.Username, u.Weight, u.AdsWatched, u.Wins})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Weight != rows[j].Weight {
			return rows[i].Weight > rows[j].Weight
		}
		return rows[i].Username < rows[j].Username
	})
	if len(rows) > 10 {
		rows = rows[:10]
	}
	st.Leaderboard = rows
	return st
}

// Me renvoie l'état du compte + sa probabilité de gagner le prochain drop.
type MeView struct {
	Username   string  `json:"username"`
	Weight     int     `json:"weight"`
	AdsWatched int     `json:"adsWatched"`
	Earned     int64   `json:"earned"`
	Wins       int     `json:"wins"`
	Referrals  int     `json:"referrals"`
	RefCode    string  `json:"refCode"`
	WinChance  float64 `json:"winChance"`
}

func (s *Store) Me(u *User) MeView {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := 0
	for _, x := range s.users {
		total += x.Weight
	}
	chance := 0.0
	if total > 0 {
		chance = float64(u.Weight) / float64(total) * 100
	}
	return MeView{
		Username:   u.Username,
		Weight:     u.Weight,
		AdsWatched: u.AdsWatched,
		Earned:     u.Earned,
		Wins:       u.Wins,
		Referrals:  u.Referrals,
		RefCode:    u.RefCode,
		WinChance:  chance,
	}
}
