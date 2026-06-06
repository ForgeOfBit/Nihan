package models

import (
	"time"

	"github.com/google/uuid"
)

// Message represents an encrypted message between two users.
type Message struct {
	ID           uuid.UUID `json:"id"`
	SenderID     uuid.UUID `json:"sender_id"`
	ReceiverID   uuid.UUID `json:"receiver_id"`
	Ciphertext   string    `json:"ciphertext"`
	Nonce        string    `json:"nonce"`
	EphemeralKey *string   `json:"ephemeral_key,omitempty"`
	MessageType  string    `json:"message_type"`
	IsRead       bool      `json:"is_read"`
	CreatedAt    time.Time `json:"created_at"`
}

// SendMessageRequest holds the data for sending a new encrypted message.
type SendMessageRequest struct {
	ReceiverID   uuid.UUID `json:"receiver_id" binding:"required"`
	Ciphertext   string    `json:"ciphertext" binding:"required"`
	Nonce        string    `json:"nonce" binding:"required"`
	EphemeralKey *string   `json:"ephemeral_key,omitempty"`
	MessageType  string    `json:"message_type" binding:"required,oneof=text image file key_exchange"`
}

// MessageHistoryQuery holds the query parameters for fetching message history.
type MessageHistoryQuery struct {
	Limit  int       `form:"limit,default=50" binding:"min=1,max=100"`
	Before *time.Time `form:"before"`
}
