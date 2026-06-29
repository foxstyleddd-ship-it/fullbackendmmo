package main

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type ChatMsg struct {
	ID   string    `json:"id"`
	User string    `json:"user"`
	Text string    `json:"text"`
	At   time.Time `json:"at"`
}

const (
	ChatMaxLen   = 240
	ChatCooldown = 3 * time.Second
	ChatKeep     = 80
)

// --- modération automatique ------------------------------------------------

var (
	reURL    = regexp.MustCompile(`(?i)(https?://|www\.|discord\.(gg|com/invite)|t\.me/|[a-z0-9-]+\.(com|net|org|fr|io|gg|me|xyz|co|app|link|shop|store|ru|tk|biz|info|live|tv))`)
	reEmail  = regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`)
	rePhone  = regexp.MustCompile(`(?:\d[\s.\-]?){9,}`)
	reSpaces = regexp.MustCompile(`\s+`)
)

// hasLongRepeat détecte un flood type "aaaaaaaa" (RE2 n'a pas de back-référence).
func hasLongRepeat(text string) bool {
	var prev rune
	run := 0
	for _, r := range text {
		if r == prev {
			if run++; run >= 6 {
				return true
			}
		} else {
			prev, run = r, 0
		}
	}
	return false
}

// Listes volontairement compactes et EXTENSIBLES. En production, brancher un
// service de modération (Perspective API, OpenAI moderation…) ou une liste
// maintenue. "hard" = message rejeté ; "soft" = mot masqué par des étoiles.
var bannedHard = []string{
	"nigger", "negro", "faggot", "nègre", "bougnoule", "youpin", "pédé", "pede",
	"sale arabe", "sale juif", "sale noir", "heil hitler",
}
var bannedSoft = []string{
	"connard", "salope", "pute", "enculé", "encule", "ducon", "bitch",
	"asshole", "fuck", "merde", "batard", "bâtard", "fdp", "ntm",
}

var leet = strings.NewReplacer(
	"0", "o", "1", "i", "3", "e", "4", "a", "5", "s", "7", "t", "@", "a", "$", "s", "ç", "c",
)

// moderate nettoie et juge un message. Renvoie (texte nettoyé, ok, raison).
func moderate(text string) (string, bool, string) {
	text = reSpaces.ReplaceAllString(strings.TrimSpace(text), " ")
	if text == "" {
		return "", false, "message vide"
	}
	if len([]rune(text)) > ChatMaxLen {
		return "", false, "message trop long (240 caractères max)"
	}
	if reURL.MatchString(text) || reEmail.MatchString(text) {
		return "", false, "liens et adresses interdits 🚫"
	}
	if rePhone.MatchString(text) {
		return "", false, "pas de numéros de téléphone 🚫"
	}
	if hasLongRepeat(text) {
		return "", false, "évite le spam de caractères"
	}

	// normalisation pour détecter les contournements (leet, accents, espaces)
	norm := leet.Replace(strings.ToLower(text))
	flat := strings.ReplaceAll(norm, " ", "")
	for _, w := range bannedHard {
		ww := strings.ReplaceAll(w, " ", "")
		if strings.Contains(norm, w) || strings.Contains(flat, ww) {
			return "", false, "message inapproprié — modéré automatiquement 🛡️"
		}
	}

	// masquage des insultes "soft"
	cleaned := text
	for _, w := range bannedSoft {
		re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(w))
		cleaned = re.ReplaceAllStringFunc(cleaned, func(m string) string {
			return strings.Repeat("*", len([]rune(m)))
		})
	}
	return cleaned, true, ""
}

// --- store -----------------------------------------------------------------

func (s *Store) PostChat(u *User, text string) (*ChatMsg, error) {
	cleaned, ok, reason := moderate(text)
	if !ok {
		return nil, errors.New(reason)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.chatLast[u.ID]; ok && time.Since(t) < ChatCooldown {
		return nil, errors.New("doucement 🐢 attends quelques secondes")
	}
	// anti-doublon : pas deux fois le même message d'affilée
	for i := len(s.chat) - 1; i >= 0; i-- {
		if s.chat[i].User == u.Username {
			if strings.EqualFold(s.chat[i].Text, cleaned) {
				return nil, errors.New("tu viens déjà d'envoyer ça")
			}
			break
		}
	}
	msg := ChatMsg{ID: token(6), User: u.Username, Text: cleaned, At: time.Now()}
	s.chat = append(s.chat, msg)
	if len(s.chat) > ChatKeep {
		s.chat = s.chat[len(s.chat)-ChatKeep:]
	}
	s.chatLast[u.ID] = time.Now()
	s.persist()
	return &msg, nil
}

func (s *Store) Chat() []ChatMsg {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ChatMsg, len(s.chat))
	copy(out, s.chat)
	return out
}
