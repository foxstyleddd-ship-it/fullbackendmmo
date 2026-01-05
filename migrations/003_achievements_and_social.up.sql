-- ============================================================================
-- HP MMO BACKEND - ACHIEVEMENTS AND SOCIAL FEATURES
-- ============================================================================

-- ACHIEVEMENTS SYSTEM
CREATE TABLE achievement_definitions (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(32) NOT NULL, -- combat, exploration, social, collection, progression
    icon_path VARCHAR(255),
    points INTEGER NOT NULL DEFAULT 10,
    is_secret BOOLEAN NOT NULL DEFAULT FALSE,
    required_count INTEGER NOT NULL DEFAULT 1,
    reward_currencies JSONB NOT NULL DEFAULT '{}',
    reward_items JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_achievement_category ON achievement_definitions(category);
CREATE INDEX idx_achievement_points ON achievement_definitions(points DESC);

-- Character achievements (tracking)
CREATE TABLE character_achievements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    achievement_id VARCHAR(64) NOT NULL REFERENCES achievement_definitions(id),
    progress INTEGER NOT NULL DEFAULT 0,
    unlocked_at TIMESTAMPTZ,
    CONSTRAINT character_achievement_unique UNIQUE (character_id, achievement_id)
);

CREATE INDEX idx_char_achievements_character ON character_achievements(character_id);
CREATE INDEX idx_char_achievements_unlocked ON character_achievements(character_id, unlocked_at) WHERE unlocked_at IS NOT NULL;

-- FRIENDS SYSTEM
CREATE TYPE friend_status AS ENUM ('pending', 'accepted', 'blocked');

CREATE TABLE friendships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    addressee_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    status friend_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT friendship_unique UNIQUE (requester_id, addressee_id),
    CONSTRAINT friendship_no_self CHECK (requester_id != addressee_id)
);

CREATE INDEX idx_friendships_requester ON friendships(requester_id);
CREATE INDEX idx_friendships_addressee ON friendships(addressee_id);
CREATE INDEX idx_friendships_status ON friendships(status);
CREATE INDEX idx_friendships_accepted ON friendships(requester_id, addressee_id) WHERE status = 'accepted';

-- LEADERBOARDS (Materialized view for performance)
CREATE MATERIALIZED VIEW leaderboard_overall AS
SELECT
    c.id,
    c.name,
    c.house,
    c.grade,
    c.level,
    c.experience,
    COALESCE(cc.amount, 0) as house_points,
    (SELECT COUNT(*) FROM character_achievements ca WHERE ca.character_id = c.id AND ca.unlocked_at IS NOT NULL) as achievement_count,
    c.total_playtime_seconds,
    c.created_at
FROM characters c
LEFT JOIN character_currencies cc ON c.id = cc.character_id AND cc.currency_id = 'house_points'
WHERE c.deleted_at IS NULL AND c.status = 'active'
ORDER BY c.level DESC, c.experience DESC;

CREATE UNIQUE INDEX idx_leaderboard_id ON leaderboard_overall(id);
CREATE INDEX idx_leaderboard_level ON leaderboard_overall(level DESC, experience DESC);
CREATE INDEX idx_leaderboard_grade ON leaderboard_overall(grade DESC);
CREATE INDEX idx_leaderboard_house ON leaderboard_overall(house, house_points DESC);

-- Function to refresh leaderboards
CREATE OR REPLACE FUNCTION refresh_leaderboards()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY leaderboard_overall;
END;
$$ LANGUAGE plpgsql;

-- Seed achievements
INSERT INTO achievement_definitions (id, name, description, category, points, required_count, reward_currencies) VALUES
-- Progression
('first_spell', 'First Spell', 'Learn your first spell', 'progression', 10, 1, '{"galleons": 100}'),
('spell_master_10', 'Spell Collector', 'Learn 10 different spells', 'progression', 50, 10, '{"galleons": 500}'),
('spell_master_25', 'Spell Master', 'Learn 25 different spells', 'progression', 100, 25, '{"galleons": 1000}'),
('level_10', 'Experienced Wizard', 'Reach level 10', 'progression', 30, 10, '{"galleons": 300}'),
('level_25', 'Master Wizard', 'Reach level 25', 'progression', 75, 25, '{"galleons": 1000}'),
('level_50', 'Legendary Wizard', 'Reach level 50', 'progression', 150, 50, '{"galleons": 5000}'),
('grade_5', 'Fifth Year', 'Reach grade 5', 'progression', 50, 5, '{"galleons": 500}'),
('grade_10', 'Graduate', 'Reach grade 10', 'progression', 100, 10, '{"galleons": 2000}'),

-- Collection
('item_hoarder', 'Item Hoarder', 'Collect 50 different items', 'collection', 50, 50, '{"galleons": 500}'),
('wealthy_wizard', 'Wealthy Wizard', 'Accumulate 10,000 galleons', 'collection', 75, 10000, '{}'),

-- Social
('first_friend', 'First Friend', 'Make your first friend', 'social', 20, 1, '{"galleons": 100}'),
('popular', 'Popular', 'Have 10 friends', 'social', 50, 10, '{"galleons": 500}'),

-- Combat
('first_duel', 'First Duel', 'Win your first duel', 'combat', 20, 1, '{"galleons": 200}'),
('duel_champion', 'Duel Champion', 'Win 50 duels', 'combat', 100, 50, '{"galleons": 2000}'),

-- Exploration
('explorer', 'Explorer', 'Visit 5 different zones', 'exploration', 30, 5, '{"galleons": 300}'),
('world_traveler', 'World Traveler', 'Visit all zones', 'exploration', 100, 10, '{"galleons": 1500}'),

-- Secret
('forbidden_knowledge', 'Forbidden Knowledge', 'Learn an Unforgivable Curse', 'progression', 200, 1, '{}'),
('azkaban_visitor', 'Azkaban Visitor', 'Visit Azkaban', 'exploration', 150, 1, '{}');

-- Update trigger for friendships
CREATE TRIGGER trigger_friendships_updated_at
    BEFORE UPDATE ON friendships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Success message
DO $$
BEGIN
    RAISE NOTICE 'HP MMO Backend: Achievements and Social features migration completed successfully!';
END $$;
