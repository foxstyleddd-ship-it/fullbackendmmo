package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/inventory"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

// ItemHandler handles item definition endpoints for admins
type ItemHandler struct {
	inventoryRepo *inventory.Repository
}

// NewItemHandler creates a new item handler
func NewItemHandler(inventoryRepo *inventory.Repository) *ItemHandler {
	return &ItemHandler{
		inventoryRepo: inventoryRepo,
	}
}

// ListItemDefinitions returns all item definitions
func (h *ItemHandler) ListItemDefinitions(c *gin.Context) {
	items, err := h.inventoryRepo.ListItemDefinitions(c.Request.Context())
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve item definitions", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"items": items,
		"count": len(items),
	})
}

// GetItemDefinition returns a specific item definition
func (h *ItemHandler) GetItemDefinition(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Item ID is required", nil)
		return
	}

	item, err := h.inventoryRepo.GetItemDefinition(c.Request.Context(), itemID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Item definition not found", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"item": item,
	})
}

// CreateItemDefinition creates a new item definition
func (h *ItemHandler) CreateItemDefinition(c *gin.Context) {
	var item inventory.ItemDefinition
	if err := c.ShouldBindJSON(&item); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	if err := h.inventoryRepo.CreateItemDefinition(c.Request.Context(), &item); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create item definition", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusCreated, gin.H{
		"item": item,
	})
}

// UpdateItemDefinition updates an existing item definition
func (h *ItemHandler) UpdateItemDefinition(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Item ID is required", nil)
		return
	}

	var item inventory.ItemDefinition
	if err := c.ShouldBindJSON(&item); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	item.ID = itemID

	if err := h.inventoryRepo.UpdateItemDefinition(c.Request.Context(), &item); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update item definition", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"item": item,
	})
}

// DeleteItemDefinition deletes an item definition
func (h *ItemHandler) DeleteItemDefinition(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Item ID is required", nil)
		return
	}

	if err := h.inventoryRepo.DeleteItemDefinition(c.Request.Context(), itemID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete item definition", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Item definition deleted successfully",
	})
}

// AddItemToInventory adds an item to a character's inventory (admin only)
func (h *ItemHandler) AddItemToInventory(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	var req struct {
		ItemDefID string `json:"item_def_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	// Add item to inventory
	if err := h.inventoryRepo.AddItem(c.Request.Context(), characterID, req.ItemDefID, req.Quantity, nil); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "ADD_FAILED", "Failed to add item to inventory", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Item added to inventory successfully",
		"character_id": characterID.String(),
		"item_def_id":  req.ItemDefID,
		"quantity":     req.Quantity,
	})
}

// RemoveItemFromInventory removes an item from a character's inventory (admin only)
func (h *ItemHandler) RemoveItemFromInventory(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	var req struct {
		ItemDefID string `json:"item_def_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	// Remove item from inventory
	if err := h.inventoryRepo.RemoveItem(c.Request.Context(), characterID, req.ItemDefID, req.Quantity); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "REMOVE_FAILED", "Failed to remove item from inventory", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Item removed from inventory successfully",
		"character_id": characterID.String(),
		"item_def_id":  req.ItemDefID,
		"quantity":     req.Quantity,
	})
}

// GetCharacterInventory returns a character's inventory (admin access)
func (h *ItemHandler) GetCharacterInventory(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	items, err := h.inventoryRepo.GetInventory(c.Request.Context(), characterID)
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
