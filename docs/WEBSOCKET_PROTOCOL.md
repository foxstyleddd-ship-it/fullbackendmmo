# Protocole WebSocket Temps Réel - Spécification Complète

## Connexion

### URL de Connexion

```
wss://zone-{zone}-{shard}.hp-mmo.game/ws?token={connection_token}&v=1
```

Paramètres:
- `token`: Token de connexion obtenu via REST API (`/zones/{zone_id}/connection`)
- `v`: Version du protocole (actuellement `1`)

### Handshake

Le client doit s'authentifier dans les 5 secondes suivant la connexion WebSocket, sinon la connexion est fermée.

---

## Format des Messages

### Enveloppe Commune (Client → Server)

```json
{
  "msg_type": "CAST_SPELL_INTENT",
  "msg_id": "uuid-v4-unique",
  "timestamp": 1704067200123,
  "payload": {
    // Données spécifiques au message
  }
}
```

| Champ | Type | Description |
|-------|------|-------------|
| `msg_type` | string | Type de message (voir catalogue ci-dessous) |
| `msg_id` | string | UUID v4 unique pour corrélation request/response |
| `timestamp` | int64 | Timestamp client (ms depuis epoch) |
| `payload` | object | Données du message |

### Enveloppe Commune (Server → Client)

```json
{
  "msg_type": "CAST_SPELL_RESULT",
  "msg_id": "uuid-v4-unique",
  "ref_msg_id": "uuid-of-request",
  "timestamp": 1704067200150,
  "success": true,
  "payload": {
    // Données spécifiques
  },
  "error": null
}
```

| Champ | Type | Description |
|-------|------|-------------|
| `msg_type` | string | Type de message |
| `msg_id` | string | UUID unique de cette réponse |
| `ref_msg_id` | string | UUID du message client (pour corrélation) |
| `timestamp` | int64 | Timestamp serveur |
| `success` | boolean | Succès ou échec |
| `payload` | object | Données |
| `error` | object/null | Détails erreur si `success=false` |

### Format Erreur

```json
{
  "error": {
    "code": "INSUFFICIENT_MANA",
    "message": "Mana insuffisant pour lancer ce sort",
    "details": {
      "required": 50,
      "current": 30
    }
  }
}
```

---

## Codes d'Erreur WebSocket

| Code | Description |
|------|-------------|
| `INVALID_MESSAGE` | Message malformé |
| `UNKNOWN_MSG_TYPE` | Type de message inconnu |
| `RATE_LIMITED` | Trop de messages |
| `UNAUTHORIZED` | Session invalide |
| `FORBIDDEN` | Action non autorisée |
| `INSUFFICIENT_GRADE` | Grade insuffisant |
| `INSUFFICIENT_MANA` | Mana insuffisant |
| `INSUFFICIENT_STAMINA` | Stamina insuffisante |
| `ON_COOLDOWN` | Action en cooldown |
| `OUT_OF_RANGE` | Cible hors de portée |
| `INVALID_TARGET` | Cible invalide |
| `ITEM_NOT_FOUND` | Item non trouvé |
| `ZONE_RESTRICTED` | Zone non accessible |
| `INTERNAL_ERROR` | Erreur serveur |

---

## Règles ACK/NACK et Retry

### Côté Client

1. **Timeout**: Si pas de réponse après 5 secondes, retry automatique (max 3 fois)
2. **Idempotence**: Réutiliser le même `msg_id` pour les retries
3. **Backoff**: Délai entre retries: 1s, 2s, 4s

### Côté Serveur

1. **Deduplication**: Le serveur garde les `msg_id` traités pendant 60 secondes
2. **Même msg_id**: Retourne le résultat original (pas de re-exécution)

---

## Catalogue des Messages

### 1. Session & Connexion

#### CONNECT (Client → Server)

Premier message après connexion WebSocket.

```json
{
  "msg_type": "CONNECT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1704067200000,
  "payload": {
    "connection_token": "temp_conn_token_xyz789",
    "character_id": "char_x1y2z3",
    "client_version": "1.2.3",
    "platform": "windows"
  }
}
```

#### CONNECT_ACK (Server → Client)

```json
{
  "msg_type": "CONNECT_ACK",
  "msg_id": "660e8400-e29b-41d4-a716-446655440001",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1704067200050,
  "success": true,
  "payload": {
    "session_id": "sess_abc123",
    "character": {
      "id": "char_x1y2z3",
      "name": "Draco Testus",
      "house": "venatrix",
      "grade": 5
    },
    "zone": {
      "id": "hogwarts_main",
      "shard": "hogwarts-main-shard-1"
    },
    "server_time": 1704067200050,
    "tick_rate_ms": 50,
    "heartbeat_interval_ms": 10000
  }
}
```

---

#### RESUME_SESSION (Client → Server)

Reconnexion après déconnexion temporaire.

```json
{
  "msg_type": "RESUME_SESSION",
  "msg_id": "550e8400-e29b-41d4-a716-446655440010",
  "timestamp": 1704067300000,
  "payload": {
    "session_id": "sess_abc123",
    "last_received_msg_id": "770e8400-e29b-41d4-a716-446655440099",
    "last_received_seq": 12345
  }
}
```

#### RESUME_SESSION_ACK (Server → Client)

```json
{
  "msg_type": "RESUME_SESSION_ACK",
  "msg_id": "660e8400-e29b-41d4-a716-446655440011",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440010",
  "timestamp": 1704067300050,
  "success": true,
  "payload": {
    "resumed": true,
    "missed_events_count": 3,
    "missed_events": [
      // Events manqués depuis last_received_msg_id
    ]
  }
}
```

---

#### HEARTBEAT (Client → Server)

Ping périodique pour maintenir la connexion.

```json
{
  "msg_type": "HEARTBEAT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440020",
  "timestamp": 1704067210000,
  "payload": {
    "client_time": 1704067210000
  }
}
```

#### HEARTBEAT_ACK (Server → Client)

```json
{
  "msg_type": "HEARTBEAT_ACK",
  "msg_id": "660e8400-e29b-41d4-a716-446655440021",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440020",
  "timestamp": 1704067210025,
  "success": true,
  "payload": {
    "server_time": 1704067210025,
    "latency_ms": 25
  }
}
```

---

### 2. World & Zones

#### ENTER_WORLD (Client → Server)

Demande d'entrée dans le monde (après CONNECT).

```json
{
  "msg_type": "ENTER_WORLD",
  "msg_id": "550e8400-e29b-41d4-a716-446655440030",
  "timestamp": 1704067200100,
  "payload": {
    "zone_id": "hogwarts_main",
    "requested_position": null
  }
}
```

#### WORLD_SNAPSHOT (Server → Client)

Snapshot initial du monde visible.

```json
{
  "msg_type": "WORLD_SNAPSHOT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440031",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440030",
  "timestamp": 1704067200150,
  "success": true,
  "payload": {
    "zone": {
      "id": "hogwarts_main",
      "display_name": "Poudlard - Château Principal",
      "is_safe_zone": true
    },
    "self": {
      "character_id": "char_x1y2z3",
      "position": { "x": 100.5, "y": 50.2, "z": 0 },
      "rotation": { "yaw": 45.0 },
      "stats": {
        "health_current": 100,
        "health_max": 120,
        "mana_current": 80,
        "mana_max": 150,
        "stamina_current": 100,
        "stamina_max": 100
      },
      "flags": ["quest.chapter1.active"],
      "cooldowns": {
        "spell.expelliarmus": 0,
        "spell.stupefy": 1704067205000
      },
      "effects": []
    },
    "nearby_players": [
      {
        "character_id": "char_abc123",
        "name": "Hermione Test",
        "house": "aerwyn",
        "grade": 7,
        "position": { "x": 105.0, "y": 52.0, "z": 0 },
        "rotation": { "yaw": 180.0 },
        "visible_equipment": {
          "wand": "wand_vine_dragon",
          "robe": "robe_aerwyn_formal"
        },
        "effects": ["effect.shield_active"]
      }
    ],
    "nearby_npcs": [
      {
        "entity_id": "npc_dumbledore_01",
        "npc_def_id": "npc_dumbledore",
        "position": { "x": 110.0, "y": 60.0, "z": 0 },
        "rotation": { "yaw": 270.0 },
        "state": "idle",
        "interactable": true
      }
    ],
    "nearby_objects": [
      {
        "entity_id": "obj_chest_001",
        "object_type": "chest",
        "position": { "x": 95.0, "y": 48.0, "z": 0 },
        "state": { "is_open": false, "locked": true }
      }
    ]
  }
}
```

---

#### ZONE_TRANSFER_REQUEST (Client → Server)

Demande de changement de zone.

```json
{
  "msg_type": "ZONE_TRANSFER_REQUEST",
  "msg_id": "550e8400-e29b-41d4-a716-446655440040",
  "timestamp": 1704067500000,
  "payload": {
    "target_zone_id": "hogsmeade",
    "entry_point": "main_gate"
  }
}
```

#### ZONE_TRANSFER_RESPONSE (Server → Client)

```json
{
  "msg_type": "ZONE_TRANSFER_RESPONSE",
  "msg_id": "660e8400-e29b-41d4-a716-446655440041",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440040",
  "timestamp": 1704067500100,
  "success": true,
  "payload": {
    "approved": true,
    "target_zone": {
      "id": "hogsmeade",
      "shard": "hogsmeade-shard-0",
      "websocket_url": "wss://zone-hogsmeade-0.hp-mmo.game/ws"
    },
    "transfer_token": "transfer_xyz789_signed",
    "token_expires_at": 1704067530000,
    "instructions": "Disconnect from current zone and connect to new URL with transfer_token"
  }
}
```

Échec (grade insuffisant):
```json
{
  "msg_type": "ZONE_TRANSFER_RESPONSE",
  "msg_id": "660e8400-e29b-41d4-a716-446655440041",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440040",
  "timestamp": 1704067500100,
  "success": false,
  "payload": null,
  "error": {
    "code": "INSUFFICIENT_GRADE",
    "message": "Grade 8 requis pour entrer à Azkaban",
    "details": {
      "required_grade": 8,
      "current_grade": 5
    }
  }
}
```

---

### 3. Presence & Players

#### PRESENCE_UPDATE (Server → Client)

Broadcast quand un joueur entre/sort/change d'état.

```json
{
  "msg_type": "PRESENCE_UPDATE",
  "msg_id": "660e8400-e29b-41d4-a716-446655440050",
  "timestamp": 1704067300000,
  "payload": {
    "event_type": "player_entered",
    "player": {
      "character_id": "char_new123",
      "name": "Ron Test",
      "house": "brumval",
      "grade": 4,
      "position": { "x": 90.0, "y": 45.0, "z": 0 },
      "rotation": { "yaw": 0 },
      "visible_equipment": {
        "wand": "wand_willow"
      }
    }
  }
}
```

Event types: `player_entered`, `player_left`, `player_status_changed`

---

### 4. Movement

#### MOVE_INTENT (Client → Server)

Le client envoie son intention de mouvement. Le serveur valide et corrige si nécessaire.

```json
{
  "msg_type": "MOVE_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440060",
  "timestamp": 1704067200200,
  "payload": {
    "input_vector": { "x": 1.0, "y": 0.5 },
    "is_sprinting": false,
    "is_jumping": false,
    "client_position": { "x": 101.5, "y": 50.7, "z": 0 },
    "client_rotation": { "yaw": 30.0 },
    "sequence_num": 12345
  }
}
```

#### POSITION_CORRECTION (Server → Client)

Si le serveur détecte une dérive trop importante.

```json
{
  "msg_type": "POSITION_CORRECTION",
  "msg_id": "660e8400-e29b-41d4-a716-446655440061",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440060",
  "timestamp": 1704067200220,
  "success": true,
  "payload": {
    "correction_type": "soft",
    "server_position": { "x": 101.3, "y": 50.5, "z": 0 },
    "server_rotation": { "yaw": 30.0 },
    "server_velocity": { "x": 2.0, "y": 1.0, "z": 0 },
    "sequence_ack": 12345,
    "reason": "position_drift_exceeded_threshold"
  }
}
```

Correction types:
- `none`: Pas de correction nécessaire
- `soft`: Interpoler vers la position serveur
- `hard`: Téléporter immédiatement (anti-cheat)

---

#### ENTITY_POSITIONS (Server → Client)

Mise à jour périodique des positions des entités proches (20 Hz typiquement).

```json
{
  "msg_type": "ENTITY_POSITIONS",
  "msg_id": "660e8400-e29b-41d4-a716-446655440070",
  "timestamp": 1704067200250,
  "payload": {
    "tick": 45678,
    "entities": [
      {
        "entity_id": "char_abc123",
        "entity_type": "player",
        "position": { "x": 106.0, "y": 53.0, "z": 0 },
        "rotation": { "yaw": 185.0 },
        "velocity": { "x": 1.0, "y": 0.5, "z": 0 },
        "state": "walking"
      },
      {
        "entity_id": "npc_hagrid_01",
        "entity_type": "npc",
        "position": { "x": 120.0, "y": 80.0, "z": 0 },
        "rotation": { "yaw": 90.0 },
        "state": "idle"
      }
    ]
  }
}
```

---

### 5. Interactions

#### INTERACT_INTENT (Client → Server)

Interagir avec un objet/NPC/porte.

```json
{
  "msg_type": "INTERACT_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440080",
  "timestamp": 1704067400000,
  "payload": {
    "target_entity_id": "npc_dumbledore_01",
    "interaction_type": "talk",
    "client_position": { "x": 108.0, "y": 58.0, "z": 0 }
  }
}
```

Interaction types: `talk`, `open`, `use`, `pickup`, `activate`, `examine`

#### INTERACT_RESULT (Server → Client)

```json
{
  "msg_type": "INTERACT_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440081",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440080",
  "timestamp": 1704067400050,
  "success": true,
  "payload": {
    "interaction_type": "talk",
    "target_entity_id": "npc_dumbledore_01",
    "result": {
      "dialogue_started": true,
      "dialogue_id": "dialogue_dumbledore_greeting",
      "dialogue_node": "node_1"
    },
    "state_diff": null
  }
}
```

Échec (trop loin):
```json
{
  "msg_type": "INTERACT_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440081",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440080",
  "timestamp": 1704067400050,
  "success": false,
  "payload": null,
  "error": {
    "code": "OUT_OF_RANGE",
    "message": "Trop loin pour interagir",
    "details": {
      "distance": 15.5,
      "max_range": 5.0
    }
  }
}
```

---

### 6. Combat & Spells

#### COMBAT_ACTION_INTENT (Client → Server)

Intention d'action de combat (attaque, sort, etc.).

```json
{
  "msg_type": "COMBAT_ACTION_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440090",
  "timestamp": 1704067450000,
  "payload": {
    "action_type": "cast_spell",
    "action_id": "spell_expelliarmus",
    "target_entity_id": "char_enemy123",
    "aim_direction": { "x": 0.8, "y": 0.6, "z": 0 },
    "client_position": { "x": 100.0, "y": 50.0, "z": 0 },
    "modifier_keys": []
  }
}
```

Action types: `cast_spell`, `melee_attack`, `ranged_attack`, `use_ability`, `dodge`, `block`

#### COMBAT_ACTION_RESULT (Server → Client)

Résultat validé par le serveur.

```json
{
  "msg_type": "COMBAT_ACTION_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440091",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440090",
  "timestamp": 1704067450030,
  "success": true,
  "payload": {
    "action_type": "cast_spell",
    "action_id": "spell_expelliarmus",
    "caster_id": "char_x1y2z3",
    "target_id": "char_enemy123",
    "hit": true,
    "damage_dealt": 45,
    "damage_type": "magical",
    "effects_applied": [
      {
        "effect_id": "effect_disarmed",
        "target_id": "char_enemy123",
        "duration_ms": 3000
      }
    ],
    "state_diff": {
      "self": {
        "mana_current": { "old": 80, "new": 50 },
        "cooldowns": {
          "spell.expelliarmus": 1704067455000
        }
      },
      "target": {
        "health_current": { "old": 100, "new": 55 },
        "effects_added": ["effect_disarmed"]
      }
    },
    "visual_data": {
      "projectile_id": "proj_001",
      "impact_position": { "x": 115.0, "y": 60.0, "z": 1.5 },
      "vfx": "vfx_expelliarmus_hit"
    }
  }
}
```

Échec (cooldown):
```json
{
  "msg_type": "COMBAT_ACTION_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440091",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440090",
  "timestamp": 1704067450030,
  "success": false,
  "payload": null,
  "error": {
    "code": "ON_COOLDOWN",
    "message": "Sort en rechargement",
    "details": {
      "spell_id": "spell_expelliarmus",
      "cooldown_remaining_ms": 2500,
      "ready_at": 1704067452500
    }
  }
}
```

Échec (grade insuffisant pour sort):
```json
{
  "msg_type": "COMBAT_ACTION_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440091",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440090",
  "timestamp": 1704067450030,
  "success": false,
  "payload": null,
  "error": {
    "code": "INSUFFICIENT_GRADE",
    "message": "Grade insuffisant pour ce sort",
    "details": {
      "spell_id": "spell_avada_kedavra",
      "required_grade": 10,
      "current_grade": 5,
      "spell_tier": 5
    }
  }
}
```

---

### 7. Inventory

#### INVENTORY_ACTION_INTENT (Client → Server)

Actions sur l'inventaire.

**Utiliser un item:**
```json
{
  "msg_type": "INVENTORY_ACTION_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440100",
  "timestamp": 1704067500000,
  "payload": {
    "action": "use",
    "item_instance_id": "inv_item_002",
    "quantity": 1,
    "target_entity_id": null
  }
}
```

**Équiper:**
```json
{
  "msg_type": "INVENTORY_ACTION_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440101",
  "timestamp": 1704067500100,
  "payload": {
    "action": "equip",
    "item_instance_id": "inv_item_001",
    "target_slot": "wand"
  }
}
```

**Split stack:**
```json
{
  "msg_type": "INVENTORY_ACTION_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440102",
  "timestamp": 1704067500200,
  "payload": {
    "action": "split",
    "item_instance_id": "inv_item_002",
    "split_quantity": 3,
    "target_slot": 15
  }
}
```

Actions: `use`, `equip`, `unequip`, `move`, `split`, `stack`, `drop`, `destroy`

#### INVENTORY_ACTION_RESULT (Server → Client)

```json
{
  "msg_type": "INVENTORY_ACTION_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440101",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440100",
  "timestamp": 1704067500050,
  "success": true,
  "payload": {
    "action": "use",
    "item_used": {
      "item_def_id": "potion_health_small",
      "quantity_consumed": 1,
      "quantity_remaining": 4
    },
    "effects_applied": [
      {
        "type": "heal",
        "value": 50
      }
    ],
    "state_diff": {
      "inventory": {
        "inv_item_002": {
          "quantity": { "old": 5, "new": 4 }
        }
      },
      "stats": {
        "health_current": { "old": 70, "new": 120 }
      },
      "cooldowns": {
        "item.potion_health_small": 1704067530000
      }
    }
  }
}
```

---

### 8. Quests & Progression

#### QUEST_INTENT (Client → Server)

Actions liées aux quêtes.

**Démarrer une quête:**
```json
{
  "msg_type": "QUEST_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440110",
  "timestamp": 1704067600000,
  "payload": {
    "action": "start",
    "quest_id": "quest_main_chapter2",
    "npc_id": "npc_mcgonagall_01"
  }
}
```

**Avancer un objectif:**
```json
{
  "msg_type": "QUEST_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440111",
  "timestamp": 1704067600100,
  "payload": {
    "action": "advance_objective",
    "quest_id": "quest_main_chapter2",
    "objective_id": "obj_collect_ingredients",
    "data": {
      "item_id": "mandrake_root",
      "quantity": 1
    }
  }
}
```

**Réclamer récompense:**
```json
{
  "msg_type": "QUEST_INTENT",
  "msg_id": "550e8400-e29b-41d4-a716-446655440112",
  "timestamp": 1704067600200,
  "payload": {
    "action": "claim_reward",
    "quest_id": "quest_main_chapter2"
  }
}
```

Actions: `start`, `advance_objective`, `claim_reward`, `abandon`

#### QUEST_RESULT (Server → Client)

```json
{
  "msg_type": "QUEST_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440113",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440112",
  "timestamp": 1704067600250,
  "success": true,
  "payload": {
    "action": "claim_reward",
    "quest_id": "quest_main_chapter2",
    "rewards_claimed": {
      "experience": 5000,
      "currencies": {
        "galleons": 50,
        "house_points": 25
      },
      "items": [
        {
          "item_def_id": "robe_formal",
          "quantity": 1,
          "instance_id": "inv_item_300"
        }
      ]
    },
    "state_diff": {
      "flags": {
        "added": ["quest.main_chapter2.completed"],
        "removed": ["quest.main_chapter2.active"]
      },
      "currencies": {
        "galleons": { "old": 100, "new": 150 },
        "house_points": { "old": 50, "new": 75 }
      },
      "experience": { "old": 120000, "new": 125000 },
      "level": { "old": 24, "new": 25 }
    },
    "unlocked": {
      "quests": ["quest_main_chapter3"],
      "zones": [],
      "spells": ["spell_patronus"]
    }
  }
}
```

---

### 9. State Diff (Serveur → Client)

#### STATE_DIFF (Server → Client)

Envoyé périodiquement ou après chaque action pour synchroniser l'état.

```json
{
  "msg_type": "STATE_DIFF",
  "msg_id": "660e8400-e29b-41d4-a716-446655440120",
  "timestamp": 1704067700000,
  "payload": {
    "diff_type": "periodic",
    "sequence": 12346,
    "stats": {
      "mana_current": { "op": "set", "value": 85 },
      "health_current": { "op": "add", "value": 5 }
    },
    "cooldowns": {
      "spell.expelliarmus": { "op": "set", "value": 0 },
      "spell.stupefy": { "op": "set", "value": 1704067750000 }
    },
    "effects": {
      "added": [
        {
          "effect_id": "effect_mana_regen",
          "duration_ms": 60000,
          "stacks": 1
        }
      ],
      "removed": ["effect_slow"],
      "updated": []
    },
    "flags": {
      "added": [],
      "removed": [],
      "updated": [
        {
          "key": "daily.login_streak",
          "value": { "count": 5 }
        }
      ]
    },
    "inventory": {
      "added": [],
      "removed": [],
      "updated": [
        {
          "instance_id": "inv_item_002",
          "changes": {
            "quantity": { "op": "set", "value": 3 }
          }
        }
      ]
    }
  }
}
```

Diff types: `periodic`, `action_result`, `external_event`, `admin_action`

Operations (`op`):
- `set`: Remplacer la valeur
- `add`: Ajouter à la valeur existante
- `remove`: Supprimer (pour collections)

---

### 10. GM Commands

#### GM_COMMAND_EXEC (Client → Server)

Exécution de commande GM (réservé aux rôles GM/Admin).

```json
{
  "msg_type": "GM_COMMAND_EXEC",
  "msg_id": "550e8400-e29b-41d4-a716-446655440200",
  "timestamp": 1704068000000,
  "payload": {
    "command": "teleport",
    "target_character_id": "char_stuck123",
    "parameters": {
      "zone_id": "hogwarts_main",
      "position": { "x": 0, "y": 0, "z": 0 }
    },
    "reason": "Joueur coincé dans la géométrie"
  }
}
```

Commands disponibles: `teleport`, `kick`, `setgrade`, `grantitem`, `granucurrency`, `announce`, `summon`, `freeze`, `unfreeze`

#### GM_COMMAND_RESULT (Server → Client)

```json
{
  "msg_type": "GM_COMMAND_RESULT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440201",
  "ref_msg_id": "550e8400-e29b-41d4-a716-446655440200",
  "timestamp": 1704068000100,
  "success": true,
  "payload": {
    "command": "teleport",
    "target": {
      "character_id": "char_stuck123",
      "character_name": "Stuck Player"
    },
    "result": {
      "previous_zone": "forbidden_forest",
      "previous_position": { "x": 999.9, "y": 888.8, "z": -50.0 },
      "new_zone": "hogwarts_main",
      "new_position": { "x": 0, "y": 0, "z": 0 }
    },
    "audit_id": "audit_gm_001"
  }
}
```

---

#### KICK (Server → Client)

Notification de kick (envoyé au joueur kické).

```json
{
  "msg_type": "KICK",
  "msg_id": "660e8400-e29b-41d4-a716-446655440210",
  "timestamp": 1704068100000,
  "payload": {
    "reason": "Comportement inapproprié",
    "kicked_by": "GameMaster",
    "can_reconnect": true,
    "reconnect_after": 0
  }
}
```

#### TELEPORT (Server → Client)

Notification de téléportation forcée.

```json
{
  "msg_type": "TELEPORT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440220",
  "timestamp": 1704068200000,
  "payload": {
    "reason": "Téléporté par un GM",
    "new_zone": "hogwarts_main",
    "new_position": { "x": 0, "y": 0, "z": 0 },
    "requires_zone_change": true,
    "transfer_token": "transfer_xyz_gm"
  }
}
```

#### ANNOUNCE (Server → Client)

Annonce serveur/GM broadcast.

```json
{
  "msg_type": "ANNOUNCE",
  "msg_id": "660e8400-e29b-41d4-a716-446655440230",
  "timestamp": 1704068300000,
  "payload": {
    "announcement_type": "gm",
    "scope": "zone",
    "title": "Événement spécial",
    "message": "Le Tournoi des Trois Sorciers commence dans 10 minutes !",
    "duration_ms": 10000,
    "priority": "high",
    "sound": "sfx_announcement"
  }
}
```

Announcement types: `system`, `gm`, `event`, `maintenance`
Scopes: `global`, `zone`, `house`, `targeted`

---

### 11. Error & Disconnect

#### ERROR (Server → Client)

Erreur fatale nécessitant une action.

```json
{
  "msg_type": "ERROR",
  "msg_id": "660e8400-e29b-41d4-a716-446655440300",
  "timestamp": 1704068400000,
  "payload": {
    "error_code": "SESSION_EXPIRED",
    "message": "Votre session a expiré",
    "action_required": "reconnect",
    "details": {}
  }
}
```

Actions requises: `reconnect`, `reauth`, `update_client`, `contact_support`

#### DISCONNECT (Server → Client)

Notification de déconnexion imminente.

```json
{
  "msg_type": "DISCONNECT",
  "msg_id": "660e8400-e29b-41d4-a716-446655440310",
  "timestamp": 1704068500000,
  "payload": {
    "reason": "server_shutdown",
    "message": "Le serveur va redémarrer pour maintenance",
    "disconnect_in_ms": 30000,
    "can_reconnect": true,
    "reconnect_url": "wss://zone-hogwarts-1.hp-mmo.game/ws"
  }
}
```

Reasons: `server_shutdown`, `maintenance`, `kicked`, `banned`, `session_expired`, `duplicate_session`, `afk_timeout`

---

## Modèle de Patch/Diff (JSON Patch alternatif)

### Format "Delta Objects"

Plutôt que JSON Patch standard (RFC 6902), utilisons un format plus lisible et optimisé pour le gaming:

```json
{
  "stats": {
    "health_current": { "op": "set", "v": 85 },
    "mana_current": { "op": "add", "v": -30 }
  },
  "inventory": {
    "inv_001": { "op": "update", "v": { "quantity": 4 } },
    "inv_002": { "op": "remove" },
    "inv_003": { "op": "add", "v": { "item_def_id": "potion_mana", "quantity": 1 } }
  },
  "flags": {
    "quest.ch1.complete": { "op": "add", "v": true },
    "buff.speed": { "op": "remove" }
  },
  "cooldowns": {
    "spell.stupefy": { "op": "set", "v": 1704067800000 }
  }
}
```

### Opérations supportées

| Op | Description | Exemple |
|----|-------------|---------|
| `set` | Remplacer valeur | `{"op": "set", "v": 100}` |
| `add` | Ajouter (numérique) | `{"op": "add", "v": -30}` |
| `remove` | Supprimer l'entrée | `{"op": "remove"}` |
| `update` | Merge partiel (objets) | `{"op": "update", "v": {"qty": 5}}` |
| `append` | Ajouter à array | `{"op": "append", "v": "effect_id"}` |
| `filter` | Retirer d'array | `{"op": "filter", "v": "effect_id"}` |

### Avantages vs JSON Patch

1. **Lisibilité**: Plus facile à debugger
2. **Type-safety**: Structure prévisible
3. **Compression**: Moins verbeux que paths JSON Patch
4. **Gaming-friendly**: Opérations numériques natives

---

## Séquence Complète : Login → Spell Cast

```
┌─────────┐                          ┌─────────────┐                    ┌────────────┐
│  Client │                          │  API Server │                    │Zone Server │
└────┬────┘                          └──────┬──────┘                    └─────┬──────┘
     │                                      │                                 │
     │  POST /auth/login                    │                                 │
     │─────────────────────────────────────>│                                 │
     │                                      │                                 │
     │  200 OK + JWT + characters[]         │                                 │
     │<─────────────────────────────────────│                                 │
     │                                      │                                 │
     │  POST /characters/{id}/select        │                                 │
     │─────────────────────────────────────>│                                 │
     │                                      │                                 │
     │  200 OK + character JWT              │                                 │
     │  + zone_connection info              │                                 │
     │<─────────────────────────────────────│                                 │
     │                                      │                                 │
     │  WebSocket Connect                   │                                 │
     │────────────────────────────────────────────────────────────────────────>
     │                                      │                                 │
     │  CONNECT {token, character_id}       │                                 │
     │────────────────────────────────────────────────────────────────────────>
     │                                      │                                 │
     │  CONNECT_ACK {session_id, zone}      │                                 │
     │<────────────────────────────────────────────────────────────────────────
     │                                      │                                 │
     │  ENTER_WORLD {zone_id}               │                                 │
     │────────────────────────────────────────────────────────────────────────>
     │                                      │                                 │
     │                                      │   Load character from DB        │
     │                                      │   Validate zone access           │
     │                                      │   Subscribe to zone events       │
     │                                      │                                 │
     │  WORLD_SNAPSHOT {self, players, npcs}│                                 │
     │<────────────────────────────────────────────────────────────────────────
     │                                      │                                 │
     │  PRESENCE_UPDATE (to other players)  │                                 │
     │                                      │   ──────────────────────────────>
     │                                      │                                 │
     │  ─── Player is now in world ───      │                                 │
     │                                      │                                 │
     │  COMBAT_ACTION_INTENT                │                                 │
     │  {cast_spell: expelliarmus,          │                                 │
     │   target: char_enemy}                │                                 │
     │────────────────────────────────────────────────────────────────────────>
     │                                      │                                 │
     │                                      │   1. Validate JWT                │
     │                                      │   2. Check cooldown (Redis)      │
     │                                      │   3. Check mana >= 30            │
     │                                      │   4. Check grade >= spell.tier   │
     │                                      │   5. Check range <= 20m          │
     │                                      │   6. Check LOS                   │
     │                                      │   7. Calculate damage            │
     │                                      │   8. Apply effects               │
     │                                      │   9. Set cooldown                │
     │                                      │   10. Persist state              │
     │                                      │                                 │
     │  COMBAT_ACTION_RESULT                │                                 │
     │  {success: true,                     │                                 │
     │   damage: 45,                        │                                 │
     │   state_diff: {                      │                                 │
     │     mana: 80→50,                     │                                 │
     │     cooldowns: {...}                 │                                 │
     │   }}                                 │                                 │
     │<────────────────────────────────────────────────────────────────────────
     │                                      │                                 │
     │  STATE_DIFF (to target player)       │                                 │
     │  {health: 100→55,                    │                                 │
     │   effects: [disarmed]}               │                                 │
     │                                      │   ──────────────────────────────>
     │                                      │                                 │
     │  Entity VFX broadcast (to zone)      │                                 │
     │                                      │   ──────────────────────────────>
     │                                      │                                 │
```

---

## Rate Limits WebSocket

| Message Type | Limit |
|--------------|-------|
| MOVE_INTENT | 60/second |
| COMBAT_ACTION_INTENT | 10/second |
| INVENTORY_ACTION_INTENT | 5/second |
| INTERACT_INTENT | 5/second |
| HEARTBEAT | 0.5/second |
| GM_COMMAND_EXEC | 2/second |

Dépassement = message `RATE_LIMITED` + ignore pendant 1 seconde.
