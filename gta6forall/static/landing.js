"use strict";

const $ = (id) => document.getElementById(id);
const euro = (m) => (m / 1_000_000).toLocaleString("fr-FR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

async function api(path, body) {
  const res = await fetch(path, {
    method: body ? "POST" : "GET",
    headers: body ? { "Content-Type": "application/json" } : {},
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || "Erreur");
  return data;
}

// --- teaser live -----------------------------------------------------------
async function refreshTeaser() {
  let st;
  try { st = await api("/api/state"); } catch { return; }
  $("t-pot").textContent = euro(st.pot) + " €";
  $("t-fill").style.width = Math.min(100, st.progress) + "%";
  $("t-games").textContent = st.gamesGiven + " GTA6 offerts";
  $("t-users").textContent = st.totalUsers;

  const wl = $("winners");
  if (st.winners && st.winners.length) {
    wl.innerHTML = st.winners.slice(0, 6).map((w) =>
      `<li><span>🎮 <b>${esc(w.username)}</b> a gagné GTA6 #${w.gameNo}</span><span class="win-time">${timeAgo(w.at)}</span></li>`).join("");
  }
}

// --- état de connexion -----------------------------------------------------
async function refreshMe() {
  let me = null;
  try { me = await api("/api/me"); } catch {}
  const mini = $("auth-mini");
  if (me) {
    // déjà connecté : on propose d'entrer dans le hub
    mini.innerHTML = `<a class="btn btn-ghost" href="/hub">Aller au hub →</a>`;
    [["cta-main", "Entrer dans le hub →"], ["cta-bottom", "Entrer dans le hub →"]].forEach(([id, txt]) => {
      const b = $(id); b.textContent = txt; b.onclick = () => (location.href = "/hub");
    });
  } else {
    mini.innerHTML = `<button class="btn btn-ghost" id="open-login">Connexion</button>`;
    $("open-login").onclick = () => openAuth("login");
  }
}

// --- auth ------------------------------------------------------------------
let authTab = "register";
function openAuth(tab) {
  authTab = tab;
  document.querySelectorAll(".tab").forEach((t) => t.classList.toggle("active", t.dataset.tab === tab));
  $("auth-submit").textContent = tab === "register" ? "Créer mon compte" : "Se connecter";
  $("f-ref").classList.toggle("hidden", tab === "login");
  $("auth-err").textContent = "";
  $("auth-modal").classList.remove("hidden");
  $("f-username").focus();
}
document.querySelectorAll(".tab").forEach((t) => t.onclick = () => openAuth(t.dataset.tab));
document.querySelectorAll("[data-close]").forEach((b) => b.onclick = () => $(b.dataset.close).classList.add("hidden"));
document.querySelectorAll(".modal").forEach((m) => m.addEventListener("click", (e) => { if (e.target === m) m.classList.add("hidden"); }));

$("cta-main").onclick = () => openAuth("register");
$("cta-bottom").onclick = () => openAuth("register");
$("cta-login").onclick = () => openAuth("login");

$("auth-submit").onclick = async () => {
  try {
    await api(authTab === "register" ? "/api/register" : "/api/login", {
      username: $("f-username").value.trim(),
      password: $("f-password").value,
      refCode: $("f-ref").value.trim(),
    });
    location.href = "/hub"; // succès → on entre dans le hub
  } catch (e) { $("auth-err").textContent = e.message; }
};

// --- utils -----------------------------------------------------------------
function esc(s) { return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])); }
function timeAgo(iso) {
  const s = Math.floor((Date.now() - new Date(iso)) / 1000);
  if (s < 60) return "à l'instant";
  if (s < 3600) return Math.floor(s / 60) + " min";
  if (s < 86400) return Math.floor(s / 3600) + " h";
  return Math.floor(s / 86400) + " j";
}

(function init() {
  const ref = new URLSearchParams(location.search).get("ref");
  if (ref) { $("f-ref").value = ref; openAuth("register"); }
  refreshMe();
  refreshTeaser();
  setInterval(refreshTeaser, 5000);
})();
