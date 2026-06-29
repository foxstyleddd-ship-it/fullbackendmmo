package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
)

// Intégration d'une vraie régie "rewarded / offerwall" (AdGate, Bitlabs, AdGem…).
//
// Flux réel :
//   1. Le joueur ouvre l'offerwall (OFFERWALL_URL) avec son id passé en "subId".
//   2. Il complète une offre / regarde une vidéo chez la régie.
//   3. La régie appelle CE serveur en server-to-server :
//        GET /api/reward/postback?userId=<id>&amount=<micro€>&txnId=<id>&sig=<hmac>
//      où sig = HMAC_SHA256(POSTBACK_SECRET, "userId:amount:txnId").
//   4. On vérifie la signature, on crédite (idempotent sur txnId) : la cagnotte
//      monte et le joueur gagne du poids.
//
// Tant que POSTBACK_SECRET n'est pas défini, le postback est refusé (sécurité).
// OFFERWALL_URL vide => le bouton "offres partenaires" reste masqué et seule la
// pub simulée de démo fonctionne.

func offerwallURL(userID string) string {
	base := os.Getenv("OFFERWALL_URL")
	if base == "" {
		return ""
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + "subId=" + userID
}

// expectedSig calcule la signature attendue pour un postback.
func expectedSig(secret, userID, amount, txn string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID + ":" + amount + ":" + txn))
	return hex.EncodeToString(mac.Sum(nil))
}

// validSig vérifie (à temps constant) la signature fournie par la régie.
func validSig(userID, amount, txn, provided string) bool {
	secret := os.Getenv("POSTBACK_SECRET")
	if secret == "" {
		return false
	}
	want := expectedSig(secret, userID, amount, txn)
	return hmac.Equal([]byte(want), []byte(strings.ToLower(strings.TrimSpace(provided))))
}

// SimulateReward (admin) déclenche une récompense comme si une régie avait
// appelé le postback — pour tester tout le pipeline avant d'avoir un vrai compte.
func (s *Store) SimulateReward(username string, micro int64) (*WatchResult, error) {
	s.mu.RLock()
	id := s.byName[strings.ToLower(strings.TrimSpace(username))]
	s.mu.RUnlock()
	if id == "" {
		return nil, errors.New("joueur introuvable")
	}
	return s.CreditReward(id, micro, "sim-"+token(8))
}

// CreditReward applique un postback de régie : crédite la cagnotte et le poids.
func (s *Store) CreditReward(userID string, amountMicro int64, txn string) (*WatchResult, error) {
	if amountMicro <= 0 || txn == "" {
		return nil, errors.New("paramètres invalides")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rewardTxns[txn] {
		return nil, errors.New("transaction déjà traitée")
	}
	u := s.users[userID]
	if u == nil {
		return nil, errors.New("joueur inconnu")
	}
	s.rewardTxns[txn] = true

	fee := int64(float64(amountMicro) * FeeRate)
	s.fee += fee
	s.pot += amountMicro - fee
	u.AdsWatched++
	u.Weight++ // une offre complétée = +1 poids (ticket)
	u.Earned += amountMicro

	res := &WatchResult{NewWins: s.runDrops(), Reward: amountMicro, Weight: u.Weight, Earned: u.Earned}
	s.persist()
	return res, nil
}
