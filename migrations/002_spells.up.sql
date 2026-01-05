-- Add spell system tables

-- Spell schools (types)
CREATE TYPE spell_school AS ENUM (
    'charms',
    'transfiguration',
    'potions',
    'defense_against_dark_arts',
    'dark_arts',
    'herbology',
    'care_of_magical_creatures',
    'divination',
    'ancient_runes',
    'arithmancy'
);

-- Spell definitions table
CREATE TABLE spell_definitions (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    spell_school spell_school NOT NULL,
    required_grade INTEGER NOT NULL DEFAULT 1,
    required_house house_type,
    mana_cost INTEGER NOT NULL DEFAULT 0,
    cooldown_seconds INTEGER NOT NULL DEFAULT 0,
    is_forbidden BOOLEAN NOT NULL DEFAULT FALSE,
    icon_path VARCHAR(255),
    properties JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_spells_school ON spell_definitions(spell_school);
CREATE INDEX idx_spells_grade ON spell_definitions(required_grade);
CREATE INDEX idx_spells_forbidden ON spell_definitions(is_forbidden);

-- Character spells (learned/unlocked)
CREATE TABLE character_spells (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    spell_id VARCHAR(64) NOT NULL REFERENCES spell_definitions(id),
    learned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    times_cast INTEGER NOT NULL DEFAULT 0,
    proficiency_level INTEGER NOT NULL DEFAULT 1 CHECK (proficiency_level >= 1 AND proficiency_level <= 10),
    CONSTRAINT character_spell_unique UNIQUE (character_id, spell_id)
);

CREATE INDEX idx_character_spells_character ON character_spells(character_id);
CREATE INDEX idx_character_spells_spell ON character_spells(spell_id);
CREATE INDEX idx_character_spells_proficiency ON character_spells(proficiency_level);

-- Trigger for updated_at
CREATE TRIGGER update_spell_definitions_updated_at
    BEFORE UPDATE ON spell_definitions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Seed common spells
INSERT INTO spell_definitions (id, name, description, spell_school, required_grade, mana_cost, cooldown_seconds, properties) VALUES
-- Charms (Grade 1-3)
('lumos', 'Lumos', 'Creates light at the wand tip', 'charms', 1, 5, 0, '{"duration_seconds": 300, "light_radius": 10}'),
('nox', 'Nox', 'Extinguishes light from Lumos', 'charms', 1, 0, 0, '{}'),
('alohomora', 'Alohomora', 'Unlocks doors and objects', 'charms', 1, 10, 5, '{"max_lock_level": 3}'),
('wingardium_leviosa', 'Wingardium Leviosa', 'Levitates objects', 'charms', 1, 15, 3, '{"max_weight_kg": 50}'),
('expelliarmus', 'Expelliarmus', 'Disarms opponent', 'charms', 2, 20, 8, '{"range_meters": 15, "damage": 10}'),
('accio', 'Accio', 'Summons objects to the caster', 'charms', 2, 25, 10, '{"max_distance_meters": 50}'),
('reparo', 'Reparo', 'Repairs broken objects', 'charms', 2, 30, 15, '{"max_damage_level": 5}'),

-- Defense Against Dark Arts (Grade 2-5)
('protego', 'Protego', 'Creates a magical shield', 'defense_against_dark_arts', 2, 35, 12, '{"shield_strength": 100, "duration_seconds": 10}'),
('expecto_patronum', 'Expecto Patronum', 'Summons a Patronus', 'defense_against_dark_arts', 5, 100, 60, '{"dementor_defense": true, "corporeal_grade": 7}'),
('riddikulus', 'Riddikulus', 'Defeats a Boggart', 'defense_against_dark_arts', 3, 20, 30, '{"fear_immunity_seconds": 5}'),
('stupefy', 'Stupefy', 'Stuns the target', 'defense_against_dark_arts', 3, 40, 10, '{"stun_duration_seconds": 5, "damage": 25}'),

-- Transfiguration (Grade 2-6)
('vera_verto', 'Vera Verto', 'Transforms animals into water goblets', 'transfiguration', 2, 30, 20, '{"difficulty": 3}'),
('avis', 'Avis', 'Conjures a flock of birds', 'transfiguration', 3, 35, 25, '{"bird_count": 5, "duration_seconds": 60}'),
('ferula', 'Ferula', 'Conjures bandages', 'transfiguration', 2, 15, 10, '{"healing_amount": 20}'),
('evanesco', 'Evanesco', 'Vanishes objects', 'transfiguration', 5, 50, 30, '{"max_object_size": "large"}'),

-- Dark Arts (Grade 7+ or Forbidden)
('crucio', 'Crucio', 'Tortures the victim (Unforgivable)', 'dark_arts', 10, 200, 120, '{"damage_per_second": 50, "unforgivable": true}'),
('imperio', 'Imperio', 'Controls the victim (Unforgivable)', 'dark_arts', 10, 250, 180, '{"duration_seconds": 60, "unforgivable": true}'),
('avada_kedavra', 'Avada Kedavra', 'Kills instantly (Unforgivable)', 'dark_arts', 10, 300, 300, '{"instant_death": true, "unforgivable": true}'),
('sectumsempra', 'Sectumsempra', 'Slashes the target', 'dark_arts', 6, 80, 45, '{"bleed_damage": 30, "duration_seconds": 10}'),

-- Potions/Herbology utility spells
('aguamenti', 'Aguamenti', 'Conjures water', 'charms', 3, 20, 5, '{"water_liters": 10}'),
('incendio', 'Incendio', 'Creates fire', 'charms', 3, 25, 8, '{"damage_per_second": 15, "duration_seconds": 5}'),
('glacius', 'Glacius', 'Freezes objects', 'charms', 4, 35, 12, '{"freeze_duration_seconds": 8}'),

-- Advanced spells
('apparate', 'Apparition', 'Teleport to a location', 'transfiguration', 7, 100, 60, '{"max_distance_km": 50, "requires_license": true}'),
('fidelius', 'Fidelius Charm', 'Hides a secret', 'charms', 9, 500, 0, '{"permanent": true, "requires_secret_keeper": true}'),
('obliviate', 'Obliviate', 'Erases memories', 'charms', 5, 60, 90, '{"memory_range_hours": 24}');

-- Mark forbidden spells
UPDATE spell_definitions SET is_forbidden = TRUE WHERE id IN ('crucio', 'imperio', 'avada_kedavra');
