# Plan MVP - 3 Sprints

## Vue d'Ensemble

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SPRINT 1: FONDATIONS                               │
│  Auth + Personnages + Enter World + Persistence + Admin Grade                │
├─────────────────────────────────────────────────────────────────────────────┤
│                           SPRINT 2: INVENTAIRE                               │
│  Inventaire Autoritaire + Transactions + State Diff + Logs/Audit            │
├─────────────────────────────────────────────────────────────────────────────┤
│                           SPRINT 3: GAMEPLAY                                 │
│  Actions Combat/Spells + Validation + Anti-Cheat + Dashboards               │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Sprint 1: Fondations

### Objectif
Pouvoir se connecter, créer un personnage, entrer dans le monde, et gérer les grades.

### Livrables

#### 1.1 Infrastructure Base

- [ ] **Setup projet Go**
  - Structure mono-repo avec modules
  - Configuration (Viper/envconfig)
  - Logging structuré (zerolog/zap)
  - Makefile avec targets: build, test, lint, docker

- [ ] **Docker Compose dev**
  ```yaml
  services:
    postgres:
      image: postgres:16
      volumes: [./init.sql:/docker-entrypoint-initdb.d/]
    redis:
      image: redis:7
    nats:
      image: nats:2.10
    api:
      build: .
      depends_on: [postgres, redis]
  ```

- [ ] **Migrations DB**
  - Outil: golang-migrate
  - Scripts initiaux depuis DATABASE_SCHEMA.sql

#### 1.2 Service Auth

- [ ] **Endpoints REST**
  - `POST /auth/register`
  - `POST /auth/login`
  - `POST /auth/refresh`
  - `POST /auth/logout`

- [ ] **JWT RS256**
  - Génération/validation tokens
  - Claims: account_id, role
  - Refresh token rotation

- [ ] **Session management**
  - Redis: `session:{session_id}` → account data
  - TTL 24h

- [ ] **Tests**
  - Unit tests auth logic
  - Integration tests endpoints

#### 1.3 Service Characters

- [ ] **Endpoints REST**
  - `GET /characters`
  - `POST /characters`
  - `GET /characters/{id}`
  - `POST /characters/{id}/select`
  - `DELETE /characters/{id}`

- [ ] **Logique métier**
  - Limite 5 personnages par compte
  - Validation nom unique
  - Attribution maison (choix joueur)
  - Grade initial = 1

- [ ] **Select character**
  - Génère JWT avec character_id
  - Retourne infos connexion zone

#### 1.4 Zone Server (Minimal)

- [ ] **WebSocket server**
  - Gorilla WebSocket
  - Connection handling
  - Heartbeat

- [ ] **Messages implémentés**
  - `CONNECT` / `CONNECT_ACK`
  - `ENTER_WORLD` / `WORLD_SNAPSHOT`
  - `HEARTBEAT` / `HEARTBEAT_ACK`
  - `PRESENCE_UPDATE`

- [ ] **État en mémoire**
  - Map des joueurs connectés
  - Positions (pas encore validées)

- [ ] **Persistence basique**
  - Sauvegarde position au disconnect
  - Chargement position au connect

#### 1.5 Admin: Grade Management

- [ ] **Endpoint REST**
  - `PUT /admin/characters/{id}/grade`
  - Validation rôle (GM/Admin)

- [ ] **WebSocket command**
  - `GM_COMMAND_EXEC` pour `setgrade`
  - `GM_COMMAND_RESULT`

- [ ] **Audit logging**
  - Écriture `audit_logs` table
  - Raison obligatoire

- [ ] **Tests**
  - Scénario complet: login → create char → enter world → GM setgrade

### Critères de Succès Sprint 1

- [ ] Un joueur peut s'inscrire et se connecter
- [ ] Un joueur peut créer un personnage avec maison
- [ ] Un joueur peut entrer dans le monde et voir sa position
- [ ] Un joueur peut voir les autres joueurs (presence)
- [ ] Un GM peut modifier le grade d'un personnage
- [ ] Les actions sont auditées

### Stack Technique Sprint 1

```
Go 1.22+
├── github.com/gin-gonic/gin          (HTTP)
├── github.com/gorilla/websocket      (WS)
├── github.com/golang-jwt/jwt/v5      (JWT)
├── github.com/jmoiron/sqlx           (SQL)
├── github.com/redis/go-redis/v9      (Redis)
├── github.com/rs/zerolog             (Logging)
└── github.com/stretchr/testify       (Testing)
```

---

## Sprint 2: Inventaire & Économie

### Objectif
Système d'inventaire complet, transactions atomiques, state diff temps réel.

### Livrables

#### 2.1 Inventaire

- [ ] **Schema complet**
  - `item_definitions` seed avec 20+ items
  - `inventory_items` opérationnel
  - `equipment_loadouts` opérationnel

- [ ] **Endpoints REST**
  - `GET /characters/{id}/inventory`
  - `POST /characters/{id}/inventory/use`
  - `POST /characters/{id}/inventory/equip`
  - `POST /characters/{id}/inventory/unequip`
  - `POST /characters/{id}/inventory/move`

- [ ] **WebSocket intents**
  - `INVENTORY_ACTION_INTENT`
  - `INVENTORY_ACTION_RESULT`

- [ ] **Validation serveur**
  - Ownership vérifié
  - Quantity suffisante
  - Slot compatible
  - Cooldowns respectés

#### 2.2 Équipement

- [ ] **10 slots équipement**
  - Wand, robe, hat, cloak, amulet, ring_left, ring_right, boots, gloves, broom

- [ ] **Calcul stats équipement**
  - Computed stats from loadout
  - Mise à jour dynamique

- [ ] **Restrictions**
  - Grade requis par item
  - Maison (pour certains)

#### 2.3 Transactions Atomiques

- [ ] **Function SQL**
  - `execute_inventory_transaction()` opérationnelle
  - Rollback automatique sur erreur

- [ ] **Idempotency**
  - Clé unique par transaction
  - Retry-safe

- [ ] **Ledger**
  - `transactions_ledger` alimenté
  - Historique complet

#### 2.4 State Diff System

- [ ] **WebSocket messages**
  - `STATE_DIFF` implémenté
  - Format delta objects

- [ ] **Diff computation**
  - Comparaison état avant/après
  - Génération patch minimal

- [ ] **Broadcast**
  - Diff envoyé aux joueurs concernés
  - Cible (self) vs observers

#### 2.5 Currencies

- [ ] **4 currencies**
  - Galleons, Sickles, Knuts, House Points

- [ ] **Endpoints**
  - `GET /characters/{id}/currencies`
  - `POST /admin/characters/{id}/economy/grant`

- [ ] **Atomic operations**
  - Add/remove atomique
  - Validation solde >= 0

#### 2.6 Audit & Logs Enrichis

- [ ] **Audit complet**
  - Toutes les mutations auditées
  - Recherche par actor/target/action

- [ ] **Endpoint audit**
  - `GET /admin/audit`
  - Filtres et pagination

- [ ] **Logs structurés**
  - Format JSON
  - Trace ID propagé
  - Intégration Loki ready

### Critères de Succès Sprint 2

- [ ] Un joueur peut voir son inventaire
- [ ] Un joueur peut utiliser un item consommable
- [ ] Un joueur peut équiper/déséquiper
- [ ] Les stats sont recalculées après équipement
- [ ] Un GM peut donner des items
- [ ] Un GM peut donner des currencies
- [ ] Les transactions sont atomiques (test de stress)
- [ ] Les state diff sont envoyés correctement
- [ ] L'audit trail est complet

### Tests Sprint 2

```go
// Test atomicité
func TestConcurrentItemUse(t *testing.T) {
    // 10 goroutines essaient d'utiliser le même item stack
    // Seules N devraient réussir (N = quantity initiale)
}

// Test idempotence
func TestIdempotentTransaction(t *testing.T) {
    // Même idempotency_key 3 fois
    // Résultat identique, exécution unique
}

// Test state diff
func TestStateDiffAfterEquip(t *testing.T) {
    // Équiper item
    // Vérifier STATE_DIFF reçu avec bon delta
}
```

---

## Sprint 3: Gameplay & Anti-Cheat

### Objectif
Actions de combat/spells validées côté serveur, anti-cheat minimal, observabilité.

### Livrables

#### 3.1 Combat Actions

- [ ] **WebSocket intents**
  - `COMBAT_ACTION_INTENT`
  - `COMBAT_ACTION_RESULT`

- [ ] **Types d'actions**
  - `cast_spell`
  - `melee_attack`
  - `use_ability`
  - `dodge`

- [ ] **10 sorts implémentés**
  ```yaml
  spells:
    - lumos (tier 1)
    - expelliarmus (tier 2)
    - stupefy (tier 3)
    - protego (tier 3)
    - expecto_patronum (tier 4)
    # ... etc
  ```

#### 3.2 Validation Combat

- [ ] **Checks complets**
  - Cooldown
  - Mana/ressources
  - Grade/tier
  - Range
  - Line of sight (simplifié)
  - House restrictions

- [ ] **Calcul dégâts**
  - Formule: base_damage * spell_power_modifier * crit
  - Résistances appliquées

- [ ] **Effets**
  - Debuffs (slow, disarm)
  - Durées gérées

#### 3.3 Cooldown System

- [ ] **Redis-based**
  - `cooldown:{character_id}:{action_id}` → expiry timestamp
  - TTL automatique

- [ ] **Sync client**
  - Cooldowns dans STATE_DIFF
  - Recovery times précis

#### 3.4 Movement Validation

- [ ] **Position checks**
  - Vitesse max par tick
  - Bounds de zone
  - Collision basique

- [ ] **Corrections**
  - `POSITION_CORRECTION` soft/hard
  - Threshold configurable

- [ ] **Anti-speedhack**
  - Détection patterns suspects
  - Logging pour review

#### 3.5 Anti-Cheat Minimal

- [ ] **Rate limiting**
  - Par type de message
  - Par joueur

- [ ] **Anomaly counters**
  - Redis increments
  - Thresholds configurables

- [ ] **Auto-actions**
  - Warning après N violations
  - Kick après 2N
  - Flag pour review

- [ ] **Logging sécurité**
  - Events suspects logués
  - Dashboard ready

#### 3.6 Quests (Basique)

- [ ] **WebSocket intents**
  - `QUEST_INTENT` (start, advance, claim)
  - `QUEST_RESULT`

- [ ] **3 quêtes test**
  - Quest tutoriel
  - Quest collecte
  - Quest combat

- [ ] **Rewards**
  - XP
  - Currencies
  - Items

#### 3.7 Observabilité

- [ ] **Prometheus metrics**
  ```
  hp_mmo_players_online_total
  hp_mmo_player_actions_total
  hp_mmo_zone_tick_duration_seconds
  hp_mmo_ws_connections_active
  hp_mmo_db_query_duration_seconds
  ```

- [ ] **Grafana dashboards**
  - Players overview
  - Zone health
  - Economy flow
  - Error rates

- [ ] **Loki logs**
  - Tous logs en JSON
  - Labels: service, zone, character_id

- [ ] **Alerting**
  - High latency
  - Error spikes
  - Zone overload

#### 3.8 GM Commands Complets

- [ ] **Commands restantes**
  - `teleport`
  - `kick`
  - `ban/unban`
  - `grantitem`
  - `grantcurrency`
  - `announce`
  - `summon`
  - `freeze/unfreeze`

- [ ] **Dashboard admin** (basique)
  - Liste joueurs online
  - Recherche joueur
  - Exécution commandes
  - Audit viewer

### Critères de Succès Sprint 3

- [ ] Un joueur peut lancer un sort sur une cible
- [ ] Le serveur valide cooldown, mana, grade, range
- [ ] Le state diff reflète les changements
- [ ] Les mouvements anormaux sont détectés et corrigés
- [ ] Les metrics Prometheus sont exposées
- [ ] Les dashboards Grafana fonctionnent
- [ ] Un GM peut utiliser toutes les commandes
- [ ] Les alertes se déclenchent correctement

### Load Testing Sprint 3

```bash
# Scénario: 500 joueurs simultanés
# - 100 actions/seconde totales
# - 50% mouvements, 30% spells, 20% inventaire

k6 run load_test.js
```

Objectifs:
- p99 latency < 100ms
- Error rate < 0.1%
- Zero data loss

---

## Structure Projet Finale

```
hp-mmo-backend/
├── cmd/
│   ├── api/                  # Service API REST
│   ├── auth/                 # Service Auth (optionnel séparé)
│   ├── zone/                 # Zone Server
│   └── admin/                # Admin tools CLI
│
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   ├── session.go
│   │   └── middleware.go
│   │
│   ├── character/
│   │   ├── model.go
│   │   ├── repository.go
│   │   └── service.go
│   │
│   ├── inventory/
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── transaction.go
│   │
│   ├── combat/
│   │   ├── spell.go
│   │   ├── validator.go
│   │   └── damage.go
│   │
│   ├── zone/
│   │   ├── server.go
│   │   ├── client.go
│   │   ├── world.go
│   │   └── movement.go
│   │
│   ├── anticheat/
│   │   ├── detector.go
│   │   └── reporter.go
│   │
│   ├── admin/
│   │   ├── commands.go
│   │   └── audit.go
│   │
│   └── pkg/
│       ├── db/
│       ├── redis/
│       ├── nats/
│       └── metrics/
│
├── api/
│   └── openapi.yaml          # OpenAPI spec
│
├── migrations/
│   ├── 001_initial.up.sql
│   ├── 001_initial.down.sql
│   └── ...
│
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   └── k8s/
│       ├── base/
│       └── overlays/
│
├── scripts/
│   ├── seed_items.sql
│   ├── seed_zones.sql
│   └── load_test.js
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── DATABASE_SCHEMA.sql
│   ├── API_REST.md
│   ├── WEBSOCKET_PROTOCOL.md
│   ├── BUSINESS_RULES.md
│   └── MVP_ROADMAP.md
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Commandes Dev

```makefile
# Makefile

.PHONY: dev test build docker

# Démarrer l'environnement dev
dev:
	docker-compose up -d postgres redis nats
	go run ./cmd/api

# Lancer les tests
test:
	go test ./... -v -cover

# Tests d'intégration
test-integration:
	docker-compose up -d
	go test ./... -tags=integration -v

# Build production
build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -o bin/zone ./cmd/zone

# Build Docker
docker:
	docker build -t hp-mmo-api:latest -f deployments/docker/Dockerfile .

# Migrations
migrate-up:
	migrate -path migrations -database "postgres://..." up

migrate-down:
	migrate -path migrations -database "postgres://..." down 1

# Lint
lint:
	golangci-lint run

# Generate mocks
mocks:
	mockgen -source=internal/character/repository.go -destination=mocks/character_repository.go
```

---

## Risques et Mitigations

| Risque | Impact | Mitigation |
|--------|--------|------------|
| Latence WebSocket | High | Connection pooling, optimize serialization |
| Race conditions inventaire | Critical | Transactions atomiques, tests de stress |
| Scaling zones | Medium | Design stateless, Redis pour état partagé |
| Anti-cheat bypass | Medium | Logging détaillé, review manuelle |
| Downtime migrations | Low | Zero-downtime migrations, feature flags |

---

## Post-MVP (Sprint 4+)

### Sprint 4: Polish
- Zone transfers complets
- Quests système étendu
- Trading entre joueurs
- Amélioration anti-cheat

### Sprint 5: Scale
- Sharding multi-zone
- Kubernetes autoscaling
- CDN pour assets
- Backup/restore procedures

### Sprint 6: Features
- Guildes/Houses leaderboards
- Events système
- Achievements
- Social features

---

## Checklist Go-Live

### Infrastructure
- [ ] PostgreSQL HA (Primary + 2 replicas)
- [ ] Redis Cluster (6 nodes)
- [ ] NATS Cluster (3 nodes)
- [ ] Kubernetes cluster (3+ nodes)
- [ ] Load balancer avec TLS
- [ ] CDN configuré

### Sécurité
- [ ] Secrets dans Vault/AWS Secrets Manager
- [ ] Network policies Kubernetes
- [ ] Rate limiting activé
- [ ] WAF configuré
- [ ] Audit logs activés

### Monitoring
- [ ] Prometheus scraping OK
- [ ] Grafana dashboards opérationnels
- [ ] Alertes configurées
- [ ] On-call rotation définie

### Documentation
- [ ] Runbook opérations
- [ ] Incident response plan
- [ ] API documentation publique
- [ ] GM training documentation

### Testing
- [ ] Load test 1000 concurrent
- [ ] Chaos testing (kill pods)
- [ ] Failover testing
- [ ] Backup restore tested
