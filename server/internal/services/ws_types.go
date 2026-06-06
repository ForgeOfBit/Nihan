package services

import "github.com/google/uuid"

// WSSendMessage is the payload structure for a "message.send" WebSocket event.
type WSSendMessage struct {
	ReceiverID   uuid.UUID `json:"receiver_id"`
	Ciphertext   string    `json:"ciphertext"`
	Nonce        string    `json:"nonce"`
	EphemeralKey *string   `json:"ephemeral_key,omitempty"`
	MessageType  string    `json:"message_type"`
}
