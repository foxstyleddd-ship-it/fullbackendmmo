package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
	"github.com/jmoiron/sqlx"
)

type LeaderboardHandler struct {
	db *sqlx.DB
}

func NewLeaderboardHandler(db *sqlx.DB) *LeaderboardHandler {
	return &LeaderboardHandler{db: db}
}

type LeaderboardEntry struct {
	ID               string `db:"id" json:"id"`
	Name             string `db:"name" json:"name"`
	House            string `db:"house" json:"house"`
	Grade            int    `db:"grade" json:"grade"`
	Level            int    `db:"level" json:"level"`
	Experience       int64  `db:"experience" json:"experience"`
	HousePoints      int64  `db:"house_points" json:"house_points"`
	AchievementCount int    `db:"achievement_count" json:"achievement_count"`
	TotalPlaytime    int64  `db:"total_playtime_seconds" json:"total_playtime_seconds"`
}

// GetLeaderboard returns the overall leaderboard
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err == nil && limit > 0 && limit <= 500 {
			// Valid limit
		} else {
			limit = 100
		}
	}

	var entries []LeaderboardEntry
	query := `SELECT * FROM leaderboard_overall LIMIT $1`

	if err := h.db.SelectContext(c.Request.Context(), &entries, query, limit); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve leaderboard", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"leaderboard": entries,
		"count":       len(entries),
	})
}

// GetHouseLeaderboard returns house-specific leaderboard
func (h *LeaderboardHandler) GetHouseLeaderboard(c *gin.Context) {
	house := c.Param("house")
	if house == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_HOUSE", "House parameter required", nil)
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err == nil && limit > 0 && limit <= 500 {
			// Valid limit
		} else {
			limit = 50
		}
	}

	var entries []LeaderboardEntry
	query := `SELECT * FROM leaderboard_overall WHERE house = $1 ORDER BY house_points DESC LIMIT $2`

	if err := h.db.SelectContext(c.Request.Context(), &entries, query, house, limit); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve house leaderboard", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"house":       house,
		"leaderboard": entries,
		"count":       len(entries),
	})
}
