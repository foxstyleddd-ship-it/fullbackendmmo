package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hp-mmo/backend/internal/auth"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

// CharacterHandler handles character endpoints
type CharacterHandler struct {
	charService *character.Service
	jwtService  *auth.JWTService
}

// NewCharacterHandler creates a new character handler
func NewCharacterHandler(charService *character.Service, jwtService *auth.JWTService) *CharacterHandler {
	return &CharacterHandler{
		charService: charService,
		jwtService:  jwtService,
	}
}

// ListCharacters handles GET /characters
func (h *CharacterHandler) ListCharacters(c *gin.Context) {
	accountIDStr, exists := c.Get("account_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Account not found", nil)
		return
	}

	accountID, err := uuid.Parse(accountIDStr.(string))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid account ID", err.Error())
		return
	}

	characters, err := h.charService.ListCharacters(c.Request.Context(), accountID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list characters", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"characters":    characters,
		"max_characters": 5,
		"slots_used":    len(characters),
	})
}

// CreateCharacter handles POST /characters
func (h *CharacterHandler) CreateCharacter(c *gin.Context) {
	accountIDStr, exists := c.Get("account_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Account not found", nil)
		return
	}

	accountID, err := uuid.Parse(accountIDStr.(string))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid account ID", err.Error())
		return
	}

	var req character.CreateCharacterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	char, err := h.charService.CreateCharacter(c.Request.Context(), accountID, req)
	if err != nil {
		if err.Error() == "maximum character limit reached (5)" {
			middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", err.Error(), nil)
			return
		}
		if err.Error() == "character name already taken" {
			middleware.ErrorResponse(c, http.StatusConflict, "CONFLICT", err.Error(), nil)
			return
		}
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create character", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusCreated, gin.H{
		"character": char,
	})
}

// GetCharacter handles GET /characters/:id
func (h *CharacterHandler) GetCharacter(c *gin.Context) {
	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid character ID", err.Error())
		return
	}

	detail, err := h.charService.GetCharacterDetail(c.Request.Context(), characterID)
	if err != nil {
		if err.Error() == "character not found" {
			middleware.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Character not found", nil)
			return
		}
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get character", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character":   detail.Character,
		"stats":       detail.Stats,
		"currencies":  detail.Currencies,
		"permissions": detail.Permissions,
	})
}

// SelectCharacter handles POST /characters/:id/select
func (h *CharacterHandler) SelectCharacter(c *gin.Context) {
	accountIDStr, exists := c.Get("account_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Account not found", nil)
		return
	}

	accountID, err := uuid.Parse(accountIDStr.(string))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid account ID", err.Error())
		return
	}

	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid character ID", err.Error())
		return
	}

	// Get character
	char, err := h.charService.GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Character not found", nil)
		return
	}

	// Verify ownership
	if char.AccountID != accountID {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Character does not belong to this account", nil)
		return
	}

	// Update last played
	if err := h.charService.SelectCharacter(c.Request.Context(), characterID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to select character", err.Error())
		return
	}

	// Get role from context
	role, _ := c.Get("role")
	sessionID, _ := c.Get("session_id")

	// Get permissions for this character
	permissions := h.charService.(*character.Service).getPermissionsForGrade(char.Grade, char.House)

	// Generate new token with character info
	accessToken, err := h.jwtService.GenerateCharacterToken(
		accountID,
		characterID,
		role.(string),
		char.House,
		char.Grade,
		sessionID.(string),
		permissions,
	)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token", err.Error())
		return
	}

	// Get zone connection info
	zoneConnection, err := h.charService.GetZoneConnection(c.Request.Context(), characterID, "wss://zone-hogwarts-0.hp-mmo.game/ws")
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get zone connection", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"access_token": accessToken,
		"character": character.CharacterInfo{
			ID:    char.ID,
			Name:  char.Name,
			House: char.House,
			Grade: char.Grade,
		},
		"zone_connection": zoneConnection,
	})
}

// DeleteCharacter handles DELETE /characters/:id
func (h *CharacterHandler) DeleteCharacter(c *gin.Context) {
	accountIDStr, exists := c.Get("account_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Account not found", nil)
		return
	}

	accountID, err := uuid.Parse(accountIDStr.(string))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid account ID", err.Error())
		return
	}

	characterID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid character ID", err.Error())
		return
	}

	// Get character
	char, err := h.charService.GetCharacter(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Character not found", nil)
		return
	}

	// Verify ownership
	if char.AccountID != accountID {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Character does not belong to this account", nil)
		return
	}

	// Delete character
	if err := h.charService.DeleteCharacter(c.Request.Context(), characterID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete character", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Character marked for deletion",
	})
}
