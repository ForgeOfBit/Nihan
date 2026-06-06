package handlers

import (
	"net/http"

	"github.com/ForgeOfBit/Nihan/server/internal/middleware"
	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/ForgeOfBit/Nihan/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MessageHandler handles message-related HTTP endpoints.
type MessageHandler struct {
	messageService *services.MessageService
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(messageService *services.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// GetHistory handles GET /api/messages/:userId
func (h *MessageHandler) GetHistory(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)

	otherIDStr := c.Param("userId")
	otherUserID, err := uuid.Parse(otherIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var query models.MessageHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.Limit == 0 {
		query.Limit = 50
	}

	messages, err := h.messageService.GetHistory(
		c.Request.Context(),
		currentUserID,
		otherUserID,
		query.Limit,
		query.Before,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get message history"})
		return
	}

	if messages == nil {
		messages = []models.Message{}
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"count":    len(messages),
	})
}
