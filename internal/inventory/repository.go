package inventory

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Repository handles inventory data access
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a new inventory repository
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// GetItemDefinition retrieves an item definition by ID
func (r *Repository) GetItemDefinition(ctx context.Context, itemDefID string) (*ItemDefinition, error) {
	var itemDef ItemDefinition
	query := `
		SELECT id, item_type, equipment_slot, display_name, description, icon_path,
		       is_stackable, max_stack_size, is_tradeable, is_droppable, is_destroyable,
		       required_grade, required_house, required_level, properties, base_value,
		       created_at, updated_at
		FROM item_definitions
		WHERE id = $1`

	err := r.db.GetContext(ctx, &itemDef, query, itemDefID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("item definition not found")
	}
	return &itemDef, err
}

// GetCharacterInventory retrieves all items in a character's inventory
func (r *Repository) GetCharacterInventory(ctx context.Context, characterID uuid.UUID) ([]InventoryItem, error) {
	var items []InventoryItem
	query := `
		SELECT id, character_id, item_def_id, quantity, inventory_slot, instance_data, acquired_at, updated_at
		FROM inventory_items
		WHERE character_id = $1
		ORDER BY inventory_slot ASC NULLS LAST, acquired_at ASC`

	err := r.db.SelectContext(ctx, &items, query, characterID)
	return items, err
}

// GetCharacterInventoryWithDefs retrieves inventory items with their definitions
func (r *Repository) GetCharacterInventoryWithDefs(ctx context.Context, characterID uuid.UUID) ([]InventoryItemWithDef, error) {
	var items []InventoryItemWithDef
	query := `
		SELECT
			i.id, i.character_id, i.item_def_id, i.quantity, i.inventory_slot, i.instance_data, i.acquired_at, i.updated_at,
			d.item_type, d.equipment_slot, d.display_name, d.description, d.icon_path,
			d.is_stackable, d.max_stack_size, d.is_tradeable, d.is_droppable, d.is_destroyable,
			d.required_grade, d.required_house, d.required_level, d.properties, d.base_value
		FROM inventory_items i
		JOIN item_definitions d ON i.item_def_id = d.id
		WHERE i.character_id = $1
		ORDER BY i.inventory_slot ASC NULLS LAST, i.acquired_at ASC`

	err := r.db.SelectContext(ctx, &items, query, characterID)
	return items, err
}

// GetInventoryItem retrieves a specific inventory item
func (r *Repository) GetInventoryItem(ctx context.Context, inventoryItemID uuid.UUID) (*InventoryItem, error) {
	var item InventoryItem
	query := `
		SELECT id, character_id, item_def_id, quantity, inventory_slot, instance_data, acquired_at, updated_at
		FROM inventory_items
		WHERE id = $1`

	err := r.db.GetContext(ctx, &item, query, inventoryItemID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("inventory item not found")
	}
	return &item, err
}

// FindStackableItem finds an existing stackable item in inventory
func (r *Repository) FindStackableItem(ctx context.Context, characterID uuid.UUID, itemDefID string) (*InventoryItem, error) {
	var item InventoryItem
	query := `
		SELECT i.id, i.character_id, i.item_def_id, i.quantity, i.inventory_slot, i.instance_data, i.acquired_at, i.updated_at
		FROM inventory_items i
		JOIN item_definitions d ON i.item_def_id = d.id
		WHERE i.character_id = $1 AND i.item_def_id = $2 AND d.is_stackable = TRUE AND i.quantity < d.max_stack_size
		ORDER BY i.quantity DESC
		LIMIT 1`

	err := r.db.GetContext(ctx, &item, query, characterID, itemDefID)
	if err == sql.ErrNoRows {
		return nil, nil // No stackable slot found
	}
	return &item, err
}

// AddInventoryItem adds a new item to inventory or stacks with existing
func (r *Repository) AddInventoryItem(ctx context.Context, tx *sqlx.Tx, req AddItemRequest) (*InventoryItem, error) {
	// Get item definition to check stackability
	itemDef, err := r.GetItemDefinition(ctx, req.ItemDefID)
	if err != nil {
		return nil, fmt.Errorf("item definition not found: %w", err)
	}

	// If stackable, try to find existing stack
	if itemDef.IsStackable {
		existingItem, err := r.FindStackableItem(ctx, req.CharacterID, req.ItemDefID)
		if err != nil {
			return nil, err
		}

		if existingItem != nil {
			// Stack with existing item
			newQuantity := existingItem.Quantity + req.Quantity
			if newQuantity > itemDef.MaxStackSize {
				newQuantity = itemDef.MaxStackSize
			}

			query := `UPDATE inventory_items SET quantity = $1, updated_at = NOW() WHERE id = $2 RETURNING *`
			var updatedItem InventoryItem
			err = tx.GetContext(ctx, &updatedItem, query, newQuantity, existingItem.ID)
			return &updatedItem, err
		}
	}

	// Create new inventory item
	item := InventoryItem{
		ID:           uuid.New(),
		CharacterID:  req.CharacterID,
		ItemDefID:    req.ItemDefID,
		Quantity:     req.Quantity,
		InstanceData: req.InstanceData,
	}

	query := `
		INSERT INTO inventory_items (id, character_id, item_def_id, quantity, instance_data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, character_id, item_def_id, quantity, inventory_slot, instance_data, acquired_at, updated_at`

	err = tx.QueryRowxContext(ctx, query, item.ID, item.CharacterID, item.ItemDefID, item.Quantity, item.InstanceData).StructScan(&item)
	return &item, err
}

// RemoveInventoryItem removes quantity from an inventory item
func (r *Repository) RemoveInventoryItem(ctx context.Context, tx *sqlx.Tx, inventoryItemID uuid.UUID, quantity int) error {
	// Get current item
	var currentQuantity int
	query := `SELECT quantity FROM inventory_items WHERE id = $1`
	err := tx.GetContext(ctx, &currentQuantity, query, inventoryItemID)
	if err != nil {
		return fmt.Errorf("inventory item not found: %w", err)
	}

	if currentQuantity < quantity {
		return fmt.Errorf("insufficient quantity: have %d, need %d", currentQuantity, quantity)
	}

	// If removing all, delete the item
	if currentQuantity == quantity {
		deleteQuery := `DELETE FROM inventory_items WHERE id = $1`
		_, err = tx.ExecContext(ctx, deleteQuery, inventoryItemID)
		return err
	}

	// Otherwise, reduce quantity
	updateQuery := `UPDATE inventory_items SET quantity = quantity - $1, updated_at = NOW() WHERE id = $2`
	_, err = tx.ExecContext(ctx, updateQuery, quantity, inventoryItemID)
	return err
}

// GetActiveLoadout retrieves the active equipment loadout
func (r *Repository) GetActiveLoadout(ctx context.Context, characterID uuid.UUID) (*EquipmentLoadout, error) {
	var loadout EquipmentLoadout
	query := `
		SELECT id, character_id, loadout_name, is_active,
		       slot_wand, slot_robe, slot_hat, slot_cloak, slot_amulet,
		       slot_ring_left, slot_ring_right, slot_boots, slot_gloves, slot_broom,
		       computed_stats, updated_at
		FROM equipment_loadouts
		WHERE character_id = $1 AND is_active = TRUE
		LIMIT 1`

	err := r.db.GetContext(ctx, &loadout, query, characterID)
	if err == sql.ErrNoRows {
		// Create default loadout if none exists
		return r.CreateDefaultLoadout(ctx, characterID)
	}
	return &loadout, err
}

// CreateDefaultLoadout creates a default equipment loadout
func (r *Repository) CreateDefaultLoadout(ctx context.Context, characterID uuid.UUID) (*EquipmentLoadout, error) {
	loadout := EquipmentLoadout{
		ID:            uuid.New(),
		CharacterID:   characterID,
		LoadoutName:   "default",
		IsActive:      true,
		ComputedStats: []byte("{}"),
	}

	query := `
		INSERT INTO equipment_loadouts (id, character_id, loadout_name, is_active, computed_stats)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING *`

	err := r.db.QueryRowxContext(ctx, query,
		loadout.ID, loadout.CharacterID, loadout.LoadoutName, loadout.IsActive, loadout.ComputedStats,
	).StructScan(&loadout)

	return &loadout, err
}

// EquipItem equips an item to a slot
func (r *Repository) EquipItem(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, inventoryItemID uuid.UUID, slot EquipmentSlot) error {
	// Get active loadout
	var loadoutID uuid.UUID
	query := `SELECT id FROM equipment_loadouts WHERE character_id = $1 AND is_active = TRUE LIMIT 1`
	err := tx.GetContext(ctx, &loadoutID, query, characterID)
	if err != nil {
		return fmt.Errorf("active loadout not found: %w", err)
	}

	// Update the slot
	slotColumn := "slot_" + string(slot)
	updateQuery := fmt.Sprintf(`UPDATE equipment_loadouts SET %s = $1, updated_at = NOW() WHERE id = $2`, slotColumn)
	_, err = tx.ExecContext(ctx, updateQuery, inventoryItemID, loadoutID)
	return err
}

// UnequipItem unequips an item from a slot
func (r *Repository) UnequipItem(ctx context.Context, tx *sqlx.Tx, characterID uuid.UUID, slot EquipmentSlot) error {
	var loadoutID uuid.UUID
	query := `SELECT id FROM equipment_loadouts WHERE character_id = $1 AND is_active = TRUE LIMIT 1`
	err := tx.GetContext(ctx, &loadoutID, query, characterID)
	if err != nil {
		return fmt.Errorf("active loadout not found: %w", err)
	}

	slotColumn := "slot_" + string(slot)
	updateQuery := fmt.Sprintf(`UPDATE equipment_loadouts SET %s = NULL, updated_at = NOW() WHERE id = $2`, slotColumn)
	_, err = tx.ExecContext(ctx, updateQuery, loadoutID)
	return err
}

// CreateTransaction creates a transaction ledger entry
func (r *Repository) CreateTransaction(ctx context.Context, tx *sqlx.Tx, ledger *TransactionLedger) error {
	ledger.ID = uuid.New()

	query := `
		INSERT INTO transactions_ledger (
			id, idempotency_key, transaction_type, status, source_character_id, target_character_id,
			initiated_by, items, currencies, metadata, reason
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`

	return tx.QueryRowContext(ctx, query,
		ledger.ID, ledger.IdempotencyKey, ledger.TransactionType, ledger.Status,
		ledger.SourceCharacterID, ledger.TargetCharacterID, ledger.InitiatedBy,
		ledger.Items, ledger.Currencies, ledger.Metadata, ledger.Reason,
	).Scan(&ledger.CreatedAt)
}

// UpdateTransactionStatus updates the status of a transaction
func (r *Repository) UpdateTransactionStatus(ctx context.Context, tx *sqlx.Tx, transactionID uuid.UUID, status TransactionStatus, errorMsg *string) error {
	query := `UPDATE transactions_ledger SET status = $1, error_message = $2, completed_at = NOW() WHERE id = $3`
	_, err := tx.ExecContext(ctx, query, status, errorMsg, transactionID)
	return err
}

// GetTransactionByIdempotencyKey retrieves a transaction by idempotency key
func (r *Repository) GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*TransactionLedger, error) {
	var ledger TransactionLedger
	query := `SELECT * FROM transactions_ledger WHERE idempotency_key = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.db.GetContext(ctx, &ledger, query, idempotencyKey)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &ledger, err
}

// CountInventoryItems counts the number of items in inventory
func (r *Repository) CountInventoryItems(ctx context.Context, characterID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM inventory_items WHERE character_id = $1`
	err := r.db.GetContext(ctx, &count, query, characterID)
	return count, err
}

// ListItemDefinitions retrieves all item definitions
func (r *Repository) ListItemDefinitions(ctx context.Context) ([]ItemDefinition, error) {
	var items []ItemDefinition
	query := `
		SELECT id, item_type, equipment_slot, display_name, description, icon_path,
		       is_stackable, max_stack_size, is_tradeable, is_droppable, is_destroyable,
		       required_grade, required_house, required_level, properties, base_value,
		       created_at, updated_at
		FROM item_definitions
		ORDER BY display_name ASC`

	err := r.db.SelectContext(ctx, &items, query)
	return items, err
}

// CreateItemDefinition creates a new item definition
func (r *Repository) CreateItemDefinition(ctx context.Context, item *ItemDefinition) error {
	query := `
		INSERT INTO item_definitions (
			id, item_type, equipment_slot, display_name, description, icon_path,
			is_stackable, max_stack_size, is_tradeable, is_droppable, is_destroyable,
			required_grade, required_house, required_level, properties, base_value
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		item.ID, item.ItemType, item.EquipmentSlot, item.DisplayName, item.Description, item.IconPath,
		item.IsStackable, item.MaxStackSize, item.IsTradeable, item.IsDroppable, item.IsDestroyable,
		item.RequiredGrade, item.RequiredHouse, item.RequiredLevel, item.Properties, item.BaseValue,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
}

// UpdateItemDefinition updates an existing item definition
func (r *Repository) UpdateItemDefinition(ctx context.Context, item *ItemDefinition) error {
	query := `
		UPDATE item_definitions SET
			item_type = $2, equipment_slot = $3, display_name = $4, description = $5,
			icon_path = $6, is_stackable = $7, max_stack_size = $8, is_tradeable = $9,
			is_droppable = $10, is_destroyable = $11, required_grade = $12,
			required_house = $13, required_level = $14, properties = $15, base_value = $16,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		item.ID, item.ItemType, item.EquipmentSlot, item.DisplayName, item.Description,
		item.IconPath, item.IsStackable, item.MaxStackSize, item.IsTradeable,
		item.IsDroppable, item.IsDestroyable, item.RequiredGrade, item.RequiredHouse,
		item.RequiredLevel, item.Properties, item.BaseValue,
	).Scan(&item.UpdatedAt)
}

// DeleteItemDefinition deletes an item definition
func (r *Repository) DeleteItemDefinition(ctx context.Context, itemDefID string) error {
	query := `DELETE FROM item_definitions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, itemDefID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item definition not found")
	}

	return nil
}

// AddItem adds an item to inventory (convenience wrapper for admin use)
func (r *Repository) AddItem(ctx context.Context, characterID uuid.UUID, itemDefID string, quantity int, instanceData []byte) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	req := AddItemRequest{
		CharacterID:  characterID,
		ItemDefID:    itemDefID,
		Quantity:     quantity,
		InstanceData: instanceData,
	}

	_, err = r.AddInventoryItem(ctx, tx, req)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// RemoveItem removes an item from inventory (convenience wrapper for admin use)
func (r *Repository) RemoveItem(ctx context.Context, characterID uuid.UUID, itemDefID string, quantity int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Find the inventory item by character ID and item def ID
	var inventoryItemID uuid.UUID
	query := `SELECT id FROM inventory_items WHERE character_id = $1 AND item_def_id = $2 LIMIT 1`
	err = tx.GetContext(ctx, &inventoryItemID, query, characterID, itemDefID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("item not found in inventory")
		}
		return err
	}

	err = r.RemoveInventoryItem(ctx, tx, inventoryItemID, quantity)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetInventory is an alias for GetCharacterInventoryWithDefs
func (r *Repository) GetInventory(ctx context.Context, characterID uuid.UUID) ([]InventoryItemWithDef, error) {
	return r.GetCharacterInventoryWithDefs(ctx, characterID)
}
