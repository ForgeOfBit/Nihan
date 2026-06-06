package handlers

import (
	"errors"
	"net/http"

	"github.com/ForgeOfBit/Nihan/server/internal/middleware"
	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/ForgeOfBit/Nihan/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP endpoints.
type UserHandler struct {
	userService *services.UserService
	discService *services.DiscriminatorService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(userService *services.UserService, discService *services.DiscriminatorService) *UserHandler {
	return &UserHandler{
		userService: userService,
		discService: discService,
	}
}

// Search handles GET /api/users/search?tag=username#disc
func (h *UserHandler) Search(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag query parameter is required"})
		return
	}

	user, err := h.userService.SearchByTag(c.Request.Context(), tag)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user.ToPublicProfile())
}

// GetProfile handles GET /api/users/:id
func (h *UserHandler) GetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		}
		return
	}

	c.JSON(http.StatusOK, user.ToPublicProfile())
}

// UpdateProfile handles PATCH /api/users/me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		}
		return
	}

	c.JSON(http.StatusOK, user.ToPublicProfile())
}

// ChangeDiscriminator handles PATCH /api/users/me/discriminator
func (h *UserHandler) ChangeDiscriminator(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.ChangeDiscriminatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.discService.Change(c.Request.Context(), userID, req.Discriminator); err != nil {
		switch {
		case errors.Is(err, services.ErrNotPremium):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrDiscriminatorTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to change discriminator"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "discriminator updated successfully"})
}
