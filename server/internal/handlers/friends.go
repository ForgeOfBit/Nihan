package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ForgeOfBit/Nihan/server/internal/middleware"
	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FriendHandler handles friendship-related HTTP endpoints.
type FriendHandler struct {
	db *pgxpool.Pool
}

// NewFriendHandler creates a new FriendHandler.
func NewFriendHandler(db *pgxpool.Pool) *FriendHandler {
	return &FriendHandler{db: db}
}

// SendRequest handles POST /api/friends/request
func (h *FriendHandler) SendRequest(c *gin.Context) {
	requesterID := middleware.GetUserID(c)

	var req models.FriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Look up the addressee by tag.
	var addresseeID uuid.UUID
	err := h.db.QueryRow(c.Request.Context(), `
		SELECT id FROM users
		WHERE username = split_part($1, '#', 1)
		  AND discriminator = split_part($1, '#', 2)
	`, req.Tag).Scan(&addresseeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		}
		return
	}

	if requesterID == addresseeID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot send a friend request to yourself"})
		return
	}

	// Check if a friendship already exists between these two users.
	var existingStatus string
	err = h.db.QueryRow(c.Request.Context(), `
		SELECT status FROM friendships
		WHERE LEAST(requester_id, addressee_id) = LEAST($1, $2)
		  AND GREATEST(requester_id, addressee_id) = GREATEST($1, $2)
	`, requesterID, addresseeID).Scan(&existingStatus)
	if err == nil {
		switch existingStatus {
		case "accepted":
			c.JSON(http.StatusConflict, gin.H{"error": "you are already friends with this user"})
		case "pending":
			c.JSON(http.StatusConflict, gin.H{"error": "a friend request already exists between you and this user"})
		case "blocked":
			c.JSON(http.StatusForbidden, gin.H{"error": "this friendship is blocked"})
		}
		return
	} else if !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing friendship"})
		return
	}

	// Create the friendship request.
	friendship := models.Friendship{}
	err = h.db.QueryRow(c.Request.Context(), `
		INSERT INTO friendships (requester_id, addressee_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, requester_id, addressee_id, status, created_at, updated_at
	`, requesterID, addresseeID,
	).Scan(
		&friendship.ID, &friendship.RequesterID, &friendship.AddresseeID,
		&friendship.Status, &friendship.CreatedAt, &friendship.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send friend request"})
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// AcceptRequest handles POST /api/friends/accept/:id
func (h *FriendHandler) AcceptRequest(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)

	friendshipIDStr := c.Param("id")
	friendshipID, err := uuid.Parse(friendshipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship ID"})
		return
	}

	// Only the addressee can accept a pending request.
	result, err := h.db.Exec(c.Request.Context(), `
		UPDATE friendships SET status = 'accepted'
		WHERE id = $1 AND addressee_id = $2 AND status = 'pending'
	`, friendshipID, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept friend request"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "friend request not found or you are not the addressee"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend request accepted"})
}

// ListFriends handles GET /api/friends
func (h *FriendHandler) ListFriends(c *gin.Context) {
	currentUserID := middleware.GetUserID(c)

	rows, err := h.db.Query(c.Request.Context(), `
		SELECT
			f.id, f.requester_id, f.addressee_id, f.status, f.created_at, f.updated_at,
			u.id, u.username, u.discriminator, u.display_name, u.avatar_url, u.status, u.bio, u.created_at
		FROM friendships f
		JOIN users u ON u.id = CASE
			WHEN f.requester_id = $1 THEN f.addressee_id
			ELSE f.requester_id
		END
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		ORDER BY f.updated_at DESC
	`, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list friends"})
		return
	}
	defer rows.Close()

	var friends []models.FriendshipWithUser
	for rows.Next() {
		var fw models.FriendshipWithUser
		if err := rows.Scan(
			&fw.ID, &fw.RequesterID, &fw.AddresseeID, &fw.Status,
			&fw.CreatedAt, &fw.UpdatedAt,
			&fw.User.ID, &fw.User.Username, &fw.User.Discriminator,
			&fw.User.DisplayName, &fw.User.AvatarURL, &fw.User.Status,
			&fw.User.Bio, &fw.User.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to scan friend: %v", err)})
			return
		}
		friends = append(friends, fw)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading friends list"})
		return
	}

	if friends == nil {
		friends = []models.FriendshipWithUser{}
	}

	c.JSON(http.StatusOK, gin.H{
		"friends": friends,
		"count":   len(friends),
	})
}
