package character

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Character represents a player character
type Character struct {
	ID                  uuid.UUID       `db:"id" json:"id"`
	AccountID           uuid.UUID       `db:"account_id" json:"account_id"`
	Name                string          `db:"name" json:"name"`
	House               string          `db:"house" json:"house"`
	Grade               int             `db:"grade" json:"grade"`
	AppearanceData      json.RawMessage `db:"appearance_data" json:"appearance_data"`
	ZoneID              string          `db:"zone_id" json:"zone_id"`
	PositionX           float32         `db:"position_x" json:"position_x"`
	PositionY           float32         `db:"position_y" json:"position_y"`
	PositionZ           float32         `db:"position_z" json:"position_z"`
	RotationYaw         float32         `db:"rotation_yaw" json:"rotation_yaw"`
	Status              string          `db:"status" json:"status"`
	Level               int             `db:"level" json:"level"`
	Experience          int64           `db:"experience" json:"experience"`
	CreatedAt           time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
	LastPlayedAt        sql.NullTime    `db:"last_played_at" json:"last_played_at,omitempty"`
	TotalPlaytimeSeconds int64          `db:"total_playtime_seconds" json:"total_playtime_seconds"`
	DeletedAt           sql.NullTime    `db:"deleted_at" json:"-"`
}

// CharacterStats represents character statistics
type CharacterStats struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	CharacterID     uuid.UUID       `db:"character_id" json:"character_id"`
	Version         int             `db:"version" json:"version"`
	HealthCurrent   int             `db:"health_current" json:"health_current"`
	HealthMax       int             `db:"health_max" json:"health_max"`
	ManaCurrent     int             `db:"mana_current" json:"mana_current"`
	ManaMax         int             `db:"mana_max" json:"mana_max"`
	StaminaCurrent  int             `db:"stamina_current" json:"stamina_current"`
	StaminaMax      int             `db:"stamina_max" json:"stamina_max"`
	Strength        int             `db:"strength" json:"strength"`
	Dexterity       int             `db:"dexterity" json:"dexterity"`
	Intelligence    int             `db:"intelligence" json:"intelligence"`
	Wisdom          int             `db:"wisdom" json:"wisdom"`
	Charisma        int             `db:"charisma" json:"charisma"`
	Luck            int             `db:"luck" json:"luck"`
	SpellPower      int             `db:"spell_power" json:"spell_power"`
	Defense         int             `db:"defense" json:"defense"`
	Speed           float32         `db:"speed" json:"speed"`
	ACFStatsBlob    json.RawMessage `db:"acf_stats_blob" json:"acf_stats_blob"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	Reason          sql.NullString  `db:"reason" json:"reason,omitempty"`
	CreatedBy       uuid.NullUUID   `db:"created_by" json:"created_by,omitempty"`
}

// CharacterCurrency represents currency balances
type CharacterCurrency struct {
	CharacterID uuid.UUID `db:"character_id" json:"character_id"`
	CurrencyID  string    `db:"currency_id" json:"currency_id"`
	Amount      int64     `db:"amount" json:"amount"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CreateCharacterRequest represents character creation request
type CreateCharacterRequest struct {
	Name           string          `json:"name" binding:"required,min=3,max=64"`
	House          string          `json:"house" binding:"required,oneof=venatrix aerwyn falcon brumval"`
	AppearanceData json.RawMessage `json:"appearance_data" binding:"required"`
}

// CharacterDetailResponse represents detailed character response
type CharacterDetailResponse struct {
	Character   Character                  `json:"character"`
	Stats       CharacterStats             `json:"stats"`
	Currencies  map[string]int64           `json:"currencies"`
	Permissions []string                   `json:"permissions"`
}

// SelectCharacterResponse represents character selection response
type SelectCharacterResponse struct {
	AccessToken    string         `json:"access_token"`
	Character      CharacterInfo  `json:"character"`
	ZoneConnection ZoneConnection `json:"zone_connection"`
}

// CharacterInfo represents minimal character info
type CharacterInfo struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	House string    `json:"house"`
	Grade int       `json:"grade"`
}

// ZoneConnection represents zone connection info
type ZoneConnection struct {
	ZoneID          string `json:"zone_id"`
	Shard           string `json:"shard"`
	WebSocketURL    string `json:"websocket_url"`
	ConnectionToken string `json:"connection_token"`
}

// Position represents a 3D position
type Position struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// Rotation represents rotation
type Rotation struct {
	Yaw float32 `json:"yaw"`
}

// NullUUID for handling nullable UUIDs
type NullUUID struct {
	uuid.UUID
	Valid bool
}

func (nu *NullUUID) Scan(value interface{}) error {
	if value == nil {
		nu.Valid = false
		return nil
	}
	nu.Valid = true
	return nu.UUID.Scan(value)
}
