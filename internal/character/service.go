package character

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Service handles character business logic
type Service struct {
	repo *Repository
}

// NewService creates a new character service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListCharacters lists all characters for an account
func (s *Service) ListCharacters(ctx context.Context, accountID uuid.UUID) ([]Character, error) {
	return s.repo.GetCharactersByAccountID(ctx, accountID)
}

// GetCharacter retrieves a character by ID
func (s *Service) GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error) {
	return s.repo.GetCharacterByID(ctx, characterID)
}

// GetCharacterDetail retrieves detailed character information
func (s *Service) GetCharacterDetail(ctx context.Context, characterID uuid.UUID) (*CharacterDetailResponse, error) {
	// Get character
	char, err := s.repo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Get stats
	stats, err := s.repo.GetLatestStats(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Get currencies
	currencies, err := s.repo.GetCharacterCurrencies(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Get permissions (based on grade and house)
	permissions := s.GetPermissionsForGrade(char.Grade, char.House)

	return &CharacterDetailResponse{
		Character:   *char,
		Stats:       *stats,
		Currencies:  currencies,
		Permissions: permissions,
	}, nil
}

// CreateCharacter creates a new character
func (s *Service) CreateCharacter(ctx context.Context, accountID uuid.UUID, req CreateCharacterRequest) (*Character, error) {
	// Check character limit (max 5 per account)
	count, err := s.repo.CountCharactersByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to count characters: %w", err)
	}
	if count >= 5 {
		return nil, fmt.Errorf("maximum character limit reached (5)")
	}

	// Check name uniqueness
	exists, err := s.repo.NameExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("character name already taken")
	}

	// Create character
	char := &Character{
		ID:             uuid.New(),
		AccountID:      accountID,
		Name:           req.Name,
		House:          req.House,
		Grade:          1,
		AppearanceData: req.AppearanceData,
		ZoneID:         "hogwarts_main",
		Status:         "active",
		Level:          1,
		Experience:     0,
	}

	if err := s.repo.CreateCharacter(ctx, char); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	// Create initial stats
	stats := &CharacterStats{
		CharacterID:    char.ID,
		HealthCurrent:  100,
		HealthMax:      100,
		ManaCurrent:    100,
		ManaMax:        100,
		StaminaCurrent: 100,
		StaminaMax:     100,
		Strength:       10,
		Dexterity:      10,
		Intelligence:   10,
		Wisdom:         10,
		Charisma:       10,
		Luck:           10,
		SpellPower:     0,
		Defense:        0,
		Speed:          1.0,
		ACFStatsBlob:   json.RawMessage(`{}`),
	}

	if err := s.repo.CreateInitialStats(ctx, stats); err != nil {
		return nil, fmt.Errorf("failed to create initial stats: %w", err)
	}

	// Initialize starting currencies
	if err := s.repo.InitializeStartingCurrencies(ctx, char.ID); err != nil {
		return nil, fmt.Errorf("failed to initialize currencies: %w", err)
	}

	return char, nil
}

// SelectCharacter selects a character for play
func (s *Service) SelectCharacter(ctx context.Context, characterID uuid.UUID) error {
	// Update last played timestamp
	return s.repo.UpdateLastPlayed(ctx, characterID)
}

// UpdatePosition updates character position
func (s *Service) UpdatePosition(ctx context.Context, characterID uuid.UUID, x, y, z, yaw float32) error {
	return s.repo.UpdateCharacterPosition(ctx, characterID, x, y, z, yaw)
}

// UpdateGrade updates character grade (admin/GM only)
func (s *Service) UpdateGrade(ctx context.Context, characterID uuid.UUID, newGrade int) error {
	if newGrade < 1 || newGrade > 100 {
		return fmt.Errorf("grade must be between 1 and 100")
	}
	return s.repo.UpdateGrade(ctx, characterID, newGrade)
}

// TeleportCharacter teleports a character to a zone (admin/GM only)
func (s *Service) TeleportCharacter(ctx context.Context, characterID uuid.UUID, zoneID string, x, y, z float32) error {
	// Verify character exists
	char, err := s.repo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Store old zone for audit log
	_ = char.ZoneID

	return s.repo.UpdateZone(ctx, characterID, zoneID, x, y, z)
}

// DeleteCharacter soft deletes a character
func (s *Service) DeleteCharacter(ctx context.Context, characterID uuid.UUID) error {
	return s.repo.SoftDeleteCharacter(ctx, characterID)
}

// GetPermissionsForGrade returns permissions based on grade and house
func (s *Service) GetPermissionsForGrade(grade int, house string) []string {
	permissions := []string{}

	// Spell tier permissions based on grade
	if grade >= 1 {
		permissions = append(permissions, "spell.tier1.cast")
	}
	if grade >= 2 {
		permissions = append(permissions, "spell.tier2.cast")
	}
	if grade >= 3 {
		permissions = append(permissions, "spell.tier3.cast", "zone.forest.enter")
	}
	if grade >= 5 {
		permissions = append(permissions, "spell.tier4.cast", "zone.ministry.enter")
	}
	if grade >= 7 {
		permissions = append(permissions, "spell.tier5.cast")
	}
	if grade >= 8 {
		permissions = append(permissions, "zone.azkaban.enter")
	}
	if grade >= 10 {
		permissions = append(permissions, "spell.unforgivable.cast")
	}

	// House-specific zones
	permissions = append(permissions, fmt.Sprintf("zone.%s_common.enter", house))

	return permissions
}

// GetZoneConnection generates zone connection info for character selection
func (s *Service) GetZoneConnection(ctx context.Context, characterID uuid.UUID, zoneServerURL string) (*ZoneConnection, error) {
	char, err := s.repo.GetCharacterByID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// Generate connection token (would use JWT in production)
	connectionToken := uuid.New().String()

	return &ZoneConnection{
		ZoneID:          char.ZoneID,
		Shard:           "shard-0",
		WebSocketURL:    zoneServerURL,
		ConnectionToken: connectionToken,
	}, nil
}
