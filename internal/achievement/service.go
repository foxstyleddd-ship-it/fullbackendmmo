package achievement

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	repo *Repository
	db   *sqlx.DB
}

func NewService(repo *Repository, db *sqlx.DB) *Service {
	return &Service{
		repo: repo,
		db:   db,
	}
}

// GetAllAchievements returns all available achievements
func (s *Service) GetAllAchievements(ctx context.Context) ([]AchievementDefinition, error) {
	return s.repo.GetAllAchievements(ctx)
}

// GetCharacterAchievements returns all achievements for a character
func (s *Service) GetCharacterAchievements(ctx context.Context, characterID uuid.UUID) ([]CharacterAchievementWithDef, error) {
	return s.repo.GetCharacterAchievements(ctx, characterID)
}

// GetAchievementStats returns achievement statistics for a character
func (s *Service) GetAchievementStats(ctx context.Context, characterID uuid.UUID) (*AchievementProgress, error) {
	return s.repo.GetAchievementStats(ctx, characterID)
}

// IncrementAchievement increments progress and auto-unlocks if complete
func (s *Service) IncrementAchievement(ctx context.Context, characterID uuid.UUID, achievementID string, increment int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get achievement definition
	achievement, err := s.repo.GetAchievementByID(ctx, achievementID)
	if err != nil {
		return err
	}

	// Get current progress
	current, err := s.repo.GetCharacterAchievementProgress(ctx, characterID, achievementID)
	newProgress := increment
	if current != nil {
		if current.UnlockedAt != nil {
			return nil // Already unlocked
		}
		newProgress = current.Progress + increment
	}

	// Update progress
	if err := s.repo.UpdateAchievementProgress(ctx, tx, characterID, achievementID, newProgress); err != nil {
		return err
	}

	// Auto-unlock if requirement met
	if newProgress >= achievement.RequiredCount {
		if err := s.repo.UnlockAchievement(ctx, tx, characterID, achievementID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CheckAndUnlockAchievements is a helper to check multiple achievements at once
func (s *Service) CheckAndUnlockAchievements(ctx context.Context, characterID uuid.UUID, achievements map[string]int) error {
	for achievementID, progress := range achievements {
		if err := s.IncrementAchievement(ctx, characterID, achievementID, progress); err != nil {
			// Log but don't fail - achievements are not critical
			continue
		}
	}
	return nil
}
