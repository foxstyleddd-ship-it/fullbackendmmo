package achievement

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetAllAchievements retrieves all achievement definitions
func (r *Repository) GetAllAchievements(ctx context.Context) ([]AchievementDefinition, error) {
	var achievements []AchievementDefinition
	query := `SELECT * FROM achievement_definitions ORDER BY category, points`

	if err := r.db.SelectContext(ctx, &achievements, query); err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}

	return achievements, nil
}

// GetAchievementByID retrieves a specific achievement
func (r *Repository) GetAchievementByID(ctx context.Context, achievementID string) (*AchievementDefinition, error) {
	var achievement AchievementDefinition
	query := `SELECT * FROM achievement_definitions WHERE id = $1`

	if err := r.db.GetContext(ctx, &achievement, query, achievementID); err != nil {
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}

	return &achievement, nil
}

// GetCharacterAchievements retrieves all achievements for a character with definitions
func (r *Repository) GetCharacterAchievements(ctx context.Context, characterID uuid.UUID) ([]CharacterAchievementWithDef, error) {
	var achievements []CharacterAchievementWithDef
	query := `
		SELECT
			ca.id, ca.character_id, ca.achievement_id, ca.progress, ca.unlocked_at,
			ad.name, ad.description, ad.category, ad.icon_path, ad.points,
			ad.is_secret, ad.required_count, ad.reward_currencies, ad.reward_items
		FROM character_achievements ca
		JOIN achievement_definitions ad ON ca.achievement_id = ad.id
		WHERE ca.character_id = $1
		ORDER BY ca.unlocked_at DESC NULLS LAST, ad.points DESC`

	if err := r.db.SelectContext(ctx, &achievements, query, characterID); err != nil {
		return nil, fmt.Errorf("failed to get character achievements: %w", err)
	}

	return achievements, nil
}

// GetCharacterAchievementProgress retrieves progress for a specific achievement
func (r *Repository) GetCharacterAchievementProgress(ctx context.Context, characterID uuid.UUID, achievementID string) (*CharacterAchievement, error) {
	var achievement CharacterAchievement
	query := `SELECT * FROM character_achievements WHERE character_id = $1 AND achievement_id = $2`

	if err := r.db.GetContext(ctx, &achievement, query, characterID, achievementID); err != nil {
		return nil, err // Not found is expected for new achievements
	}

	return &achievement, nil
}

// UpdateAchievementProgress updates or creates achievement progress
func (r *Repository) UpdateAchievementProgress(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, achievementID string, progress int) error {
	query := `
		INSERT INTO character_achievements (character_id, achievement_id, progress)
		VALUES ($1, $2, $3)
		ON CONFLICT (character_id, achievement_id)
		DO UPDATE SET progress = $3
		WHERE character_achievements.unlocked_at IS NULL`

	executor := r.getExecutor(tx)
	_, err := executor.ExecContext(ctx, query, characterID, achievementID, progress)
	if err != nil {
		return fmt.Errorf("failed to update achievement progress: %w", err)
	}

	return nil
}

// UnlockAchievement marks an achievement as unlocked
func (r *Repository) UnlockAchievement(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, achievementID string) error {
	query := `
		UPDATE character_achievements
		SET unlocked_at = NOW()
		WHERE character_id = $1 AND achievement_id = $2 AND unlocked_at IS NULL`

	executor := r.getExecutor(tx)
	result, err := executor.ExecContext(ctx, query, characterID, achievementID)
	if err != nil {
		return fmt.Errorf("failed to unlock achievement: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("achievement already unlocked or not found")
	}

	return nil
}

// GetAchievementStats retrieves achievement statistics for a character
func (r *Repository) GetAchievementStats(ctx context.Context, characterID uuid.UUID) (*AchievementProgress, error) {
	var stats AchievementProgress

	query := `
		SELECT
			COUNT(*) as total,
			COUNT(ca.unlocked_at) as unlocked,
			COALESCE(SUM(CASE WHEN ca.unlocked_at IS NOT NULL THEN ad.points ELSE 0 END), 0) as points
		FROM achievement_definitions ad
		LEFT JOIN character_achievements ca ON ad.id = ca.achievement_id AND ca.character_id = $1`

	if err := r.db.GetContext(ctx, &stats, query, characterID); err != nil {
		return nil, fmt.Errorf("failed to get achievement stats: %w", err)
	}

	return &stats, nil
}

func (r *Repository) getExecutor(tx *sqlx.Tx) sqlx.ExtContext {
	if tx != nil {
		return tx
	}
	return r.db
}
