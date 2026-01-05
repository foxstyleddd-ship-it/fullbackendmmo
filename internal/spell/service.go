package spell

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/pkg/db"
	"github.com/jmoiron/sqlx"
)

// Service handles spell business logic
type Service struct {
	repo     *Repository
	charRepo *character.Repository
	db       *sqlx.DB
}

// NewService creates a new spell service
func NewService(repo *Repository, charRepo *character.Repository, db *sqlx.DB) *Service {
	return &Service{
		repo:     repo,
		charRepo: charRepo,
		db:       db,
	}
}

// GetAllSpells retrieves all available spells
func (s *Service) GetAllSpells(ctx context.Context) ([]SpellDefinition, error) {
	return s.repo.GetAllSpellDefinitions(ctx)
}

// GetSpellsBySchool retrieves spells by school
func (s *Service) GetSpellsBySchool(ctx context.Context, school SpellSchool) ([]SpellDefinition, error) {
	return s.repo.GetSpellsBySchool(ctx, school)
}

// GetCharacterSpells retrieves spells learned by a character
func (s *Service) GetCharacterSpells(ctx context.Context, characterID uuid.UUID) ([]CharacterSpellWithDef, error) {
	return s.repo.GetCharacterSpellsWithDefs(ctx, characterID)
}

// GrantSpell grants a spell to a character with validation
func (s *Service) GrantSpell(ctx context.Context, characterID uuid.UUID, spellID string) error {
	// Validate spell exists
	spellDef, err := s.repo.GetSpellDefinition(ctx, spellID)
	if err != nil {
		return fmt.Errorf("spell not found: %w", err)
	}

	// Get character to check grade/house requirements
	char, err := s.charRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Check grade requirement
	if char.Grade < spellDef.RequiredGrade {
		return fmt.Errorf("character grade %d is below required grade %d", char.Grade, spellDef.RequiredGrade)
	}

	// Check house requirement if specified
	if spellDef.RequiredHouse != nil && *spellDef.RequiredHouse != char.House {
		return fmt.Errorf("spell requires house %s, character is in %s", *spellDef.RequiredHouse, char.House)
	}

	// Grant spell in transaction
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		_, err := s.repo.GrantSpell(ctx, tx, characterID, spellID)
		return err
	})
}

// RemoveSpell removes a spell from a character
func (s *Service) RemoveSpell(ctx context.Context, characterID uuid.UUID, spellID string) error {
	// Verify character exists
	_, err := s.charRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Verify spell exists
	_, err = s.repo.GetSpellDefinition(ctx, spellID)
	if err != nil {
		return fmt.Errorf("spell not found: %w", err)
	}

	// Remove spell in transaction
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		return s.repo.RemoveSpell(ctx, tx, characterID, spellID)
	})
}

// GrantSpellBypass grants a spell bypassing requirements (admin/GM only)
func (s *Service) GrantSpellBypass(ctx context.Context, characterID uuid.UUID, spellID string) error {
	// Validate spell exists
	_, err := s.repo.GetSpellDefinition(ctx, spellID)
	if err != nil {
		return fmt.Errorf("spell not found: %w", err)
	}

	// Verify character exists
	_, err = s.charRepo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Grant spell without validation (bypass requirements)
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		_, err := s.repo.GrantSpell(ctx, tx, characterID, spellID)
		return err
	})
}
