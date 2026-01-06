-- Rollback achievements and social features

DROP TRIGGER IF EXISTS trigger_friendships_updated_at ON friendships;
DROP FUNCTION IF EXISTS refresh_leaderboards();
DROP MATERIALIZED VIEW IF EXISTS leaderboard_overall;
DROP TABLE IF EXISTS friendships;
DROP TABLE IF EXISTS character_achievements;
DROP TABLE IF EXISTS achievement_definitions;
DROP TYPE IF EXISTS friend_status;

DO $$
BEGIN
    RAISE NOTICE 'HP MMO Backend: Achievements and Social features rolled back successfully!';
END $$;
