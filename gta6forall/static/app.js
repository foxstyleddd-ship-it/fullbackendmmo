"use strict";

const $ = (id) => document.getElementById(id);
const euro = (micro) => (micro / 1_000_000).toLocaleString("fr-FR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

let me = null;          // compte connecté
let currentAd = null;   // pub en cours { adId, needCaptcha }
let knownGames = null;  // pour détecter les nouveaux drops

// --- API -------------------------------------------------------------------
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

// --- État global (cagnotte / feeds) ---------------------------------------
async function refreshState() {
  let st;
  try { st = await api("/api/state"); } catch { return; }

  $("pot").textContent = euro(st.pot);
  $("fee").textContent = euro(st.fee) + " €";
  $("pot-pct").textContent = st.progress.toFixed(1) + "%";
  $("bar-fill").style.width = Math.min(100, st.progress) + "%";
  $("total-users").textContent = st.totalUsers;
  $("games-given").textContent = st.gamesGiven + " GTA6 déjà offerts";

  // détecter un nouveau drop pour l'animation
  if (knownGames !== null && st.gamesGiven > knownGames && st.winners[0]) {
    showDrop(st.winners[0].username);
  }
  knownGames = st.gamesGiven;

  // gagnants
  const wl = $("winners");
  if (st.winners.length === 0) {
    wl.innerHTML = '<li class="muted">Personne n\'a encore gagné… sois le premier !</li>';
  } else {
    wl.innerHTML = st.winners.map((w) =>
      `<li><span>🎮 <b>${esc(w.username)}</b> a gagné GTA6 #${w.gameNo}</span><span class="win-time">${timeAgo(w.at)}</span></li>`
    ).join("");
  }

  // classement
  const lb = $("leaderboard");
  if (st.leaderboard.length === 0) {
    lb.innerHTML = '<li class="muted">Aucun joueur pour le moment.</li>';
  } else {
    lb.innerHTML = st.leaderboard.map((r) =>
      `<li><span>${esc(r.username)}</span><span>${r.weight} poids · ${r.wins} 🎮</span></li>`
    ).join("");
  }
}

// --- Compte ----------------------------------------------------------------
async function refreshMe() {
  try { me = await api("/api/me"); } catch { me = null; }
  renderAuth();
}

function renderAuth() {
  const mini = $("auth-mini");
  if (me) {
    mini.innerHTML = `<span class="neon-cyan">${esc(me.username)}</span> · ${me.weight} poids <button class="btn btn-ghost" id="logout-btn">Déconnexion</button>`;
    $("logout-btn").onclick = logout;
    $("dash").classList.remove("hidden");
    $("hero-cta").textContent = "Regarder une pub 🍿";
    fillMe();
  } else {
    mini.innerHTML = `<button class="btn btn-ghost" id="open-auth">Connexion</button>`;
    $("open-auth").onclick = () => openAuth("login");
    $("dash").classList.add("hidden");
    $("hero-cta").textContent = "Créer mon compte & gagner →";
  }
}

function fillMe() {
  if (!me) return;
  $("my-weight").textContent = me.weight;
  $("my-ads").textContent = me.adsWatched;
  $("my-chance").textContent = me.winChance.toFixed(2) + "%";
  $("my-wins").textContent = me.wins;
  $("ref-link").value = `${location.origin}/?ref=${me.refCode}`;
}

// --- Auth modal ------------------------------------------------------------
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
function closeAuth() { $("auth-modal").classList.add("hidden"); }

document.querySelectorAll(".tab").forEach((t) => t.onclick = () => openAuth(t.dataset.tab));
$("auth-close").onclick = closeAuth;

$("auth-submit").onclick = async () => {
  const username = $("f-username").value.trim();
  const password = $("f-password").value;
  const refCode = $("f-ref").value.trim();
  try {
    me = await api(authTab === "register" ? "/api/register" : "/api/login", { username, password, refCode });
    closeAuth();
    renderAuth();
    refreshState();
  } catch (e) {
    $("auth-err").textContent = e.message;
  }
};

async function logout() {
  await api("/api/logout", {});
  me = null;
  renderAuth();
}

$("hero-cta").onclick = () => { me ? startAd() : openAuth("register"); };
$("copy-ref").onclick = () => {
  navigator.clipboard.writeText($("ref-link").value);
  $("copy-ref").textContent = "Copié ✓";
  setTimeout(() => ($("copy-ref").textContent = "Copier"), 1500);
};

// --- Flow pub --------------------------------------------------------------
$("watch-btn").onclick = startAd;
let adTimer = null;

async function startAd() {
  $("watch-msg").textContent = "";
  let ad;
  try { ad = await api("/api/ad/start", {}); }
  catch (e) { return setMsg(e.message, false); }

  currentAd = { adId: ad.adId, needCaptcha: ad.needCaptcha };
  $("ad-sponsor").textContent = ad.sponsor;
  $("ad-err").textContent = "";
  $("ad-captcha").classList.toggle("hidden", !ad.needCaptcha);
  if (ad.needCaptcha) { $("ad-captcha-q").textContent = ad.captcha; $("ad-captcha-a").value = ""; }

  const claim = $("ad-claim");
  claim.disabled = true;
  $("ad-modal").classList.remove("hidden");

  let left = ad.duration;
  $("ad-timer").textContent = left;
  claim.textContent = `Patiente… ${left}s`;
  clearInterval(adTimer);
  adTimer = setInterval(() => {
    left--;
    $("ad-timer").textContent = Math.max(0, left);
    if (left <= 0) {
      clearInterval(adTimer);
      claim.disabled = false;
      claim.textContent = "Récupérer ma récompense 🎁";
    } else {
      claim.textContent = `Patiente… ${left}s`;
    }
  }, 1000);
}

$("ad-claim").onclick = async () => {
  if (!currentAd) return;
  const captcha = currentAd.needCaptcha ? $("ad-captcha-a").value.trim() : "";
  try {
    const res = await api("/api/ad/complete", { adId: currentAd.adId, captcha });
    $("ad-modal").classList.add("hidden");
    currentAd = null;
    setMsg(`+1 poids · tu as généré ${euro(res.reward)} € pour la cagnotte 🎉`, true);
    await refreshMe();
    await refreshState();
    if (res.newWins && res.newWins.length) showDrop(res.newWins[0].username);
  } catch (e) {
    $("ad-err").textContent = e.message;
  }
};

function setMsg(msg, ok) {
  const el = $("watch-msg");
  el.textContent = msg;
  el.className = "watch-msg " + (ok ? "ok" : "bad");
}

// --- Drop toast ------------------------------------------------------------
function showDrop(username) {
  const t = $("drop-toast");
  t.textContent = `🎮 DROP ! ${username} vient de gagner un GTA6 !`;
  t.classList.remove("hidden");
  clearTimeout(t._h);
  t._h = setTimeout(() => t.classList.add("hidden"), 4200);
}

// --- utils -----------------------------------------------------------------
function esc(s) { return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])); }
function timeAgo(iso) {
  const s = Math.floor((Date.now() - new Date(iso)) / 1000);
  if (s < 60) return "à l'instant";
  if (s < 3600) return Math.floor(s / 60) + " min";
  if (s < 86400) return Math.floor(s / 3600) + " h";
  return Math.floor(s / 86400) + " j";
}

// --- init ------------------------------------------------------------------
(function init() {
  // pré-remplir le code de parrainage depuis l'URL
  const ref = new URLSearchParams(location.search).get("ref");
  if (ref) { $("f-ref").value = ref; }

  refreshMe();
  refreshState();
  setInterval(refreshState, 4000); // feed live
})();
