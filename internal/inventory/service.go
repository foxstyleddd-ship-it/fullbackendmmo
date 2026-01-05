package inventory

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/pkg/db"
	"github.com/jmoiron/sqlx"
)

const (
	MaxInventorySlots = 100
)

// Service handles inventory business logic
type Service struct {
	repo     *Repository
	charRepo *character.Repository
	db       *sqlx.DB
}

// NewService creates a new inventory service
func NewService(repo *Repository, charRepo *character.Repository, db *sqlx.DB) *Service {
	return &Service{
		repo:     repo,
		charRepo: charRepo,
		db:       db,
	}
}

// GetInventory retrieves a character's full inventory
func (s *Service) GetInventory(ctx context.Context, characterID uuid.UUID) ([]InventoryItemWithDef, error) {
	return s.repo.GetCharacterInventoryWithDefs(ctx, characterID)
}

// GetEquipment retrieves a character's equipped items
func (s *Service) GetEquipment(ctx context.Context, characterID uuid.UUID) (*EquipmentLoadout, error) {
	return s.repo.GetActiveLoadout(ctx, characterID)
}

// AddItem adds an item to inventory with validation
func (s *Service) AddItem(ctx context.Context, req AddItemRequest) (*InventoryItem, error) {
	// Validate item definition exists
	itemDef, err := s.repo.GetItemDefinition(ctx, req.ItemDefID)
	if err != nil {
		return nil, fmt.Errorf("invalid item: %w", err)
	}

	// Check inventory capacity
	count, err := s.repo.CountInventoryItems(ctx, req.CharacterID)
	if err != nil {
		return nil, err
	}

	if count >= MaxInventorySlots && !itemDef.IsStackable {
		return nil, fmt.Errorf("inventory full (max %d slots)", MaxInventorySlots)
	}

	// Execute in transaction
	var addedItem *InventoryItem
	err = db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		item, err := s.repo.AddInventoryItem(ctx, tx, req)
		if err != nil {
			return err
		}
		addedItem = item
		return nil
	})

	return addedItem, err
}

// RemoveItem removes an item from inventory
func (s *Service) RemoveItem(ctx context.Context, inventoryItemID uuid.UUID, quantity int) error {
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		return s.repo.RemoveInventoryItem(ctx, tx, inventoryItemID, quantity)
	})
}

// EquipItem equips an item to a slot
func (s *Service) EquipItem(ctx context.Context, characterID uuid.UUID, req EquipItemRequest) error {
	// Verify ownership
	item, err := s.repo.GetInventoryItem(ctx, req.InventoryItemID)
	if err != nil {
		return err
	}

	if item.CharacterID != characterID {
		return fmt.Errorf("item does not belong to character")
	}

	// Get item definition to verify slot
	itemDef, err := s.repo.GetItemDefinition(ctx, item.ItemDefID)
	if err != nil {
		return err
	}

	if itemDef.EquipmentSlot == nil {
		return fmt.Errorf("item is not equippable")
	}

	if *itemDef.EquipmentSlot != req.Slot {
		return fmt.Errorf("item cannot be equipped to slot %s (requires %s)", req.Slot, *itemDef.EquipmentSlot)
	}

	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		return s.repo.EquipItem(ctx, tx, characterID, req.InventoryItemID, req.Slot)
	})
}

// UnequipItem unequips an item from a slot
func (s *Service) UnequipItem(ctx context.Context, characterID uuid.UUID, req UnequipItemRequest) error {
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		return s.repo.UnequipItem(ctx, tx, characterID, req.Slot)
	})
}

// PurchaseItem handles purchasing an item from a vendor
func (s *Service) PurchaseItem(ctx context.Context, characterID uuid.UUID, req PurchaseItemRequest, idempotencyKey string) error {
	// Check for existing transaction with same idempotency key
	existingTx, err := s.repo.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return err
	}

	if existingTx != nil {
		if existingTx.Status == TransactionStatusCompleted {
			return nil // Already processed
		}
		if existingTx.Status == TransactionStatusFailed {
			return fmt.Errorf("transaction previously failed: %s", *existingTx.ErrorMessage)
		}
	}

	// Get item definition
	itemDef, err := s.repo.GetItemDefinition(ctx, req.ItemDefID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	// Calculate total cost
	totalCost := itemDef.BaseValue * req.Quantity

	// Get character currencies
	currencies, err := s.charRepo.GetCharacterCurrencies(ctx, characterID)
	if err != nil {
		return err
	}

	galleons := currencies["galleons"]
	if galleons < int64(totalCost) {
		return fmt.Errorf("insufficient galleons: have %d, need %d", galleons, totalCost)
	}

	// Create transaction in atomic operation
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		// Create transaction ledger
		items, _ := json.Marshal([]map[string]interface{}{
			{"item_def_id": req.ItemDefID, "quantity": req.Quantity},
		})
		currenciesJSON, _ := json.Marshal(map[string]int{"galleons": -totalCost})
		metadata, _ := json.Marshal(map[string]string{"vendor_id": req.VendorID})

		ledger := &TransactionLedger{
			IdempotencyKey:    idempotencyKey,
			TransactionType:   TransactionTypePurchase,
			Status:            TransactionStatusPending,
			TargetCharacterID: &characterID,
			Items:             items,
			Currencies:        currenciesJSON,
			Metadata:          metadata,
		}

		err := s.repo.CreateTransaction(ctx, tx, ledger)
		if err != nil {
			return err
		}

		// Deduct currency
		updateCurrencyQuery := `
			UPDATE character_currencies
			SET amount = amount - $1, updated_at = NOW()
			WHERE character_id = $2 AND currency_id = 'galleons'`
		_, err = tx.ExecContext(ctx, updateCurrencyQuery, totalCost, characterID)
		if err != nil {
			errMsg := err.Error()
			_ = s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusFailed, &errMsg)
			return fmt.Errorf("failed to deduct currency: %w", err)
		}

		// Add item to inventory
		addReq := AddItemRequest{
			CharacterID: characterID,
			ItemDefID:   req.ItemDefID,
			Quantity:    req.Quantity,
		}

		_, err = s.repo.AddInventoryItem(ctx, tx, addReq)
		if err != nil {
			errMsg := err.Error()
			_ = s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusFailed, &errMsg)
			return fmt.Errorf("failed to add item: %w", err)
		}

		// Mark transaction as completed
		return s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusCompleted, nil)
	})
}

// SellItem handles selling an item to a vendor
func (s *Service) SellItem(ctx context.Context, characterID uuid.UUID, req SellItemRequest, idempotencyKey string) error {
	// Check for existing transaction
	existingTx, err := s.repo.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
	if err != nil {
		return err
	}

	if existingTx != nil {
		if existingTx.Status == TransactionStatusCompleted {
			return nil
		}
		if existingTx.Status == TransactionStatusFailed {
			return fmt.Errorf("transaction previously failed: %s", *existingTx.ErrorMessage)
		}
	}

	// Get inventory item
	item, err := s.repo.GetInventoryItem(ctx, req.InventoryItemID)
	if err != nil {
		return err
	}

	if item.CharacterID != characterID {
		return fmt.Errorf("item does not belong to character")
	}

	if item.Quantity < req.Quantity {
		return fmt.Errorf("insufficient quantity: have %d, need %d", item.Quantity, req.Quantity)
	}

	// Get item definition
	itemDef, err := s.repo.GetItemDefinition(ctx, item.ItemDefID)
	if err != nil {
		return err
	}

	if !itemDef.IsTradeable {
		return fmt.Errorf("item is not tradeable")
	}

	// Calculate sell price (50% of base value)
	sellPrice := (itemDef.BaseValue * req.Quantity) / 2

	// Execute in transaction
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		// Create transaction ledger
		items, _ := json.Marshal([]map[string]interface{}{
			{"item_def_id": item.ItemDefID, "quantity": req.Quantity},
		})
		currenciesJSON, _ := json.Marshal(map[string]int{"galleons": sellPrice})

		ledger := &TransactionLedger{
			IdempotencyKey:    idempotencyKey,
			TransactionType:   TransactionTypeSale,
			Status:            TransactionStatusPending,
			SourceCharacterID: &characterID,
			Items:             items,
			Currencies:        currenciesJSON,
		}

		err := s.repo.CreateTransaction(ctx, tx, ledger)
		if err != nil {
			return err
		}

		// Remove item from inventory
		err = s.repo.RemoveInventoryItem(ctx, tx, req.InventoryItemID, req.Quantity)
		if err != nil {
			errMsg := err.Error()
			_ = s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusFailed, &errMsg)
			return fmt.Errorf("failed to remove item: %w", err)
		}

		// Add currency
		updateCurrencyQuery := `
			UPDATE character_currencies
			SET amount = amount + $1, updated_at = NOW()
			WHERE character_id = $2 AND currency_id = 'galleons'`
		_, err = tx.ExecContext(ctx, updateCurrencyQuery, sellPrice, characterID)
		if err != nil {
			errMsg := err.Error()
			_ = s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusFailed, &errMsg)
			return fmt.Errorf("failed to add currency: %w", err)
		}

		// Mark transaction as completed
		return s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusCompleted, nil)
	})
}

// GrantItem grants an item to a character (admin command)
func (s *Service) GrantItem(ctx context.Context, characterID uuid.UUID, itemDefID string, quantity int, grantedBy uuid.UUID) error {
	// Validate item exists
	_, err := s.repo.GetItemDefinition(ctx, itemDefID)
	if err != nil {
		return fmt.Errorf("item not found: %w", err)
	}

	// Execute in transaction
	return db.InTransaction(s.db, func(tx *sqlx.Tx) error {
		// Create transaction ledger
		items, _ := json.Marshal([]map[string]interface{}{
			{"item_def_id": itemDefID, "quantity": quantity},
		})

		ledger := &TransactionLedger{
			IdempotencyKey:    uuid.New().String(), // Unique per grant
			TransactionType:   TransactionTypeGrant,
			Status:            TransactionStatusPending,
			TargetCharacterID: &characterID,
			InitiatedBy:       &grantedBy,
			Items:             items,
			Currencies:        []byte("{}"),
			Metadata:          []byte("{}"),
		}

		err := s.repo.CreateTransaction(ctx, tx, ledger)
		if err != nil {
			return err
		}

		// Add item to inventory
		addReq := AddItemRequest{
			CharacterID: characterID,
			ItemDefID:   itemDefID,
			Quantity:    quantity,
		}

		_, err = s.repo.AddInventoryItem(ctx, tx, addReq)
		if err != nil {
			errMsg := err.Error()
			_ = s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusFailed, &errMsg)
			return fmt.Errorf("failed to grant item: %w", err)
		}

		// Mark transaction as completed
		return s.repo.UpdateTransactionStatus(ctx, tx, ledger.ID, TransactionStatusCompleted, nil)
	})
}
