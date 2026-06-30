package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed static/*
var staticFS embed.FS

var store *Store

func main() {
	addr := ":" + env("PORT", "8080")
	store = NewStore(env("DATA_PATH", "data.json"))

	sub, _ := fs.Sub(staticFS, "static")
	mux := http.NewServeMux()
	mux.Handle("/", noCache(http.FileServer(http.FS(sub))))
	mux.HandleFunc("/hub", servePage("hub.html")) // page dédiée au hub de spots

	mux.HandleFunc("/api/register", handleRegister)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/me", handleMe)
	mux.HandleFunc("/api/state", handleState)
	mux.HandleFunc("/api/ad/start", handleAdStart)
	mux.HandleFunc("/api/ad/complete", handleAdComplete)
	mux.HandleFunc("/api/ads", handleAds)
	mux.HandleFunc("/api/ads/offerwall", handleOfferwall)
	mux.HandleFunc("/api/reward/postback", handlePostback)
	mux.HandleFunc("/api/chat", handleChat)
	mux.HandleFunc("/api/admin/slot", handleAdminSlot)
	mux.HandleFunc("/api/admin/slot/clear", handleAdminClear)
	mux.HandleFunc("/api/admin/simulate-reward", handleAdminSimulate)
	mux.HandleFunc("/admin", servePage("admin.html"))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })

	// écriture disque périodique en arrière-plan (hors du chemin des requêtes)
	store.StartFlusher(2 * time.Second)

	srv := &http.Server{
		Addr:              addr,
		Handler:           recoverPanic(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	// arrêt propre : on flush les données avant de quitter (redéploiement Render)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
		<-stop
		log.Println("arrêt en cours, sauvegarde…")
		store.flush()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("🎮 GTA6forall en écoute sur http://localhost%s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	store.flush()
}

// recoverPanic empêche qu'une panique dans un handler ne perturbe le serveur.
func recoverPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panique récupérée sur %s: %v", r.URL.Path, rec)
				http.Error(w, `{"error":"erreur interne"}`, http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// noCache empêche les navigateurs de servir une vieille version des assets
// pendant qu'on itère (sinon on ne voit pas ses changements).
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	})
}

// servePage rend une page HTML embarquée à une URL propre (ex: /hub).
func servePage(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := staticFS.ReadFile("static/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// --- helpers ---------------------------------------------------------------

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<16)).Decode(v)
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func setSession(w http.ResponseWriter, tok string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 60 * 24 * 30,
	})
}

func currentUser(r *http.Request) *User {
	c, err := r.Cookie("session")
	if err != nil {
		return nil
	}
	return store.UserByToken(c.Value)
}

// --- handlers --------------------------------------------------------------

type authReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RefCode  string `json:"refCode"`
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	var req authReq
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	u, tok, err := store.Register(req.Username, req.Password, strings.TrimSpace(req.RefCode), clientIP(r))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	setSession(w, tok)
	writeJSON(w, 200, store.Me(u))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	var req authReq
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	u, tok, err := store.Login(req.Username, req.Password)
	if err != nil {
		writeErr(w, 401, err.Error())
		return
	}
	setSession(w, tok)
	writeJSON(w, 200, store.Me(u))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("session"); err == nil {
		store.Logout(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if u == nil {
		writeErr(w, 401, "non connecté")
		return
	}
	writeJSON(w, 200, store.Me(u))
}

func handleState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, store.State())
}

func handleAdStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	u := currentUser(r)
	if u == nil {
		writeErr(w, 401, "non connecté")
		return
	}
	ad, err := store.StartAd(u.ID)
	if err != nil {
		writeErr(w, 429, err.Error())
		return
	}
	writeJSON(w, 200, ad)
}

type completeReq struct {
	AdID    string `json:"adId"`
	Captcha string `json:"captcha"`
}

func handleAdComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	u := currentUser(r)
	if u == nil {
		writeErr(w, 401, "non connecté")
		return
	}
	var req completeReq
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	cap, _ := strconv.Atoi(strings.TrimSpace(req.Captcha))
	res, err := store.CompleteAd(u.ID, req.AdID, cap)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, res)
}

func handleAds(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"slots":     store.Slots(),
		"spotPrice": int64(SpotPrice),
	})
}

// --- offerwall / régie réelle ---------------------------------------------

func handleOfferwall(w http.ResponseWriter, r *http.Request) {
	u := currentUser(r)
	if u == nil {
		writeErr(w, 401, "non connecté")
		return
	}
	url := offerwallURL(u.ID)
	writeJSON(w, 200, map[string]any{"enabled": url != "", "url": url})
}

// handlePostback reçoit le callback serveur-à-serveur signé de la régie.
func handlePostback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID, amount, txn, signature := q.Get("userId"), q.Get("amount"), q.Get("txnId"), q.Get("sig")
	if !validSig(userID, amount, txn, signature) {
		writeErr(w, 403, "signature invalide")
		return
	}
	micro, err := strconv.ParseInt(amount, 10, 64)
	if err != nil {
		writeErr(w, 400, "montant invalide")
		return
	}
	if _, err := store.CreditReward(userID, micro, txn); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	w.Write([]byte("OK")) // la plupart des régies attendent "OK" en réponse
}

// --- tchat communautaire ---------------------------------------------------

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]any{"messages": store.Chat()})
		return
	}
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	u := currentUser(r)
	if u == nil {
		writeErr(w, 401, "connecte-toi pour discuter")
		return
	}
	var req struct {
		Text string `json:"text"`
	}
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	msg, err := store.PostChat(u, req.Text)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, msg)
}

// --- admin -----------------------------------------------------------------

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	tok := os.Getenv("ADMIN_TOKEN")
	if tok == "" {
		writeErr(w, 503, "admin non configuré (définis ADMIN_TOKEN)")
		return false
	}
	if r.Header.Get("X-Admin-Token") != tok {
		writeErr(w, 401, "token admin invalide")
		return false
	}
	return true
}

type adminSlotReq struct {
	ID    int    `json:"id"`
	Brand string `json:"brand"`
	Title string `json:"title"`
	Emoji string `json:"emoji"`
	Color string `json:"color"`
	Link  string `json:"link"`
	Price int64  `json:"price"` // en micro-euros ; >0 alimente la cagnotte
}

func handleAdminSlot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	var req adminSlotReq
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	res, err := store.AdminUpsertSlot(req.ID, req.Brand, req.Title, req.Emoji, req.Color, req.Link, req.Price)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, res)
}

func handleAdminClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	var req struct {
		ID int `json:"id"`
	}
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	if err := store.AdminClearSlot(req.ID); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func handleAdminSimulate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	if !requireAdmin(w, r) {
		return
	}
	var req struct {
		Username string  `json:"username"`
		Euros    float64 `json:"euros"`
	}
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	micro := int64(req.Euros * 1_000_000)
	res, err := store.SimulateReward(req.Username, micro)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, res)
}
