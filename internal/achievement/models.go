package achievement

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AchievementCategory string

const (
	CategoryCombat      AchievementCategory = "combat"
	CategoryExploration AchievementCategory = "exploration"
	CategorySocial      AchievementCategory = "social"
	CategoryCollection  AchievementCategory = "collection"
	CategoryProgression AchievementCategory = "progression"
)

type AchievementDefinition struct {
	ID               string              `db:"id" json:"id"`
	Name             string              `db:"name" json:"name"`
	Description      string              `db:"description" json:"description"`
	Category         AchievementCategory `db:"category" json:"category"`
	IconPath         string              `db:"icon_path" json:"icon_path"`
	Points           int                 `db:"points" json:"points"`
	IsSecret         bool                `db:"is_secret" json:"is_secret"`
	RequiredCount    int                 `db:"required_count" json:"required_count"`
	RewardCurrencies json.RawMessage     `db:"reward_currencies" json:"reward_currencies"`
	RewardItems      json.RawMessage     `db:"reward_items" json:"reward_items"`
	Metadata         json.RawMessage     `db:"metadata" json:"metadata"`
	CreatedAt        time.Time           `db:"created_at" json:"created_at"`
}

type CharacterAchievement struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	CharacterID   uuid.UUID  `db:"character_id" json:"character_id"`
	AchievementID string     `db:"achievement_id" json:"achievement_id"`
	Progress      int        `db:"progress" json:"progress"`
	UnlockedAt    *time.Time `db:"unlocked_at" json:"unlocked_at,omitempty"`
}

type CharacterAchievementWithDef struct {
	CharacterAchievement
	Name             string              `db:"name" json:"name"`
	Description      string              `db:"description" json:"description"`
	Category         AchievementCategory `db:"category" json:"category"`
	IconPath         string              `db:"icon_path" json:"icon_path"`
	Points           int                 `db:"points" json:"points"`
	IsSecret         bool                `db:"is_secret" json:"is_secret"`
	RequiredCount    int                 `db:"required_count" json:"required_count"`
	RewardCurrencies json.RawMessage     `db:"reward_currencies" json:"reward_currencies"`
	RewardItems      json.RawMessage     `db:"reward_items" json:"reward_items"`
}

type AchievementProgress struct {
	Total    int `json:"total"`
	Unlocked int `json:"unlocked"`
	Points   int `json:"points"`
}
