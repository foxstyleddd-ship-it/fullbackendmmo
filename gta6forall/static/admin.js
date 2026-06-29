"use strict";

const $ = (id) => document.getElementById(id);
const euro = (m) => (m / 1_000_000).toLocaleString("fr-FR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
let token = localStorage.getItem("gta6_admin_token") || "";

async function api(path, body, admin) {
  const headers = {};
  if (body) headers["Content-Type"] = "application/json";
  if (admin) headers["X-Admin-Token"] = token;
  const res = await fetch(path, { method: body ? "POST" : "GET", headers, body: body ? JSON.stringify(body) : undefined });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || "Erreur");
  return data;
}

function esc(s) { return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c])); }

// --- connexion -------------------------------------------------------------
$("admin-enter").onclick = async () => {
  token = $("admin-token").value.trim();
  $("admin-err").textContent = "";
  try {
    // on teste le token via un upsert no-op ? Non : on charge les slots (public) puis on tente une action protégée légère.
    await loadSlots();
    // vérifie le token avec une requête protégée inoffensive (clear d'un slot inexistant => 401 si mauvais token, 400 sinon)
    await api("/api/admin/slot/clear", { id: -999 }, true).catch((e) => {
      if (/token/i.test(e.message)) throw e; // mauvais token
    });
    localStorage.setItem("gta6_admin_token", token);
    $("admin-login").classList.add("hidden");
    $("admin-panel").classList.remove("hidden");
  } catch (e) { $("admin-err").textContent = e.message; }
};

// --- spots -----------------------------------------------------------------
async function loadSlots() {
  const data = await api("/api/ads");
  renderSlots(data.slots);
}

function renderSlots(slots) {
  $("slots-admin").innerHTML = slots.map((s) => `
    <div class="slot-edit" data-id="${s.id}">
      <div class="slot-edit-head">Spot #${s.id} ${s.filled ? `<span class="filled-tag" style="color:${esc(s.color)}">● occupé</span>` : '<span class="muted">libre</span>'}</div>
      <input class="f-brand" placeholder="Marque" maxlength="22" value="${s.filled ? esc(s.brand) : ""}" />
      <input class="f-title" placeholder="Message" maxlength="60" value="${s.filled ? esc(s.title) : ""}" />
      <div class="slot-edit-row">
        <input class="f-emoji" placeholder="Emoji" maxlength="8" value="${s.filled ? esc(s.emoji) : ""}" />
        <label class="color-pick">Couleur <input class="f-color" type="color" value="${s.filled ? esc(s.color) : "#7a1fff"}" /></label>
        <input class="f-price" type="number" min="0" step="1" placeholder="Prix €" value="${s.price ? (s.price / 1000000) : ""}" />
      </div>
      <input class="f-link" placeholder="Lien (optionnel)" maxlength="200" value="${s.filled ? esc(s.link) : ""}" />
      <div class="slot-edit-actions">
        <button class="btn btn-primary save-btn">💾 Enregistrer</button>
        ${s.filled ? '<button class="btn btn-ghost clear-btn">🗑️ Libérer</button>' : ""}
      </div>
      <div class="slot-msg form-err"></div>
    </div>`).join("");

  document.querySelectorAll(".slot-edit").forEach((el) => {
    const id = +el.dataset.id;
    const msg = el.querySelector(".slot-msg");
    el.querySelector(".save-btn").onclick = async () => {
      msg.textContent = "";
      msg.style.color = "";
      try {
        const priceEuros = parseFloat(el.querySelector(".f-price").value) || 0;
        const res = await api("/api/admin/slot", {
          id,
          brand: el.querySelector(".f-brand").value.trim(),
          title: el.querySelector(".f-title").value.trim(),
          emoji: el.querySelector(".f-emoji").value.trim(),
          color: el.querySelector(".f-color").value,
          link: el.querySelector(".f-link").value.trim(),
          price: Math.round(priceEuros * 1000000),
        }, true);
        msg.style.color = "var(--cyan)";
        msg.textContent = res.funded > 0 ? `✓ Enregistré · +${euro(res.funded)} € dans la cagnotte` : "✓ Enregistré";
        loadSlots();
      } catch (e) { msg.textContent = e.message; }
    };
    const clearBtn = el.querySelector(".clear-btn");
    if (clearBtn) clearBtn.onclick = async () => {
      try { await api("/api/admin/slot/clear", { id }, true); loadSlots(); }
      catch (e) { msg.textContent = e.message; }
    };
  });
}

// auto-login si token déjà mémorisé
(function init() {
  if (token) {
    $("admin-token").value = token;
    $("admin-enter").click();
  }
})();
