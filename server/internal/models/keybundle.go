package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// KeyBundle represents a user's Signal Protocol key bundle for E2EE.
type KeyBundle struct {
	ID               uuid.UUID       `json:"id"`
	UserID           uuid.UUID       `json:"user_id"`
	IdentityKey      string          `json:"identity_key"`
	SignedPreKey     string          `json:"signed_pre_key"`
	SignedPreKeySig  string          `json:"signed_pre_key_sig"`
	OneTimePreKeys   json.RawMessage `json:"one_time_pre_keys"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// UploadKeyBundleRequest holds the data for uploading a new key bundle.
type UploadKeyBundleRequest struct {
	IdentityKey    string          `json:"identity_key" binding:"required"`
	SignedPreKey   string          `json:"signed_pre_key" binding:"required"`
	SignedPreKeySig string         `json:"signed_pre_key_sig" binding:"required"`
	OneTimePreKeys json.RawMessage `json:"one_time_pre_keys" binding:"required"`
}
