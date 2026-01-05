-- ============================================================================
-- HP MMO BACKEND - SEED ITEMS
-- ============================================================================

-- Wands (Equipment)
INSERT INTO item_definitions (id, item_type, equipment_slot, display_name, description, is_stackable, required_grade, properties, base_value) VALUES
('wand_holly', 'equipment', 'wand', 'Baguette en Houx', 'Baguette classique en bois de houx', FALSE, 1, '{"spell_power": 10, "affinity": "light"}', 7),
('wand_vine_dragon', 'equipment', 'wand', 'Baguette Vigne et Dragon', 'Baguette en bois de vigne avec ventricule de dragon', FALSE, 3, '{"spell_power": 20, "affinity": "fire"}', 50),
('wand_phoenix_feather', 'equipment', 'wand', 'Baguette Plume de Phénix', 'Baguette rare avec plume de phénix', FALSE, 5, '{"spell_power": 30, "affinity": "light"}', 150),
('wand_elder', 'equipment', 'wand', 'Baguette de Sureau', 'Baguette légendaire très puissante', FALSE, 10, '{"spell_power": 50, "affinity": "all"}', 1000),

-- Robes (Equipment)
('robe_basic', 'equipment', 'robe', 'Robe de Sorcier Basique', 'Robe standard de première année', FALSE, 1, '{"defense": 5}', 5),
('robe_venatrix_formal', 'equipment', 'robe', 'Robe de Venatrix', 'Robe officielle de la maison Venatrix', FALSE, 1, '{"defense": 8}', 20),
('robe_aerwyn_formal', 'equipment', 'robe', 'Robe de Aerwyn', 'Robe officielle de la maison Aerwyn', FALSE, 1, '{"defense": 8}', 20),
('robe_falcon_formal', 'equipment', 'robe', 'Robe de Falcon', 'Robe officielle de la maison Falcon', FALSE, 1, '{"defense": 8}', 20),
('robe_brumval_formal', 'equipment', 'robe', 'Robe de Brumval', 'Robe officielle de la maison Brumval', FALSE, 1, '{"defense": 8}', 20),
('robe_advanced', 'equipment', 'robe', 'Robe Avancée', 'Robe renforcée pour sorciers expérimentés', FALSE, 5, '{"defense": 15, "magic_resist": 10}', 100),

-- Brooms (Equipment)
('broom_comet', 'equipment', 'broom', 'Comète 260', 'Balai de vol standard', FALSE, 1, '{"speed": 1.2}', 50),
('broom_nimbus_2000', 'equipment', 'broom', 'Nimbus 2000', 'Balai de course professionnel', FALSE, 2, '{"speed": 1.5}', 200),
('broom_firebolt', 'equipment', 'broom', 'Éclair de Feu', 'Le balai le plus rapide du monde', FALSE, 5, '{"speed': 2.0}', 1000),

-- Potions (Consumables)
('potion_health_small', 'consumable', NULL, 'Petite Potion de Soin', 'Restaure 50 points de vie', TRUE, 1, '{"heal_amount": 50, "cooldown_ms": 30000}', 10),
('potion_health_large', 'consumable', NULL, 'Grande Potion de Soin', 'Restaure 150 points de vie', TRUE, 3, '{"heal_amount": 150, "cooldown_ms": 30000}', 50),
('potion_mana_small', 'consumable', NULL, 'Petite Potion de Mana', 'Restaure 50 points de mana', TRUE, 1, '{"mana_amount": 50, "cooldown_ms": 30000}', 10),
('potion_mana_large', 'consumable', NULL, 'Grande Potion de Mana', 'Restaure 150 points de mana', TRUE, 3, '{"mana_amount": 150, "cooldown_ms": 30000}', 50),
('potion_felix_felicis', 'consumable', NULL, 'Felix Felicis', 'Potion de chance suprême', TRUE, 10, '{"luck_boost": 50, "duration_ms": 300000, "cooldown_ms": 86400000}', 5000),

-- Quest Items
('ingredient_mandrake_root', 'quest', NULL, 'Racine de Mandragore', 'Ingrédient pour potions', TRUE, 1, '{}', 5),
('ingredient_unicorn_hair', 'quest', NULL, 'Crin de Licorne', 'Ingrédient rare et précieux', TRUE, 3, '{}', 50),
('key_chamber_secrets', 'quest', NULL, 'Clé de la Chambre des Secrets', 'Clé mystérieuse', FALSE, 5, '{}', 0),

-- Materials
('material_wood_oak', 'material', NULL, 'Bois de Chêne', 'Matériau de fabrication commun', TRUE, 1, '{}', 2),
('material_gemstone_ruby', 'material', NULL, 'Rubis', 'Pierre précieuse', TRUE, 3, '{}', 100),

-- Cosmetics
('hat_sorting', 'cosmetic', 'hat', 'Choixpeau Magique', 'Réplique décorative', FALSE, 1, '{}', 500),
('scarf_venatrix', 'cosmetic', NULL, 'Écharpe Venatrix', 'Écharpe aux couleurs de la maison', FALSE, 1, '{}', 15),
('scarf_aerwyn', 'cosmetic', NULL, 'Écharpe Aerwyn', 'Écharpe aux couleurs de la maison', FALSE, 1, '{}', 15),
('scarf_falcon', 'cosmetic', NULL, 'Écharpe Falcon', 'Écharpe aux couleurs de la maison', FALSE, 1, '{}', 15),
('scarf_brumval', 'cosmetic', NULL, 'Écharpe Brumval', 'Écharpe aux couleurs de la maison', FALSE, 1, '{}', 15)

ON CONFLICT (id) DO NOTHING;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'Items seeded successfully!';
END $$;
