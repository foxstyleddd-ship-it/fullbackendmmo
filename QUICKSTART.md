# HP MMO Backend - Quick Start Guide

## 🎯 Ce qui est implémenté (Sprint 1 - 80%)

### ✅ Infrastructure
- Docker Compose (PostgreSQL, Redis, NATS, Prometheus, Grafana)
- Migrations de base de données complètes
- Scripts de seed (zones et items)
- Configuration via variables d'environnement
- Logging structuré avec Zerolog

### ✅ Service d'Authentification
- Inscription avec validation
- Login avec bcrypt
- JWT (Access Token 24h + Refresh Token 30j)
- Session management avec Redis
- Logout

### ✅ Service Characters
- Création de personnages (max 5 par compte)
- Maisons : Venatrix, Aerwyn, Falcon, Brumval
- Grades (1-100) avec permissions
- Stats initiales et monnaies
- Sélection de personnage pour jouer
- Soft delete

### ✅ API REST
- Endpoints Auth (`/auth/*`)
- Endpoints Characters (`/characters/*`)
- Middleware d'authentification JWT
- Format de réponse standardisé
- Gestion d'erreurs

### ✅ Zone Server (WebSocket)
- Connexion WebSocket
- Messages : CONNECT, HEARTBEAT, ENTER_WORLD
- Gestion des joueurs en mémoire
- Broadcast de présence
- Latency tracking

---

## 🚀 Démarrage Rapide (5 minutes)

### Prérequis
- Docker & Docker Compose
- Go 1.22+ (pour build local)

### Étape 1 : Démarrer l'infrastructure

```bash
# Cloner le repo (si pas déjà fait)
cd fullbackendmmo

# Démarrer PostgreSQL, Redis, NATS, Prometheus, Grafana
make dev

# Attendre que les services soient prêts (15-20 secondes)
# Vérifier les logs
make logs
```

### Étape 2 : Initialiser la base de données

```bash
# Appliquer les migrations
make migrate-up

# (Optionnel) Seed les données de test (zones + items)
make seed
```

### Étape 3 : Démarrer les serveurs

**Terminal 1 - API Server:**
```bash
make run-api
# Serveur API démarré sur http://localhost:8080
```

**Terminal 2 - Zone Server:**
```bash
make run-zone
# Zone Server démarré sur ws://localhost:8081
```

---

## 🧪 Tester l'API

### 1. Inscription

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "harry@hogwarts.edu",
    "username": "harry_potter",
    "password": "password123",
    "display_name": "Harry Potter"
  }'
```

**Réponse attendue:**
```json
{
  "success": true,
  "data": {
    "account_id": "...",
    "username": "harry_potter",
    "role": "player",
    "tokens": {
      "access_token": "eyJhbGc...",
      "refresh_token": "...",
      "expires_in": 86400,
      "token_type": "Bearer"
    },
    "characters": []
  }
}
```

### 2. Login

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "harry@hogwarts.edu",
    "password": "password123"
  }'
```

### 3. Créer un personnage

```bash
# Remplacer <ACCESS_TOKEN> par le token obtenu au login
curl -X POST http://localhost:8080/v1/characters \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ACCESS_TOKEN>" \
  -d '{
    "name": "Harry Potter",
    "house": "falcon",
    "appearance_data": {
      "body_type": 1,
      "face": 2,
      "hair": 3,
      "hair_color": "#000000",
      "skin_tone": 2
    }
  }'
```

### 4. Lister les personnages

```bash
curl -X GET http://localhost:8080/v1/characters \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

### 5. Sélectionner un personnage

```bash
# Remplacer <CHARACTER_ID> par l'ID retourné à la création
curl -X POST http://localhost:8080/v1/characters/<CHARACTER_ID>/select \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

**Réponse importante:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbG... (nouveau token avec character_id)",
    "character": {
      "id": "...",
      "name": "Harry Potter",
      "house": "falcon",
      "grade": 1
    },
    "zone_connection": {
      "zone_id": "hogwarts_main",
      "shard": "shard-0",
      "websocket_url": "ws://localhost:8081/ws",
      "connection_token": "..."
    }
  }
}
```

---

## 👑 Tester les Commandes GM (Admin)

### Prérequis
Pour utiliser les commandes GM, vous devez avoir un compte avec le rôle `gm`, `admin` ou `superadmin`.

**Créer un compte admin via SQL:**
```bash
# Se connecter à PostgreSQL
make psql

# Mettre à jour le rôle d'un compte existant
UPDATE accounts SET role = 'admin' WHERE email = 'harry@hogwarts.edu';
```

### 1. Commande SetGrade

Change le grade (année) d'un personnage (1-100).

```bash
# Remplacer <ADMIN_TOKEN> par votre token admin
# Remplacer <CHARACTER_ID> par l'ID du personnage cible
curl -X POST http://localhost:8080/v1/admin/commands/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "command": "setgrade",
    "target_id": "<CHARACTER_ID>",
    "parameters": {
      "grade": 5
    }
  }'
```

**Réponse:**
```json
{
  "success": true,
  "data": {
    "success": true,
    "message": "Grade updated from 1 to 5 for character Harry Potter",
    "data": {
      "character_id": "...",
      "old_grade": 1,
      "new_grade": 5
    }
  }
}
```

### 2. Commande Teleport

Téléporte un personnage vers une zone.

```bash
curl -X POST http://localhost:8080/v1/admin/commands/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "command": "teleport",
    "target_id": "<CHARACTER_ID>",
    "parameters": {
      "zone_id": "diagon_alley",
      "x": 100.0,
      "y": 50.0,
      "z": 0.0
    }
  }'
```

**Zones disponibles** (après seed):
- `hogwarts_main`
- `hogsmeade`
- `forbidden_forest`
- `quidditch_pitch`
- `ministry_of_magic`
- `diagon_alley`
- `azkaban`

### 3. Commande GrantItem (Preview - Sprint 2)

Donne un item à un personnage. Note: Le système d'inventaire sera implémenté au Sprint 2.

```bash
curl -X POST http://localhost:8080/v1/admin/commands/execute \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <ADMIN_TOKEN>" \
  -d '{
    "command": "grantitem",
    "target_id": "<CHARACTER_ID>",
    "parameters": {
      "item_id": "550e8400-e29b-41d4-a716-446655440001",
      "quantity": 1
    }
  }'
```

### 4. Consulter l'Audit Log d'un Personnage

```bash
curl -X GET "http://localhost:8080/v1/admin/audit/character/<CHARACTER_ID>?limit=20" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

**Réponse:**
```json
{
  "success": true,
  "data": {
    "character_id": "...",
    "logs": [
      {
        "id": "...",
        "event_type": "grade_change",
        "actor_id": "...",
        "actor_type": "account",
        "target_id": "...",
        "target_type": "character",
        "action": "update_grade",
        "details": "Grade changed from 1 to 5",
        "old_value": "1",
        "new_value": "5",
        "ip_address": "127.0.0.1",
        "success": true,
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "count": 1
  }
}
```

### 5. Consulter ses Propres Actions Admin

```bash
curl -X GET "http://localhost:8080/v1/admin/audit/my-actions?limit=50" \
  -H "Authorization: Bearer <ADMIN_TOKEN>"
```

---

## 🎮 Tester le Zone Server (WebSocket)

Utiliser un client WebSocket (exemple avec `websocat` ou client JavaScript).

### Installation de websocat (optionnel)
```bash
# Linux
sudo apt install websocat

# Mac
brew install websocat
```

### Connexion WebSocket

```bash
websocat ws://localhost:8081/ws
```

**Envoyer le message CONNECT:**
```json
{
  "msg_type": "CONNECT",
  "msg_id": "msg-001",
  "timestamp": 1704067200000,
  "payload": {
    "connection_token": "token_from_select_character",
    "character_id": "your-character-id",
    "client_version": "1.0.0",
    "platform": "test"
  }
}
```

**Réponse CONNECT_ACK:**
```json
{
  "msg_type": "CONNECT_ACK",
  "msg_id": "...",
  "ref_msg_id": "msg-001",
  "timestamp": 1704067200050,
  "success": true,
  "payload": {
    "session_id": "...",
    "character": {
      "id": "...",
      "name": "Harry Potter",
      "house": "falcon",
      "grade": 1
    },
    "zone": {
      "id": "hogwarts_main",
      "shard": "shard-0"
    },
    "server_time": 1704067200050,
    "tick_rate_ms": 50,
    "heartbeat_interval_ms": 10000
  }
}
```

**Envoyer ENTER_WORLD:**
```json
{
  "msg_type": "ENTER_WORLD",
  "msg_id": "msg-002",
  "timestamp": 1704067201000,
  "payload": {
    "zone_id": "hogwarts_main"
  }
}
```

**Heartbeat (à envoyer toutes les 10s):**
```json
{
  "msg_type": "HEARTBEAT",
  "msg_id": "msg-003",
  "timestamp": 1704067210000,
  "payload": {
    "client_time": 1704067210000
  }
}
```

---

## 📊 Services de Monitoring

### Grafana
- URL: http://localhost:3000
- Login: `admin` / `admin`
- Dashboards (à configurer)

### Prometheus
- URL: http://localhost:9090
- Metrics disponibles (quand implémentées)

### PostgreSQL
```bash
# Se connecter
make psql

# Vérifier les comptes
SELECT id, username, email, role FROM accounts;

# Vérifier les personnages
SELECT id, name, house, grade, level FROM characters;
```

### Redis
```bash
# Se connecter
make redis-cli

# Lister les sessions
KEYS session:*

# Voir une session
GET session:uuid-here
```

---

## 🏗️ Structure du Projet

```
fullbackendmmo/
├── cmd/
│   ├── api/
│   │   ├── handlers/      # HTTP handlers
│   │   └── main.go        # API server
│   └── zone/
│       └── main.go        # Zone server (WebSocket)
│
├── internal/
│   ├── auth/              # Service d'authentification
│   │   ├── models.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── jwt.go
│   │
│   ├── character/         # Service personnages
│   │   ├── models.go
│   │   ├── repository.go
│   │   └── service.go
│   │
│   ├── zone/              # WebSocket zone
│   │   └── models.go
│   │
│   └── pkg/               # Packages utilitaires
│       ├── config/
│       ├── db/
│       ├── redis/
│       ├── logger/
│       └── middleware/
│
├── migrations/            # Migrations DB
├── scripts/               # Scripts seed
├── deployments/           # Docker & K8s
└── docs/                  # Documentation
```

---

## 🔧 Commandes Utiles

```bash
# Voir les logs de tous les services
make logs

# Voir les logs d'un service spécifique
docker-compose logs -f postgres
docker-compose logs -f redis

# Rebuild les services Go
make build-local

# Lancer les tests (quand implémentés)
make test

# Formatter le code
make fmt

# Linter
make lint

# Nettoyer
make clean

# Arrêter tout
make dev-down
```

---

## 🐛 Troubleshooting

### Le serveur API ne démarre pas
```bash
# Vérifier que PostgreSQL est prêt
docker-compose ps
docker-compose logs postgres

# Vérifier que Redis est prêt
docker-compose logs redis

# Vérifier les variables d'environnement
cat .env
```

### Erreur "database not found"
```bash
# Recréer la base
make migrate-down
make migrate-up
```

### WebSocket connection refused
```bash
# Vérifier que le Zone Server est lancé
ps aux | grep zone

# Vérifier le port
lsof -i :8081
```

---

## 📋 Checklist de Validation Sprint 1

- [x] Un joueur peut s'inscrire
- [x] Un joueur peut se connecter
- [x] Un joueur peut créer un personnage avec maison
- [x] Un joueur peut lister ses personnages
- [x] Un joueur peut sélectionner un personnage
- [x] Le JWT contient character_id, house, grade, permissions
- [x] Un joueur peut se connecter au Zone Server via WebSocket
- [x] Le Zone Server charge l'état du personnage depuis la DB
- [x] Le serveur envoie WORLD_SNAPSHOT avec état du personnage
- [x] Heartbeat fonctionne avec calcul de latence
- [x] Un GM peut modifier le grade
- [x] Un GM peut téléporter un personnage
- [x] Les actions GM sont auditées avec IP et timestamp
- [x] Les audit logs sont consultables par caractère et par admin
- [ ] Tests d'intégration (TODO)

---

## 🎯 Prochaines Étapes

### Sprint 1 (derniers items)
- [x] Endpoints admin pour GM commands (setgrade, teleport, grantitem)
- [x] Audit logging pour toutes les actions admin
- [ ] Tests d'intégration
- [ ] Documentation OpenAPI/Swagger

### Sprint 2
- [ ] Système d'inventaire complet
- [ ] Transactions atomiques
- [ ] State diff système
- [ ] Équipement et stats recalculées

### Sprint 3
- [ ] Actions de combat validées
- [ ] Cooldown système
- [ ] Anti-cheat basique
- [ ] Dashboards Grafana

---

## 📞 Support

Pour toute question ou problème:
1. Vérifier les logs : `make logs`
2. Vérifier la documentation : `/docs`
3. Vérifier l'état des services : `docker-compose ps`

Bon développement ! 🧙‍♂️✨
