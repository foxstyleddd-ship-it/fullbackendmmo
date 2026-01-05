package spell

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles spell data access
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new spell repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetSpellDefinition retrieves a spell definition by ID
func (r *Repository) GetSpellDefinition(ctx context.Context, spellID string) (*SpellDefinition, error) {
	var spell SpellDefinition
	query := `
		SELECT id, name, description, spell_school, required_grade, required_house,
		       mana_cost, cooldown_seconds, is_forbidden, icon_path, properties,
		       created_at, updated_at
		FROM spell_definitions
		WHERE id = $1`

	err := r.db.GetContext(ctx, &spell, query, spellID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("spell not found")
	}
	return &spell, err
}

// GetAllSpellDefinitions retrieves all spell definitions
func (r *Repository) GetAllSpellDefinitions(ctx context.Context) ([]SpellDefinition, error) {
	var spells []SpellDefinition
	query := `
		SELECT id, name, description, spell_school, required_grade, required_house,
		       mana_cost, cooldown_seconds, is_forbidden, icon_path, properties,
		       created_at, updated_at
		FROM spell_definitions
		ORDER BY required_grade ASC, spell_school ASC, name ASC`

	err := r.db.SelectContext(ctx, &spells, query)
	return spells, err
}

// GetSpellsBySchool retrieves spells by school
func (r *Repository) GetSpellsBySchool(ctx context.Context, school SpellSchool) ([]SpellDefinition, error) {
	var spells []SpellDefinition
	query := `
		SELECT id, name, description, spell_school, required_grade, required_house,
		       mana_cost, cooldown_seconds, is_forbidden, icon_path, properties,
		       created_at, updated_at
		FROM spell_definitions
		WHERE spell_school = $1
		ORDER BY required_grade ASC, name ASC`

	err := r.db.SelectContext(ctx, &spells, query, school)
	return spells, err
}

// GetCharacterSpells retrieves all spells learned by a character
func (r *Repository) GetCharacterSpells(ctx context.Context, characterID uuid.UUID) ([]CharacterSpell, error) {
	var spells []CharacterSpell
	query := `
		SELECT id, character_id, spell_id, learned_at, times_cast, proficiency_level
		FROM character_spells
		WHERE character_id = $1
		ORDER BY learned_at DESC`

	err := r.db.SelectContext(ctx, &spells, query, characterID)
	return spells, err
}

// GetCharacterSpellsWithDefs retrieves character spells with their definitions
func (r *Repository) GetCharacterSpellsWithDefs(ctx context.Context, characterID uuid.UUID) ([]CharacterSpellWithDef, error) {
	var spells []CharacterSpellWithDef
	query := `
		SELECT
			cs.id, cs.character_id, cs.spell_id, cs.learned_at, cs.times_cast, cs.proficiency_level,
			sd.name, sd.description, sd.spell_school, sd.required_grade, sd.required_house,
			sd.mana_cost, sd.cooldown_seconds, sd.is_forbidden, sd.icon_path, sd.properties,
			sd.created_at, sd.updated_at
		FROM character_spells cs
		JOIN spell_definitions sd ON cs.spell_id = sd.id
		WHERE cs.character_id = $1
		ORDER BY cs.learned_at DESC`

	err := r.db.SelectContext(ctx, &spells, query, characterID)
	return spells, err
}

// GrantSpell grants a spell to a character
func (r *Repository) GrantSpell(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, spellID string) (*CharacterSpell, error) {
	// Check if already has spell
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM character_spells WHERE character_id = $1 AND spell_id = $2)`
	err := tx.GetContext(ctx, &exists, checkQuery, characterID, spellID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("character already has this spell")
	}

	// Grant spell
	spell := CharacterSpell{
		ID:               uuid.New(),
		CharacterID:      characterID,
		SpellID:          spellID,
		ProficiencyLevel: 1,
		TimesCast:        0,
	}

	query := `
		INSERT INTO character_spells (id, character_id, spell_id, proficiency_level, times_cast)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING learned_at`

	err = tx.QueryRowContext(ctx, query,
		spell.ID, spell.CharacterID, spell.SpellID, spell.ProficiencyLevel, spell.TimesCast,
	).Scan(&spell.LearnedAt)

	return &spell, err
}

// RemoveSpell removes a spell from a character
func (r *Repository) RemoveSpell(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, spellID string) error {
	query := `DELETE FROM character_spells WHERE character_id = $1 AND spell_id = $2`
	result, err := tx.ExecContext(ctx, query, characterID, spellID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("character does not have this spell")
	}

	return nil
}

// HasSpell checks if a character has a specific spell
func (r *Repository) HasSpell(ctx context.Context, characterID uuid.UUID, spellID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM character_spells WHERE character_id = $1 AND spell_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, characterID, spellID)
	return exists, err
}

// CountCharacterSpells counts spells learned by a character
func (r *Repository) CountCharacterSpells(ctx context.Context, characterID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM character_spells WHERE character_id = $1`
	err := r.db.GetContext(ctx, &count, query, characterID)
	return count, err
}

// IncrementSpellCasts increments the times_cast counter
func (r *Repository) IncrementSpellCasts(ctx context.Context, characterID uuid.UUID, spellID string) error {
	query := `UPDATE character_spells SET times_cast = times_cast + 1 WHERE character_id = $1 AND spell_id = $2`
	_, err := r.db.ExecContext(ctx, query, characterID, spellID)
	return err
}

// UpdateProficiency updates spell proficiency level
func (r *Repository) UpdateProficiency(ctx context.Context, characterID uuid.UUID, spellID string, level int) error {
	if level < 1 || level > 10 {
		return fmt.Errorf("proficiency level must be between 1 and 10")
	}

	query := `UPDATE character_spells SET proficiency_level = $1 WHERE character_id = $2 AND spell_id = $3`
	_, err := r.db.ExecContext(ctx, query, level, characterID, spellID)
	return err
}
