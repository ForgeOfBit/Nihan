package handlers

import (
	"encoding/json"
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

// KeyHandler handles key bundle HTTP endpoints.
type KeyHandler struct {
	db *pgxpool.Pool
}

// NewKeyHandler creates a new KeyHandler.
func NewKeyHandler(db *pgxpool.Pool) *KeyHandler {
	return &KeyHandler{db: db}
}

// UploadBundle handles POST /api/keys/bundle
func (h *KeyHandler) UploadBundle(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req models.UploadKeyBundleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that one_time_pre_keys is a valid JSON array.
	var keys []json.RawMessage
	if err := json.Unmarshal(req.OneTimePreKeys, &keys); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "one_time_pre_keys must be a valid JSON array"})
		return
	}

	bundle := models.KeyBundle{}
	err := h.db.QueryRow(c.Request.Context(), `
		INSERT INTO key_bundles (user_id, identity_key, signed_pre_key, signed_pre_key_sig, one_time_pre_keys)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			identity_key = EXCLUDED.identity_key,
			signed_pre_key = EXCLUDED.signed_pre_key,
			signed_pre_key_sig = EXCLUDED.signed_pre_key_sig,
			one_time_pre_keys = EXCLUDED.one_time_pre_keys
		RETURNING id, user_id, identity_key, signed_pre_key, signed_pre_key_sig,
		          one_time_pre_keys, created_at, updated_at
	`, userID, req.IdentityKey, req.SignedPreKey, req.SignedPreKeySig, req.OneTimePreKeys,
	).Scan(
		&bundle.ID, &bundle.UserID, &bundle.IdentityKey,
		&bundle.SignedPreKey, &bundle.SignedPreKeySig,
		&bundle.OneTimePreKeys, &bundle.CreatedAt, &bundle.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload key bundle"})
		return
	}

	c.JSON(http.StatusOK, bundle)
}

// GetBundle handles GET /api/keys/bundle/:userId
// It returns the key bundle and consumes (pops) one one-time pre-key.
func (h *KeyHandler) GetBundle(c *gin.Context) {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	bundle := models.KeyBundle{}
	err = h.db.QueryRow(c.Request.Context(), `
		SELECT id, user_id, identity_key, signed_pre_key, signed_pre_key_sig,
		       one_time_pre_keys, created_at, updated_at
		FROM key_bundles WHERE user_id = $1
	`, userID,
	).Scan(
		&bundle.ID, &bundle.UserID, &bundle.IdentityKey,
		&bundle.SignedPreKey, &bundle.SignedPreKeySig,
		&bundle.OneTimePreKeys, &bundle.CreatedAt, &bundle.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "key bundle not found for this user"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get key bundle"})
		}
		return
	}

	// Consume one one-time pre-key if available.
	var otpKeys []json.RawMessage
	if err := json.Unmarshal(bundle.OneTimePreKeys, &otpKeys); err == nil && len(otpKeys) > 0 {
		consumedKey := otpKeys[0]
		remainingKeys := otpKeys[1:]

		remainingJSON, err := json.Marshal(remainingKeys)
		if err == nil {
			// Update the database to remove the consumed key.
			_, _ = h.db.Exec(c.Request.Context(),
				"UPDATE key_bundles SET one_time_pre_keys = $1 WHERE id = $2",
				remainingJSON, bundle.ID,
			)
		}

		// Return only the consumed key to the requester.
		response := gin.H{
			"identity_key":      bundle.IdentityKey,
			"signed_pre_key":    bundle.SignedPreKey,
			"signed_pre_key_sig": bundle.SignedPreKeySig,
			"one_time_pre_key":  consumedKey,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// No one-time pre-keys available; return without one.
	c.JSON(http.StatusOK, gin.H{
		"identity_key":      bundle.IdentityKey,
		"signed_pre_key":    bundle.SignedPreKey,
		"signed_pre_key_sig": bundle.SignedPreKeySig,
		"one_time_pre_key":  nil,
		"warning":           fmt.Sprintf("user %s has no one-time pre-keys left", userID),
	})
}
