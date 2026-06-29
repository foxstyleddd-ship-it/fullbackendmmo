package main

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	return NewStore(filepath.Join(t.TempDir(), "data.json"))
}

func TestRegisterLoginAndReferral(t *testing.T) {
	s := newTestStore(t)
	u1, _, err := s.Register("Capitaine", "secret", "", "1.1.1.1")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, _, err := s.Register("Capitaine", "another", "", "1.1.1.1"); err != ErrTaken {
		t.Fatalf("expected ErrTaken, got %v", err)
	}
	if _, _, err := s.Login("Capitaine", "wrong"); err != ErrBadCred {
		t.Fatalf("expected ErrBadCred, got %v", err)
	}
	if _, tok, err := s.Login("capitaine", "secret"); err != nil || tok == "" {
		t.Fatalf("login (case-insensitive) failed: %v", err)
	}

	// parrainage : les deux comptes gagnent du poids
	u2, _, err := s.Register("Razor", "secret", u1.RefCode, "2.2.2.2")
	if err != nil {
		t.Fatalf("register referral: %v", err)
	}
	if u2.Weight != RefBonusYou {
		t.Fatalf("filleul weight = %d, want %d", u2.Weight, RefBonusYou)
	}
	if got := s.users[u1.ID].Weight; got != RefBonusRef {
		t.Fatalf("parrain weight = %d, want %d", got, RefBonusRef)
	}
	if s.users[u1.ID].Referrals != 1 {
		t.Fatalf("referrals = %d, want 1", s.users[u1.ID].Referrals)
	}
}

func TestDrawWinnerRespectsWeight(t *testing.T) {
	s := newTestStore(t)
	a, _, _ := s.Register("Heavy", "pwpw", "", "1.1.1.1")
	b, _, _ := s.Register("Light", "pwpw", "", "1.1.1.2")
	s.users[a.ID].Weight = 99
	s.users[b.ID].Weight = 1

	wins := map[string]int{}
	for i := 0; i < 400; i++ {
		wins[s.drawWinner()]++
	}
	if wins["Heavy"] <= wins["Light"] {
		t.Fatalf("le poids n'est pas respecté: %+v", wins)
	}
	// avec zéro poids total, pas de gagnant valide
	s.users[a.ID].Weight, s.users[b.ID].Weight = 0, 0
	if w := s.drawWinner(); w != "—" {
		t.Fatalf("expected no winner, got %q", w)
	}
}

func TestStateProgressAndFee(t *testing.T) {
	s := newTestStore(t)
	s.pot = GamePrice / 2
	s.fee = 1_000_000
	st := s.State()
	if st.Progress < 49 || st.Progress > 51 {
		t.Fatalf("progress = %.2f, want ~50", st.Progress)
	}
	if st.Price != GamePrice {
		t.Fatalf("price = %d", st.Price)
	}
}
