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

### 📓 Carnets d'École et Notes
- 📔 **Carnets multiples** : Chaque personnage peut avoir plusieurs carnets
- 📝 **Contenu** : Titre, contenu texte, matière optionnelle (13 matières)
- 📊 **Notes académiques** : Par matière et année d'étude (1-7)
- 📈 **Examens** : Contrôle continu, examens partiels, finaux, BUSES, ASPICS
- 🎓 **6 niveaux de notes** : Optimal (O), Effort Exceptionnel (E), Acceptable (A), Piètre (P), Désolant (D), Troll (T)
- 📜 **Bulletins** : Rapport de fin d'année avec commentaires des professeurs et signature du directeur
- 🔒 **Permissions** : Carnets CRUD par les joueurs, notes admin/GM uniquement

### 📋 Audit Logs
- 📖 **Logs complets** : Toutes les actions admin/GM enregistrées
- 🔍 **Métadonnées** : Qui, Quoi, Quand, Où, Résultat
- 🎯 **Filtrage** : Par personnage, admin, date
- ♾️ **Rétention permanente**

### 🛡️ Commandes GM/Admin
`setgrade`, `teleport`, `grantitem`, `removeitem`, `grantspell`, `removespell`, `ban`, `kick`, `whitelist`, `addhousepoints`, `removehousepoints`

## 🎨 Admin Panel (Frontend React)

### 📱 12 Vues Complètes

1. **📋 Liste des Personnages** - Recherche, filtrage, actions CRUD
2. **👤 Détails Personnage** - Info complète + Inventaire + Sorts
3. **👥 Liste des Comptes** - Recherche, filtrage par statut/rôle, stats
4. **➕ Créer Compte** - Formulaire de registration
5. **🧙 Créer Personnage** - Nom, maison, apparence
6. **🏆 Achievements** - Vue globale + progression par personnage
7. **👥 Friends** - Gestion complète des amis
8. **📊 Leaderboard** - Global et par maison
9. **🏠 House Points** - Interface d'ajout/retrait
10. **📓 Carnets** - Gestion des carnets d'école avec CRUD complet
11. **📊 Notes & Bulletins** - Visualisation des notes et bulletins scolaires
12. **📝 Audit Logs** - Historique des actions

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
    ├── notebook_handler.go
    └── admin_handler.go

internal/
├── auth/              # Authentification & comptes
├── character/         # Gestion personnages
├── inventory/         # Inventaire & shop
├── spell/             # Système de sorts
├── achievement/       # Achievements
├── friends/           # Système d'amis
├── notebook/          # Carnets d'école & notes académiques
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
├── 004_add_account_discord_and_status.up.sql
└── 005_school_notebooks_and_grades.up.sql
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
│   ├── Notebooks.tsx           # Carnets d'école
│   ├── GradeReport.tsx         # Notes et bulletins
│   └── AuditLogsViewer.tsx
├── services/
│   └── api.ts                  # Client API
└── types/
    └── index.ts                # Types TypeScript
```

## 🚀 Installation & Démarrage

### Prérequis
- **Go 1.22+** - [Télécharger](https://go.dev/dl/)
- **PostgreSQL 16** (via Docker recommandé)
- **Redis 7** (via Docker recommandé)
- **Node.js 18+** - [Télécharger](https://nodejs.org/)
- **golang-migrate** - [Installation](https://github.com/golang-migrate/migrate)

### 1. Configuration de l'Environnement

```bash
# Cloner le projet
git clone <repository-url>
cd fullbackendmmo

# Copier le fichier d'environnement exemple
cp .env.example .env

# Éditer .env avec vos valeurs (les valeurs par défaut sont déjà configurées pour Docker)
# Les valeurs importantes à vérifier :
# DB_PASSWORD=hpmmo_dev_password
# REDIS_PASSWORD=hpmmo_redis_password
```

**⚠️ Important** : Le mot de passe PostgreSQL par défaut est `hpmmo_dev_password`, pas `hpmmo` !

### 2. Démarrage des Services (Docker)

```bash
# Démarrer PostgreSQL, Redis, NATS, Prometheus et Grafana
docker-compose up -d

# Vérifier que les services sont démarrés
docker-compose ps

# Devrait afficher tous les services avec le statut "Up"
```

**Services disponibles** :
- PostgreSQL : `localhost:5432`
- Redis : `localhost:6379`
- NATS : `localhost:4222`
- Prometheus : `http://localhost:9090`
- Grafana : `http://localhost:3000` (admin/admin)

### 3. Installation de golang-migrate (si pas déjà installé)

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/migrate

# Windows (via Scoop)
scoop install migrate

# Ou télécharger depuis : https://github.com/golang-migrate/migrate/releases
```

### 4. Exécution des Migrations

```bash
# IMPORTANT : Utiliser le bon mot de passe !
migrate -path migrations -database "postgresql://hpmmo:hpmmo_dev_password@localhost:5432/hpmmo?sslmode=disable" up

# Pour vérifier l'état des migrations
migrate -path migrations -database "postgresql://hpmmo:hpmmo_dev_password@localhost:5432/hpmmo?sslmode=disable" version

# Pour voir les migrations appliquées (optionnel)
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo -c "\dt"
```

**En cas d'erreur d'authentification** :
```bash
# Si vous obtenez "password authentication failed for user hpmmo"
# Vérifiez que vous utilisez le bon mot de passe : hpmmo_dev_password

# Vérifier que PostgreSQL est bien démarré
docker-compose ps postgres

# Voir les logs PostgreSQL
docker-compose logs postgres

# Se connecter manuellement pour tester
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo
# Mot de passe : hpmmo_dev_password
```

### 5. Backend (API Server)

```bash
# Installer les dépendances Go
go mod download

# Compiler le serveur API
go build -o bin/api ./cmd/api

# Lancer le serveur (en mode développement)
./bin/api

# Ou directement avec go run
go run ./cmd/api
```

Le serveur API démarre sur `http://localhost:8080`

**Vérifier que l'API fonctionne** :
```bash
curl http://localhost:8080/health
# Devrait retourner : {"status":"ok"}
```

### 6. Frontend (Admin Panel)

```bash
cd admin-panel

# Copier le fichier d'environnement
cp .env.example .env

# Vérifier que l'URL de l'API est correcte dans .env
# VITE_API_BASE_URL=http://localhost:8080/v1

# Installer les dépendances npm
npm install

# Lancer en mode développement (avec hot-reload)
npm run dev

# Ou compiler pour la production
npm run build
npm run preview
```

Le panneau d'admin sera accessible sur `http://localhost:5173`

### 7. Créer un Compte Admin Initial

```bash
# Option 1 : Via API
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@hpmmo.com",
    "username": "admin",
    "password": "Admin123!"
  }'

# Option 2 : Directement en base de données
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo -c \
  "UPDATE accounts SET role = 'superadmin' WHERE email = 'admin@hpmmo.com';"
```

### 8. Se Connecter au Panneau Admin

1. Ouvrir `http://localhost:5173`
2. Utiliser les identifiants créés à l'étape 7
3. Vous devriez voir le Dashboard avec tous les onglets

## 🔧 Dépannage

### Problème : "password authentication failed for user hpmmo"

**Solution** : Utilisez `hpmmo_dev_password` et non `hpmmo` comme mot de passe.

```bash
# Bonne commande
migrate -path migrations -database "postgresql://hpmmo:hpmmo_dev_password@localhost:5432/hpmmo?sslmode=disable" up

# Mauvaise commande (va échouer)
migrate -path migrations -database "postgresql://hpmmo:hpmmo@localhost:5432/hpmmo?sslmode=disable" up
```

### Problème : "database connection refused"

```bash
# Vérifier que Docker est démarré
docker-compose ps

# Redémarrer les services
docker-compose down
docker-compose up -d

# Attendre quelques secondes que PostgreSQL soit prêt
sleep 5
```

### Problème : "migration: no change" ou table déjà existe

```bash
# Réinitialiser complètement la base de données
docker-compose down -v
docker-compose up -d
sleep 5
migrate -path migrations -database "postgresql://hpmmo:hpmmo_dev_password@localhost:5432/hpmmo?sslmode=disable" up
```

### Problème : Frontend ne se connecte pas au backend

1. Vérifier que le backend est démarré : `curl http://localhost:8080/health`
2. Vérifier le fichier `admin-panel/.env` : `VITE_API_BASE_URL=http://localhost:8080/v1`
3. Vérifier les CORS dans le backend (déjà configuré pour localhost:5173)

## 📊 Commandes Utiles

```bash
# Voir les logs du backend
./bin/api

# Voir les logs PostgreSQL
docker-compose logs -f postgres

# Voir les logs Redis
docker-compose logs -f redis

# Accéder à PostgreSQL en ligne de commande
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo

# Lister les tables
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo -c "\dt"

# Compter les personnages
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo -c "SELECT COUNT(*) FROM characters;"

# Voir les migrations appliquées
docker exec -it hpmmo-postgres psql -U hpmmo -d hpmmo -c "SELECT * FROM schema_migrations;"

# Arrêter tous les services
docker-compose down

# Arrêter et supprimer les volumes (⚠️ perte de données)
docker-compose down -v
```

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

### Carnets d'École
- `GET /v1/characters/:id/notebooks` - Carnets d'un personnage
- `POST /v1/notebooks` - Créer carnet
- `GET /v1/notebooks/:id` - Détails carnet
- `PUT /v1/notebooks/:id` - Modifier carnet
- `DELETE /v1/notebooks/:id` - Supprimer carnet

### Notes Académiques (Admin/GM uniquement pour création)
- `GET /v1/characters/:id/grades` - Notes d'un personnage (?year=N pour filtrer)
- `POST /v1/grades` - Créer note
- `DELETE /v1/grades/:id` - Supprimer note

### Bulletins Scolaires (Admin/GM uniquement)
- `GET /v1/characters/:id/report-cards` - Tous les bulletins
- `GET /v1/characters/:id/report-cards/:year` - Bulletin d'une année
- `POST /v1/report-cards` - Créer bulletin
- `DELETE /v1/report-cards/:id` - Supprimer bulletin

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
- `school_notebooks` - Carnets d'école des personnages
- `academic_grades` - Notes académiques par matière et année
- `report_cards` - Bulletins de fin d'année
- `sanctions` - Bans/kicks
- `audit_logs` - Logs d'audit
- `leaderboard_overall` - Materialized view pour leaderboard

### Enums Personnalisés
- `school_subject` - 13 matières (Charms, Transfiguration, Potions, etc.)
- `grade_value` - 6 niveaux de notes (Outstanding, Exceeds Expectations, Acceptable, Poor, Dreadful, Troll)

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
