package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MessageService handles encrypted message storage and retrieval.
type MessageService struct {
	db *pgxpool.Pool
}

// NewMessageService creates a new MessageService.
func NewMessageService(db *pgxpool.Pool) *MessageService {
	return &MessageService{db: db}
}

// Send stores a new encrypted message in the database.
func (s *MessageService) Send(ctx context.Context, senderID uuid.UUID, req models.SendMessageRequest) (*models.Message, error) {
	msg := models.Message{}
	err := s.db.QueryRow(ctx, `
		INSERT INTO messages (sender_id, receiver_id, ciphertext, nonce, ephemeral_key, message_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, sender_id, receiver_id, ciphertext, nonce, ephemeral_key,
		          message_type, is_read, created_at
	`, senderID, req.ReceiverID, req.Ciphertext, req.Nonce, req.EphemeralKey, req.MessageType,
	).Scan(
		&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Ciphertext,
		&msg.Nonce, &msg.EphemeralKey, &msg.MessageType, &msg.IsRead, &msg.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("message: failed to send: %w", err)
	}
	return &msg, nil
}

// GetHistory retrieves the encrypted message history between two users,
// paginated by a cursor timestamp and limited to a maximum count.
func (s *MessageService) GetHistory(ctx context.Context, userID, otherUserID uuid.UUID, limit int, before *time.Time) ([]models.Message, error) {
	var query string
	var args []interface{}

	if before != nil {
		query = `
			SELECT id, sender_id, receiver_id, ciphertext, nonce, ephemeral_key,
			       message_type, is_read, created_at
			FROM messages
			WHERE ((sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1))
			  AND created_at < $3
			ORDER BY created_at DESC
			LIMIT $4
		`
		args = []interface{}{userID, otherUserID, *before, limit}
	} else {
		query = `
			SELECT id, sender_id, receiver_id, ciphertext, nonce, ephemeral_key,
			       message_type, is_read, created_at
			FROM messages
			WHERE (sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1)
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{userID, otherUserID, limit}
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("message: failed to get history: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Ciphertext,
			&msg.Nonce, &msg.EphemeralKey, &msg.MessageType, &msg.IsRead, &msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("message: failed to scan row: %w", err)
		}
		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("message: row iteration error: %w", err)
	}

	return messages, nil
}

// MarkAsRead marks all messages from the sender to the receiver as read.
func (s *MessageService) MarkAsRead(ctx context.Context, receiverID, senderID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE messages SET is_read = TRUE
		WHERE sender_id = $1 AND receiver_id = $2 AND is_read = FALSE
	`, senderID, receiverID)
	if err != nil {
		return fmt.Errorf("message: failed to mark as read: %w", err)
	}
	return nil
}
