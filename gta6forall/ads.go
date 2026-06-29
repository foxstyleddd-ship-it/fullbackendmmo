package main

import (
	"errors"
	"strings"
)

// Le hub de spots publicitaires : des marques louent un emplacement, et leur
// argent (moins 5 % de commission) remplit directement la cagnotte GTA6.
const (
	NumSlots  = 8
	SpotPrice = 25_000_000 // 25 € le spot (démo). En vrai : tarif/CPM négocié ou via un ad server (Kevel).
)

type AdSlot struct {
	ID     int    `json:"id"`
	Brand  string `json:"brand"`
	Title  string `json:"title"`
	Emoji  string `json:"emoji"`
	Color  string `json:"color"`
	Link   string `json:"link"`
	Price  int64  `json:"price"`
	Filled bool   `json:"filled"`
}

// seedSlots crée le hub de départ avec des annonces "types" pour montrer le
// rendu. (Le owner les remplace par ses vrais deals dans /admin.)
func seedSlots() []*AdSlot {
	demo := []*AdSlot{
		{Brand: "Sprunk", Title: "La boisson n°1 de Los Santos 🥤", Emoji: "🥤", Color: "#1fe0ff"},
		{Brand: "Maze Bank", Title: "Ouvre un compte, +5000$ offerts", Emoji: "🏦", Color: "#ff2e88"},
		{Brand: "LS Customs", Title: "Tune ta caisse, -20% ce mois-ci", Emoji: "🔧", Color: "#ffd23f"},
		{Brand: "Ammu-Nation", Title: "Soldes sur tout l'arsenal 💥", Emoji: "🔫", Color: "#ff6b35"},
		{Brand: "eCola", Title: "Plus de bulles, plus de fun", Emoji: "🥤", Color: "#e63946"},
		{Brand: "Vinewood Auto", Title: "La supercar de tes rêves", Emoji: "🏎️", Color: "#a06bff"},
		{Brand: "Lifeinvader", Title: "Connecte-toi à Los Santos", Emoji: "📱", Color: "#2ecc71"},
		{Brand: "Bean Machine", Title: "Ton café offert ce matin ☕", Emoji: "☕", Color: "#c08552"},
	}
	slots := make([]*AdSlot, NumSlots)
	for i := 0; i < NumSlots; i++ {
		if i < len(demo) {
			slots[i] = demo[i]
			slots[i].Filled = true
		} else {
			slots[i] = &AdSlot{}
		}
		slots[i].ID = i
	}
	return slots
}

// Slots renvoie une copie de l'état des emplacements.
func (s *Store) Slots() []*AdSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*AdSlot, len(s.slots))
	for i, sl := range s.slots {
		cp := *sl
		out[i] = &cp
	}
	return out
}

func sanitizeColor(c string) string {
	c = strings.TrimSpace(c)
	if len(c) != 7 || c[0] != '#' {
		return "#7a1fff"
	}
	for _, r := range c[1:] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return "#7a1fff"
		}
	}
	return c
}

func clip(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		return s[:max]
	}
	return s
}

type PlaceResult struct {
	Slot    *AdSlot  `json:"slot"`
	Funded  int64    `json:"funded"` // montant ajouté à la cagnotte (micro-€)
	NewWins []Winner `json:"newWins"`
}

// AdminUpsertSlot (réservé à l'admin) crée/édite un spot vendu en direct à une
// marque. Si priceMicro > 0, ce revenu alimente la cagnotte (moins 5 %) — c'est
// ainsi que les deals signés financent les GTA6.
func (s *Store) AdminUpsertSlot(id int, brand, title, emoji, color, link string, priceMicro int64) (*PlaceResult, error) {
	brand = clip(brand, 22)
	title = clip(title, 60)
	if brand == "" || title == "" {
		return nil, errors.New("nom de la marque et message requis")
	}
	if priceMicro < 0 {
		priceMicro = 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if id < 0 || id >= len(s.slots) {
		return nil, errors.New("emplacement inconnu")
	}
	slot := s.slots[id]
	slot.Brand = brand
	slot.Title = title
	slot.Emoji = clip(emoji, 8)
	if slot.Emoji == "" {
		slot.Emoji = "📢"
	}
	slot.Color = sanitizeColor(color)
	slot.Link = clip(link, 200)
	slot.Price = priceMicro
	slot.Filled = true

	var funded int64
	if priceMicro > 0 {
		fee := int64(float64(priceMicro) * FeeRate)
		s.fee += fee
		funded = priceMicro - fee
		s.pot += funded
	}
	res := &PlaceResult{Funded: funded, NewWins: s.runDrops()}
	s.persist()
	cp := *slot
	res.Slot = &cp
	return res, nil
}

// AdminClearSlot (réservé à l'admin) libère un emplacement.
func (s *Store) AdminClearSlot(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id < 0 || id >= len(s.slots) {
		return errors.New("emplacement inconnu")
	}
	s.slots[id] = &AdSlot{ID: id}
	s.persist()
	return nil
}
