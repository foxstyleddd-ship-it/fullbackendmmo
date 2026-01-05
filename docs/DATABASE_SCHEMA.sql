-- ============================================================================
-- HP MMO BACKEND - SCHÉMA DE BASE DE DONNÉES POSTGRESQL
-- ============================================================================
-- Source of Truth pour toutes les données persistantes
-- Version: 1.0.0
-- ============================================================================

-- Extensions nécessaires
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- Pour recherche fuzzy

-- ============================================================================
-- TYPES ENUM
-- ============================================================================

CREATE TYPE house_type AS ENUM ('venatrix', 'aerwyn', 'falcon', 'brumval');
CREATE TYPE role_type AS ENUM ('player', 'vip', 'moderator', 'gm', 'admin', 'superadmin');
CREATE TYPE character_status AS ENUM ('active', 'deleted', 'banned', 'locked');
CREATE TYPE item_type AS ENUM ('consumable', 'equipment', 'quest', 'currency', 'material', 'cosmetic');
CREATE TYPE equipment_slot AS ENUM ('wand', 'robe', 'hat', 'cloak', 'amulet', 'ring_left', 'ring_right', 'boots', 'gloves', 'broom');
CREATE TYPE transaction_type AS ENUM ('grant', 'purchase', 'trade', 'quest_reward', 'drop', 'craft', 'destroy', 'admin_adjust', 'transfer');
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed', 'rolled_back');
CREATE TYPE sanction_type AS ENUM ('warning', 'mute', 'kick', 'temp_ban', 'perma_ban', 'character_lock');
CREATE TYPE audit_action AS ENUM (
    'grade_change', 'house_change', 'role_change', 'item_grant', 'item_remove',
    'currency_adjust', 'teleport', 'kick', 'ban', 'unban', 'mute', 'unmute',
    'character_create', 'character_delete', 'stat_modify', 'flag_set', 'flag_unset'
);

-- ============================================================================
-- TABLE: accounts
-- ============================================================================
-- Comptes utilisateurs (1 compte = N personnages)

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Authentification
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255) NOT NULL,  -- bcrypt hash

    -- Identité
    username VARCHAR(32) NOT NULL,
    display_name VARCHAR(64),

    -- Rôle et permissions globales
    role role_type NOT NULL DEFAULT 'player',

    -- Métadonnées
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,

    -- Sécurité
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(64),

    -- Soft delete
    deleted_at TIMESTAMPTZ,

    CONSTRAINT accounts_email_unique UNIQUE (email),
    CONSTRAINT accounts_username_unique UNIQUE (username),
    CONSTRAINT accounts_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT accounts_username_format CHECK (username ~* '^[a-zA-Z0-9_]{3,32}$')
);

CREATE INDEX idx_accounts_email ON accounts(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_username ON accounts(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_role ON accounts(role);
CREATE INDEX idx_accounts_last_login ON accounts(last_login_at);

-- Exemple d'enregistrement:
-- INSERT INTO accounts (email, password_hash, username, role) VALUES
-- ('harry@hogwarts.edu', '$2b$12$...', 'harry_potter', 'player');

-- ============================================================================
-- TABLE: characters
-- ============================================================================
-- Personnages jouables (1 compte = max 5 personnages)

CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,

    -- Identité RP
    name VARCHAR(64) NOT NULL,
    house house_type NOT NULL,
    grade INTEGER NOT NULL DEFAULT 1 CHECK (grade >= 1 AND grade <= 100),

    -- Apparence (seed pour génération client)
    appearance_data JSONB NOT NULL DEFAULT '{}',
    -- Exemple: {"body_type": 2, "face": 5, "hair": 12, "hair_color": "#8B4513", "skin_tone": 3}

    -- Position monde
    zone_id VARCHAR(64) NOT NULL DEFAULT 'hogwarts_main',
    position_x REAL NOT NULL DEFAULT 0,
    position_y REAL NOT NULL DEFAULT 0,
    position_z REAL NOT NULL DEFAULT 0,
    rotation_yaw REAL NOT NULL DEFAULT 0,

    -- Statut
    status character_status NOT NULL DEFAULT 'active',

    -- Progression
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience BIGINT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_played_at TIMESTAMPTZ,
    total_playtime_seconds BIGINT NOT NULL DEFAULT 0,

    -- Soft delete
    deleted_at TIMESTAMPTZ,

    CONSTRAINT characters_name_unique UNIQUE (name),
    CONSTRAINT characters_name_format CHECK (name ~* '^[a-zA-Z][a-zA-Z\s'']{2,63}$'),
    CONSTRAINT characters_per_account CHECK (
        (SELECT COUNT(*) FROM characters c WHERE c.account_id = characters.account_id AND c.deleted_at IS NULL) <= 5
    )
);

CREATE INDEX idx_characters_account ON characters(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_name ON characters(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_house ON characters(house) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_zone ON characters(zone_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_grade ON characters(grade);
CREATE INDEX idx_characters_status ON characters(status) WHERE status != 'deleted';

-- Exemple:
-- INSERT INTO characters (account_id, name, house, grade, zone_id) VALUES
-- ('uuid-account', 'Hermione Granger', 'aerwyn', 7, 'hogwarts_main');

-- ============================================================================
-- TABLE: character_stats_snapshots
-- ============================================================================
-- Snapshots versionnés des stats (pour historique et rollback)

CREATE TABLE character_stats_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- Version du snapshot
    version INTEGER NOT NULL,

    -- Stats de base (valeurs actuelles)
    health_current INTEGER NOT NULL DEFAULT 100,
    health_max INTEGER NOT NULL DEFAULT 100,
    mana_current INTEGER NOT NULL DEFAULT 100,
    mana_max INTEGER NOT NULL DEFAULT 100,
    stamina_current INTEGER NOT NULL DEFAULT 100,
    stamina_max INTEGER NOT NULL DEFAULT 100,

    -- Attributs
    strength INTEGER NOT NULL DEFAULT 10,
    dexterity INTEGER NOT NULL DEFAULT 10,
    intelligence INTEGER NOT NULL DEFAULT 10,
    wisdom INTEGER NOT NULL DEFAULT 10,
    charisma INTEGER NOT NULL DEFAULT 10,
    luck INTEGER NOT NULL DEFAULT 10,

    -- Stats dérivées (calculées mais cachées pour perf)
    spell_power INTEGER NOT NULL DEFAULT 0,
    defense INTEGER NOT NULL DEFAULT 0,
    speed REAL NOT NULL DEFAULT 1.0,

    -- Données ACF (format flexible pour compatibilité)
    acf_stats_blob JSONB NOT NULL DEFAULT '{}',
    -- Exemple: {"attack_speed": 1.2, "crit_chance": 0.05, "magic_resist": 15}

    -- Métadonnées
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason VARCHAR(255),  -- Pourquoi ce snapshot a été créé
    created_by UUID REFERENCES accounts(id),  -- Qui l'a créé (NULL = système)

    CONSTRAINT character_stats_version_unique UNIQUE (character_id, version)
);

CREATE INDEX idx_stats_character ON character_stats_snapshots(character_id);
CREATE INDEX idx_stats_version ON character_stats_snapshots(character_id, version DESC);
CREATE INDEX idx_stats_created ON character_stats_snapshots(created_at);

-- Fonction pour auto-incrémenter la version
CREATE OR REPLACE FUNCTION increment_stats_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version := COALESCE(
        (SELECT MAX(version) + 1 FROM character_stats_snapshots WHERE character_id = NEW.character_id),
        1
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_stats_version
    BEFORE INSERT ON character_stats_snapshots
    FOR EACH ROW EXECUTE FUNCTION increment_stats_version();

-- ============================================================================
-- TABLE: character_flags
-- ============================================================================
-- Flags/Tags gameplay (quêtes complétées, états, achievements)

CREATE TABLE character_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- Flag identifiant
    flag_key VARCHAR(128) NOT NULL,  -- Ex: "quest.main.chapter1.completed"

    -- Valeur (optionnelle, pour flags avec données)
    flag_value JSONB DEFAULT 'true',
    -- Exemples:
    -- true (flag simple)
    -- {"progress": 3, "max": 5} (progression)
    -- {"unlocked_at": "2024-01-01T12:00:00Z"} (timestamp)

    -- Métadonnées
    set_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,  -- NULL = permanent
    set_by UUID REFERENCES accounts(id),  -- NULL = système

    CONSTRAINT character_flags_unique UNIQUE (character_id, flag_key)
);

CREATE INDEX idx_flags_character ON character_flags(character_id);
CREATE INDEX idx_flags_key ON character_flags(flag_key);
CREATE INDEX idx_flags_expires ON character_flags(expires_at) WHERE expires_at IS NOT NULL;

-- Exemple:
-- INSERT INTO character_flags (character_id, flag_key, flag_value) VALUES
-- ('char-uuid', 'quest.sorting_ceremony.completed', '{"house_assigned": "venatrix"}');

-- ============================================================================
-- TABLE: item_definitions
-- ============================================================================
-- Définitions des items (catalogue, pas instances)

CREATE TABLE item_definitions (
    id VARCHAR(64) PRIMARY KEY,  -- Ex: "wand_elder", "potion_health_large"

    -- Catégorisation
    item_type item_type NOT NULL,
    equipment_slot equipment_slot,  -- NULL si non équipable

    -- Métadonnées
    display_name VARCHAR(128) NOT NULL,
    description TEXT,
    icon_path VARCHAR(255),

    -- Propriétés
    is_stackable BOOLEAN NOT NULL DEFAULT FALSE,
    max_stack_size INTEGER DEFAULT 1,
    is_tradeable BOOLEAN NOT NULL DEFAULT TRUE,
    is_droppable BOOLEAN NOT NULL DEFAULT TRUE,
    is_destroyable BOOLEAN NOT NULL DEFAULT TRUE,

    -- Restrictions
    required_grade INTEGER DEFAULT 1,
    required_house house_type,  -- NULL = toutes maisons
    required_level INTEGER DEFAULT 1,

    -- Données gameplay (flexible pour ACF)
    properties JSONB NOT NULL DEFAULT '{}',
    -- Exemple: {"damage": 50, "mana_cost": 20, "cooldown_ms": 5000}

    -- Économie
    base_value INTEGER NOT NULL DEFAULT 0,

    -- Métadonnées
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_type ON item_definitions(item_type);
CREATE INDEX idx_items_slot ON item_definitions(equipment_slot) WHERE equipment_slot IS NOT NULL;

-- Exemples:
-- INSERT INTO item_definitions (id, item_type, equipment_slot, display_name, is_stackable, properties) VALUES
-- ('wand_phoenix_feather', 'equipment', 'wand', 'Baguette Plume de Phénix', FALSE, '{"spell_power": 25, "affinity": "fire"}'),
-- ('potion_health_small', 'consumable', NULL, 'Petite Potion de Soin', TRUE, '{"heal_amount": 50, "cooldown_ms": 30000}');

-- ============================================================================
-- TABLE: inventory_items
-- ============================================================================
-- Instances d'items possédés par les personnages

CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Propriétaire
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- Définition
    item_def_id VARCHAR(64) NOT NULL REFERENCES item_definitions(id),

    -- Quantité (pour stackables)
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 1),

    -- Slot inventaire (NULL = pas dans slot spécifique)
    inventory_slot INTEGER CHECK (inventory_slot >= 0 AND inventory_slot < 100),

    -- Métadonnées instance (enchantements, durabilité, etc.)
    instance_data JSONB NOT NULL DEFAULT '{}',
    -- Exemple: {"durability": 85, "enchantment": "fire_resist_2", "bound": true}

    -- Timestamps
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT inventory_slot_unique UNIQUE (character_id, inventory_slot)
);

CREATE INDEX idx_inventory_character ON inventory_items(character_id);
CREATE INDEX idx_inventory_item_def ON inventory_items(item_def_id);
CREATE INDEX idx_inventory_slot ON inventory_items(character_id, inventory_slot);

-- ============================================================================
-- TABLE: equipment_loadouts
-- ============================================================================
-- Équipement actuellement porté

CREATE TABLE equipment_loadouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- Loadout name (pour multi-loadouts futurs)
    loadout_name VARCHAR(32) NOT NULL DEFAULT 'default',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- Slots équipés (référence vers inventory_items)
    slot_wand UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_robe UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_hat UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_cloak UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_amulet UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_ring_left UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_ring_right UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_boots UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_gloves UUID REFERENCES inventory_items(id) ON DELETE SET NULL,
    slot_broom UUID REFERENCES inventory_items(id) ON DELETE SET NULL,

    -- Stats calculées du loadout (cache)
    computed_stats JSONB NOT NULL DEFAULT '{}',

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT equipment_loadout_unique UNIQUE (character_id, loadout_name)
);

CREATE INDEX idx_equipment_character ON equipment_loadouts(character_id);
CREATE INDEX idx_equipment_active ON equipment_loadouts(character_id, is_active) WHERE is_active = TRUE;

-- ============================================================================
-- TABLE: currencies
-- ============================================================================
-- Différentes monnaies du jeu

CREATE TABLE currencies (
    id VARCHAR(32) PRIMARY KEY,  -- Ex: "galleons", "sickles", "knuts", "house_points"
    display_name VARCHAR(64) NOT NULL,
    description TEXT,
    icon_path VARCHAR(255),
    is_tradeable BOOLEAN NOT NULL DEFAULT TRUE,
    max_amount BIGINT DEFAULT 9999999999,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed currencies
INSERT INTO currencies (id, display_name, is_tradeable) VALUES
('galleons', 'Gallions', TRUE),
('sickles', 'Mornilles', TRUE),
('knuts', 'Noises', TRUE),
('house_points', 'Points de Maison', FALSE);

-- ============================================================================
-- TABLE: character_currencies
-- ============================================================================
-- Soldes de monnaies par personnage

CREATE TABLE character_currencies (
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    currency_id VARCHAR(32) NOT NULL REFERENCES currencies(id),

    amount BIGINT NOT NULL DEFAULT 0 CHECK (amount >= 0),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (character_id, currency_id)
);

CREATE INDEX idx_currencies_character ON character_currencies(character_id);

-- ============================================================================
-- TABLE: transactions_ledger
-- ============================================================================
-- Journal de toutes les transactions économiques (audit trail)

CREATE TABLE transactions_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Clé d'idempotence (prévient les doublons)
    idempotency_key VARCHAR(128) NOT NULL,

    -- Type et statut
    transaction_type transaction_type NOT NULL,
    status transaction_status NOT NULL DEFAULT 'pending',

    -- Acteurs
    source_character_id UUID REFERENCES characters(id),
    target_character_id UUID REFERENCES characters(id),
    initiated_by UUID REFERENCES accounts(id),  -- Pour actions admin

    -- Contenu de la transaction
    items JSONB NOT NULL DEFAULT '[]',
    -- Format: [{"item_def_id": "potion_health", "quantity": 5, "direction": "add/remove"}]

    currencies JSONB NOT NULL DEFAULT '[]',
    -- Format: [{"currency_id": "galleons", "amount": 100, "direction": "add/remove"}]

    -- Métadonnées
    metadata JSONB NOT NULL DEFAULT '{}',
    -- Ex: {"quest_id": "main_quest_1", "shop_id": "diagon_alley_wands"}

    reason VARCHAR(255),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    -- Erreur si échouée
    error_message TEXT,

    CONSTRAINT transactions_idempotency_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_transactions_source ON transactions_ledger(source_character_id);
CREATE INDEX idx_transactions_target ON transactions_ledger(target_character_id);
CREATE INDEX idx_transactions_type ON transactions_ledger(transaction_type);
CREATE INDEX idx_transactions_status ON transactions_ledger(status);
CREATE INDEX idx_transactions_created ON transactions_ledger(created_at);
CREATE INDEX idx_transactions_idempotency ON transactions_ledger(idempotency_key);

-- ============================================================================
-- TABLE: permissions
-- ============================================================================
-- Définitions des permissions

CREATE TABLE permissions (
    id VARCHAR(128) PRIMARY KEY,  -- Ex: "spell.tier3.cast", "zone.forest.enter"

    description TEXT,
    category VARCHAR(64),  -- Ex: "spell", "zone", "item", "admin"

    -- Conditions par défaut
    default_min_grade INTEGER DEFAULT 1,
    default_min_role role_type DEFAULT 'player',
    allowed_houses house_type[],  -- NULL = toutes

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_category ON permissions(category);

-- Seed permissions de base
INSERT INTO permissions (id, description, category, default_min_grade) VALUES
('spell.tier1.cast', 'Lancer sorts niveau 1', 'spell', 1),
('spell.tier2.cast', 'Lancer sorts niveau 2', 'spell', 3),
('spell.tier3.cast', 'Lancer sorts niveau 3', 'spell', 5),
('spell.tier4.cast', 'Lancer sorts niveau 4 (avancé)', 'spell', 7),
('spell.unforgivable.cast', 'Lancer Sortilèges Impardonnables', 'spell', 10),
('zone.forest.enter', 'Entrer dans la Forêt Interdite', 'zone', 3),
('zone.ministry.enter', 'Entrer au Ministère', 'zone', 5),
('zone.azkaban.enter', 'Entrer à Azkaban', 'zone', 8),
('item.restricted.use', 'Utiliser objets restreints', 'item', 6),
('admin.kick', 'Expulser un joueur', 'admin', NULL),
('admin.ban', 'Bannir un joueur', 'admin', NULL),
('admin.grant_item', 'Donner des objets', 'admin', NULL),
('admin.set_grade', 'Modifier un grade', 'admin', NULL),
('admin.teleport', 'Téléporter', 'admin', NULL);

-- ============================================================================
-- TABLE: role_permissions
-- ============================================================================
-- Permissions associées aux rôles

CREATE TABLE role_permissions (
    role role_type NOT NULL,
    permission_id VARCHAR(128) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

    PRIMARY KEY (role, permission_id)
);

-- Seed: Admins et GMs ont toutes les permissions admin
INSERT INTO role_permissions (role, permission_id)
SELECT 'admin', id FROM permissions WHERE category = 'admin';

INSERT INTO role_permissions (role, permission_id)
SELECT 'superadmin', id FROM permissions WHERE category = 'admin';

INSERT INTO role_permissions (role, permission_id)
SELECT 'gm', id FROM permissions WHERE id IN ('admin.kick', 'admin.grant_item', 'admin.set_grade', 'admin.teleport');

-- ============================================================================
-- TABLE: zones
-- ============================================================================
-- Définition des zones du monde

CREATE TABLE zones (
    id VARCHAR(64) PRIMARY KEY,

    display_name VARCHAR(128) NOT NULL,
    description TEXT,

    -- Type de zone
    is_safe_zone BOOLEAN NOT NULL DEFAULT FALSE,  -- Pas de PvP
    is_instanced BOOLEAN NOT NULL DEFAULT FALSE,  -- Instance privée
    max_players INTEGER DEFAULT 500,

    -- Restrictions d'accès
    required_grade INTEGER DEFAULT 1,
    required_permission VARCHAR(128) REFERENCES permissions(id),
    allowed_houses house_type[],  -- NULL = toutes

    -- Position par défaut (spawn point)
    spawn_x REAL NOT NULL DEFAULT 0,
    spawn_y REAL NOT NULL DEFAULT 0,
    spawn_z REAL NOT NULL DEFAULT 0,

    -- Zones connectées (pour navigation)
    connected_zones VARCHAR(64)[] DEFAULT '{}',

    -- Métadonnées
    metadata JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed zones
INSERT INTO zones (id, display_name, is_safe_zone, required_grade, connected_zones) VALUES
('hogwarts_main', 'Poudlard - Château Principal', TRUE, 1, ARRAY['hogsmeade', 'forbidden_forest', 'quidditch_pitch']),
('hogsmeade', 'Pré-au-Lard', TRUE, 1, ARRAY['hogwarts_main']),
('forbidden_forest', 'Forêt Interdite', FALSE, 3, ARRAY['hogwarts_main']),
('quidditch_pitch', 'Terrain de Quidditch', TRUE, 1, ARRAY['hogwarts_main']),
('ministry_of_magic', 'Ministère de la Magie', TRUE, 5, ARRAY['diagon_alley']),
('diagon_alley', 'Chemin de Traverse', TRUE, 1, ARRAY['ministry_of_magic']),
('azkaban', 'Azkaban', FALSE, 8, ARRAY[]);

-- ============================================================================
-- TABLE: audit_logs
-- ============================================================================
-- Journal d'audit pour actions admin/GM

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Qui a fait l'action
    actor_account_id UUID NOT NULL REFERENCES accounts(id),
    actor_role role_type NOT NULL,
    actor_ip INET,

    -- Sur qui/quoi
    target_account_id UUID REFERENCES accounts(id),
    target_character_id UUID REFERENCES characters(id),

    -- Action
    action audit_action NOT NULL,

    -- Détails
    old_value JSONB,
    new_value JSONB,
    reason TEXT,

    -- Métadonnées
    metadata JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor ON audit_logs(actor_account_id);
CREATE INDEX idx_audit_target_account ON audit_logs(target_account_id);
CREATE INDEX idx_audit_target_character ON audit_logs(target_character_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_created ON audit_logs(created_at);

-- Partition par mois pour performance (optionnel en prod)
-- CREATE TABLE audit_logs_2024_01 PARTITION OF audit_logs FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- ============================================================================
-- TABLE: sanctions
-- ============================================================================
-- Bans, mutes, et autres sanctions

CREATE TABLE sanctions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Cible
    account_id UUID NOT NULL REFERENCES accounts(id),
    character_id UUID REFERENCES characters(id),  -- NULL = sanction compte entier

    -- Type et durée
    sanction_type sanction_type NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,  -- NULL = permanent

    -- Qui et pourquoi
    issued_by UUID NOT NULL REFERENCES accounts(id),
    reason TEXT NOT NULL,

    -- Levée de sanction
    lifted_at TIMESTAMPTZ,
    lifted_by UUID REFERENCES accounts(id),
    lift_reason TEXT,

    -- Métadonnées
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_sanctions_account ON sanctions(account_id);
CREATE INDEX idx_sanctions_character ON sanctions(character_id);
CREATE INDEX idx_sanctions_type ON sanctions(sanction_type);
CREATE INDEX idx_sanctions_active ON sanctions(account_id, sanction_type)
    WHERE lifted_at IS NULL AND (expires_at IS NULL OR expires_at > NOW());

-- ============================================================================
-- TABLE: sessions
-- ============================================================================
-- Sessions de connexion (backup de Redis, principalement pour audit)

CREATE TABLE sessions (
    id VARCHAR(128) PRIMARY KEY,  -- Session ID (aussi stocké dans Redis)

    account_id UUID NOT NULL REFERENCES accounts(id),
    character_id UUID REFERENCES characters(id),  -- NULL si pas de perso sélectionné

    -- Connexion
    zone_id VARCHAR(64),
    zone_shard VARCHAR(64),
    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at TIMESTAMPTZ,

    -- Client info
    client_ip INET,
    client_version VARCHAR(32),
    platform VARCHAR(32),

    -- Métadonnées
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_sessions_account ON sessions(account_id);
CREATE INDEX idx_sessions_character ON sessions(character_id);
CREATE INDEX idx_sessions_active ON sessions(account_id) WHERE disconnected_at IS NULL;

-- ============================================================================
-- TABLE: world_state
-- ============================================================================
-- État persistant du monde (objets, portes, événements)

CREATE TABLE world_state (
    id VARCHAR(128) PRIMARY KEY,  -- Ex: "hogwarts.great_hall.door_main"
    zone_id VARCHAR(64) NOT NULL REFERENCES zones(id),

    -- Type d'entité
    entity_type VARCHAR(32) NOT NULL,  -- "door", "chest", "lever", "spawn_point"

    -- État actuel
    state JSONB NOT NULL DEFAULT '{}',
    -- Ex: {"is_open": false, "locked": true, "requires_key": "key_great_hall"}

    -- Position (pour entités positionnées)
    position_x REAL,
    position_y REAL,
    position_z REAL,

    -- Métadonnées
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by UUID REFERENCES accounts(id)
);

CREATE INDEX idx_world_state_zone ON world_state(zone_id);
CREATE INDEX idx_world_state_type ON world_state(entity_type);

-- ============================================================================
-- TABLE: creature_spawns
-- ============================================================================
-- État des spawns de créatures/NPCs (si persistant)

CREATE TABLE creature_spawns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Définition
    creature_def_id VARCHAR(64) NOT NULL,  -- Ex: "acromantula", "dementor"
    spawn_point_id VARCHAR(128),  -- Référence world_state ou NULL pour dynamique

    -- Zone
    zone_id VARCHAR(64) NOT NULL REFERENCES zones(id),

    -- Position actuelle
    position_x REAL NOT NULL,
    position_y REAL NOT NULL,
    position_z REAL NOT NULL,

    -- État
    is_alive BOOLEAN NOT NULL DEFAULT TRUE,
    health_current INTEGER,
    health_max INTEGER,

    -- Respawn
    died_at TIMESTAMPTZ,
    respawn_at TIMESTAMPTZ,

    -- Métadonnées
    state_data JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_creatures_zone ON creature_spawns(zone_id);
CREATE INDEX idx_creatures_alive ON creature_spawns(zone_id, is_alive) WHERE is_alive = TRUE;
CREATE INDEX idx_creatures_respawn ON creature_spawns(respawn_at) WHERE respawn_at IS NOT NULL;

-- ============================================================================
-- FONCTIONS UTILITAIRES
-- ============================================================================

-- Fonction pour vérifier si un personnage a une permission
CREATE OR REPLACE FUNCTION check_character_permission(
    p_character_id UUID,
    p_permission_id VARCHAR(128)
)
RETURNS BOOLEAN AS $$
DECLARE
    v_character RECORD;
    v_account RECORD;
    v_permission RECORD;
    v_has_role_perm BOOLEAN;
BEGIN
    -- Récupérer le personnage et son compte
    SELECT c.*, a.role INTO v_character
    FROM characters c
    JOIN accounts a ON c.account_id = a.id
    WHERE c.id = p_character_id AND c.status = 'active';

    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;

    -- Récupérer la permission
    SELECT * INTO v_permission FROM permissions WHERE id = p_permission_id;
    IF NOT FOUND THEN
        RETURN FALSE;
    END IF;

    -- Vérifier le rôle a la permission directement
    SELECT EXISTS(
        SELECT 1 FROM role_permissions
        WHERE role = v_character.role AND permission_id = p_permission_id
    ) INTO v_has_role_perm;

    IF v_has_role_perm THEN
        RETURN TRUE;
    END IF;

    -- Vérifier grade minimum
    IF v_permission.default_min_grade IS NOT NULL AND v_character.grade < v_permission.default_min_grade THEN
        RETURN FALSE;
    END IF;

    -- Vérifier maison autorisée
    IF v_permission.allowed_houses IS NOT NULL AND v_character.house != ALL(v_permission.allowed_houses) THEN
        RETURN FALSE;
    END IF;

    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- Fonction pour transaction atomique d'inventaire
CREATE OR REPLACE FUNCTION execute_inventory_transaction(
    p_idempotency_key VARCHAR(128),
    p_transaction_type transaction_type,
    p_source_character_id UUID,
    p_target_character_id UUID,
    p_items JSONB,
    p_currencies JSONB,
    p_metadata JSONB,
    p_reason VARCHAR(255)
)
RETURNS UUID AS $$
DECLARE
    v_transaction_id UUID;
    v_item JSONB;
    v_currency JSONB;
BEGIN
    -- Vérifier idempotence
    SELECT id INTO v_transaction_id
    FROM transactions_ledger
    WHERE idempotency_key = p_idempotency_key;

    IF FOUND THEN
        RETURN v_transaction_id;  -- Transaction déjà effectuée
    END IF;

    -- Créer la transaction
    INSERT INTO transactions_ledger (
        idempotency_key, transaction_type, source_character_id,
        target_character_id, items, currencies, metadata, reason, status
    ) VALUES (
        p_idempotency_key, p_transaction_type, p_source_character_id,
        p_target_character_id, p_items, p_currencies, p_metadata, p_reason, 'pending'
    ) RETURNING id INTO v_transaction_id;

    -- Traiter les items
    FOR v_item IN SELECT * FROM jsonb_array_elements(p_items)
    LOOP
        IF v_item->>'direction' = 'add' THEN
            INSERT INTO inventory_items (character_id, item_def_id, quantity, instance_data)
            VALUES (
                p_target_character_id,
                v_item->>'item_def_id',
                (v_item->>'quantity')::INTEGER,
                COALESCE(v_item->'instance_data', '{}'::JSONB)
            );
        ELSIF v_item->>'direction' = 'remove' THEN
            -- Supprimer ou décrémenter
            UPDATE inventory_items
            SET quantity = quantity - (v_item->>'quantity')::INTEGER,
                updated_at = NOW()
            WHERE character_id = p_source_character_id
              AND item_def_id = v_item->>'item_def_id'
              AND quantity >= (v_item->>'quantity')::INTEGER;

            -- Supprimer si quantité <= 0
            DELETE FROM inventory_items
            WHERE character_id = p_source_character_id
              AND item_def_id = v_item->>'item_def_id'
              AND quantity <= 0;
        END IF;
    END LOOP;

    -- Traiter les currencies
    FOR v_currency IN SELECT * FROM jsonb_array_elements(p_currencies)
    LOOP
        IF v_currency->>'direction' = 'add' THEN
            INSERT INTO character_currencies (character_id, currency_id, amount)
            VALUES (p_target_character_id, v_currency->>'currency_id', (v_currency->>'amount')::BIGINT)
            ON CONFLICT (character_id, currency_id)
            DO UPDATE SET amount = character_currencies.amount + (v_currency->>'amount')::BIGINT,
                          updated_at = NOW();
        ELSIF v_currency->>'direction' = 'remove' THEN
            UPDATE character_currencies
            SET amount = amount - (v_currency->>'amount')::BIGINT,
                updated_at = NOW()
            WHERE character_id = p_source_character_id
              AND currency_id = v_currency->>'currency_id'
              AND amount >= (v_currency->>'amount')::BIGINT;
        END IF;
    END LOOP;

    -- Marquer comme complétée
    UPDATE transactions_ledger
    SET status = 'completed', completed_at = NOW()
    WHERE id = v_transaction_id;

    RETURN v_transaction_id;

EXCEPTION WHEN OTHERS THEN
    -- Marquer comme échouée
    UPDATE transactions_ledger
    SET status = 'failed', error_message = SQLERRM
    WHERE id = v_transaction_id;

    RAISE;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_characters_updated_at
    BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_inventory_updated_at
    BEFORE UPDATE ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_equipment_updated_at
    BEFORE UPDATE ON equipment_loadouts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- EXEMPLES D'ENREGISTREMENTS
-- ============================================================================

-- Compte exemple
/*
INSERT INTO accounts (id, email, password_hash, username, role) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'test@hogwarts.edu',
 '$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4qIuWZIpUDNTpTt.', -- "password123"
 'test_wizard',
 'player');

-- Personnage exemple
INSERT INTO characters (id, account_id, name, house, grade, zone_id) VALUES
('c1d2e3f4-5678-90ab-cdef-123456789012',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Draco Testus',
 'venatrix',
 5,
 'hogwarts_main');

-- Stats snapshot exemple
INSERT INTO character_stats_snapshots (character_id, health_max, mana_max, intelligence, acf_stats_blob)
VALUES (
    'c1d2e3f4-5678-90ab-cdef-123456789012',
    120, 150, 15,
    '{"spell_power": 35, "crit_chance": 0.08}'
);

-- Item inventaire exemple
INSERT INTO inventory_items (character_id, item_def_id, quantity, instance_data) VALUES
('c1d2e3f4-5678-90ab-cdef-123456789012', 'wand_phoenix_feather', 1,
 '{"durability": 100, "enchantment": "lumos_enhanced"}'),
('c1d2e3f4-5678-90ab-cdef-123456789012', 'potion_health_small', 5, '{}');

-- Currency exemple
INSERT INTO character_currencies (character_id, currency_id, amount) VALUES
('c1d2e3f4-5678-90ab-cdef-123456789012', 'galleons', 150),
('c1d2e3f4-5678-90ab-cdef-123456789012', 'sickles', 340);

-- Flag exemple
INSERT INTO character_flags (character_id, flag_key, flag_value) VALUES
('c1d2e3f4-5678-90ab-cdef-123456789012', 'quest.sorting_ceremony.completed',
 '{"completed_at": "2024-01-01T10:00:00Z", "house_assigned": "venatrix"}'),
('c1d2e3f4-5678-90ab-cdef-123456789012', 'achievement.first_spell', 'true');
*/
