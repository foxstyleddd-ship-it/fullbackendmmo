package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/auth"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
	"github.com/jmoiron/sqlx"
)

type AccountHandler struct {
	db       *sqlx.DB
	authRepo *auth.Repository
	charRepo *character.Repository
}

func NewAccountHandler(db *sqlx.DB, authRepo *auth.Repository, charRepo *character.Repository) *AccountHandler {
	return &AccountHandler{
		db:       db,
		authRepo: authRepo,
		charRepo: charRepo,
	}
}

// ListAccounts returns a list of all accounts with pagination
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	var accounts []auth.Account
	query := `
		SELECT id, email, email_verified, username, display_name, discord_id, status, role,
		       created_at, updated_at, last_login_at, last_login_ip, two_factor_enabled
		FROM accounts
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 100
	`

	if err := h.db.SelectContext(c.Request.Context(), &accounts, query); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to fetch accounts", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"accounts": accounts,
		"count":    len(accounts),
	})
}

// GetAccount returns details of a specific account including characters
func (h *AccountHandler) GetAccount(c *gin.Context) {
	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid account ID format", err.Error())
		return
	}

	// Get account
	account, err := h.authRepo.GetAccountByID(c.Request.Context(), accountID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found", err.Error())
		return
	}

	// Get characters for this account
	characters, err := h.charRepo.GetCharactersByAccountID(c.Request.Context(), accountID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to fetch characters", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"account":    account,
		"characters": characters,
	})
}

// UpdateAccountStatus updates the status of an account (for moderation)
func (h *AccountHandler) UpdateAccountStatus(c *gin.Context) {
	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid account ID format", err.Error())
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active whitelisted banned kicked_out"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	// Update account status
	query := `
		UPDATE accounts
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, email, username, status, updated_at
	`

	var updatedAccount struct {
		ID        uuid.UUID         `db:"id" json:"id"`
		Email     string            `db:"email" json:"email"`
		Username  string            `db:"username" json:"username"`
		Status    auth.AccountStatus `db:"status" json:"status"`
		UpdatedAt string            `db:"updated_at" json:"updated_at"`
	}

	if err := h.db.GetContext(c.Request.Context(), &updatedAccount, query, req.Status, accountID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update account status", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"account": updatedAccount,
		"message": "Account status updated successfully",
	})
}

// UpdateAccountDiscord updates the Discord ID of an account
func (h *AccountHandler) UpdateAccountDiscord(c *gin.Context) {
	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_UUID", "Invalid account ID format", err.Error())
		return
	}

	var req struct {
		DiscordID string `json:"discord_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", err.Error())
		return
	}

	query := `
		UPDATE accounts
		SET discord_id = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id, email, username, discord_id, updated_at
	`

	var updatedAccount struct {
		ID        uuid.UUID `db:"id" json:"id"`
		Email     string    `db:"email" json:"email"`
		Username  string    `db:"username" json:"username"`
		DiscordID *string   `db:"discord_id" json:"discord_id"`
		UpdatedAt string    `db:"updated_at" json:"updated_at"`
	}

	discordIDValue := &req.DiscordID
	if req.DiscordID == "" {
		discordIDValue = nil
	}

	if err := h.db.GetContext(c.Request.Context(), &updatedAccount, query, discordIDValue, accountID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update Discord ID", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"account": updatedAccount,
		"message": "Discord ID updated successfully",
	})
}

// SearchAccounts searches accounts by email or username
func (h *AccountHandler) SearchAccounts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		middleware.ErrorResponse(c, http.StatusBadRequest, "MISSING_QUERY", "Search query is required", "")
		return
	}

	var accounts []auth.Account
	sqlQuery := `
		SELECT id, email, email_verified, username, display_name, discord_id, status, role,
		       created_at, updated_at, last_login_at, last_login_ip, two_factor_enabled
		FROM accounts
		WHERE deleted_at IS NULL
		  AND (email ILIKE $1 OR username ILIKE $1 OR discord_id ILIKE $1)
		ORDER BY created_at DESC
		LIMIT 50
	`

	searchPattern := "%" + query + "%"
	if err := h.db.SelectContext(c.Request.Context(), &accounts, sqlQuery, searchPattern); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to search accounts", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"accounts": accounts,
		"count":    len(accounts),
		"query":    query,
	})
}
