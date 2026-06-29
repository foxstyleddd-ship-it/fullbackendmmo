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

// seedSlots crée le hub de départ : quelques marques "démo" + des spots libres.
func seedSlots() []*AdSlot {
	demo := []*AdSlot{
		{Brand: "Sprunk", Title: "La boisson n°1 de Los Santos", Emoji: "🥤", Color: "#1fe0ff", Filled: true, Price: SpotPrice},
		{Brand: "Maze Bank", Title: "Ouvre un compte, +5000$ offerts", Emoji: "🏦", Color: "#ff2e88", Filled: true, Price: SpotPrice},
		{Brand: "LS Customs", Title: "Tune ta caisse, -20% ce mois-ci", Emoji: "🔧", Color: "#ffd23f", Filled: true, Price: SpotPrice},
	}
	slots := make([]*AdSlot, NumSlots)
	for i := 0; i < NumSlots; i++ {
		if i < len(demo) {
			slots[i] = demo[i]
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

// PlaceAd loue un emplacement pour une marque. slotID < 0 = premier spot libre.
// Le prix payé alimente la cagnotte (moins la commission de 5 %).
func (s *Store) PlaceAd(slotID int, brand, title, emoji, color, link string) (*PlaceResult, error) {
	brand = clip(brand, 22)
	title = clip(title, 60)
	if brand == "" || title == "" {
		return nil, errors.New("nom de la marque et message requis")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	var slot *AdSlot
	if slotID >= 0 && slotID < len(s.slots) {
		slot = s.slots[slotID]
		if slot.Filled {
			return nil, errors.New("ce spot est déjà pris, choisis-en un autre")
		}
	} else {
		for _, sl := range s.slots {
			if !sl.Filled {
				slot = sl
				break
			}
		}
	}
	if slot == nil {
		return nil, errors.New("plus aucun spot libre — le hub est complet 🎉")
	}

	slot.Brand = brand
	slot.Title = title
	slot.Emoji = clip(emoji, 8)
	if slot.Emoji == "" {
		slot.Emoji = "📢"
	}
	slot.Color = sanitizeColor(color)
	slot.Link = clip(link, 200)
	slot.Price = SpotPrice
	slot.Filled = true

	fee := int64(float64(SpotPrice) * FeeRate)
	s.fee += fee
	s.pot += SpotPrice - fee

	res := &PlaceResult{Funded: SpotPrice - fee, NewWins: s.runDrops()}
	s.persist()
	cp := *slot
	res.Slot = &cp
	return res, nil
}
