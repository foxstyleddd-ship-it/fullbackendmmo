package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hp-mmo/backend/internal/friends"
	"github.com/hp-mmo/backend/internal/pkg/middleware"
)

type FriendsHandler struct {
	friendsService *friends.Service
}

func NewFriendsHandler(friendsService *friends.Service) *FriendsHandler {
	return &FriendsHandler{
		friendsService: friendsService,
	}
}

// GetFriends returns the character's friends list
func (h *FriendsHandler) GetFriends(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists || contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot access another character's friends", nil)
		return
	}

	friendsList, err := h.friendsService.GetFriends(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve friends", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character_id": characterID.String(),
		"friends":      friendsList,
		"count":        len(friendsList),
	})
}

// GetPendingRequests returns pending friend requests
func (h *FriendsHandler) GetPendingRequests(c *gin.Context) {
	characterIDStr := c.Param("id")
	characterID, err := uuid.Parse(characterIDStr)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid character ID", err.Error())
		return
	}

	// Verify ownership
	contextCharID, exists := c.Get("character_id")
	if !exists || contextCharID.(string) != characterIDStr {
		middleware.ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Cannot access another character's requests", nil)
		return
	}

	requests, err := h.friendsService.GetPendingRequests(c.Request.Context(), characterID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "QUERY_FAILED", "Failed to retrieve pending requests", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"character_id": characterID.String(),
		"requests":     requests,
		"count":        len(requests),
	})
}

// SendFriendRequest sends a friend request
func (h *FriendsHandler) SendFriendRequest(c *gin.Context) {
	var req struct {
		AddresseeID string `json:"addressee_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	requesterIDStr, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	requesterID, _ := uuid.Parse(requesterIDStr.(string))
	addresseeID, err := uuid.Parse(req.AddresseeID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid addressee ID", err.Error())
		return
	}

	if err := h.friendsService.SendFriendRequest(c.Request.Context(), requesterID, addresseeID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "OPERATION_FAILED", "Failed to send friend request", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Friend request sent",
		"requester_id": requesterID.String(),
		"addressee_id": addresseeID.String(),
	})
}

// AcceptFriendRequest accepts a friend request
func (h *FriendsHandler) AcceptFriendRequest(c *gin.Context) {
	var req struct {
		FriendshipID string `json:"friendship_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	friendshipID, err := uuid.Parse(req.FriendshipID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid friendship ID", err.Error())
		return
	}

	if err := h.friendsService.AcceptFriendRequest(c.Request.Context(), friendshipID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "OPERATION_FAILED", "Failed to accept friend request", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message":       "Friend request accepted",
		"friendship_id": friendshipID.String(),
	})
}

// DeclineFriendRequest declines a friend request
func (h *FriendsHandler) DeclineFriendRequest(c *gin.Context) {
	var req struct {
		FriendshipID string `json:"friendship_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	friendshipID, err := uuid.Parse(req.FriendshipID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_ID", "Invalid friendship ID", err.Error())
		return
	}

	if err := h.friendsService.DeclineFriendRequest(c.Request.Context(), friendshipID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "OPERATION_FAILED", "Failed to decline friend request", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message":       "Friend request declined",
		"friendship_id": friendshipID.String(),
	})
}

// RemoveFriend removes a friend
func (h *FriendsHandler) RemoveFriend(c *gin.Context) {
	var req struct {
		FriendID string `json:"friend_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body", err.Error())
		return
	}

	characterIDStr, exists := c.Get("character_id")
	if !exists {
		middleware.ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Character ID not found in context", nil)
		return
	}

	characterID, _ := uuid.Parse(characterIDStr.(string))
	friendID, err := uuid.Parse(req.FriendID)
	if err != nil {
		middleware.ErrorResponse(c, http.StatusBadRequest, "INVALID_CHARACTER_ID", "Invalid friend ID", err.Error())
		return
	}

	if err := h.friendsService.RemoveFriend(c.Request.Context(), characterID, friendID); err != nil {
		middleware.ErrorResponse(c, http.StatusInternalServerError, "OPERATION_FAILED", "Failed to remove friend", err.Error())
		return
	}

	middleware.SuccessResponse(c, http.StatusOK, gin.H{
		"message":   "Friend removed",
		"friend_id": friendID.String(),
	})
}
