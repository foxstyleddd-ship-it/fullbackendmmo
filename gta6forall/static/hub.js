"use strict";

const $ = (id) => document.getElementById(id);
const euro = (m) => (m / 1_000_000).toLocaleString("fr-FR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });

let me = null;
let currentAd = null;
let knownGames = null;
let spotPrice = 25_000_000;

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

// --- garde d'authentification ---------------------------------------------
async function requireAuth() {
  try { me = await api("/api/me"); }
  catch { location.href = "/"; return false; } // pas connecté → page d'accueil
  renderAccount();
  return true;
}

function renderAccount() {
  $("auth-mini").innerHTML = `<span class="neon-cyan">${esc(me.username)}</span> · ${me.weight} poids <button class="btn btn-ghost" id="logout-btn">Quitter</button>`;
  $("logout-btn").onclick = async () => { await api("/api/logout", {}); location.href = "/"; };
  $("my-weight").textContent = me.weight;
  $("my-ads").textContent = me.adsWatched;
  $("my-chance").textContent = me.winChance.toFixed(2) + "%";
  $("my-wins").textContent = me.wins;
  $("ref-link").value = `${location.origin}/?ref=${me.refCode}`;
}
async function refreshMe() { try { me = await api("/api/me"); renderAccount(); } catch {} }

// --- état global -----------------------------------------------------------
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

  $("winners").innerHTML = st.winners.length === 0
    ? '<li class="muted">Personne n\'a encore gagné… sois le premier !</li>'
    : st.winners.map((w) => `<li><span>🎮 <b>${esc(w.username)}</b> a gagné GTA6 #${w.gameNo}</span><span class="win-time">${timeAgo(w.at)}</span></li>`).join("");

  $("leaderboard").innerHTML = st.leaderboard.length === 0
    ? '<li class="muted">Aucun joueur pour le moment.</li>'
    : st.leaderboard.map((r) => `<li><span>${esc(r.username)}</span><span>${r.weight} poids · ${r.wins} 🎮</span></li>`).join("");
}

// --- hub de spots ----------------------------------------------------------
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
    return `<div class="spot empty" data-advertise="1">
      <div class="plus">＋</div><div class="lbl">Votre pub ici</div><div class="sub">À vendre · nous contacter</div>
    </div>`;
  }).join("");
  grid.querySelectorAll("[data-watch]").forEach((el) => el.onclick = startAd);
  grid.querySelectorAll("[data-advertise]").forEach((el) => el.onclick = () => open("advertise-modal"));
}

// --- modales (helpers) -----------------------------------------------------
function open(id) { $(id).classList.remove("hidden"); }
function close(id) { $(id).classList.add("hidden"); }
document.querySelectorAll("[data-close]").forEach((b) => b.onclick = () => close(b.dataset.close));
document.querySelectorAll(".modal").forEach((m) => m.addEventListener("click", (e) => { if (e.target === m) m.classList.add("hidden"); }));

// --- flow pub (regarder) ---------------------------------------------------
$("watch-btn").onclick = startAd;
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

// --- annoncer (contact, plus de paiement en self-service) ------------------
$("advertise-btn").onclick = () => open("advertise-modal");
(function setupContact() {
  const email = $("contact-email").textContent.trim();
  $("contact-link").href = `mailto:${email}?subject=${encodeURIComponent("Réserver un spot sur GTA6forall")}`;
})();

// --- offres partenaires (vraie régie rewarded/offerwall) -------------------
async function setupOfferwall() {
  let ow;
  try { ow = await api("/api/ads/offerwall"); } catch { return; }
  if (ow.enabled && ow.url) {
    const btn = $("offerwall-btn");
    btn.classList.remove("hidden");
    btn.onclick = () => window.open(ow.url, "_blank", "noopener");
  }
}

// --- tchat communautaire ---------------------------------------------------
let lastChatId = null;
async function refreshChat() {
  let data;
  try { data = await api("/api/chat"); } catch { return; }
  const box = $("chat-box");
  const atBottom = box.scrollHeight - box.scrollTop - box.clientHeight < 40;
  if (!data.messages.length) {
    box.innerHTML = '<div class="muted">Sois le premier à écrire 👋</div>';
    return;
  }
  const last = data.messages[data.messages.length - 1];
  if (last.id === lastChatId) return; // rien de neuf
  lastChatId = last.id;
  box.innerHTML = data.messages.map((m) =>
    `<div class="chat-msg"><span class="chat-user">${esc(m.user)}</span><span class="chat-text">${esc(m.text)}</span></div>`
  ).join("");
  if (atBottom) box.scrollTop = box.scrollHeight;
}

$("chat-form").onsubmit = async (e) => {
  e.preventDefault();
  const input = $("chat-input");
  const text = input.value.trim();
  if (!text) return;
  $("chat-err").textContent = "";
  try {
    await api("/api/chat", { text });
    input.value = "";
    lastChatId = null; // force le re-render
    await refreshChat();
    $("chat-box").scrollTop = $("chat-box").scrollHeight;
  } catch (e) { $("chat-err").textContent = e.message; }
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
(async function init() {
  if (!(await requireAuth())) return; // redirige si pas connecté
  refreshState(); refreshAds(); refreshChat(); setupOfferwall();
  setInterval(() => { refreshState(); refreshAds(); }, 4000);
  setInterval(refreshChat, 3000);
})();
