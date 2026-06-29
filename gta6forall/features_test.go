package main

import "testing"

func TestModeration(t *testing.T) {
	cases := []struct {
		in     string
		ok     bool
		reason string
	}{
		{"Salut la team, ça hype GTA6 !", true, ""},
		{"viens sur http://arnaque.com", false, "lien"},
		{"écris-moi a@b.fr", false, "email"},
		{"appelle 06 12 34 56 78", false, "téléphone"},
		{"aaaaaaaaaa", false, "flood"},
		{"", false, "vide"},
	}
	for _, c := range cases {
		_, ok, _ := moderate(c.in)
		if ok != c.ok {
			t.Errorf("moderate(%q) ok=%v, want %v (%s)", c.in, ok, c.ok, c.reason)
		}
	}
	// masquage des insultes "soft"
	if out, ok, _ := moderate("espèce de connard"); !ok || out == "espèce de connard" {
		t.Errorf("insulte soft non masquée: %q ok=%v", out, ok)
	}
	// rejet des slurs (même avec leetspeak)
	if _, ok, _ := moderate("sale n3gro"); ok {
		t.Error("slur en leetspeak non détecté")
	}
}

func TestChatFlowAndAntiSpam(t *testing.T) {
	s := newTestStore(t)
	u, _, _ := s.Register("Joueur", "pwpw", "", "1.1.1.1")
	if _, err := s.PostChat(u, "premier message"); err != nil {
		t.Fatalf("post: %v", err)
	}
	// cooldown : second message immédiat refusé
	if _, err := s.PostChat(u, "deuxieme"); err == nil {
		t.Error("le cooldown anti-spam devrait bloquer")
	}
	if len(s.Chat()) != 1 {
		t.Errorf("chat len = %d, want 1", len(s.Chat()))
	}
}

func TestRewardPostbackSignatureAndIdempotency(t *testing.T) {
	t.Setenv("POSTBACK_SECRET", "topsecret")
	s := newTestStore(t)
	u, _, _ := s.Register("Gagnant", "pwpw", "", "1.1.1.1")

	if validSig(u.ID, "1000000", "tx1", "deadbeef") {
		t.Error("signature bidon acceptée")
	}
	good := expectedSig("topsecret", u.ID, "1000000", "tx1")
	if !validSig(u.ID, "1000000", "tx1", good) {
		t.Fatal("signature valide refusée")
	}

	if _, err := s.CreditReward(u.ID, 1_000_000, "tx1"); err != nil {
		t.Fatalf("credit: %v", err)
	}
	if s.users[u.ID].Weight != 1 {
		t.Errorf("weight = %d, want 1", s.users[u.ID].Weight)
	}
	// idempotence : même txn refusé
	if _, err := s.CreditReward(u.ID, 1_000_000, "tx1"); err == nil {
		t.Error("le même txn devrait être rejeté (idempotence)")
	}
}

func TestAdminSlotFundsPot(t *testing.T) {
	s := newTestStore(t)
	before := s.State().Pot
	res, err := s.AdminUpsertSlot(0, "NikeLS", "Just buy it", "👟", "#ff0000", "", 20_000_000)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if res.Funded != 19_000_000 { // 20€ - 5%
		t.Errorf("funded = %d, want 19000000", res.Funded)
	}
	if s.State().Pot != before+19_000_000 {
		t.Error("la cagnotte n'a pas été alimentée par le deal admin")
	}
	if err := s.AdminClearSlot(0); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if s.Slots()[0].Filled {
		t.Error("le spot devrait être libre après clear")
	}
}
