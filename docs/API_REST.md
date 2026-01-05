# API REST - Spécification Complète

## Base URL

```
Production: https://api.hp-mmo.game/v1
Staging:    https://api-staging.hp-mmo.game/v1
```

## Headers Communs

### Requêtes

```http
Content-Type: application/json
Accept: application/json
Authorization: Bearer <jwt_token>
X-Request-ID: <uuid>  # Pour traçabilité
X-Idempotency-Key: <unique_key>  # Pour mutations
X-Client-Version: 1.2.3
```

### Réponses

```http
Content-Type: application/json
X-Request-ID: <uuid>
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704067200
```

## Format de Réponse Standard

### Succès

```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-01-01T12:00:00.000Z",
    "version": "v1"
  }
}
```

### Erreur

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_GRADE",
    "message": "Grade 5 requis pour cette action",
    "details": {
      "required_grade": 5,
      "current_grade": 3
    }
  },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2024-01-01T12:00:00.000Z"
  }
}
```

## Codes d'Erreur

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_REQUEST` | 400 | Requête malformée |
| `VALIDATION_ERROR` | 400 | Validation des champs échouée |
| `UNAUTHORIZED` | 401 | Token manquant ou invalide |
| `TOKEN_EXPIRED` | 401 | Token expiré |
| `FORBIDDEN` | 403 | Permission refusée |
| `INSUFFICIENT_GRADE` | 403 | Grade insuffisant |
| `INSUFFICIENT_ROLE` | 403 | Rôle insuffisant |
| `HOUSE_RESTRICTED` | 403 | Maison non autorisée |
| `NOT_FOUND` | 404 | Ressource non trouvée |
| `CONFLICT` | 409 | Conflit (ex: nom déjà pris) |
| `IDEMPOTENCY_CONFLICT` | 409 | Clé d'idempotence déjà utilisée |
| `RATE_LIMITED` | 429 | Trop de requêtes |
| `INTERNAL_ERROR` | 500 | Erreur serveur |
| `SERVICE_UNAVAILABLE` | 503 | Service temporairement indisponible |

---

## 1. Authentication

### POST /auth/register

Créer un nouveau compte.

**Request:**

```json
{
  "email": "wizard@hogwarts.edu",
  "username": "young_wizard",
  "password": "SecurePass123!",
  "display_name": "Young Wizard"
}
```

**Response 201:**

```json
{
  "success": true,
  "data": {
    "account_id": "acc_a1b2c3d4",
    "email": "wizard@hogwarts.edu",
    "username": "young_wizard",
    "display_name": "Young Wizard",
    "role": "player",
    "created_at": "2024-01-01T12:00:00.000Z",
    "tokens": {
      "access_token": "eyJhbGciOiJSUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
      "expires_in": 86400,
      "token_type": "Bearer"
    }
  }
}
```

**Errors:**
- `VALIDATION_ERROR` - Champs invalides
- `CONFLICT` - Email ou username déjà utilisé

---

### POST /auth/login

Authentification.

**Request:**

```json
{
  "email": "wizard@hogwarts.edu",
  "password": "SecurePass123!"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "account_id": "acc_a1b2c3d4",
    "username": "young_wizard",
    "display_name": "Young Wizard",
    "role": "player",
    "last_login_at": "2024-01-01T10:00:00.000Z",
    "tokens": {
      "access_token": "eyJhbGciOiJSUzI1NiIs...",
      "refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
      "expires_in": 86400,
      "token_type": "Bearer"
    },
    "characters": [
      {
        "id": "char_x1y2z3",
        "name": "Draco Testus",
        "house": "venatrix",
        "grade": 5,
        "level": 25,
        "last_played_at": "2024-01-01T09:00:00.000Z"
      }
    ]
  }
}
```

**Errors:**
- `UNAUTHORIZED` - Identifiants incorrects
- `FORBIDDEN` - Compte banni

---

### POST /auth/refresh

Rafraîchir le token d'accès.

**Request:**

```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJl..."
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "refresh_token": "bmV3IHJlZnJlc2ggdG9r...",
    "expires_in": 86400,
    "token_type": "Bearer"
  }
}
```

---

### POST /auth/logout

Déconnexion (invalide la session).

**Headers:**
```http
Authorization: Bearer <access_token>
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "message": "Session terminée"
  }
}
```

---

## 2. Characters

### GET /characters

Liste des personnages du compte connecté.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "characters": [
      {
        "id": "char_x1y2z3",
        "name": "Draco Testus",
        "house": "venatrix",
        "grade": 5,
        "level": 25,
        "experience": 125000,
        "zone_id": "hogwarts_main",
        "status": "active",
        "created_at": "2023-06-15T14:30:00.000Z",
        "last_played_at": "2024-01-01T09:00:00.000Z",
        "total_playtime_seconds": 360000,
        "appearance_preview_url": "https://cdn.hp-mmo.game/avatars/char_x1y2z3.png"
      }
    ],
    "max_characters": 5,
    "slots_used": 1
  }
}
```

---

### POST /characters

Créer un nouveau personnage.

**Request:**

```json
{
  "name": "Hermione Testus",
  "house": "aerwyn",
  "appearance_data": {
    "body_type": 1,
    "face": 3,
    "hair": 8,
    "hair_color": "#8B4513",
    "skin_tone": 2,
    "eye_color": "#654321"
  }
}
```

**Response 201:**

```json
{
  "success": true,
  "data": {
    "character": {
      "id": "char_a1b2c3",
      "name": "Hermione Testus",
      "house": "aerwyn",
      "grade": 1,
      "level": 1,
      "experience": 0,
      "zone_id": "hogwarts_main",
      "position": { "x": 0, "y": 0, "z": 0 },
      "status": "active",
      "created_at": "2024-01-01T12:00:00.000Z",
      "appearance_data": { ... }
    }
  }
}
```

**Errors:**
- `CONFLICT` - Nom déjà pris
- `VALIDATION_ERROR` - Maison invalide ou données apparence incorrectes
- `FORBIDDEN` - Maximum de personnages atteint

---

### GET /characters/{character_id}

Détails d'un personnage.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "character": {
      "id": "char_x1y2z3",
      "name": "Draco Testus",
      "house": "venatrix",
      "grade": 5,
      "level": 25,
      "experience": 125000,
      "experience_to_next_level": 150000,
      "zone_id": "hogwarts_main",
      "position": { "x": 100.5, "y": 50.2, "z": 0 },
      "status": "active",
      "created_at": "2023-06-15T14:30:00.000Z",
      "last_played_at": "2024-01-01T09:00:00.000Z",
      "total_playtime_seconds": 360000,
      "appearance_data": { ... }
    },
    "stats": {
      "health_current": 100,
      "health_max": 120,
      "mana_current": 80,
      "mana_max": 150,
      "stamina_current": 100,
      "stamina_max": 100,
      "strength": 12,
      "dexterity": 14,
      "intelligence": 18,
      "wisdom": 15,
      "charisma": 10,
      "luck": 8,
      "spell_power": 45,
      "defense": 20,
      "speed": 1.1
    },
    "currencies": {
      "galleons": 150,
      "sickles": 340,
      "knuts": 1200,
      "house_points": 75
    },
    "permissions": [
      "spell.tier1.cast",
      "spell.tier2.cast",
      "spell.tier3.cast",
      "zone.forest.enter"
    ]
  }
}
```

---

### POST /characters/{character_id}/select

Sélectionner un personnage pour jouer (génère un JWT avec character_id).

**Response 200:**

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIs...",
    "character": {
      "id": "char_x1y2z3",
      "name": "Draco Testus",
      "house": "venatrix",
      "grade": 5
    },
    "zone_connection": {
      "zone_id": "hogwarts_main",
      "shard": "hogwarts-main-shard-0",
      "websocket_url": "wss://zone-hogwarts-0.hp-mmo.game/ws",
      "connection_token": "temp_token_for_ws_auth"
    }
  }
}
```

---

### DELETE /characters/{character_id}

Supprimer un personnage (soft delete avec délai de grâce).

**Request:**

```json
{
  "confirmation": "DELETE_Draco Testus"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "message": "Personnage marqué pour suppression",
    "deleted_at": "2024-01-01T12:00:00.000Z",
    "permanent_deletion_at": "2024-01-08T12:00:00.000Z",
    "can_restore_until": "2024-01-08T12:00:00.000Z"
  }
}
```

---

## 3. House & Grade

### GET /characters/{character_id}/house

Information sur la maison du personnage.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "house": "venatrix",
    "house_display_name": "Venatrix",
    "house_description": "La maison des chasseurs audacieux...",
    "house_colors": ["#1a472a", "#silver"],
    "house_mascot": "serpent",
    "house_points_contributed": 75,
    "house_total_points": 12500,
    "house_rank": 2
  }
}
```

---

### GET /houses/leaderboard

Classement des maisons.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "leaderboard": [
      { "house": "aerwyn", "points": 15200, "rank": 1 },
      { "house": "venatrix", "points": 12500, "rank": 2 },
      { "house": "falcon", "points": 11800, "rank": 3 },
      { "house": "brumval", "points": 10500, "rank": 4 }
    ],
    "season": "2024-Q1",
    "ends_at": "2024-03-31T23:59:59.000Z"
  }
}
```

---

### PUT /admin/characters/{character_id}/grade

**[ADMIN/GM]** Modifier le grade d'un personnage.

**Headers:**
```http
X-Idempotency-Key: grade_change_char_x1y2z3_1704067200
```

**Request:**

```json
{
  "new_grade": 7,
  "reason": "Promotion suite à réussite des examens RP"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "character_id": "char_x1y2z3",
    "previous_grade": 5,
    "new_grade": 7,
    "new_permissions": [
      "spell.tier4.cast"
    ],
    "audit_id": "audit_123456"
  }
}
```

**Errors:**
- `INSUFFICIENT_ROLE` - Rôle GM/Admin requis
- `VALIDATION_ERROR` - Grade invalide (1-100)

---

## 4. Inventory

### GET /characters/{character_id}/inventory

Liste l'inventaire complet.

**Query Parameters:**
- `page` (default: 1)
- `per_page` (default: 50, max: 100)
- `type` (filter: consumable, equipment, quest, etc.)

**Response 200:**

```json
{
  "success": true,
  "data": {
    "items": [
      {
        "instance_id": "inv_item_001",
        "item_def_id": "wand_phoenix_feather",
        "display_name": "Baguette Plume de Phénix",
        "item_type": "equipment",
        "equipment_slot": "wand",
        "quantity": 1,
        "inventory_slot": 0,
        "is_equipped": true,
        "instance_data": {
          "durability": 85,
          "enchantment": "lumos_enhanced"
        },
        "properties": {
          "spell_power": 25,
          "affinity": "fire"
        },
        "acquired_at": "2023-09-01T10:00:00.000Z"
      },
      {
        "instance_id": "inv_item_002",
        "item_def_id": "potion_health_small",
        "display_name": "Petite Potion de Soin",
        "item_type": "consumable",
        "quantity": 5,
        "inventory_slot": 1,
        "is_equipped": false,
        "properties": {
          "heal_amount": 50,
          "cooldown_ms": 30000
        },
        "acquired_at": "2024-01-01T08:00:00.000Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 50,
      "total_items": 2,
      "total_pages": 1
    },
    "inventory_capacity": 100,
    "slots_used": 2
  }
}
```

---

### POST /characters/{character_id}/inventory/use

Utiliser un item consommable.

**Headers:**
```http
X-Idempotency-Key: use_item_inv_item_002_1704067200
```

**Request:**

```json
{
  "instance_id": "inv_item_002",
  "quantity": 1
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "action": "use_item",
    "item_used": {
      "item_def_id": "potion_health_small",
      "quantity_used": 1,
      "quantity_remaining": 4
    },
    "effects_applied": [
      {
        "type": "heal",
        "stat": "health_current",
        "delta": 50,
        "new_value": 100
      }
    ],
    "cooldown_until": "2024-01-01T12:00:30.000Z"
  }
}
```

**Errors:**
- `NOT_FOUND` - Item non trouvé
- `FORBIDDEN` - Item non utilisable ou grade insuffisant
- `CONFLICT` - Cooldown actif

---

### POST /characters/{character_id}/inventory/equip

Équiper un item.

**Request:**

```json
{
  "instance_id": "inv_item_001",
  "slot": "wand"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "action": "equip",
    "equipped_item": {
      "instance_id": "inv_item_001",
      "slot": "wand"
    },
    "unequipped_item": null,
    "stats_delta": {
      "spell_power": { "old": 20, "new": 45 }
    }
  }
}
```

---

### POST /characters/{character_id}/inventory/unequip

Déséquiper un item.

**Request:**

```json
{
  "slot": "wand"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "action": "unequip",
    "unequipped_item": {
      "instance_id": "inv_item_001",
      "new_inventory_slot": 5
    },
    "stats_delta": {
      "spell_power": { "old": 45, "new": 20 }
    }
  }
}
```

---

### POST /characters/{character_id}/inventory/move

Déplacer un item dans l'inventaire.

**Request:**

```json
{
  "instance_id": "inv_item_002",
  "target_slot": 10
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "action": "move",
    "item_moved": {
      "instance_id": "inv_item_002",
      "from_slot": 1,
      "to_slot": 10
    }
  }
}
```

---

### POST /admin/characters/{character_id}/inventory/grant

**[ADMIN/GM]** Donner des items à un personnage.

**Headers:**
```http
X-Idempotency-Key: grant_item_char_x1y2z3_wand_elder_1704067200
```

**Request:**

```json
{
  "items": [
    {
      "item_def_id": "wand_elder",
      "quantity": 1,
      "instance_data": {
        "enchantment": "power_overwhelming"
      }
    },
    {
      "item_def_id": "potion_felix_felicis",
      "quantity": 3
    }
  ],
  "reason": "Récompense event Halloween 2024"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "transaction_id": "txn_abc123",
    "items_granted": [
      {
        "instance_id": "inv_item_100",
        "item_def_id": "wand_elder",
        "quantity": 1
      },
      {
        "instance_id": "inv_item_101",
        "item_def_id": "potion_felix_felicis",
        "quantity": 3
      }
    ],
    "audit_id": "audit_789012"
  }
}
```

---

## 5. Economy

### GET /characters/{character_id}/currencies

Soldes des monnaies.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "currencies": {
      "galleons": 150,
      "sickles": 340,
      "knuts": 1200,
      "house_points": 75
    }
  }
}
```

---

### POST /characters/{character_id}/economy/purchase

Effectuer un achat (boutique NPC, etc.).

**Headers:**
```http
X-Idempotency-Key: purchase_char_x1y2z3_shop_wands_1704067200
```

**Request:**

```json
{
  "shop_id": "ollivanders",
  "items": [
    {
      "item_def_id": "wand_holly",
      "quantity": 1
    }
  ],
  "currency": "galleons"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "transaction_id": "txn_def456",
    "items_received": [
      {
        "instance_id": "inv_item_200",
        "item_def_id": "wand_holly",
        "quantity": 1
      }
    ],
    "cost": {
      "galleons": 7
    },
    "new_balance": {
      "galleons": 143
    }
  }
}
```

**Errors:**
- `INSUFFICIENT_FUNDS` - Solde insuffisant
- `NOT_FOUND` - Shop ou item non trouvé

---

### POST /admin/characters/{character_id}/economy/grant

**[ADMIN/GM]** Ajouter des monnaies.

**Headers:**
```http
X-Idempotency-Key: grant_currency_char_x1y2z3_1704067200
```

**Request:**

```json
{
  "currencies": {
    "galleons": 100,
    "house_points": 50
  },
  "reason": "Récompense quête principale chapitre 5"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "transaction_id": "txn_ghi789",
    "currencies_granted": {
      "galleons": 100,
      "house_points": 50
    },
    "new_balances": {
      "galleons": 250,
      "house_points": 125
    },
    "audit_id": "audit_345678"
  }
}
```

---

## 6. Admin / GM Commands

### POST /admin/commands/execute

**[ADMIN/GM]** Exécuter une commande GM.

**Request - Set Grade:**

```json
{
  "command": "setgrade",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "grade": 8
  },
  "reason": "Promotion exceptionnelle"
}
```

**Request - Teleport:**

```json
{
  "command": "teleport",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "zone_id": "hogsmeade",
    "position": { "x": 100, "y": 50, "z": 0 }
  },
  "reason": "Déblocage joueur coincé"
}
```

**Request - Grant Item:**

```json
{
  "command": "grantitem",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "item_def_id": "broom_firebolt",
    "quantity": 1,
    "instance_data": {}
  },
  "reason": "Récompense tournoi Quidditch"
}
```

**Request - Kick:**

```json
{
  "command": "kick",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "message": "Comportement inapproprié - avertissement"
  },
  "reason": "Spam chat - 1er avertissement"
}
```

**Request - Ban:**

```json
{
  "command": "ban",
  "target_account_id": "acc_a1b2c3d4",
  "parameters": {
    "duration_hours": 24,
    "type": "temp_ban"
  },
  "reason": "Harcèlement répété - ban 24h"
}
```

**Response 200:**

```json
{
  "success": true,
  "data": {
    "command": "setgrade",
    "executed_at": "2024-01-01T12:00:00.000Z",
    "result": {
      "previous_grade": 5,
      "new_grade": 8
    },
    "audit_id": "audit_901234",
    "affected_character": {
      "id": "char_x1y2z3",
      "name": "Draco Testus"
    }
  }
}
```

---

### GET /admin/audit

**[ADMIN]** Récupérer les logs d'audit.

**Query Parameters:**
- `actor_id` - Filtrer par acteur
- `target_id` - Filtrer par cible
- `action` - Type d'action
- `from` - Date début (ISO8601)
- `to` - Date fin (ISO8601)
- `page`, `per_page`

**Response 200:**

```json
{
  "success": true,
  "data": {
    "audit_logs": [
      {
        "id": "audit_901234",
        "actor": {
          "account_id": "acc_gm001",
          "username": "GameMaster_Alice",
          "role": "gm"
        },
        "target": {
          "account_id": "acc_a1b2c3d4",
          "character_id": "char_x1y2z3",
          "character_name": "Draco Testus"
        },
        "action": "grade_change",
        "old_value": { "grade": 5 },
        "new_value": { "grade": 8 },
        "reason": "Promotion exceptionnelle",
        "ip_address": "192.168.1.100",
        "created_at": "2024-01-01T12:00:00.000Z"
      }
    ],
    "pagination": {
      "page": 1,
      "per_page": 50,
      "total_items": 1,
      "total_pages": 1
    }
  }
}
```

---

### GET /admin/sanctions

**[ADMIN/GM]** Liste des sanctions actives.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "sanctions": [
      {
        "id": "sanc_001",
        "account_id": "acc_banned01",
        "username": "bad_player",
        "sanction_type": "temp_ban",
        "reason": "Harcèlement",
        "issued_by": {
          "account_id": "acc_gm001",
          "username": "GameMaster_Alice"
        },
        "started_at": "2024-01-01T10:00:00.000Z",
        "expires_at": "2024-01-02T10:00:00.000Z",
        "is_active": true
      }
    ]
  }
}
```

---

## 7. Presence

### GET /presence/online

Liste des joueurs en ligne (pour systèmes externes voix/chat).

**Query Parameters:**
- `zone_id` - Filtrer par zone
- `house` - Filtrer par maison
- `friends_only` - Seulement les amis

**Response 200:**

```json
{
  "success": true,
  "data": {
    "online_count": 1247,
    "players": [
      {
        "character_id": "char_x1y2z3",
        "name": "Draco Testus",
        "house": "venatrix",
        "grade": 5,
        "zone_id": "hogwarts_main",
        "status": "online",
        "last_activity": "2024-01-01T12:00:00.000Z"
      }
    ],
    "by_zone": {
      "hogwarts_main": 523,
      "hogsmeade": 312,
      "forbidden_forest": 89,
      "quidditch_pitch": 145,
      "other": 178
    },
    "by_house": {
      "venatrix": 312,
      "aerwyn": 298,
      "falcon": 345,
      "brumval": 292
    }
  }
}
```

---

## 8. Zones

### GET /zones

Liste des zones accessibles.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "zones": [
      {
        "id": "hogwarts_main",
        "display_name": "Poudlard - Château Principal",
        "is_safe_zone": true,
        "required_grade": 1,
        "current_players": 523,
        "max_players": 1500,
        "connected_zones": ["hogsmeade", "forbidden_forest", "quidditch_pitch"],
        "is_accessible": true
      },
      {
        "id": "forbidden_forest",
        "display_name": "Forêt Interdite",
        "is_safe_zone": false,
        "required_grade": 3,
        "current_players": 89,
        "max_players": 300,
        "connected_zones": ["hogwarts_main"],
        "is_accessible": true,
        "access_reason": "Grade 3+ requis"
      },
      {
        "id": "azkaban",
        "display_name": "Azkaban",
        "is_safe_zone": false,
        "required_grade": 8,
        "current_players": 12,
        "max_players": 100,
        "is_accessible": false,
        "access_reason": "Grade 8 requis (vous avez grade 5)"
      }
    ]
  }
}
```

---

### GET /zones/{zone_id}/connection

Obtenir les informations de connexion pour une zone.

**Response 200:**

```json
{
  "success": true,
  "data": {
    "zone_id": "hogwarts_main",
    "assigned_shard": "hogwarts-main-shard-1",
    "websocket_url": "wss://zone-hogwarts-1.hp-mmo.game/ws",
    "connection_token": "temp_conn_token_xyz789",
    "token_expires_at": "2024-01-01T12:01:00.000Z"
  }
}
```

**Errors:**
- `FORBIDDEN` - Zone non accessible (grade, maison, etc.)

---

## Rate Limits

| Endpoint Category | Rate Limit |
|-------------------|------------|
| Auth (login/register) | 5/minute |
| Auth (refresh) | 30/minute |
| Read endpoints | 100/minute |
| Write endpoints | 30/minute |
| Admin commands | 60/minute |
| Inventory actions | 60/minute |

---

## Idempotency

Toutes les mutations (POST/PUT/DELETE avec effets) **doivent** inclure `X-Idempotency-Key`.

Format recommandé:
```
{action}_{resource_id}_{timestamp_epoch_seconds}_{optional_nonce}
```

Exemples:
- `use_item_inv_001_1704067200`
- `grant_currency_char_x1y2z3_1704067200_abc`
- `purchase_shop_ollivanders_1704067200`

Le serveur conserve les clés pendant 24h. Une requête avec une clé déjà utilisée retourne le résultat original.
