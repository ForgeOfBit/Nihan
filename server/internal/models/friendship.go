package models

import (
	"time"

	"github.com/google/uuid"
)

// FriendshipStatus represents the state of a friendship.
type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipBlocked  FriendshipStatus = "blocked"
)

// Friendship represents a relationship between two users.
type Friendship struct {
	ID          uuid.UUID        `json:"id"`
	RequesterID uuid.UUID        `json:"requester_id"`
	AddresseeID uuid.UUID        `json:"addressee_id"`
	Status      FriendshipStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// FriendshipWithUser includes the friendship data alongside the other user's profile.
type FriendshipWithUser struct {
	Friendship
	User PublicProfile `json:"user"`
}

// FriendRequest holds the data for sending a friend request.
type FriendRequest struct {
	// Tag is the target user's tag in "username#discriminator" format.
	Tag string `json:"tag" binding:"required"`
}
