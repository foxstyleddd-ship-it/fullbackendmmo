package inventory

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ItemType represents the type of an item
type ItemType string

const (
	ItemTypeWeapon      ItemType = "weapon"
	ItemTypeArmor       ItemType = "armor"
	ItemTypeAccessory   ItemType = "accessory"
	ItemTypeConsumable  ItemType = "consumable"
	ItemTypeQuestItem   ItemType = "quest_item"
	ItemTypeMaterial    ItemType = "material"
	ItemTypeCosmetic    ItemType = "cosmetic"
	ItemTypeCurrency    ItemType = "currency"
	ItemTypeBroom       ItemType = "broom"
	ItemTypePet         ItemType = "pet"
)

// EquipmentSlot represents where an item can be equipped
type EquipmentSlot string

const (
	EquipmentSlotWand      EquipmentSlot = "wand"
	EquipmentSlotRobe      EquipmentSlot = "robe"
	EquipmentSlotHat       EquipmentSlot = "hat"
	EquipmentSlotCloak     EquipmentSlot = "cloak"
	EquipmentSlotAmulet    EquipmentSlot = "amulet"
	EquipmentSlotRingLeft  EquipmentSlot = "ring_left"
	EquipmentSlotRingRight EquipmentSlot = "ring_right"
	EquipmentSlotBoots     EquipmentSlot = "boots"
	EquipmentSlotGloves    EquipmentSlot = "gloves"
	EquipmentSlotBroom     EquipmentSlot = "broom"
)

// ItemDefinition represents an item template
type ItemDefinition struct {
	ID             string            `db:"id" json:"id"`
	ItemType       ItemType          `db:"item_type" json:"item_type"`
	EquipmentSlot  *EquipmentSlot    `db:"equipment_slot" json:"equipment_slot,omitempty"`
	DisplayName    string            `db:"display_name" json:"display_name"`
	Description    string            `db:"description" json:"description"`
	IconPath       string            `db:"icon_path" json:"icon_path"`
	IsStackable    bool              `db:"is_stackable" json:"is_stackable"`
	MaxStackSize   int               `db:"max_stack_size" json:"max_stack_size"`
	IsTradeable    bool              `db:"is_tradeable" json:"is_tradeable"`
	IsDroppable    bool              `db:"is_droppable" json:"is_droppable"`
	IsDestroyable  bool              `db:"is_destroyable" json:"is_destroyable"`
	RequiredGrade  int               `db:"required_grade" json:"required_grade"`
	RequiredHouse  *string           `db:"required_house" json:"required_house,omitempty"`
	RequiredLevel  int               `db:"required_level" json:"required_level"`
	Properties     json.RawMessage   `db:"properties" json:"properties"`
	BaseValue      int               `db:"base_value" json:"base_value"`
	CreatedAt      time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at" json:"updated_at"`
}

// InventoryItem represents an item instance in a character's inventory
type InventoryItem struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	CharacterID   uuid.UUID       `db:"character_id" json:"character_id"`
	ItemDefID     string          `db:"item_def_id" json:"item_def_id"`
	Quantity      int             `db:"quantity" json:"quantity"`
	InventorySlot *int            `db:"inventory_slot" json:"inventory_slot,omitempty"`
	InstanceData  json.RawMessage `db:"instance_data" json:"instance_data"`
	AcquiredAt    time.Time       `db:"acquired_at" json:"acquired_at"`
	UpdatedAt     time.Time       `db:"updated_at" json:"updated_at"`
}

// InventoryItemWithDef combines an inventory item with its definition
type InventoryItemWithDef struct {
	InventoryItem
	ItemDefinition
}

// EquipmentLoadout represents a character's equipped items
type EquipmentLoadout struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	CharacterID   uuid.UUID       `db:"character_id" json:"character_id"`
	LoadoutName   string          `db:"loadout_name" json:"loadout_name"`
	IsActive      bool            `db:"is_active" json:"is_active"`
	SlotWand      *uuid.UUID      `db:"slot_wand" json:"slot_wand,omitempty"`
	SlotRobe      *uuid.UUID      `db:"slot_robe" json:"slot_robe,omitempty"`
	SlotHat       *uuid.UUID      `db:"slot_hat" json:"slot_hat,omitempty"`
	SlotCloak     *uuid.UUID      `db:"slot_cloak" json:"slot_cloak,omitempty"`
	SlotAmulet    *uuid.UUID      `db:"slot_amulet" json:"slot_amulet,omitempty"`
	SlotRingLeft  *uuid.UUID      `db:"slot_ring_left" json:"slot_ring_left,omitempty"`
	SlotRingRight *uuid.UUID      `db:"slot_ring_right" json:"slot_ring_right,omitempty"`
	SlotBoots     *uuid.UUID      `db:"slot_boots" json:"slot_boots,omitempty"`
	SlotGloves    *uuid.UUID      `db:"slot_gloves" json:"slot_gloves,omitempty"`
	SlotBroom     *uuid.UUID      `db:"slot_broom" json:"slot_broom,omitempty"`
	ComputedStats json.RawMessage `db:"computed_stats" json:"computed_stats"`
	UpdatedAt     time.Time       `db:"updated_at" json:"updated_at"`
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypePurchase TransactionType = "purchase"
	TransactionTypeSale     TransactionType = "sale"
	TransactionTypeTrade    TransactionType = "trade"
	TransactionTypeGrant    TransactionType = "grant"
	TransactionTypeConsume  TransactionType = "consume"
	TransactionTypeDrop     TransactionType = "drop"
	TransactionTypePickup   TransactionType = "pickup"
	TransactionTypeQuest    TransactionType = "quest_reward"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
)

// TransactionLedger represents a transaction record
type TransactionLedger struct {
	ID                 uuid.UUID         `db:"id" json:"id"`
	IdempotencyKey     string            `db:"idempotency_key" json:"idempotency_key"`
	TransactionType    TransactionType   `db:"transaction_type" json:"transaction_type"`
	Status             TransactionStatus `db:"status" json:"status"`
	SourceCharacterID  *uuid.UUID        `db:"source_character_id" json:"source_character_id,omitempty"`
	TargetCharacterID  *uuid.UUID        `db:"target_character_id" json:"target_character_id,omitempty"`
	InitiatedBy        *uuid.UUID        `db:"initiated_by" json:"initiated_by,omitempty"`
	Items              json.RawMessage   `db:"items" json:"items"`
	Currencies         json.RawMessage   `db:"currencies" json:"currencies"`
	Metadata           json.RawMessage   `db:"metadata" json:"metadata"`
	Reason             *string           `db:"reason" json:"reason,omitempty"`
	ErrorMessage       *string           `db:"error_message" json:"error_message,omitempty"`
	CreatedAt          time.Time         `db:"created_at" json:"created_at"`
	CompletedAt        *time.Time        `db:"completed_at" json:"completed_at,omitempty"`
}

// AddItemRequest represents a request to add an item to inventory
type AddItemRequest struct {
	CharacterID  uuid.UUID       `json:"character_id"`
	ItemDefID    string          `json:"item_def_id"`
	Quantity     int             `json:"quantity"`
	InstanceData json.RawMessage `json:"instance_data,omitempty"`
}

// RemoveItemRequest represents a request to remove an item from inventory
type RemoveItemRequest struct {
	CharacterID uuid.UUID `json:"character_id"`
	ItemDefID   string    `json:"item_def_id"`
	Quantity    int       `json:"quantity"`
}

// EquipItemRequest represents a request to equip an item
type EquipItemRequest struct {
	InventoryItemID uuid.UUID     `json:"inventory_item_id" binding:"required"`
	Slot            EquipmentSlot `json:"slot" binding:"required"`
}

// UnequipItemRequest represents a request to unequip an item
type UnequipItemRequest struct {
	Slot EquipmentSlot `json:"slot" binding:"required"`
}

// PurchaseItemRequest represents a request to purchase an item
type PurchaseItemRequest struct {
	ItemDefID string `json:"item_def_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
	VendorID  string `json:"vendor_id" binding:"required"`
}

// SellItemRequest represents a request to sell an item
type SellItemRequest struct {
	InventoryItemID uuid.UUID `json:"inventory_item_id" binding:"required"`
	Quantity        int       `json:"quantity" binding:"required,min=1"`
}
