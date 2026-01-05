-- ============================================================================
-- HP MMO BACKEND - INITIAL MIGRATION
-- ============================================================================
-- This migration creates all initial tables and functions
-- Source: docs/DATABASE_SCHEMA.sql
-- ============================================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- TYPES
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

-- TABLE: accounts
CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(32) NOT NULL,
    display_name VARCHAR(64),
    role role_type NOT NULL DEFAULT 'player',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMPTZ,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(64),
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

-- TABLE: characters
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    name VARCHAR(64) NOT NULL,
    house house_type NOT NULL,
    grade INTEGER NOT NULL DEFAULT 1 CHECK (grade >= 1 AND grade <= 100),
    appearance_data JSONB NOT NULL DEFAULT '{}',
    zone_id VARCHAR(64) NOT NULL DEFAULT 'hogwarts_main',
    position_x REAL NOT NULL DEFAULT 0,
    position_y REAL NOT NULL DEFAULT 0,
    position_z REAL NOT NULL DEFAULT 0,
    rotation_yaw REAL NOT NULL DEFAULT 0,
    status character_status NOT NULL DEFAULT 'active',
    level INTEGER NOT NULL DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_played_at TIMESTAMPTZ,
    total_playtime_seconds BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT characters_name_unique UNIQUE (name),
    CONSTRAINT characters_name_format CHECK (name ~* '^[a-zA-Z][a-zA-Z\s'']{2,63}$')
);

CREATE INDEX idx_characters_account ON characters(account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_name ON characters(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_house ON characters(house) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_zone ON characters(zone_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_characters_grade ON characters(grade);
CREATE INDEX idx_characters_status ON characters(status) WHERE status != 'deleted';

-- TABLE: character_stats_snapshots
CREATE TABLE character_stats_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    health_current INTEGER NOT NULL DEFAULT 100,
    health_max INTEGER NOT NULL DEFAULT 100,
    mana_current INTEGER NOT NULL DEFAULT 100,
    mana_max INTEGER NOT NULL DEFAULT 100,
    stamina_current INTEGER NOT NULL DEFAULT 100,
    stamina_max INTEGER NOT NULL DEFAULT 100,
    strength INTEGER NOT NULL DEFAULT 10,
    dexterity INTEGER NOT NULL DEFAULT 10,
    intelligence INTEGER NOT NULL DEFAULT 10,
    wisdom INTEGER NOT NULL DEFAULT 10,
    charisma INTEGER NOT NULL DEFAULT 10,
    luck INTEGER NOT NULL DEFAULT 10,
    spell_power INTEGER NOT NULL DEFAULT 0,
    defense INTEGER NOT NULL DEFAULT 0,
    speed REAL NOT NULL DEFAULT 1.0,
    acf_stats_blob JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reason VARCHAR(255),
    created_by UUID REFERENCES accounts(id),
    CONSTRAINT character_stats_version_unique UNIQUE (character_id, version)
);

CREATE INDEX idx_stats_character ON character_stats_snapshots(character_id);
CREATE INDEX idx_stats_version ON character_stats_snapshots(character_id, version DESC);
CREATE INDEX idx_stats_created ON character_stats_snapshots(created_at);

-- Auto-increment version trigger
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

-- TABLE: character_flags
CREATE TABLE character_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    flag_key VARCHAR(128) NOT NULL,
    flag_value JSONB DEFAULT 'true',
    set_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    set_by UUID REFERENCES accounts(id),
    CONSTRAINT character_flags_unique UNIQUE (character_id, flag_key)
);

CREATE INDEX idx_flags_character ON character_flags(character_id);
CREATE INDEX idx_flags_key ON character_flags(flag_key);
CREATE INDEX idx_flags_expires ON character_flags(expires_at) WHERE expires_at IS NOT NULL;

-- TABLE: currencies
CREATE TABLE currencies (
    id VARCHAR(32) PRIMARY KEY,
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

-- TABLE: character_currencies
CREATE TABLE character_currencies (
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    currency_id VARCHAR(32) NOT NULL REFERENCES currencies(id),
    amount BIGINT NOT NULL DEFAULT 0 CHECK (amount >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (character_id, currency_id)
);

CREATE INDEX idx_currencies_character ON character_currencies(character_id);

-- TABLE: item_definitions
CREATE TABLE item_definitions (
    id VARCHAR(64) PRIMARY KEY,
    item_type item_type NOT NULL,
    equipment_slot equipment_slot,
    display_name VARCHAR(128) NOT NULL,
    description TEXT,
    icon_path VARCHAR(255),
    is_stackable BOOLEAN NOT NULL DEFAULT FALSE,
    max_stack_size INTEGER DEFAULT 1,
    is_tradeable BOOLEAN NOT NULL DEFAULT TRUE,
    is_droppable BOOLEAN NOT NULL DEFAULT TRUE,
    is_destroyable BOOLEAN NOT NULL DEFAULT TRUE,
    required_grade INTEGER DEFAULT 1,
    required_house house_type,
    required_level INTEGER DEFAULT 1,
    properties JSONB NOT NULL DEFAULT '{}',
    base_value INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_type ON item_definitions(item_type);
CREATE INDEX idx_items_slot ON item_definitions(equipment_slot) WHERE equipment_slot IS NOT NULL;

-- TABLE: inventory_items
CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    item_def_id VARCHAR(64) NOT NULL REFERENCES item_definitions(id),
    quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity >= 1),
    inventory_slot INTEGER CHECK (inventory_slot >= 0 AND inventory_slot < 100),
    instance_data JSONB NOT NULL DEFAULT '{}',
    acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_slot_unique UNIQUE (character_id, inventory_slot)
);

CREATE INDEX idx_inventory_character ON inventory_items(character_id);
CREATE INDEX idx_inventory_item_def ON inventory_items(item_def_id);
CREATE INDEX idx_inventory_slot ON inventory_items(character_id, inventory_slot);

-- TABLE: equipment_loadouts
CREATE TABLE equipment_loadouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    loadout_name VARCHAR(32) NOT NULL DEFAULT 'default',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
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
    computed_stats JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT equipment_loadout_unique UNIQUE (character_id, loadout_name)
);

CREATE INDEX idx_equipment_character ON equipment_loadouts(character_id);
CREATE INDEX idx_equipment_active ON equipment_loadouts(character_id, is_active) WHERE is_active = TRUE;

-- TABLE: transactions_ledger
CREATE TABLE transactions_ledger (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    idempotency_key VARCHAR(128) NOT NULL,
    transaction_type transaction_type NOT NULL,
    status transaction_status NOT NULL DEFAULT 'pending',
    source_character_id UUID REFERENCES characters(id),
    target_character_id UUID REFERENCES characters(id),
    initiated_by UUID REFERENCES accounts(id),
    items JSONB NOT NULL DEFAULT '[]',
    currencies JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    reason VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    CONSTRAINT transactions_idempotency_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_transactions_source ON transactions_ledger(source_character_id);
CREATE INDEX idx_transactions_target ON transactions_ledger(target_character_id);
CREATE INDEX idx_transactions_type ON transactions_ledger(transaction_type);
CREATE INDEX idx_transactions_status ON transactions_ledger(status);
CREATE INDEX idx_transactions_created ON transactions_ledger(created_at);
CREATE INDEX idx_transactions_idempotency ON transactions_ledger(idempotency_key);

-- TABLE: permissions
CREATE TABLE permissions (
    id VARCHAR(128) PRIMARY KEY,
    description TEXT,
    category VARCHAR(64),
    default_min_grade INTEGER DEFAULT 1,
    default_min_role role_type DEFAULT 'player',
    allowed_houses house_type[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_permissions_category ON permissions(category);

-- Seed permissions
INSERT INTO permissions (id, description, category, default_min_grade) VALUES
('spell.tier1.cast', 'Lancer sorts niveau 1', 'spell', 1),
('spell.tier2.cast', 'Lancer sorts niveau 2', 'spell', 2),
('spell.tier3.cast', 'Lancer sorts niveau 3', 'spell', 3),
('spell.tier4.cast', 'Lancer sorts niveau 4 (avancé)', 'spell', 5),
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

-- TABLE: role_permissions
CREATE TABLE role_permissions (
    role role_type NOT NULL,
    permission_id VARCHAR(128) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role, permission_id)
);

-- Seed role permissions
INSERT INTO role_permissions (role, permission_id)
SELECT 'admin', id FROM permissions WHERE category = 'admin';

INSERT INTO role_permissions (role, permission_id)
SELECT 'superadmin', id FROM permissions WHERE category = 'admin';

INSERT INTO role_permissions (role, permission_id)
SELECT 'gm', id FROM permissions WHERE id IN ('admin.kick', 'admin.grant_item', 'admin.set_grade', 'admin.teleport');

-- TABLE: zones
CREATE TABLE zones (
    id VARCHAR(64) PRIMARY KEY,
    display_name VARCHAR(128) NOT NULL,
    description TEXT,
    is_safe_zone BOOLEAN NOT NULL DEFAULT FALSE,
    is_instanced BOOLEAN NOT NULL DEFAULT FALSE,
    max_players INTEGER DEFAULT 500,
    required_grade INTEGER DEFAULT 1,
    required_permission VARCHAR(128) REFERENCES permissions(id),
    allowed_houses house_type[],
    spawn_x REAL NOT NULL DEFAULT 0,
    spawn_y REAL NOT NULL DEFAULT 0,
    spawn_z REAL NOT NULL DEFAULT 0,
    connected_zones VARCHAR(64)[] DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed zones (will be in separate seed file)

-- TABLE: audit_logs
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_account_id UUID NOT NULL REFERENCES accounts(id),
    actor_role role_type NOT NULL,
    actor_ip INET,
    target_account_id UUID REFERENCES accounts(id),
    target_character_id UUID REFERENCES characters(id),
    action audit_action NOT NULL,
    old_value JSONB,
    new_value JSONB,
    reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_actor ON audit_logs(actor_account_id);
CREATE INDEX idx_audit_target_account ON audit_logs(target_account_id);
CREATE INDEX idx_audit_target_character ON audit_logs(target_character_id);
CREATE INDEX idx_audit_action ON audit_logs(action);
CREATE INDEX idx_audit_created ON audit_logs(created_at);

-- TABLE: sanctions
CREATE TABLE sanctions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    character_id UUID REFERENCES characters(id),
    sanction_type sanction_type NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    issued_by UUID NOT NULL REFERENCES accounts(id),
    reason TEXT NOT NULL,
    lifted_at TIMESTAMPTZ,
    lifted_by UUID REFERENCES accounts(id),
    lift_reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_sanctions_account ON sanctions(account_id);
CREATE INDEX idx_sanctions_character ON sanctions(character_id);
CREATE INDEX idx_sanctions_type ON sanctions(sanction_type);
CREATE INDEX idx_sanctions_active ON sanctions(account_id, sanction_type)
    WHERE lifted_at IS NULL AND (expires_at IS NULL OR expires_at > NOW());

-- TABLE: sessions
CREATE TABLE sessions (
    id VARCHAR(128) PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id),
    character_id UUID REFERENCES characters(id),
    zone_id VARCHAR(64),
    zone_shard VARCHAR(64),
    connected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_activity_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    disconnected_at TIMESTAMPTZ,
    client_ip INET,
    client_version VARCHAR(32),
    platform VARCHAR(32),
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_sessions_account ON sessions(account_id);
CREATE INDEX idx_sessions_character ON sessions(character_id);
CREATE INDEX idx_sessions_active ON sessions(account_id) WHERE disconnected_at IS NULL;

-- Auto-update updated_at trigger
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

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'HP MMO Backend: Initial migration completed successfully!';
END $$;
