package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/audit"
	"github.com/hp-mmo/backend/internal/character"
	"github.com/hp-mmo/backend/internal/inventory"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
	"github.com/hp-mmo/backend/internal/spell"
)

// AdminHandler handles admin and GM commands
type AdminHandler struct {
	charService      *character.Service
	charRepo         *character.Repository
	auditService     *audit.Service
	inventoryService *inventory.Service
	spellService     *spell.Service
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(charService *character.Service, charRepo *character.Repository, auditService *audit.Service, inventoryService *inventory.Service, spellService *spell.Service) *AdminHandler {
	return &AdminHandler{
		charService:      charService,
		charRepo:         charRepo,
		auditService:     auditService,
		inventoryService: inventoryService,
		spellService:     spellService,
	}
}

// ExecuteCommand handles GM command execution
type ExecuteCommandRequest struct {
	Command    string                 `json:"command" binding:"required"`
	TargetID   string                 `json:"target_id" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ExecuteCommandResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ExecuteCommand executes a GM command
func (h *AdminHandler) ExecuteCommand(c *gin.Context) {
	var req ExecuteCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format", err.Error())
		return
	}

	// Get admin account ID from context
	adminID, exists := c.Get("account_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin account ID not found", nil)
		return
	}

	accountID, ok := adminID.(string)
	if !ok {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid account ID", nil)
		return
	}

	adminUUID, err := uuid.Parse(accountID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADMIN_ID", "Invalid admin ID", err.Error())
		return
	}

	// Parse target ID
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_TARGET_ID", "Invalid target character ID", err.Error())
		return
	}

	// Get client info for audit log
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Execute command based on type
	var response *ExecuteCommandResponse
	switch req.Command {
	case "setgrade":
		response, err = h.executeSetGrade(c, adminUUID, targetID, req.Parameters, ip, userAgent)
	case "teleport":
		response, err = h.executeTeleport(c, adminUUID, targetID, req.Parameters, ip, userAgent)
	case "grantitem":
		response, err = h.executeGrantItem(c, adminUUID, targetID, req.Parameters, ip, userAgent)
	case "grantspell":
		response, err = h.executeGrantSpell(c, adminUUID, targetID, req.Parameters, ip, userAgent)
	case "removespell":
		response, err = h.executeRemoveSpell(c, adminUUID, targetID, req.Parameters, ip, userAgent)
	case "addhousepoints":
		// For house points, target_id is the house name
		response, err = h.executeAddHousePoints(c, adminUUID, req.TargetID, req.Parameters, ip, userAgent)
	case "removehousepoints":
		// For house points, target_id is the house name
		response, err = h.executeRemoveHousePoints(c, adminUUID, req.TargetID, req.Parameters, ip, userAgent)
	default:
		middleware.ErrorResponse(c, http.StatusBadRequest, "UNKNOWN_COMMAND", fmt.Sprintf("Unknown command: %s", req.Command), nil)
		return
	}

	if err != nil {
		// Log failed attempt
		errorMsg := err.Error()
		_ = h.auditService.LogAdminAction(c.Request.Context(), adminUUID, targetID, req.Command, fmt.Sprintf("Failed: %s", errorMsg), false, &errorMsg, ip, userAgent)

		middleware.ErrorResponse(c, http.StatusInternalServerError, "COMMAND_FAILED", err.Error(), nil)
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, response)
}

// executeSetGrade handles the setgrade command
func (h *AdminHandler) executeSetGrade(c *gin.Context, adminID, targetID uuid.UUID, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	// Parse new grade
	newGradeFloat, ok := params["grade"].(float64)
	if !ok {
		return nil, fmt.Errorf("grade parameter is required and must be a number")
	}
	newGrade := int(newGradeFloat)

	// Get current character info
	char, err := h.charRepo.GetCharacterByID(c.Request.Context(), targetID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}
	oldGrade := char.Grade

	// Update grade
	if err := h.charService.UpdateGrade(c.Request.Context(), targetID, newGrade); err != nil {
		return nil, fmt.Errorf("failed to update grade: %w", err)
	}

	// Log audit
	if err := h.auditService.LogGradeChange(c.Request.Context(), adminID, targetID, oldGrade, newGrade, ip, userAgent); err != nil {
		// Log error but don't fail the command
		fmt.Printf("Failed to log grade change audit: %v\n", err)
	}

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Grade updated from %d to %d for character %s", oldGrade, newGrade, char.Name),
		Data: gin.H{
			"character_id": targetID.String(),
			"old_grade":    oldGrade,
			"new_grade":    newGrade,
		},
	}, nil
}

// executeTeleport handles the teleport command
func (h *AdminHandler) executeTeleport(c *gin.Context, adminID, targetID uuid.UUID, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	// Parse zone ID
	zoneID, ok := params["zone_id"].(string)
	if !ok || zoneID == "" {
		return nil, fmt.Errorf("zone_id parameter is required")
	}

	// Parse coordinates (optional, defaults to zone spawn)
	x := float32(0.0)
	y := float32(0.0)
	z := float32(0.0)

	if xVal, ok := params["x"].(float64); ok {
		x = float32(xVal)
	}
	if yVal, ok := params["y"].(float64); ok {
		y = float32(yVal)
	}
	if zVal, ok := params["z"].(float64); ok {
		z = float32(zVal)
	}

	// Get current character info
	char, err := h.charRepo.GetCharacterByID(c.Request.Context(), targetID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}
	oldZone := char.ZoneID

	// Teleport character
	if err := h.charService.TeleportCharacter(c.Request.Context(), targetID, zoneID, x, y, z); err != nil {
		return nil, fmt.Errorf("failed to teleport: %w", err)
	}

	// Log audit
	if err := h.auditService.LogTeleport(c.Request.Context(), adminID, targetID, oldZone, zoneID, ip, userAgent); err != nil {
		fmt.Printf("Failed to log teleport audit: %v\n", err)
	}

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Teleported character %s from %s to %s", char.Name, oldZone, zoneID),
		Data: gin.H{
			"character_id": targetID.String(),
			"old_zone":     oldZone,
			"new_zone":     zoneID,
			"position": gin.H{
				"x": x,
				"y": y,
				"z": z,
			},
		},
	}, nil
}

// executeGrantItem handles the grantitem command
func (h *AdminHandler) executeGrantItem(c *gin.Context, adminID, targetID uuid.UUID, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	// Parse item definition ID (string, not UUID)
	itemDefID, ok := params["item_def_id"].(string)
	if !ok || itemDefID == "" {
		return nil, fmt.Errorf("item_def_id parameter is required (use item definition ID like 'elder_wand')")
	}

	// Parse quantity
	quantityFloat, ok := params["quantity"].(float64)
	if !ok {
		quantityFloat = 1
	}
	quantity := int(quantityFloat)

	// Get character info
	char, err := h.charRepo.GetCharacterByID(c.Request.Context(), targetID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Grant item using inventory service
	err = h.inventoryService.GrantItem(c.Request.Context(), targetID, itemDefID, quantity, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to grant item: %w", err)
	}

	// Log audit (the inventory service already creates a transaction ledger)
	// We use UUID zero for the item_id in audit since it's an item_def_id
	if err := h.auditService.LogItemGrant(c.Request.Context(), adminID, targetID, uuid.Nil, quantity, ip, userAgent); err != nil {
		fmt.Printf("Failed to log item grant audit: %v\n", err)
	}

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Granted %d x %s to character %s", quantity, itemDefID, char.Name),
		Data: gin.H{
			"character_id": targetID.String(),
			"item_def_id":  itemDefID,
			"quantity":     quantity,
		},
	}, nil
}

// executeGrantSpell handles the grantspell command
func (h *AdminHandler) executeGrantSpell(c *gin.Context, adminID, targetID uuid.UUID, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	// Parse spell ID
	spellID, ok := params["spell_id"].(string)
	if !ok || spellID == "" {
		return nil, fmt.Errorf("spell_id parameter is required (e.g., 'lumos', 'expelliarmus')")
	}

	// Get character info
	char, err := h.charRepo.GetCharacterByID(c.Request.Context(), targetID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Grant spell using spell service (bypasses requirements for GM)
	err = h.spellService.GrantSpellBypass(c.Request.Context(), targetID, spellID)
	if err != nil {
		return nil, fmt.Errorf("failed to grant spell: %w", err)
	}

	// Log admin action
	details := fmt.Sprintf("Granted spell %s to character %s", spellID, char.Name)
	_ = h.auditService.LogAdminAction(c.Request.Context(), adminID, targetID, "grant_spell", details, true, nil, ip, userAgent)

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Granted spell '%s' to character %s", spellID, char.Name),
		Data: gin.H{
			"character_id": targetID.String(),
			"spell_id":     spellID,
		},
	}, nil
}

// executeRemoveSpell handles the removespell command
func (h *AdminHandler) executeRemoveSpell(c *gin.Context, adminID, targetID uuid.UUID, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	// Parse spell ID
	spellID, ok := params["spell_id"].(string)
	if !ok || spellID == "" {
		return nil, fmt.Errorf("spell_id parameter is required")
	}

	// Get character info
	char, err := h.charRepo.GetCharacterByID(c.Request.Context(), targetID)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}

	// Remove spell
	err = h.spellService.RemoveSpell(c.Request.Context(), targetID, spellID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove spell: %w", err)
	}

	// Log admin action
	details := fmt.Sprintf("Removed spell %s from character %s", spellID, char.Name)
	_ = h.auditService.LogAdminAction(c.Request.Context(), adminID, targetID, "remove_spell", details, true, nil, ip, userAgent)

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Removed spell '%s' from character %s", spellID, char.Name),
		Data: gin.H{
			"character_id": targetID.String(),
			"spell_id":     spellID,
		},
	}, nil
}

// GetCharacterAuditLogs retrieves audit logs for a character
func (h *AdminHandler) GetCharacterAuditLogs(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
	}

	logs, err := h.auditService.GetCharacterAuditLogs(c.Request.Context(), characterID, limit)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve audit logs", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character_id": characterID.String(),
		"logs":         logs,
		"count":        len(logs),
	})
}

// GetAdminAuditLogs retrieves audit logs for an admin user
func (h *AdminHandler) GetAdminAuditLogs(c *gin.Context) {
	adminID, exists := c.Get("account_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Admin account ID not found", nil)
		return
	}

	accountID, ok := adminID.(string)
	if !ok {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid account ID", nil)
		return
	}

	adminUUID, err := uuid.Parse(accountID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_ADMIN_ID", "Invalid admin ID", err.Error())
		return
	}

	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		fmt.Sscanf(limitParam, "%d", &limit)
	}

	logs, err := h.auditService.GetAdminAuditLogs(c.Request.Context(), adminUUID, limit)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve audit logs", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"admin_id": adminUUID.String(),
		"logs":     logs,
		"count":    len(logs),
	})
}

// executeAddHousePoints adds house points to ALL members of a house
func (h *AdminHandler) executeAddHousePoints(c *gin.Context, adminID uuid.UUID, house string, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	points, ok := params["points"].(float64)
	if !ok || points <= 0 {
		return nil, fmt.Errorf("points parameter is required and must be positive")
	}

	pointsInt := int(points)

	// Get all characters in the house
	query := `
		UPDATE character_currencies cc
		SET amount = amount + $1, updated_at = NOW()
		FROM characters c
		WHERE cc.character_id = c.id
		AND c.house = $2
		AND cc.currency_id = 'house_points'
		AND c.deleted_at IS NULL`

	result, err := h.charRepo.GetDB().ExecContext(c.Request.Context(), query, pointsInt, house)
	if err != nil {
		return nil, fmt.Errorf("failed to add house points: %w", err)
	}

	affected, _ := result.RowsAffected()

	details := fmt.Sprintf("Added %d house points to %s (affected %d characters)", pointsInt, house, affected)
	_ = h.auditService.LogAdminAction(c.Request.Context(), adminID, uuid.Nil, "add_house_points", details, true, nil, ip, userAgent)

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Added %d points to house %s", pointsInt, house),
		Data: gin.H{
			"house":              house,
			"points_added":       pointsInt,
			"characters_updated": affected,
		},
	}, nil
}

// executeRemoveHousePoints removes house points from ALL members of a house
func (h *AdminHandler) executeRemoveHousePoints(c *gin.Context, adminID uuid.UUID, house string, params map[string]interface{}, ip, userAgent string) (*ExecuteCommandResponse, error) {
	points, ok := params["points"].(float64)
	if !ok || points <= 0 {
		return nil, fmt.Errorf("points parameter is required and must be positive")
	}

	pointsInt := int(points)

	// Remove points (but don't go below 0)
	query := `
		UPDATE character_currencies cc
		SET amount = GREATEST(amount - $1, 0), updated_at = NOW()
		FROM characters c
		WHERE cc.character_id = c.id
		AND c.house = $2
		AND cc.currency_id = 'house_points'
		AND c.deleted_at IS NULL`

	result, err := h.charRepo.GetDB().ExecContext(c.Request.Context(), query, pointsInt, house)
	if err != nil {
		return nil, fmt.Errorf("failed to remove house points: %w", err)
	}

	affected, _ := result.RowsAffected()

	details := fmt.Sprintf("Removed %d house points from %s (affected %d characters)", pointsInt, house, affected)
	_ = h.auditService.LogAdminAction(c.Request.Context(), adminID, uuid.Nil, "remove_house_points", details, true, nil, ip, userAgent)

	return &ExecuteCommandResponse{
		Success: true,
		Message: fmt.Sprintf("Removed %d points from house %s", pointsInt, house),
		Data: gin.H{
			"house":              house,
			"points_removed":     pointsInt,
			"characters_updated": affected,
		},
	}, nil
}
