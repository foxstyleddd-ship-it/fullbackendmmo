"use strict";

const $ = (id) => document.getElementById(id);
const euro = (m) => (m / 1_000_000).toLocaleString("fr-FR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

let me = null;
let currentAd = null;
let knownGames = null;
let spotPrice = 25_000_000;

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

// --- État global -----------------------------------------------------------
async function refreshState() {
  let st;
  try { st = await api("/api/state"); } catch { return; }

  $("pot").textContent = euro(st.pot);
  $("fee").textContent = euro(st.fee) + " €";
  $("pot-pct").textContent = st.progress.toFixed(1) + "%";
  $("bar-fill").style.width = Math.min(100, st.progress) + "%";
  $("total-users").textContent = st.totalUsers;
  $("games-given").textContent = st.gamesGiven + " GTA6 offerts";

  if (knownGames !== null && st.gamesGiven > knownGames && st.winners[0]) showDrop(st.winners[0].username);
  knownGames = st.gamesGiven;

  const wl = $("winners");
  wl.innerHTML = st.winners.length === 0
    ? '<li class="muted">Personne n\'a encore gagné… sois le premier !</li>'
    : st.winners.map((w) => `<li><span>🎮 <b>${esc(w.username)}</b> a gagné GTA6 #${w.gameNo}</span><span class="win-time">${timeAgo(w.at)}</span></li>`).join("");

  const lb = $("leaderboard");
  lb.innerHTML = st.leaderboard.length === 0
    ? '<li class="muted">Aucun joueur pour le moment.</li>'
    : st.leaderboard.map((r) => `<li><span>${esc(r.username)}</span><span>${r.weight} poids · ${r.wins} 🎮</span></li>`).join("");
}

// --- Hub de spots ----------------------------------------------------------
async function refreshAds() {
  let data;
  try { data = await api("/api/ads"); } catch { return; }
  spotPrice = data.spotPrice;
  $("spot-price").textContent = euro(spotPrice) + " €";
  $("place-submit").textContent = `Payer ${euro(spotPrice)} € & afficher ma pub`;
  renderSpots(data.slots);
}

function renderSpots(slots) {
  const grid = $("spots");
  grid.innerHTML = slots.map((s) => {
    if (s.filled) {
      return `<div class="spot filled" style="--spot:${esc(s.color)}" data-watch="1">
        <div class="spot-tag">SPONSORISÉ</div>
        <div class="spot-emoji">${esc(s.emoji)}</div>
        <div>
          <div class="spot-brand">${esc(s.brand)}</div>
          <div class="spot-title">${esc(s.title)}</div>
        </div>
        <div class="spot-cta">▶ Regarder · +1 poids</div>
      </div>`;
    }
    return `<div class="spot empty" data-place="${s.id}">
      <div class="plus">＋</div>
      <div class="lbl">Votre pub ici</div>
      <div class="sub">Réserver ce spot</div>
    </div>`;
  }).join("");

  grid.querySelectorAll("[data-watch]").forEach((el) => el.onclick = onWatchClick);
  grid.querySelectorAll("[data-place]").forEach((el) => el.onclick = () => openPlace(+el.dataset.place));
}

// --- Compte ----------------------------------------------------------------
async function refreshMe() {
  try { me = await api("/api/me"); } catch { me = null; }
  renderAuth();
}

function renderAuth() {
  const mini = $("auth-mini");
  if (me) {
    mini.innerHTML = `<span class="neon-cyan">${esc(me.username)}</span> · ${me.weight} poids <button class="btn btn-ghost" id="logout-btn">Quitter</button>`;
    $("logout-btn").onclick = logout;
    $("dash").classList.remove("hidden");
    fillMe();
  } else {
    mini.innerHTML = `<button class="btn btn-ghost" id="open-auth">Connexion</button>`;
    $("open-auth").onclick = () => openAuth("login");
    $("dash").classList.add("hidden");
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

// --- Modales (helpers) -----------------------------------------------------
function open(id) { $(id).classList.remove("hidden"); }
function close(id) { $(id).classList.add("hidden"); }
document.querySelectorAll("[data-close]").forEach((b) => b.onclick = () => close(b.dataset.close));
document.querySelectorAll(".modal").forEach((m) => m.addEventListener("click", (e) => { if (e.target === m) m.classList.add("hidden"); }));

// --- Auth ------------------------------------------------------------------
let authTab = "register";
function openAuth(tab) {
  authTab = tab;
  document.querySelectorAll(".tab").forEach((t) => t.classList.toggle("active", t.dataset.tab === tab));
  $("auth-submit").textContent = tab === "register" ? "Créer mon compte" : "Se connecter";
  $("f-ref").classList.toggle("hidden", tab === "login");
  $("auth-err").textContent = "";
  open("auth-modal");
  $("f-username").focus();
}
document.querySelectorAll(".tab").forEach((t) => t.onclick = () => openAuth(t.dataset.tab));

$("auth-submit").onclick = async () => {
  try {
    me = await api(authTab === "register" ? "/api/register" : "/api/login", {
      username: $("f-username").value.trim(),
      password: $("f-password").value,
      refCode: $("f-ref").value.trim(),
    });
    close("auth-modal");
    renderAuth();
    refreshState();
  } catch (e) { $("auth-err").textContent = e.message; }
};

async function logout() { await api("/api/logout", {}); me = null; renderAuth(); }

// --- Flow pub (regarder) ---------------------------------------------------
function onWatchClick() { me ? startAd() : openAuth("register"); }
$("watch-btn").onclick = onWatchClick;

let adTimer = null;
async function startAd() {
  setMsg("", true);
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
  open("ad-modal");

  let left = ad.duration;
  $("ad-timer").textContent = left;
  claim.textContent = `Patiente… ${left}s`;
  clearInterval(adTimer);
  adTimer = setInterval(() => {
    left--;
    $("ad-timer").textContent = Math.max(0, left);
    if (left <= 0) { clearInterval(adTimer); claim.disabled = false; claim.textContent = "Récupérer ma récompense 🎁"; }
    else claim.textContent = `Patiente… ${left}s`;
  }, 1000);
}

$("ad-claim").onclick = async () => {
  if (!currentAd) return;
  const captcha = currentAd.needCaptcha ? $("ad-captcha-a").value.trim() : "";
  try {
    const res = await api("/api/ad/complete", { adId: currentAd.adId, captcha });
    close("ad-modal");
    currentAd = null;
    setMsg(`+1 poids · tu as généré ${euro(res.reward)} € pour la cagnotte 🎉`, true);
    await refreshMe(); await refreshState();
    if (res.newWins && res.newWins.length) showDrop(res.newWins[0].username);
  } catch (e) { $("ad-err").textContent = e.message; }
};

// --- Flow marque (réserver un spot) ----------------------------------------
let placeSlot = -1;
function openPlace(slot) {
  placeSlot = (typeof slot === "number" && slot >= 0) ? slot : -1;
  $("place-err").textContent = "";
  open("place-modal");
  $("p-brand").focus();
}
$("place-btn").onclick = () => openPlace(-1);

$("place-submit").onclick = async () => {
  try {
    const res = await api("/api/ads/place", {
      slot: placeSlot,
      brand: $("p-brand").value.trim(),
      title: $("p-title").value.trim(),
      emoji: $("p-emoji").value.trim(),
      color: $("p-color").value,
      link: $("p-link").value.trim(),
    });
    close("place-modal");
    ["p-brand", "p-title", "p-emoji", "p-link"].forEach((id) => ($(id).value = ""));
    setMsg(`📢 Pub affichée ! +${euro(res.funded)} € injectés dans la cagnotte 🎉`, true);
    await refreshAds(); await refreshState();
    if (res.newWins && res.newWins.length) showDrop(res.newWins[0].username);
  } catch (e) { $("place-err").textContent = e.message; }
};

// --- divers ----------------------------------------------------------------
$("copy-ref").onclick = () => {
  navigator.clipboard.writeText($("ref-link").value);
  $("copy-ref").textContent = "Copié ✓";
  setTimeout(() => ($("copy-ref").textContent = "Copier"), 1500);
};

function setMsg(msg, ok) { const el = $("watch-msg"); el.textContent = msg; el.className = "watch-msg " + (ok ? "ok" : "bad"); }
function showDrop(u) {
  const t = $("drop-toast");
  t.textContent = `🎮 DROP ! ${u} vient de gagner un GTA6 !`;
  t.classList.remove("hidden");
  clearTimeout(t._h);
  t._h = setTimeout(() => t.classList.add("hidden"), 4200);
}
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
  const ref = new URLSearchParams(location.search).get("ref");
  if (ref) $("f-ref").value = ref;
  refreshMe(); refreshState(); refreshAds();
  setInterval(() => { refreshState(); refreshAds(); }, 4000);
})();
