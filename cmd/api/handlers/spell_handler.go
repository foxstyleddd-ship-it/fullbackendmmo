package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
	"github.com/hp-mmo/backend/internal/spell"
)

// SpellHandler handles spell-related endpoints
type SpellHandler struct {
	spellService *spell.Service
}

// NewSpellHandler creates a new spell handler
func NewSpellHandler(spellService *spell.Service) *SpellHandler {
	return &SpellHandler{
		spellService: spellService,
	}
}

// GetAllSpells retrieves all available spells
func (h *SpellHandler) GetAllSpells(c *gin.Context) {
	spells, err := h.spellService.GetAllSpells(c.Request.Context())
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve spells", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"spells": spells,
		"count":  len(spells),
	})
}

// GetCharacterSpells retrieves spells learned by a character
func (h *SpellHandler) GetCharacterSpells(c *gin.Context) {
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
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot access another character's spells", nil)
		return
	}

	spells, err := h.spellService.GetCharacterSpells(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve character spells", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character_id": characterID.String(),
		"spells":       spells,
		"count":        len(spells),
	})
}
