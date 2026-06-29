# GTA6forall 🎮

> Regarde des pubs → remplis une **cagnotte commune** → dès qu'elle atteint le prix d'un GTA6, le site « achète » le jeu et le distribue **au hasard, pondéré par ton poids**. Plus tu regardes, plus tu as de chances.

MVP **démo jouable** : économie 100 % simulée, pas d'argent ni de jeu réel. Backend Go (stdlib uniquement), frontend embarqué, persistance JSON. Aucune dépendance externe.

## Lancer

```bash
cd gta6forall
go run .
# → http://localhost:8080
```

### Variables d'environnement

| Variable | Rôle | Défaut |
|---|---|---|
| `PORT` | Port HTTP | `8080` |
| `DATA_PATH` | Fichier de persistance JSON | `data.json` |
| `ADMIN_TOKEN` | Active la page `/admin` et protège ses appels. **Vide = admin désactivé.** | — |
| `OFFERWALL_URL` | URL de l'offerwall d'une vraie régie (AdGate, Bitlabs…). Affiche le bouton « Offres partenaires ». Vide = bouton masqué. | — |
| `POSTBACK_SECRET` | Secret HMAC pour vérifier le postback serveur-à-serveur de la régie. **Vide = postback refusé.** | — |
| `CONTACT_EMAIL` | E-mail affiché pour réserver un spot (info uniquement). | `contact@gta6forall.com` |

### Pages

- `/` — landing qui explique le principe (visiteurs non connectés).
- `/hub` — hub de spots + tchat + compte (connectés ; redirige vers `/` sinon).
- `/admin` — gestion des emplacements (token `ADMIN_TOKEN`).

### Brancher une vraie régie « rewarded / offerwall »

Le mécanisme réel est déjà câblé (il ne manque qu'un compte éditeur validé) :

1. Crée un compte chez une régie qui autorise l'incentivé : **AdGate Media, Bitlabs, AdGem, Lootably…** (jamais AdSense, qui l'interdit).
2. Mets son URL d'offerwall dans `OFFERWALL_URL` (on y ajoute `subId=<idJoueur>` automatiquement).
3. Configure son **postback serveur-à-serveur** vers :
   `GET /api/reward/postback?userId=<subId>&amount=<micro€>&txnId=<id>&sig=<hmac>`
   avec `sig = HMAC_SHA256(POSTBACK_SECRET, "userId:amount:txnId")`.
4. Le joueur complète une offre → la régie appelle le postback → la cagnotte monte et le joueur gagne du poids. Idempotent sur `txnId`.

Tests :

```bash
go test ./...
```

## Ce qui est implémenté

| Mécanique | Détail |
|---|---|
| **Inscription / connexion** | Pseudo + mot de passe, session par cookie. |
| **Mur de pubs** | Modale de pub simulée (sponsors façon Los Santos). |
| **Anti-fraude (démo)** | Durée minimale regardée avant validation, **cooldown** entre deux pubs, **captcha** 1 pub sur 3. |
| **Poids = actions vérifiées** | Chaque pub validée = **+1 poids (ticket)**. Le poids sert au tirage. |
| **Cagnotte live** | Barre de progression en temps réel vers le prochain GTA6 (poll toutes les 4 s). |
| **Drop automatique** | Cagnotte ≥ 69,99 € → tirage au sort **pondéré par le poids**, animation de drop. |
| **5 % de bénéfices** | Une commission de 5 % est prélevée à chaque récompense et affichée. |
| **Parrainage** | Lien `?ref=CODE` : +5 poids pour le parrain, +2 pour le filleul (boucle virale TikTok). |
| **Feed gagnants + classement** | Derniers GTA6 offerts et top 10 des poids, en direct. |
| **Régie réelle (rewarded/offerwall)** | Bouton « Offres partenaires » + **postback HMAC** serveur-à-serveur idempotent (voir ci-dessous). |
| **Spots gérés en admin** | Plus de paiement self-service : le owner saisit les deals vendus aux marques dans `/admin`, le revenu finance la cagnotte. |
| **Tchat communautaire auto-modéré** | Réservé aux connectés. Anti-liens/emails/téléphones, filtre slurs (rejet) + insultes (masquées), anti-flood, cooldown 3 s, anti-doublon. |

## Économie (volontairement « démo »)

Les valeurs sont tunées pour que des drops arrivent pendant une session/un TikTok
(récompense 0,10–0,80 € par pub). **Dans la vraie vie** :

- AdSense **interdit** la pub incentivée → il faut passer par une régie **rewarded / offerwall**
  (AdGate, Bitlabs, AdGem, Unity/ironSource…), seules autorisées pour le « regarde → gagne ».
- Une vue rapporte ~0,001–0,01 € : la cagnotte se remplit **beaucoup** plus lentement.
- La marge doit garantir *revenu par joueur > valeur moyenne du lot par joueur*, sinon le site perd de l'argent.
- L'anti-fraude réel (device fingerprint, vérification, postback signé de la régie) est indispensable :
  c'est le risque n°1 du modèle.

Les hooks (flux start/complete de la pub, calcul du poids, déclenchement du drop) sont
isolés pour brancher une vraie régie ensuite sans tout réécrire.

## Architecture

```
gta6forall/
├── main.go          # serveur HTTP + routes API + sert le frontend embarqué
├── store.go         # logique métier : comptes, économie, tirage pondéré, persistance
├── store_test.go    # tests (auth, parrainage, tirage pondéré, progression)
└── static/          # frontend (HTML / CSS / JS vanilla), embarqué via go:embed
    ├── index.html
    ├── style.css
    └── app.js
```

### API

| Méthode | Route | Rôle |
|---|---|---|
| POST | `/api/register` | Inscription (`username`, `password`, `refCode?`) |
| POST | `/api/login` | Connexion |
| POST | `/api/logout` | Déconnexion |
| GET | `/api/me` | Mon compte + ma probabilité au prochain drop |
| GET | `/api/state` | Cagnotte, progression, gagnants, classement |
| POST | `/api/ad/start` | Démarre une pub (renvoie durée + captcha éventuel) |
| POST | `/api/ad/complete` | Valide la pub regardée, crédite la cagnotte, déclenche les drops |

> ⚠️ Projet fun, non affilié à Rockstar Games. « GTA » et « GTA6 » appartiennent à leurs propriétaires respectifs.
