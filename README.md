# Harry Potter MMO - Backend & Admin Panel

Backend complet et panneau d'administration clés en main pour un MMO Harry Potter avec gestion des comptes, personnages, inventaire, sorts, achievements, friends, et leaderboards.

## 🎮 Fonctionnalités Implémentées

### ✅ Système de Comptes
- ✨ **Gestion complète** : Email, pseudo, mot de passe (hashé), Discord ID
- 🔐 **Statuts** : `active`, `whitelisted`, `banned`, `kicked_out`
- 👑 **Rôles** : `player`, `gm`, `admin`, `superadmin`
- 🎟️ **JWT Authentication** avec refresh tokens
- 🛡️ **2FA** support
- 📊 **Soft delete** + **Audit trail**

### 🧙 Système de Personnages
- 🏰 **Multi-personnages** par compte
- 🎓 **Maisons** : Gryffindor, Slytherin, Hufflepuff, Ravenclaw
- 📈 **Progression** : Niveau, XP, année d'étude (grade 1-7)
- 💪 **Statistiques complètes** : Health, Mana, Stamina, Strength, Intelligence, etc.
- 🔑 **Permissions** basées sur grade et maison
- 🗺️ **Zones** avec système de sharding

### 🎒 Système d'Inventaire
- **6 types d'objets** :
  - 💰 `currency` - Gallions, house points
  - 🧹 `balais` - Balais volants
  - 👔 `equipment` - Robes, chapeaux, etc.
  - 🪄 `wand` - Baguettes magiques
  - 🌿 `ingredient` - Ingrédients de potions
  - 🧪 `consumable` - Potions, nourriture
- ⚔️ **Équipement** : Slots (head, chest, wand, broom, etc.)
- 📦 **Stack** : Objets empilables avec limite
- 🛒 **Shop** : Achat/vente avec prix dynamiques

### ✨ Système de Sorts
- 📚 **10 écoles de magie** : Charms, Transfiguration, Potions, DADA, Dark Arts, Herbology, etc.
- 🎯 **4 niveaux de maîtrise** (0-3) : Novice → Beginner → Intermediate → Advanced
- 📊 **Statistiques** : Times cast, mana cost, cooldown
- 🚫 **Restrictions** : Grade requis, maison requise
- ⚡ **Sorts interdits** : Dark Arts sous conditions

### 🏆 Système d'Achievements
- **18 achievements** pré-configurés
- **5 catégories** : Combat, Exploration, Social, Collection, Progression
- ✅ **Auto-unlock** : Déblocage automatique quand objectif atteint
- 🎁 **Récompenses** : Currencies (gallions, house points, etc.)
- 🤫 **Achievements secrets** cachés

### 👥 Système d'Amis
- 📨 **Demandes d'amis** : Envoi et acceptation
- 🟢 **Statuts** : pending, accepted, blocked
- 📋 **Liste** : Avec dernière connexion, maison, niveau
- ⚙️ **Gestion** : Accepter, décliner, supprimer

### 📊 Leaderboards
- 🌍 **Global** : Tous les joueurs classés par niveau/XP
- 🏠 **Par maison** : Top joueurs de chaque maison
- 📈 **Métriques** : Niveau, XP, house points, achievements, temps de jeu
- ⚡ **Materialized views** pour performance optimale
- 🔝 **Filtres** : Top 10/25/50/100

### 🏠 Points de Maison
- ➕ **Gestion globale** : Ajout/retrait pour toute une maison
- 📝 **Raisons obligatoires** pour accountability
- 👤 **Nom du personnage** optionnel pour audit
- 📊 **Audit automatique** de toutes les modifications

### 📋 Audit Logs
- 📖 **Logs complets** : Toutes les actions admin/GM enregistrées
- 🔍 **Métadonnées** : Qui, Quoi, Quand, Où, Résultat
- 🎯 **Filtrage** : Par personnage, admin, date
- ♾️ **Rétention permanente**

### 🛡️ Commandes GM/Admin
`setgrade`, `teleport`, `grantitem`, `removeitem`, `grantspell`, `removespell`, `ban`, `kick`, `whitelist`, `addhousepoints`, `removehousepoints`

## 🎨 Admin Panel (Frontend React)

### 📱 10 Vues Complètes

1. **📋 Liste des Personnages** - Recherche, filtrage, actions CRUD
2. **👤 Détails Personnage** - Info complète + Inventaire + Sorts
3. **👥 Liste des Comptes** - Recherche, filtrage par statut/rôle, stats
4. **➕ Créer Compte** - Formulaire de registration
5. **🧙 Créer Personnage** - Nom, maison, apparence
6. **🏆 Achievements** - Vue globale + progression par personnage
7. **👥 Friends** - Gestion complète des amis
8. **📊 Leaderboard** - Global et par maison
9. **🏠 House Points** - Interface d'ajout/retrait
10. **📝 Audit Logs** - Historique des actions

## 🏗️ Architecture

### Backend (Go 1.22+)
```
cmd/api/
├── main.go
└── handlers/
    ├── auth_handler.go
    ├── character_handler.go
    ├── inventory_handler.go
    ├── spell_handler.go
    ├── achievement_handler.go
    ├── friends_handler.go
    ├── leaderboard_handler.go
    ├── account_handler.go
    └── admin_handler.go

internal/
├── auth/              # Authentification & comptes
├── character/         # Gestion personnages
├── inventory/         # Inventaire & shop
├── spell/             # Système de sorts
├── achievement/       # Achievements
├── friends/           # Système d'amis
├── audit/             # Audit logs
└── pkg/
    ├── config/        # Configuration
    ├── db/            # PostgreSQL connection
    ├── logger/        # Logging
    ├── middleware/    # JWT, CORS, etc.
    └── redis/         # Cache Redis

migrations/
├── 001_initial.up.sql
├── 002_spells.up.sql
├── 003_achievements_and_social.up.sql
└── 004_add_account_discord_and_status.up.sql
```

### Frontend (React + TypeScript + Vite)
```
admin-panel/src/
├── components/
│   ├── Dashboard.tsx           # Conteneur principal
│   ├── CharacterManager.tsx    # Gestion personnages
│   ├── CharacterDetail.tsx     # Détails + inventaire + sorts
│   ├── AccountList.tsx         # Liste des comptes
│   ├── CreateAccount.tsx
│   ├── CreateCharacter.tsx
│   ├── Achievements.tsx
│   ├── Friends.tsx
│   ├── Leaderboard.tsx
│   ├── HousePoints.tsx
│   └── AuditLogsViewer.tsx
├── services/
│   └── api.ts                  # Client API
└── types/
    └── index.ts                # Types TypeScript
```

## 🚀 Installation & Démarrage

### Prérequis
- Go 1.22+
- PostgreSQL 16
- Redis 7
- Node.js 18+

### 1. Database Setup
```bash
# Démarrer PostgreSQL et Redis
docker-compose up -d

# Exécuter les migrations
migrate -path migrations -database "postgresql://hpmmo:hpmmo@localhost:5432/hpmmo?sslmode=disable" up
```

### 2. Backend
```bash
# Installer les dépendances
go mod download

# Compiler
go build -o bin/api ./cmd/api

# Lancer le serveur
./bin/api
```

Le serveur API démarre sur `http://localhost:8080`

### 3. Frontend (Admin Panel)
```bash
cd admin-panel

# Installer les dépendances
npm install

# Development mode
npm run dev

# Build pour production
npm run build
```

Le panneau d'admin sera accessible sur `http://localhost:5173`

## 📡 API Endpoints Principaux

### Auth
- `POST /v1/auth/register` - Créer un compte
- `POST /v1/auth/login` - Se connecter
- `POST /v1/auth/refresh` - Rafraîchir token
- `POST /v1/auth/logout` - Se déconnecter

### Personnages
- `GET /v1/characters` - Liste
- `POST /v1/characters` - Créer
- `GET /v1/characters/:id` - Détails
- `DELETE /v1/characters/:id` - Supprimer

### Inventaire
- `GET /v1/characters/:id/inventory` - Inventaire
- `GET /v1/characters/:id/equipment` - Équipement
- `POST /v1/characters/:id/purchase` - Acheter
- `POST /v1/characters/:id/sell` - Vendre

### Sorts
- `GET /v1/spells` - Catalogue
- `GET /v1/characters/:id/spells` - Sorts du personnage

### Achievements
- `GET /v1/achievements` - Tous les achievements
- `GET /v1/characters/:id/achievements` - Achievements du personnage

### Friends
- `GET /v1/characters/:id/friends` - Liste d'amis
- `POST /v1/friends/request` - Envoyer demande
- `POST /v1/friends/accept/:id` - Accepter
- `DELETE /v1/friends/:id` - Supprimer

### Leaderboards
- `GET /v1/leaderboard` - Global
- `GET /v1/leaderboard/house/:house` - Par maison

### Admin (Requiert GM/Admin/Superadmin)
- `POST /v1/admin/commands/execute` - Exécuter commande
- `GET /v1/admin/accounts` - Liste des comptes
- `GET /v1/admin/accounts/:id` - Détails compte
- `PUT /v1/admin/accounts/:id/status` - Modifier statut
- `GET /v1/admin/audit/character/:id` - Logs d'un personnage

## 🗄️ Base de Données

### Tables Principales
- `accounts` - Comptes utilisateurs (avec Discord ID et status)
- `characters` - Personnages
- `sessions` - Sessions actives
- `character_stats` - Statistiques
- `character_currencies` - Monnaies (gallions, house points)
- `inventory_items` - Objets
- `item_definitions` - Définitions d'objets
- `equipment_loadouts` - Équipements équipés
- `spell_definitions` - Définitions de sorts
- `character_spells` - Sorts appris (avec proficiency level)
- `achievement_definitions` - Définitions d'achievements
- `character_achievements` - Achievements des personnages
- `friendships` - Relations d'amitié
- `sanctions` - Bans/kicks
- `audit_logs` - Logs d'audit
- `leaderboard_overall` - Materialized view pour leaderboard

## 🔒 Sécurité

- ✅ JWT Authentication avec expiration
- ✅ Password Hashing (bcrypt)
- ✅ Role-Based Access Control
- ✅ SQL Injection Protection (requêtes paramétrées)
- ✅ CORS configuré
- ✅ Audit Logs complet

## 🎯 Intégration Unreal Engine

Pour utiliser ce backend avec Unreal Engine :
1. Authentification via HTTP REST (`POST /auth/login`)
2. Récupération des personnages et sélection
3. Connexion WebSocket pour le temps réel (à implémenter)
4. Les sorts, inventaire, stats sont exposés via l'API

Voir `UNREAL_INTEGRATION.md` pour plus de détails.

## 📝 Documentation

| Document | Description |
|----------|-------------|
| [UNREAL_INTEGRATION.md](UNREAL_INTEGRATION.md) | Guide d'intégration avec Unreal Engine |
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Architecture technique |
| [DATABASE_SCHEMA.sql](docs/DATABASE_SCHEMA.sql) | Schéma PostgreSQL complet |
| [API_REST.md](docs/API_REST.md) | Documentation API REST |

## 📄 License

MIT

## 🐛 Bug Reports

Pour rapporter un bug, créez une issue avec:
- Description du problème
- Steps to reproduce
- Expected vs actual behavior
- Screenshots si applicable

---

**Backend clés en main pour MMO Harry Potter ✨**
