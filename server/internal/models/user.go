package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the Nihan messaging system.
type User struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"` // Never serialize the password hash
	DisplayName   *string   `json:"display_name,omitempty"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	IsPremium     bool      `json:"is_premium"`
	Status        string    `json:"status"`
	Bio           *string   `json:"bio,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserTag returns the full user tag in the format "username#discriminator".
func (u *User) UserTag() string {
	return u.Username + "#" + u.Discriminator
}

// PublicProfile is the public-facing representation of a user,
// excluding sensitive fields like email and password hash.
type PublicProfile struct {
	ID            uuid.UUID `json:"id"`
	Username      string    `json:"username"`
	Discriminator string    `json:"discriminator"`
	DisplayName   *string   `json:"display_name,omitempty"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	Status        string    `json:"status"`
	Bio           *string   `json:"bio,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToPublicProfile converts a User to its public-facing representation.
func (u *User) ToPublicProfile() PublicProfile {
	return PublicProfile{
		ID:            u.ID,
		Username:      u.Username,
		Discriminator: u.Discriminator,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		Status:        u.Status,
		Bio:           u.Bio,
		CreatedAt:     u.CreatedAt,
	}
}

// RegisterRequest holds the data needed to register a new user.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// LoginRequest holds the data needed to log in.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest holds the data for updating a user profile.
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,max=64"`
	AvatarURL   *string `json:"avatar_url" binding:"omitempty,url"`
	Bio         *string `json:"bio" binding:"omitempty,max=256"`
	Status      *string `json:"status" binding:"omitempty,oneof=online offline idle dnd invisible"`
}

// ChangeDiscriminatorRequest holds the data for changing a discriminator.
type ChangeDiscriminatorRequest struct {
	Discriminator string `json:"discriminator" binding:"required,len=4,numeric"`
}

// AuthResponse is returned after successful login or token refresh.
type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	User         PublicProfile `json:"user"`
}

// RefreshRequest holds a refresh token for obtaining new tokens.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
