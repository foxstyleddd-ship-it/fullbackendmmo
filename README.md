# HP MMO Backend - Architecture Technique

Backend autoritaire pour MMO roleplay Harry Potter sous Unreal Engine + ACF (Ascent Combat Framework).

## 🚀 État du Projet

**Sprint 1 en cours** - Infrastructure de base et services core

✅ Architecture documentée
✅ Schéma de base de données complet
✅ API REST spécifiée
✅ Protocole WebSocket défini
✅ Structure du projet créée
✅ Docker Compose configuré
✅ Migrations de base de données
🚧 Implémentation Go en cours

## Principe Fondamental

**Le client envoie des intentions, le serveur valide et renvoie des diffs d'état.**

```
Client (Unreal/ACF)          Zone Server (Go)              Database
       │                           │                           │
       │  INTENT {action}          │                           │
       │──────────────────────────>│                           │
       │                           │                           │
       │                    ┌──────┴──────┐                    │
       │                    │  VALIDATE   │                    │
       │                    │  • Auth     │                    │
       │                    │  • Rules    │                    │
       │                    │  • State    │                    │
       │                    └──────┬──────┘                    │
       │                           │                           │
       │                           │────────────────────────────>
       │                           │          PERSIST          │
       │                           │<────────────────────────────
       │                           │                           │
       │  RESULT + STATE_DIFF      │                           │
       │<──────────────────────────│                           │
```

## Documentation

| Document | Description |
|----------|-------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Diagrammes services, choix techniques, scaling |
| [DATABASE_SCHEMA.sql](docs/DATABASE_SCHEMA.sql) | Schéma PostgreSQL complet avec exemples |
| [API_REST.md](docs/API_REST.md) | Endpoints REST, payloads, codes d'erreur |
| [WEBSOCKET_PROTOCOL.md](docs/WEBSOCKET_PROTOCOL.md) | Messages temps réel, state diff |
| [BUSINESS_RULES.md](docs/BUSINESS_RULES.md) | Maisons, grades, permissions, GM commands |
| [MVP_ROADMAP.md](docs/MVP_ROADMAP.md) | Plan 3 sprints avec livrables |

## Stack Technique

| Composant | Technologie |
|-----------|-------------|
| Services | Go 1.22+ |
| Base de données | PostgreSQL 16 |
| Cache/Sessions | Redis 7 Cluster |
| Message Bus | NATS JetStream |
| WebSocket | gorilla/websocket |
| Auth | JWT RS256 |
| Orchestration | Kubernetes |
| Observabilité | Prometheus + Grafana + Loki + Jaeger |

## Maisons

| Maison | Description |
|--------|-------------|
| **Venatrix** | Les chasseurs rusés et ambitieux |
| **Aerwyn** | Les érudits et les sages |
| **Falcon** | Les courageux et téméraires |
| **Brumval** | Les loyaux et persévérants |

## Grades

Le grade (1-100) détermine les permissions: sorts, zones, items.

| Grade | Équivalent | Permissions clés |
|-------|------------|------------------|
| 1-2 | Années 1-2 | Sorts tier 1-2 |
| 3-4 | Années 3-4 | Forêt interdite, tier 3 |
| 5-7 | Années 5-7 | Ministère, tier 4 |
| 8-10 | Formation avancée | Azkaban, tier 5 |
| 11+ | Maître/Expert | Tout |

## 🏁 Quick Start (Dev)

### Prérequis

- Docker & Docker Compose
- Go 1.22+ (pour développement local)
- Make

### Démarrage Rapide

```bash
# 1. Cloner le repository
git clone <repo-url>
cd fullbackendmmo

# 2. Copier le fichier d'environnement
cp .env.example .env

# 3. Démarrer l'infrastructure (PostgreSQL, Redis, NATS, Prometheus, Grafana)
make dev

# 4. Appliquer les migrations de base de données
make migrate-up

# 5. (Optionnel) Seed la base de données avec des données de test
make seed

# 6. Lancer le serveur API (quand implémenté)
make run-api

# 7. Dans un autre terminal, lancer le Zone Server (quand implémenté)
make run-zone
```

### Accès aux Services

| Service | URL | Credentials |
|---------|-----|-------------|
| PostgreSQL | `localhost:5432` | user: `hpmmo`, password: `hpmmo_dev_password` |
| Redis | `localhost:6379` | password: `hpmmo_redis_password` |
| NATS | `localhost:4222` | - |
| Prometheus | `http://localhost:9090` | - |
| Grafana | `http://localhost:3000` | user: `admin`, password: `admin` |
| API Server | `http://localhost:8080` | (when running) |
| Zone Server | `ws://localhost:8081` | (when running) |

### Commandes Utiles

```bash
# Voir les logs
make logs

# Se connecter à PostgreSQL
make psql

# Se connecter à Redis
make redis-cli

# Lancer les tests
make test

# Lancer le linter
make lint

# Formater le code
make fmt

# Nettoyer les builds
make clean

# Arrêter tous les services
make dev-down
```

## Commandes GM

```bash
# Exemples via API
POST /admin/commands/execute

# setgrade
{"command": "setgrade", "target_character_id": "char_x1y2z3", "parameters": {"grade": 8}, "reason": "Promotion"}

# teleport
{"command": "teleport", "target_character_id": "char_x1y2z3", "parameters": {"zone_id": "hogsmeade"}, "reason": "Déblocage"}

# grantitem
{"command": "grantitem", "target_character_id": "char_x1y2z3", "parameters": {"items": [{"item_def_id": "broom_firebolt", "quantity": 1}]}, "reason": "Récompense event"}
```

## Flow Exemple: Login → Cast Spell

```
1. POST /auth/login
   → JWT + liste personnages

2. POST /characters/{id}/select
   → JWT avec character_id + zone connection info

3. WebSocket Connect + CONNECT message
   → CONNECT_ACK avec session

4. ENTER_WORLD
   → WORLD_SNAPSHOT (position, stats, nearby entities)

5. COMBAT_ACTION_INTENT {cast_spell: "expelliarmus", target: "char_enemy"}
   → Server validates: cooldown ✓, mana ✓, grade ✓, range ✓
   → COMBAT_ACTION_RESULT {success: true, damage: 45, state_diff: {...}}
```

## License

Proprietary - All rights reserved
