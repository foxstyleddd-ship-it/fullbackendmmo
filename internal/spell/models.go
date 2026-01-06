package spell

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SpellSchool represents the school of magic
type SpellSchool string

const (
	SpellSchoolCharms                 SpellSchool = "charms"
	SpellSchoolTransfiguration        SpellSchool = "transfiguration"
	SpellSchoolPotions                SpellSchool = "potions"
	SpellSchoolDefenseAgainstDarkArts SpellSchool = "defense_against_dark_arts"
	SpellSchoolDarkArts               SpellSchool = "dark_arts"
	SpellSchoolHerbology              SpellSchool = "herbology"
	SpellSchoolCareOfMagicalCreatures SpellSchool = "care_of_magical_creatures"
	SpellSchoolDivination             SpellSchool = "divination"
	SpellSchoolAncientRunes           SpellSchool = "ancient_runes"
	SpellSchoolArithmancy             SpellSchool = "arithmancy"
)

// SpellDefinition represents a spell template
type SpellDefinition struct {
	ID              string          `db:"id" json:"id"`
	Name            string          `db:"name" json:"name"`
	Description     string          `db:"description" json:"description"`
	SpellSchool     SpellSchool     `db:"spell_school" json:"spell_school"`
	RequiredGrade   int             `db:"required_grade" json:"required_grade"`
	RequiredHouse   *string         `db:"required_house" json:"required_house,omitempty"`
	ManaCost        int             `db:"mana_cost" json:"mana_cost"`
	CooldownSeconds int             `db:"cooldown_seconds" json:"cooldown_seconds"`
	IsForbidden     bool            `db:"is_forbidden" json:"is_forbidden"`
	IconPath        *string         `db:"icon_path" json:"icon_path,omitempty"`
	Properties      json.RawMessage `db:"properties" json:"properties"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

// CharacterSpell represents a spell learned by a character
type CharacterSpell struct {
	ID               uuid.UUID `db:"id" json:"id"`
	CharacterID      uuid.UUID `db:"character_id" json:"character_id"`
	SpellID          string    `db:"spell_id" json:"spell_id"`
	LearnedAt        time.Time `db:"learned_at" json:"learned_at"`
	TimesCast        int       `db:"times_cast" json:"times_cast"`
	ProficiencyLevel int       `db:"proficiency_level" json:"proficiency_level"`
}

// CharacterSpellWithDef combines a character spell with its definition
type CharacterSpellWithDef struct {
	CharacterSpell
	SpellDefinition
}

// GrantSpellRequest represents a request to grant a spell
type GrantSpellRequest struct {
	CharacterID uuid.UUID `json:"character_id"`
	SpellID     string    `json:"spell_id" binding:"required"`
}

// RemoveSpellRequest represents a request to remove a spell
type RemoveSpellRequest struct {
	CharacterID uuid.UUID `json:"character_id"`
	SpellID     string    `json:"spell_id" binding:"required"`
}
