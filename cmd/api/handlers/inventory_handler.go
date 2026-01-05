package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/inventory"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

// InventoryHandler handles inventory-related endpoints
type InventoryHandler struct {
	inventoryService *inventory.Service
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService *inventory.Service) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// GetInventory retrieves a character's full inventory
func (h *InventoryHandler) GetInventory(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot access another character's inventory", nil)
		return
	}

	items, err := h.inventoryService.GetInventory(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve inventory", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character_id": characterID.String(),
		"items":        items,
		"count":        len(items),
	})
}

// GetEquipment retrieves a character's equipped items
func (h *InventoryHandler) GetEquipment(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot access another character's equipment", nil)
		return
	}

	equipment, err := h.inventoryService.GetEquipment(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve equipment", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, equipment)
}

// EquipItem equips an item to a slot
func (h *InventoryHandler) EquipItem(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot equip items for another character", nil)
		return
	}

	var req inventory.EquipItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}

	err = h.inventoryService.EquipItem(c.Request.Context(), characterID, req)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "EQUIP_FAILED", err.Error(), nil)
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Item equipped successfully",
	})
}

// UnequipItem unequips an item from a slot
func (h *InventoryHandler) UnequipItem(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot unequip items for another character", nil)
		return
	}

	var req inventory.UnequipItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}

	err = h.inventoryService.UnequipItem(c.Request.Context(), characterID, req)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "UNEQUIP_FAILED", err.Error(), nil)
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Item unequipped successfully",
	})
}

// PurchaseItem handles purchasing an item from a vendor
func (h *InventoryHandler) PurchaseItem(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot purchase for another character", nil)
		return
	}

	var req inventory.PurchaseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}

	// Generate idempotency key from request ID
	requestID, _ := c.Get("request_id")
	idempotencyKey := requestID.(string)

	err = h.inventoryService.PurchaseItem(c.Request.Context(), characterID, req, idempotencyKey)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "PURCHASE_FAILED", err.Error(), nil)
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Item purchased successfully",
	})
}

// SellItem handles selling an item to a vendor
func (h *InventoryHandler) SellItem(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot sell for another character", nil)
		return
	}

	var req inventory.SellItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}

	// Generate idempotency key from request ID
	requestID, _ := c.Get("request_id")
	idempotencyKey := requestID.(string)

	err = h.inventoryService.SellItem(c.Request.Context(), characterID, req, idempotencyKey)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "SELL_FAILED", err.Error(), nil)
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Item sold successfully",
	})
}
