package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/hp-mmo/backend/internal/auth"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
	charRepo    *character.Repository
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service, charRepo *character.Repository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		charRepo:    charRepo,
	}
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	account, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "email already registered" || err.Error() == "username already taken" {
			middleware.ErrorResponse(c, http.StatusConflict, "CONFLICT", err.Error(), nil)
			return
		}
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to register", err.Error())
		return
	}

	// Auto-login after registration
	loginReq := auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	ip := c.ClientIP()
	loginResp, err := h.authService.Login(c.Request.Context(), loginReq, ip)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Registration successful but login failed", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusCreated, loginResp)
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	ip := c.ClientIP()
	resp, err := h.authService.Login(c.Request.Context(), req, ip)
	if err != nil {
		if err.Error() == "invalid credentials" {
			middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password", nil)
			return
		}
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Login failed", err.Error())
		return
	}

	// Fetch characters
	accountID, _ := uuid.Parse(resp.AccountID.String())
	characters, err := h.charRepo.GetCharactersByAccountID(c.Request.Context(), accountID)
	if err == nil {
		resp.Characters = make([]auth.CharacterInfo, len(characters))
		for i, char := range characters {
			resp.Characters[i] = auth.CharacterInfo{
				ID:    char.ID,
				Name:  char.Name,
				House: char.House,
				Grade: char.Grade,
				Level: char.Level,
			}
			if char.LastPlayedAt.Valid {
				lastPlayed := char.LastPlayedAt.Time
				resp.Characters[i].LastPlayedAt = &lastPlayed
			}
		}
	}

	middleware.SuccessResponse(c, http.StatusOK, resp)
}

// Refresh handles POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req auth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Invalid or expired refresh token", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, resp)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, exists := c.Get("session_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Session not found", nil)
		return
	}

	err := h.authService.Logout(c.Request.Context(), sessionID.(string))
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Logout failed", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Session terminated successfully",
	})
}
