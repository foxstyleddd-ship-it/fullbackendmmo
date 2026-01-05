# Backend MMO Harry Potter - Architecture Technique Complète

## Vue d'ensemble

Backend autoritaire pour MMO roleplay Harry Potter sous Unreal Engine + ACF.
**Principe fondamental** : Le client envoie des intentions, le serveur valide et renvoie des diffs d'état.

---

## 1. Diagramme d'Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                   CLIENTS (Unreal Engine + ACF)                      │
│                        Envoient uniquement des INTENTIONS                            │
└──────────────────────────────────┬──────────────────────────────────────────────────┘
                                   │ WSS/HTTPS
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY / LOAD BALANCER                             │
│                         (Traefik / Kong / AWS ALB)                                   │
│  - TLS Termination                                                                   │
│  - Rate Limiting                                                                     │
│  - JWT Validation                                                                    │
│  - Routing vers services                                                             │
└───────┬─────────────────┬─────────────────┬─────────────────┬───────────────────────┘
        │                 │                 │                 │
        ▼                 ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────────────────────┐
│  AUTH SERVICE │ │  API SERVICE  │ │ ADMIN SERVICE │ │      ZONE ROUTER SERVICE      │
│     (Go)      │ │     (Go)      │ │     (Go)      │ │           (Go)                │
│               │ │               │ │               │ │                               │
│ - Register    │ │ - Characters  │ │ - GM Commands │ │ - Player → Zone mapping       │
│ - Login/JWT   │ │ - Inventory   │ │ - Audit logs  │ │ - Transfer tokens             │
│ - Refresh     │ │ - Profiles    │ │ - Bans        │ │ - Load balancing zones        │
│ - Sessions    │ │ - Economy     │ │ - Broadcasts  │ │ - Health check zone servers   │
└───────┬───────┘ └───────┬───────┘ └───────┬───────┘ └───────────────┬───────────────┘
        │                 │                 │                         │
        │                 │                 │                         │
        ▼                 ▼                 ▼                         ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              MESSAGE BUS (NATS JetStream)                            │
│  - Events inter-services                                                             │
│  - Player state sync                                                                 │
│  - Cross-zone communication                                                          │
│  - GM broadcasts                                                                     │
│  Topics: player.*, zone.*, admin.*, economy.*, inventory.*                          │
└───────┬─────────────────┬─────────────────┬─────────────────┬───────────────────────┘
        │                 │                 │                 │
        ▼                 ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           ZONE SERVERS (Autoritaires)                                │
│                                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│  │  HOGWARTS   │  │ PRE-AU-LARD │  │   FOREST    │  │  QUIDDITCH  │                 │
│  │   Zone 1    │  │   Zone 2    │  │   Zone 3    │  │   Zone 4    │                 │
│  │             │  │             │  │             │  │             │                 │
│  │ - State     │  │ - State     │  │ - State     │  │ - State     │                 │
│  │ - Validation│  │ - Validation│  │ - Validation│  │ - Validation│                 │
│  │ - Physics   │  │ - Physics   │  │ - Physics   │  │ - Physics   │                 │
│  │ - Combat    │  │ - Combat    │  │ - Combat    │  │ - Combat    │                 │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘                 │
│                                                                                      │
│  Chaque zone server:                                                                 │
│  - WebSocket server pour clients connectés                                           │
│  - Valide TOUTES les intentions                                                      │
│  - Gère l'état local des entités (joueurs, NPCs, objets)                            │
│  - Envoie des STATE_DIFF aux clients                                                │
│  - Publie événements vers NATS                                                       │
└───────┬─────────────────────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              PERSISTENCE LAYER                                       │
│                                                                                      │
│  ┌─────────────────────────────┐    ┌─────────────────────────────────────────────┐ │
│  │      POSTGRESQL             │    │              REDIS CLUSTER                  │ │
│  │   (Source of Truth)         │    │         (Cache + Sessions + Presence)       │ │
│  │                             │    │                                             │ │
│  │ - Accounts                  │    │ - Session tokens (TTL 24h)                  │ │
│  │ - Characters                │    │ - Player presence (zone, status)            │ │
│  │ - Inventory                 │    │ - Rate limiting counters                    │ │
│  │ - Transactions              │    │ - Hot data cache (character stats)          │ │
│  │ - Audit logs                │    │ - Transfer tokens (TTL 30s)                 │ │
│  │ - World state               │    │ - Cooldowns                                 │ │
│  │ - Permissions               │    │ - Pub/Sub presence updates                  │ │
│  └─────────────────────────────┘    └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              OBSERVABILITY STACK                                     │
│                                                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────────┐ │
│  │   Grafana   │  │ Prometheus  │  │    Loki     │  │         Jaeger              │ │
│  │ Dashboards  │  │   Metrics   │  │    Logs     │  │    Distributed Traces       │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Choix Technologiques

| Composant | Technologie | Justification |
|-----------|-------------|---------------|
| **Services Backend** | Go 1.22+ | Performance, concurrence native (goroutines), faible latence, excellent pour gamedev backend |
| **API Gateway** | Traefik / Kong | Load balancing, TLS, rate limiting, routing dynamique |
| **Base de données** | PostgreSQL 16 | ACID, JSONB pour données flexibles, extensions (pg_cron, timescaledb) |
| **Cache/Sessions** | Redis 7 Cluster | Sub-ms latency, Pub/Sub, TTL natif, structures de données riches |
| **Message Bus** | NATS JetStream | Léger, rapide, persistence, parfait pour gaming (< 1ms latency) |
| **WebSocket** | gorilla/websocket | Standard Go, haute performance |
| **Auth** | JWT RS256 | Stateless, claims personnalisés, rotation de clés |
| **Orchestration** | Kubernetes | Scaling horizontal, rolling updates, health checks |
| **Observabilité** | Prometheus + Grafana + Loki + Jaeger | Stack standard, bien intégrée |

### Pourquoi Go plutôt que Rust/Node/etc ?

1. **Performance** : Comparable à Rust pour ce use case, bien meilleur que Node
2. **Concurrence** : Goroutines = modèle mental simple pour gérer milliers de connexions
3. **Écosystème gaming** : Nombreux backends de jeux en Go (Heroic Labs/Nakama est en Go)
4. **Productivité** : Compilation rapide, tooling excellent, courbe d'apprentissage raisonnable
5. **Deployment** : Binaire statique, image Docker < 20MB

---

## 3. Stratégie "Monde Unique" Scalable

### 3.1 Concept de Zones

Le monde est divisé en **zones logiques** (pas de loading screens visibles pour le joueur) :

```
ZONES:
├── hogwarts_main (Zone ID: 1)
│   ├── great_hall
│   ├── dungeons
│   ├── towers
│   └── courtyards
├── hogsmeade (Zone ID: 2)
│   ├── three_broomsticks
│   ├── honeydukes
│   └── shrieking_shack
├── forbidden_forest (Zone ID: 3)
│   ├── centaur_territory
│   ├── acromantula_nest
│   └── unicorn_glade
├── quidditch_pitch (Zone ID: 4)
└── ministry_of_magic (Zone ID: 5)
```

### 3.2 Zone Server Scaling

Chaque zone peut avoir **N instances** (shards) pour le scaling horizontal :

```
Zone "hogwarts_main":
  - hogwarts-main-shard-0 (players 0-500)
  - hogwarts-main-shard-1 (players 501-1000)
  - hogwarts-main-shard-2 (players 1001-1500)
```

Le **Zone Router** décide du shard en fonction de :
- Charge actuelle des shards
- Préférence de garder les groupes ensemble (house, party)
- Localisation géographique du joueur

### 3.3 Migration Inter-Zone (Transfer Protocol)

```
┌─────────┐         ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│  Client │         │ Zone Server │         │ Zone Router │         │  New Zone   │
│         │         │   Current   │         │             │         │   Server    │
└────┬────┘         └──────┬──────┘         └──────┬──────┘         └──────┬──────┘
     │                     │                       │                       │
     │ ZONE_TRANSFER_REQ   │                       │                       │
     │ (target_zone_id)    │                       │                       │
     │────────────────────>│                       │                       │
     │                     │                       │                       │
     │                     │ Persist player state  │                       │
     │                     │ to PostgreSQL         │                       │
     │                     │                       │                       │
     │                     │ REQUEST_TRANSFER_TOKEN│                       │
     │                     │──────────────────────>│                       │
     │                     │                       │                       │
     │                     │                       │ Select best shard     │
     │                     │                       │ Create signed token   │
     │                     │                       │ Store in Redis (30s)  │
     │                     │                       │                       │
     │                     │  TRANSFER_TOKEN       │                       │
     │                     │<──────────────────────│                       │
     │                     │                       │                       │
     │ ZONE_TRANSFER_RESP  │                       │                       │
     │ (token, new_addr)   │                       │                       │
     │<────────────────────│                       │                       │
     │                     │                       │                       │
     │ Disconnect from current zone                │                       │
     │                     │                       │                       │
     │ CONNECT + token     │                       │                       │
     │─────────────────────────────────────────────────────────────────────>│
     │                     │                       │                       │
     │                     │                       │     Validate token    │
     │                     │                       │     Load state from   │
     │                     │                       │     PostgreSQL        │
     │                     │                       │                       │
     │ WORLD_SNAPSHOT      │                       │                       │
     │<─────────────────────────────────────────────────────────────────────│
     │                     │                       │                       │
```

### 3.4 Transfer Token Structure

```json
{
  "token_id": "uuid-v4",
  "character_id": "char_123",
  "account_id": "acc_456",
  "source_zone": "hogwarts_main",
  "target_zone": "hogsmeade",
  "target_shard": "hogsmeade-shard-0",
  "target_address": "wss://zone-hogsmeade-0.game.internal:8443",
  "issued_at": 1704067200,
  "expires_at": 1704067230,
  "signature": "RS256_signature_base64"
}
```

---

## 4. Flow de Données Principal

```
CLIENT                    ZONE SERVER                 PERSISTENCE
  │                           │                           │
  │  TRY_CAST_SPELL          │                           │
  │  {spell: "expelliarmus"} │                           │
  │─────────────────────────>│                           │
  │                           │                           │
  │                           │ 1. Validate JWT           │
  │                           │ 2. Check cooldown (Redis) │
  │                           │ 3. Check grade permission │
  │                           │ 4. Check mana/resources   │
  │                           │ 5. Check range/LOS        │
  │                           │ 6. Execute spell logic    │
  │                           │ 7. Apply effects          │
  │                           │ 8. Update cooldown        │
  │                           │                           │
  │                           │  Update character state   │
  │                           │──────────────────────────>│
  │                           │                           │
  │  STATE_DIFF              │                           │
  │  {stats: {mana: -30},    │                           │
  │   cooldowns: {...},      │                           │
  │   effects: [...]}        │                           │
  │<─────────────────────────│                           │
  │                           │                           │
  │                           │  Broadcast to nearby      │
  │                           │  players                  │
  │                           │                           │
```

---

## 5. Sécurité & Anti-Cheat

### 5.1 Principes

1. **Server Authoritative** : Le client n'a JAMAIS raison par défaut
2. **Validation systématique** : Chaque intention validée contre l'état serveur
3. **Rate Limiting** : Par action, par endpoint, par IP
4. **Anomaly Detection** : Patterns suspects = flag + review

### 5.2 Validations Critiques

| Action | Validations |
|--------|-------------|
| MOVE | Distance max/tick, collision, zone bounds, speed hacks |
| CAST_SPELL | Cooldown, mana, grade permission, range, LOS |
| USE_ITEM | Ownership, stackable qty, cooldown, zone restriction |
| EQUIP | Ownership, slot compatibility, grade requirement |
| INTERACT | Distance, permission, state de l'objet |
| TRADE | Both online, distance, item ownership, qty |

### 5.3 JWT Claims

```json
{
  "sub": "acc_12345",
  "iss": "hp-mmo-auth",
  "iat": 1704067200,
  "exp": 1704153600,
  "jti": "unique-token-id",
  "account_id": "acc_12345",
  "character_id": "char_67890",
  "role": "player",
  "house": "venatrix",
  "grade": 5,
  "permissions": ["spell.tier1", "spell.tier2", "zone.forest"],
  "session_id": "sess_abc123"
}
```

---

## 6. Kubernetes Deployment Strategy

```yaml
# Namespace structure
namespaces:
  - hp-mmo-prod
  - hp-mmo-staging
  - hp-mmo-observability

# Services scaling
deployments:
  auth-service:
    replicas: 3
    resources:
      requests: { cpu: 500m, memory: 512Mi }
      limits: { cpu: 1000m, memory: 1Gi }

  api-service:
    replicas: 5
    resources:
      requests: { cpu: 1000m, memory: 1Gi }
      limits: { cpu: 2000m, memory: 2Gi }

  zone-router:
    replicas: 3
    resources:
      requests: { cpu: 500m, memory: 512Mi }
      limits: { cpu: 1000m, memory: 1Gi }

  zone-server-hogwarts:
    replicas: 3  # 3 shards
    resources:
      requests: { cpu: 2000m, memory: 4Gi }
      limits: { cpu: 4000m, memory: 8Gi }

  zone-server-hogsmeade:
    replicas: 2
    resources:
      requests: { cpu: 1000m, memory: 2Gi }
      limits: { cpu: 2000m, memory: 4Gi }

# StatefulSets
statefulsets:
  postgresql:
    replicas: 3  # Primary + 2 replicas
    storage: 500Gi SSD

  redis:
    replicas: 6  # 3 masters + 3 replicas
    storage: 50Gi SSD

  nats:
    replicas: 3
    storage: 100Gi SSD
```

---

## 7. Observabilité

### 7.1 Metrics Prometheus

```
# Player metrics
hp_mmo_players_online_total{zone, shard, house}
hp_mmo_player_actions_total{action_type, status, zone}
hp_mmo_player_latency_seconds{zone, percentile}

# Zone metrics
hp_mmo_zone_capacity_ratio{zone, shard}
hp_mmo_zone_tick_duration_seconds{zone}
hp_mmo_zone_transfers_total{from_zone, to_zone, status}

# Economy metrics
hp_mmo_transactions_total{type, status}
hp_mmo_currency_flow{direction, currency_type}

# System metrics
hp_mmo_ws_connections_active{zone}
hp_mmo_db_query_duration_seconds{query_type}
hp_mmo_redis_operations_total{operation}
```

### 7.2 Alerting Rules

```yaml
alerts:
  - name: HighPlayerLatency
    expr: hp_mmo_player_latency_seconds{percentile="p99"} > 0.1
    for: 5m
    severity: warning

  - name: ZoneOverloaded
    expr: hp_mmo_zone_capacity_ratio > 0.9
    for: 2m
    severity: critical

  - name: HighErrorRate
    expr: rate(hp_mmo_player_actions_total{status="error"}[5m]) > 0.01
    for: 5m
    severity: warning
```

### 7.3 Structured Logging

```json
{
  "timestamp": "2024-01-01T12:00:00.123Z",
  "level": "info",
  "service": "zone-server-hogwarts",
  "shard": "shard-0",
  "trace_id": "abc123",
  "span_id": "def456",
  "account_id": "acc_123",
  "character_id": "char_456",
  "action": "CAST_SPELL",
  "spell": "expelliarmus",
  "result": "success",
  "duration_ms": 2.5,
  "mana_cost": 30,
  "target": "char_789"
}
```
