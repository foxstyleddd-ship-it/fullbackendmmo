package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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

	mux.HandleFunc("/api/register", handleRegister)
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/logout", handleLogout)
	mux.HandleFunc("/api/me", handleMe)
	mux.HandleFunc("/api/state", handleState)
	mux.HandleFunc("/api/ad/start", handleAdStart)
	mux.HandleFunc("/api/ad/complete", handleAdComplete)
	mux.HandleFunc("/api/ads", handleAds)
	mux.HandleFunc("/api/ads/place", handleAdsPlace)

	log.Printf("🎮 GTA6forall en écoute sur http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// noCache empêche les navigateurs de servir une vieille version des assets
// pendant qu'on itère (sinon on ne voit pas ses changements).
func noCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	})
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

type placeReq struct {
	Slot  int    `json:"slot"`
	Brand string `json:"brand"`
	Title string `json:"title"`
	Emoji string `json:"emoji"`
	Color string `json:"color"`
	Link  string `json:"link"`
}

func handleAdsPlace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "méthode non autorisée")
		return
	}
	var req placeReq
	if decode(r, &req) != nil {
		writeErr(w, 400, "requête invalide")
		return
	}
	res, err := store.PlaceAd(req.Slot, req.Brand, req.Title, req.Emoji, req.Color, req.Link)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, res)
}
