# Règles Métiers - Spécification Complète

## 1. Système de Maisons

### 1.1 Les Quatre Maisons

| Maison | ID | Couleurs | Mascotte | Description |
|--------|-----|----------|----------|-------------|
| **Venatrix** | `venatrix` | Vert émeraude, Argent | Serpent | Maison des chasseurs rusés et ambitieux |
| **Aerwyn** | `aerwyn` | Bleu saphir, Bronze | Aigle | Maison des érudits et des sages |
| **Falcon** | `falcon` | Rouge rubis, Or | Faucon | Maison des courageux et téméraires |
| **Brumval** | `brumval` | Jaune topaze, Noir | Blaireau | Maison des loyaux et persévérants |

### 1.2 Attribution de la Maison

```
RÈGLE: La maison est définie à la CRÉATION du personnage et ne peut être modifiée.

Exception:
- Admin/SuperAdmin peut forcer un changement via commande GM (cas exceptionnel)
- Nécessite audit_log obligatoire avec justification
```

**Flow de création:**
1. Joueur crée personnage → choisit apparence
2. **Option A**: Joueur choisit directement sa maison
3. **Option B**: Cérémonie de répartition RP (si implémentée)
   - NPC "Chapeau Magique" pose des questions
   - Réponses influencent la recommandation
   - Joueur peut accepter ou choisir différemment

### 1.3 Impact de la Maison

#### Zones Exclusives
```yaml
venatrix:
  exclusive_zones:
    - "venatrix_common_room"
    - "venatrix_dormitory"
  blocked_zones: []

aerwyn:
  exclusive_zones:
    - "aerwyn_common_room"
    - "aerwyn_dormitory"
  blocked_zones: []

falcon:
  exclusive_zones:
    - "falcon_common_room"
    - "falcon_dormitory"
  blocked_zones: []

brumval:
  exclusive_zones:
    - "brumval_common_room"
    - "brumval_dormitory"
  blocked_zones: []
```

#### Items Exclusifs
```yaml
house_exclusive_items:
  venatrix:
    - "robe_venatrix_*"
    - "scarf_venatrix"
    - "badge_venatrix_prefect"
  aerwyn:
    - "robe_aerwyn_*"
    - "scarf_aerwyn"
    - "badge_aerwyn_prefect"
  # ... etc
```

#### Sorts Exclusifs (Optionnel)
```yaml
house_exclusive_spells:
  venatrix:
    - spell_serpent_summon  # Grade 7+
  aerwyn:
    - spell_wisdom_aura     # Grade 6+
  falcon:
    - spell_courage_boost   # Grade 5+
  brumval:
    - spell_earth_shield    # Grade 6+
```

### 1.4 Points de Maison

```sql
-- Stockage dans character_currencies
currency_id = 'house_points'

-- Agrégation pour classement
SELECT
  c.house,
  SUM(cc.amount) as total_points
FROM characters c
JOIN character_currencies cc ON c.id = cc.character_id
WHERE cc.currency_id = 'house_points'
  AND c.status = 'active'
GROUP BY c.house
ORDER BY total_points DESC;
```

**Règles d'attribution:**
- Quêtes complétées: +5 à +50 points selon difficulté
- Actions RP positives (GM): +1 à +20 points
- Victoire Quidditch: +150 points équipe gagnante
- Cours réussi: +10 points
- Comportement négatif: -5 à -50 points (GM)

---

## 2. Système de Grades

### 2.1 Définition

Le **grade** représente le niveau de progression RP et académique du personnage (équivalent "année scolaire" étendue).

```
GRADE: Entier de 1 à 100
- 1-7: Années standard Poudlard
- 8-10: Formation avancée
- 11+: Maître, Expert, Légende (progression endgame)
```

### 2.2 Stockage

```sql
-- Dans table characters
grade INTEGER NOT NULL DEFAULT 1 CHECK (grade >= 1 AND grade <= 100)
```

### 2.3 Qui Peut Modifier un Grade

| Rôle | Peut modifier | Plage autorisée |
|------|---------------|-----------------|
| `player` | Non | - |
| `vip` | Non | - |
| `moderator` | Non | - |
| `gm` | Oui | 1-20 |
| `admin` | Oui | 1-50 |
| `superadmin` | Oui | 1-100 |

### 2.4 Règles de Modification

```
RÈGLE 1: Toute modification de grade DOIT être auditée
RÈGLE 2: Une raison textuelle est OBLIGATOIRE
RÈGLE 3: Un GM ne peut pas modifier son propre grade
RÈGLE 4: La modification est immédiatement effective
```

**Audit log obligatoire:**
```json
{
  "action": "grade_change",
  "actor_account_id": "acc_gm001",
  "actor_role": "gm",
  "target_character_id": "char_x1y2z3",
  "old_value": { "grade": 5 },
  "new_value": { "grade": 7 },
  "reason": "Promotion suite aux examens RP du 15/01",
  "created_at": "2024-01-15T18:30:00Z"
}
```

### 2.5 Effets du Grade

#### Accès aux Sorts

```yaml
spell_tiers:
  tier_1:
    min_grade: 1
    spells:
      - lumos
      - nox
      - alohomora
      - reparo

  tier_2:
    min_grade: 2
    spells:
      - wingardium_leviosa
      - incendio
      - aguamenti
      - expelliarmus

  tier_3:
    min_grade: 3
    spells:
      - stupefy
      - protego
      - accio
      - riddikulus

  tier_4:
    min_grade: 5
    spells:
      - expecto_patronum
      - petrificus_totalus
      - sectumsempra

  tier_5:
    min_grade: 7
    spells:
      - fiendfyre
      - apparition

  tier_unforgivable:
    min_grade: 10
    requires_permission: "spell.unforgivable.cast"
    spells:
      - imperio
      - crucio
      - avada_kedavra
```

#### Accès aux Zones

```yaml
zones:
  hogwarts_main:
    min_grade: 1

  hogsmeade:
    min_grade: 1
    note: "Accessible uniquement certains week-ends pour grade 1-2"

  forbidden_forest:
    min_grade: 3

  ministry_of_magic:
    min_grade: 5

  knockturn_alley:
    min_grade: 4
    note: "Zone de roleplay sombre"

  azkaban:
    min_grade: 8
    requires_permission: "zone.azkaban.enter"

  room_of_requirement:
    min_grade: 4
    special: true
```

#### Accès aux Items

```yaml
item_restrictions:
  broom_nimbus_2000:
    min_grade: 2

  broom_firebolt:
    min_grade: 5

  time_turner:
    min_grade: 8
    requires_permission: "item.restricted.use"

  elder_wand:
    min_grade: 10
    unique: true
    quest_reward: "quest_deathly_hallows"
```

### 2.6 Progression de Grade

**Méthodes d'augmentation:**
1. **Automatique**: Fin d'année scolaire RP (event global)
2. **Quest**: Complétion de quêtes principales
3. **GM/Admin**: Promotion manuelle avec justification
4. **Event**: Victoire tournois spéciaux

**Règle importante:**
```
Le grade NE DIMINUE JAMAIS automatiquement.
Seule action admin peut réduire un grade (sanction extrême).
```

---

## 3. Système de Permissions

### 3.1 Modèle de Permissions

```
Permission = {role_based} + {grade_based} + {house_based} + {flag_based}
```

### 3.2 Évaluation des Permissions

```python
def has_permission(character, permission_id):
    permission = get_permission(permission_id)
    account = get_account(character.account_id)

    # 1. Vérifier si le rôle a la permission directement
    if permission_id in get_role_permissions(account.role):
        return True

    # 2. Vérifier grade minimum
    if permission.min_grade and character.grade < permission.min_grade:
        return False

    # 3. Vérifier maison autorisée
    if permission.allowed_houses and character.house not in permission.allowed_houses:
        return False

    # 4. Vérifier flags requis
    if permission.required_flags:
        for flag in permission.required_flags:
            if not character_has_flag(character.id, flag):
                return False

    return True
```

### 3.3 Permissions de Base

```yaml
permissions:
  # Sorts
  spell.tier1.cast: { min_grade: 1 }
  spell.tier2.cast: { min_grade: 2 }
  spell.tier3.cast: { min_grade: 3 }
  spell.tier4.cast: { min_grade: 5 }
  spell.tier5.cast: { min_grade: 7 }
  spell.unforgivable.cast: { min_grade: 10, roles: [admin, superadmin] }

  # Zones
  zone.forest.enter: { min_grade: 3 }
  zone.ministry.enter: { min_grade: 5 }
  zone.azkaban.enter: { min_grade: 8 }
  zone.venatrix_common.enter: { houses: [venatrix] }

  # Items
  item.broom.use: { min_grade: 1 }
  item.restricted.use: { min_grade: 6 }

  # Admin
  admin.kick: { roles: [moderator, gm, admin, superadmin] }
  admin.ban: { roles: [gm, admin, superadmin] }
  admin.grant_item: { roles: [gm, admin, superadmin] }
  admin.set_grade: { roles: [gm, admin, superadmin] }
  admin.teleport: { roles: [gm, admin, superadmin] }
  admin.modify_world: { roles: [admin, superadmin] }
```

---

## 4. Anti-Duplication & Atomicité

### 4.1 Principes

```
RÈGLE 1: Toute opération modifiant l'inventaire ou l'économie DOIT être atomique
RÈGLE 2: Toute opération DOIT avoir une clé d'idempotence
RÈGLE 3: Les transactions sont journalisées dans transactions_ledger
```

### 4.2 Clé d'Idempotence

**Format:**
```
{action}_{source}_{target}_{timestamp_epoch_ms}_{nonce}
```

**Exemples:**
```
use_item_char123_null_1704067200123_abc
trade_char123_char456_1704067200123_def
grant_gm001_char789_1704067200123_ghi
```

### 4.3 Transaction Atomique

```sql
-- Voir function execute_inventory_transaction() dans DATABASE_SCHEMA.sql

BEGIN;

-- 1. Vérifier idempotence
SELECT id FROM transactions_ledger WHERE idempotency_key = ?;
IF FOUND THEN RETURN existing_id; END IF;

-- 2. Créer transaction en 'pending'
INSERT INTO transactions_ledger (...) VALUES (...);

-- 3. Exécuter les modifications
-- (débit/crédit items, currencies)

-- 4. Marquer 'completed'
UPDATE transactions_ledger SET status = 'completed';

COMMIT;
```

### 4.4 Prévention des Race Conditions

```sql
-- Lock optimiste sur inventory_items
UPDATE inventory_items
SET quantity = quantity - 1,
    updated_at = NOW()
WHERE id = ?
  AND character_id = ?
  AND quantity >= 1
  AND updated_at = ?;  -- Version check

IF ROW_COUNT = 0 THEN
  ROLLBACK;
  RETURN 'CONFLICT';
END IF;
```

---

## 5. Commandes GM Détaillées

### 5.1 setgrade

**Description:** Modifier le grade d'un personnage.

**Permission requise:** `admin.set_grade`

**Payload:**
```json
{
  "command": "setgrade",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "grade": 8
  },
  "reason": "Promotion suite à réussite des examens de fin d'année RP"
}
```

**Validations:**
1. Target character existe et est actif
2. Acteur a le rôle suffisant pour le grade cible
   - GM: max 20
   - Admin: max 50
   - SuperAdmin: max 100
3. Acteur ≠ Target (ne peut pas se promouvoir soi-même)
4. Reason non vide (min 10 caractères)

**Effets:**
1. `characters.grade` mis à jour
2. Permissions recalculées
3. Nouveau stats_snapshot créé (optionnel si progression auto)
4. Audit log créé
5. Notification au joueur cible

**Audit:**
```json
{
  "id": "audit_001",
  "action": "grade_change",
  "actor_account_id": "acc_gm001",
  "actor_role": "gm",
  "actor_ip": "192.168.1.100",
  "target_character_id": "char_x1y2z3",
  "old_value": { "grade": 5 },
  "new_value": { "grade": 8 },
  "reason": "Promotion suite à réussite des examens de fin d'année RP",
  "created_at": "2024-01-15T18:30:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "command": "setgrade",
    "target": {
      "character_id": "char_x1y2z3",
      "character_name": "Draco Testus"
    },
    "result": {
      "previous_grade": 5,
      "new_grade": 8,
      "new_permissions": [
        "spell.tier4.cast",
        "zone.azkaban.enter"
      ],
      "removed_restrictions": [
        "zone.ministry.enter"
      ]
    },
    "audit_id": "audit_001"
  }
}
```

---

### 5.2 teleport

**Description:** Téléporter un personnage vers une position/zone.

**Permission requise:** `admin.teleport`

**Payload - Même zone:**
```json
{
  "command": "teleport",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "position": { "x": 100.0, "y": 50.0, "z": 0 }
  },
  "reason": "Joueur coincé dans la géométrie"
}
```

**Payload - Autre zone:**
```json
{
  "command": "teleport",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "zone_id": "hogsmeade",
    "position": { "x": 0, "y": 0, "z": 0 },
    "use_spawn_point": true
  },
  "reason": "Debug zone transfer"
}
```

**Validations:**
1. Target character existe, actif, et en ligne
2. Zone cible existe
3. Position dans les bounds de la zone
4. Si target offline → position sauvegardée pour prochain login

**Effets:**
1. Si même zone:
   - Position mise à jour immédiatement
   - Message `TELEPORT` envoyé au client
2. Si autre zone:
   - Transfer token généré
   - Message `TELEPORT` avec `requires_zone_change: true`
   - Client doit reconnecter à la nouvelle zone

**Audit:**
```json
{
  "action": "teleport",
  "actor_account_id": "acc_gm001",
  "target_character_id": "char_x1y2z3",
  "old_value": {
    "zone_id": "forbidden_forest",
    "position": { "x": 500.0, "y": 300.0, "z": -10.0 }
  },
  "new_value": {
    "zone_id": "hogwarts_main",
    "position": { "x": 0, "y": 0, "z": 0 }
  },
  "reason": "Joueur coincé dans la géométrie"
}
```

---

### 5.3 grantitem

**Description:** Donner des items à un personnage.

**Permission requise:** `admin.grant_item`

**Payload:**
```json
{
  "command": "grantitem",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "items": [
      {
        "item_def_id": "broom_firebolt",
        "quantity": 1,
        "instance_data": {
          "enchantment": "speed_boost_2"
        }
      },
      {
        "item_def_id": "potion_felix_felicis",
        "quantity": 5
      }
    ]
  },
  "reason": "Récompense event Halloween 2024"
}
```

**Validations:**
1. Target character existe et actif
2. Tous les item_def_id existent
3. Quantities > 0
4. Inventaire a la capacité
5. Items respectent les restrictions de maison (si applicable)

**Effets:**
1. Items créés dans `inventory_items`
2. Transaction créée dans `transactions_ledger`
3. Audit log créé
4. Notification au joueur
5. STATE_DIFF envoyé si en ligne

**Audit:**
```json
{
  "action": "item_grant",
  "actor_account_id": "acc_gm001",
  "target_character_id": "char_x1y2z3",
  "new_value": {
    "items": [
      { "item_def_id": "broom_firebolt", "quantity": 1, "instance_id": "inv_item_500" },
      { "item_def_id": "potion_felix_felicis", "quantity": 5, "instance_id": "inv_item_501" }
    ]
  },
  "reason": "Récompense event Halloween 2024",
  "metadata": {
    "transaction_id": "txn_abc123",
    "idempotency_key": "grant_gm001_char_x1y2z3_1704067200_event"
  }
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "command": "grantitem",
    "target": {
      "character_id": "char_x1y2z3",
      "character_name": "Draco Testus"
    },
    "result": {
      "items_granted": [
        {
          "instance_id": "inv_item_500",
          "item_def_id": "broom_firebolt",
          "quantity": 1
        },
        {
          "instance_id": "inv_item_501",
          "item_def_id": "potion_felix_felicis",
          "quantity": 5
        }
      ],
      "transaction_id": "txn_abc123"
    },
    "audit_id": "audit_003"
  }
}
```

---

### 5.4 Autres Commandes GM

#### kick
```json
{
  "command": "kick",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "message": "Comportement inapproprié - premier avertissement"
  },
  "reason": "Spam répété dans le chat global"
}
```

#### ban
```json
{
  "command": "ban",
  "target_account_id": "acc_badplayer",
  "parameters": {
    "type": "temp_ban",
    "duration_hours": 24
  },
  "reason": "Harcèlement envers d'autres joueurs"
}
```

#### announce
```json
{
  "command": "announce",
  "parameters": {
    "scope": "global",
    "title": "Maintenance",
    "message": "Le serveur redémarre dans 15 minutes pour maintenance.",
    "priority": "high"
  }
}
```

#### summon
```json
{
  "command": "summon",
  "target_character_id": "char_x1y2z3",
  "parameters": {},
  "reason": "Convocation pour discussion RP"
}
```
*Téléporte le joueur vers la position du GM.*

#### freeze/unfreeze
```json
{
  "command": "freeze",
  "target_character_id": "char_x1y2z3",
  "parameters": {
    "duration_seconds": 60
  },
  "reason": "Stabilisation pour investigation anti-cheat"
}
```
*Empêche le joueur de bouger/agir.*

---

## 6. Sécurité JWT

### 6.1 Structure JWT

**Header:**
```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key_2024_01"
}
```

**Payload:**
```json
{
  "iss": "hp-mmo-auth",
  "sub": "acc_a1b2c3d4",
  "iat": 1704067200,
  "exp": 1704153600,
  "jti": "jwt_unique_id",

  "account_id": "acc_a1b2c3d4",
  "character_id": "char_x1y2z3",
  "role": "player",
  "house": "venatrix",
  "grade": 5,

  "permissions": [
    "spell.tier1.cast",
    "spell.tier2.cast",
    "spell.tier3.cast",
    "zone.forest.enter"
  ],

  "session_id": "sess_abc123",
  "scopes": ["game:play", "chat:read", "chat:write"]
}
```

### 6.2 Tokens

| Token | Durée | Stockage | Usage |
|-------|-------|----------|-------|
| Access Token | 24h | Client memory | Authentification API/WS |
| Refresh Token | 30 jours | Secure storage | Renouveler access token |
| Connection Token | 30s | Redis | Connexion WS one-time |
| Transfer Token | 30s | Redis | Migration inter-zone |

### 6.3 Validation

```python
def validate_jwt(token):
    # 1. Vérifier signature RS256
    # 2. Vérifier expiration
    # 3. Vérifier issuer = "hp-mmo-auth"
    # 4. Vérifier jti pas révoqué (Redis blacklist)
    # 5. Vérifier session_id active (Redis)
    # 6. Retourner claims
```

### 6.4 Révocation

```python
def revoke_session(session_id):
    # Ajouter tous les JWT de cette session à la blacklist
    redis.sadd(f"revoked:{session_id}", jwt_id, ex=86400)

    # Supprimer la session
    redis.delete(f"session:{session_id}")

    # Déconnecter si en ligne
    publish_to_nats("session.revoked", session_id)
```

---

## 7. Anti-Cheat Minimal

### 7.1 Validations Serveur

#### Movement
```python
def validate_move(character, intent):
    # 1. Calculer distance depuis dernière position
    distance = calc_distance(character.position, intent.client_position)
    time_delta = intent.timestamp - character.last_move_time

    # 2. Calculer vitesse max autorisée
    max_speed = character.stats.speed * BASE_SPEED
    if intent.is_sprinting:
        max_speed *= SPRINT_MULTIPLIER

    max_distance = max_speed * time_delta

    # 3. Tolérance pour latence
    tolerance = max_distance * 1.2  # 20% tolerance

    if distance > tolerance:
        return PositionCorrection(type="hard", reason="speed_hack_detected")

    # 4. Vérifier collision avec world geometry
    if collides_with_world(intent.client_position):
        return PositionCorrection(type="hard", reason="collision_detected")

    return PositionCorrection(type="none")
```

#### Combat
```python
def validate_combat_action(character, intent):
    spell = get_spell(intent.action_id)

    # 1. Vérifier cooldown
    if is_on_cooldown(character.id, spell.id):
        return Error("ON_COOLDOWN", remaining_ms=get_cooldown_remaining())

    # 2. Vérifier ressources
    if character.stats.mana_current < spell.mana_cost:
        return Error("INSUFFICIENT_MANA")

    # 3. Vérifier grade
    if character.grade < spell.required_grade:
        return Error("INSUFFICIENT_GRADE")

    # 4. Vérifier permission maison
    if spell.house_exclusive and character.house != spell.house_exclusive:
        return Error("HOUSE_RESTRICTED")

    # 5. Vérifier range
    target = get_entity(intent.target_entity_id)
    if calc_distance(character.position, target.position) > spell.max_range:
        return Error("OUT_OF_RANGE")

    # 6. Vérifier LOS (si applicable)
    if spell.requires_los and not has_line_of_sight(character, target):
        return Error("NO_LINE_OF_SIGHT")

    return Success()
```

### 7.2 Détection d'Anomalies

```python
# Compteurs dans Redis
anomaly_counters = {
    "speed_violations": {"threshold": 10, "window": 300},  # 10 en 5min
    "invalid_actions": {"threshold": 20, "window": 60},    # 20 en 1min
    "packet_flood": {"threshold": 200, "window": 10},      # 200 en 10s
}

def check_anomaly(character_id, anomaly_type):
    key = f"anomaly:{character_id}:{anomaly_type}"
    count = redis.incr(key)
    redis.expire(key, anomaly_counters[anomaly_type]["window"])

    if count >= anomaly_counters[anomaly_type]["threshold"]:
        flag_for_review(character_id, anomaly_type, count)
        if count >= anomaly_counters[anomaly_type]["threshold"] * 2:
            auto_kick(character_id, f"Suspicious activity: {anomaly_type}")
```

### 7.3 Logging pour Review

```json
{
  "type": "anticheat_flag",
  "character_id": "char_x1y2z3",
  "anomaly_type": "speed_violations",
  "count": 15,
  "window_seconds": 300,
  "samples": [
    {"timestamp": "...", "expected_max": 5.2, "actual": 12.8},
    {"timestamp": "...", "expected_max": 5.2, "actual": 15.1}
  ],
  "client_info": {
    "ip": "192.168.1.50",
    "version": "1.2.3"
  },
  "action_taken": "flagged_for_review"
}
```

---

## 8. Règles de Validation ACF-Compatible

### 8.1 Mapping Concepts ACF → Backend

| Concept ACF | Représentation Backend |
|-------------|------------------------|
| Character Stats | `character_stats_snapshots.acf_stats_blob` |
| Inventory | `inventory_items` + `equipment_loadouts` |
| Abilities | Permissions + Flags + Grade |
| Effects/Buffs | `character_flags` avec TTL |
| Cooldowns | Redis avec TTL |
| Quest State | `character_flags` |

### 8.2 Stats Snapshot Format

```json
{
  "version": 1,
  "base_stats": {
    "health_max": 120,
    "mana_max": 150,
    "stamina_max": 100,
    "strength": 12,
    "dexterity": 14,
    "intelligence": 18
  },
  "acf_stats_blob": {
    "attack_speed": 1.2,
    "cast_speed": 1.1,
    "critical_chance": 0.08,
    "critical_damage": 1.5,
    "magic_resistance": 15,
    "physical_resistance": 10,
    "movement_speed_modifier": 1.0,
    "cooldown_reduction": 0.05
  },
  "computed_from_equipment": {
    "spell_power_bonus": 25,
    "defense_bonus": 15
  }
}
```

### 8.3 Intention → Validation Flow

```
CLIENT                    ZONE SERVER
   │                           │
   │  INTENT {action, data}    │
   │──────────────────────────>│
   │                           │
   │                    ┌──────┴──────┐
   │                    │  VALIDATOR  │
   │                    │             │
   │                    │ 1. Parse    │
   │                    │ 2. Auth     │
   │                    │ 3. Rules    │
   │                    │ 4. Execute  │
   │                    │ 5. Persist  │
   │                    └──────┬──────┘
   │                           │
   │  RESULT {success, diff}   │
   │<──────────────────────────│
   │                           │
```

Chaque intention passe par:
1. **Parse**: Désérialisation et validation format
2. **Auth**: JWT valide, character_id match, session active
3. **Rules**: Toutes les règles métiers (grade, permissions, cooldowns, etc.)
4. **Execute**: Logique de jeu (calculs dégâts, effets, etc.)
5. **Persist**: Sauvegarde état (PostgreSQL + Redis)
