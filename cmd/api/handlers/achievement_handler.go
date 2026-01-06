package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/achievement"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

type AchievementHandler struct {
	achievementService *achievement.Service
}

func NewAchievementHandler(achievementService *achievement.Service) *AchievementHandler {
	return &AchievementHandler{
		achievementService: achievementService,
	}
}

// GetAllAchievements returns all available achievements
func (h *AchievementHandler) GetAllAchievements(c *gin.Context) {
	achievements, err := h.achievementService.GetAllAchievements(c.Request.Context())
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve achievements", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"achievements": achievements,
		"count":        len(achievements),
	})
}

// GetCharacterAchievements returns achievements for a specific character
func (h *AchievementHandler) GetCharacterAchievements(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership or admin
	contextCharID, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	if contextCharID.(string) != characterIDStr {
		// Check if admin
		role, _ := c.Get("role")
		if role != "admin" && role != "gm" && role != "superadmin" {
			middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot access another character's achievements", nil)
			return
		}
	}

	achievements, err := h.achievementService.GetCharacterAchievements(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve character achievements", err.Error())
		return
	}

	stats, err := h.achievementService.GetAchievementStats(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve achievement stats", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character_id":  characterID.String(),
		"achievements":  achievements,
		"count":         len(achievements),
		"stats":         stats,
	})
}
