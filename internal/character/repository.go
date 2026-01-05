package character

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles database operations for characters
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new character repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// CreateCharacter creates a new character
func (r *Repository) CreateCharacter(ctx context.Context, char *Character) error {
	query := `
		INSERT INTO characters (id, account_id, name, house, grade, appearance_data, zone_id, status, level, experience)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(
		ctx, query,
		char.ID,
		char.AccountID,
		char.Name,
		char.House,
		char.Grade,
		char.AppearanceData,
		char.ZoneID,
		char.Status,
		char.Level,
		char.Experience,
	).Scan(&char.CreatedAt, &char.UpdatedAt)
}

// CreateInitialStats creates initial character stats
func (r *Repository) CreateInitialStats(ctx context.Context, stats *CharacterStats) error {
	query := `
		INSERT INTO character_stats_snapshots (
			character_id, health_current, health_max, mana_current, mana_max,
			stamina_current, stamina_max, strength, dexterity, intelligence,
			wisdom, charisma, luck, spell_power, defense, speed, acf_stats_blob, reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, version, created_at`

	return r.db.QueryRowContext(
		ctx, query,
		stats.CharacterID,
		stats.HealthCurrent, stats.HealthMax,
		stats.ManaCurrent, stats.ManaMax,
		stats.StaminaCurrent, stats.StaminaMax,
		stats.Strength, stats.Dexterity, stats.Intelligence,
		stats.Wisdom, stats.Charisma, stats.Luck,
		stats.SpellPower, stats.Defense, stats.Speed,
		stats.ACFStatsBlob,
		"Initial character creation",
	).Scan(&stats.ID, &stats.Version, &stats.CreatedAt)
}

// GetCharactersByAccountID retrieves all characters for an account
func (r *Repository) GetCharactersByAccountID(ctx context.Context, accountID uuid.UUID) ([]Character, error) {
	var characters []Character
	query := `
		SELECT id, account_id, name, house, grade, appearance_data, zone_id,
		       position_x, position_y, position_z, rotation_yaw, status, level, experience,
		       created_at, updated_at, last_played_at, total_playtime_seconds
		FROM characters
		WHERE account_id = $1 AND deleted_at IS NULL
		ORDER BY last_played_at DESC NULLS LAST, created_at DESC`

	err := r.db.SelectContext(ctx, &characters, query, accountID)
	return characters, err
}

// GetCharacterByID retrieves a character by ID
func (r *Repository) GetCharacterByID(ctx context.Context, id uuid.UUID) (*Character, error) {
	var char Character
	query := `
		SELECT id, account_id, name, house, grade, appearance_data, zone_id,
		       position_x, position_y, position_z, rotation_yaw, status, level, experience,
		       created_at, updated_at, last_played_at, total_playtime_seconds
		FROM characters
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.GetContext(ctx, &char, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("character not found")
	}
	return &char, err
}

// GetLatestStats retrieves the latest stats snapshot for a character
func (r *Repository) GetLatestStats(ctx context.Context, characterID uuid.UUID) (*CharacterStats, error) {
	var stats CharacterStats
	query := `
		SELECT id, character_id, version, health_current, health_max, mana_current, mana_max,
		       stamina_current, stamina_max, strength, dexterity, intelligence, wisdom,
		       charisma, luck, spell_power, defense, speed, acf_stats_blob, created_at, reason
		FROM character_stats_snapshots
		WHERE character_id = $1
		ORDER BY version DESC
		LIMIT 1`

	err := r.db.GetContext(ctx, &stats, query, characterID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stats not found")
	}
	return &stats, err
}

// GetCharacterCurrencies retrieves all currencies for a character
func (r *Repository) GetCharacterCurrencies(ctx context.Context, characterID uuid.UUID) (map[string]int64, error) {
	var currencies []CharacterCurrency
	query := `
		SELECT character_id, currency_id, amount, updated_at
		FROM character_currencies
		WHERE character_id = $1`

	if err := r.db.SelectContext(ctx, &currencies, query, characterID); err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, c := range currencies {
		result[c.CurrencyID] = c.Amount
	}

	return result, nil
}

// InitializeStartingCurrencies initializes starting currencies for a new character
func (r *Repository) InitializeStartingCurrencies(ctx context.Context, characterID uuid.UUID) error {
	query := `
		INSERT INTO character_currencies (character_id, currency_id, amount)
		VALUES
			($1, 'galleons', 10),
			($1, 'sickles', 50),
			($1, 'knuts', 100),
			($1, 'house_points', 0)`

	_, err := r.db.ExecContext(ctx, query, characterID)
	return err
}

// NameExists checks if a character name is already taken
func (r *Repository) NameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM characters WHERE name = $1 AND deleted_at IS NULL)`
	err := r.db.GetContext(ctx, &exists, query, name)
	return exists, err
}

// CountCharactersByAccount counts active characters for an account
func (r *Repository) CountCharactersByAccount(ctx context.Context, accountID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM characters WHERE account_id = $1 AND deleted_at IS NULL`
	err := r.db.GetContext(ctx, &count, query, accountID)
	return count, err
}

// UpdateCharacterPosition updates character position
func (r *Repository) UpdateCharacterPosition(ctx context.Context, characterID uuid.UUID, x, y, z, yaw float32) error {
	query := `
		UPDATE characters
		SET position_x = $2, position_y = $3, position_z = $4, rotation_yaw = $5, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, characterID, x, y, z, yaw)
	return err
}

// UpdateLastPlayed updates the last played timestamp
func (r *Repository) UpdateLastPlayed(ctx context.Context, characterID uuid.UUID) error {
	query := `UPDATE characters SET last_played_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, characterID)
	return err
}

// UpdateGrade updates character grade (for GM commands)
func (r *Repository) UpdateGrade(ctx context.Context, characterID uuid.UUID, newGrade int) error {
	query := `UPDATE characters SET grade = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, characterID, newGrade)
	return err
}

// UpdateZone updates character zone (for teleport)
func (r *Repository) UpdateZone(ctx context.Context, characterID uuid.UUID, zoneID string, x, y, z float32) error {
	query := `
		UPDATE characters
		SET zone_id = $2, position_x = $3, position_y = $4, position_z = $5, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, characterID, zoneID, x, y, z)
	return err
}

// SoftDeleteCharacter soft deletes a character
func (r *Repository) SoftDeleteCharacter(ctx context.Context, characterID uuid.UUID) error {
	query := `UPDATE characters SET deleted_at = NOW(), status = 'deleted', updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, characterID)
	return err
}
