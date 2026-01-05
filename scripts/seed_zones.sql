-- ============================================================================
-- HP MMO BACKEND - SEED ZONES
-- ============================================================================

INSERT INTO zones (id, display_name, description, is_safe_zone, required_grade, connected_zones, spawn_x, spawn_y, spawn_z) VALUES
('hogwarts_main', 'Poudlard - Château Principal', 'Le château magique de Poudlard', TRUE, 1, ARRAY['hogsmeade', 'forbidden_forest', 'quidditch_pitch'], 0, 0, 0),
('hogsmeade', 'Pré-au-Lard', 'Village sorcier près de Poudlard', TRUE, 1, ARRAY['hogwarts_main'], 500, 0, 0),
('forbidden_forest', 'Forêt Interdite', 'Forêt dangereuse remplie de créatures magiques', FALSE, 3, ARRAY['hogwarts_main'], -300, 200, 0),
('quidditch_pitch', 'Terrain de Quidditch', 'Terrain de Quidditch officiel de Poudlard', TRUE, 1, ARRAY['hogwarts_main'], 200, -100, 50),
('ministry_of_magic', 'Ministère de la Magie', 'Centre du gouvernement sorcier', TRUE, 5, ARRAY['diagon_alley'], 1000, 0, -50),
('diagon_alley', 'Chemin de Traverse', 'Rue commerçante magique de Londres', TRUE, 1, ARRAY['ministry_of_magic'], 750, 0, 0),
('azkaban', 'Azkaban', 'Prison des sorciers gardée par les Détraqueurs', FALSE, 8, ARRAY[], 2000, 500, -100)
ON CONFLICT (id) DO NOTHING;

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'Zones seeded successfully!';
END $$;
