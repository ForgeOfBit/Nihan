package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// UserService handles user profile operations.
type UserService struct {
	db *pgxpool.Pool
}

// NewUserService creates a new UserService.
func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

// GetByID retrieves a user by their UUID.
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := models.User{}
	err := s.db.QueryRow(ctx, `
		SELECT id, username, discriminator, email, password_hash, display_name,
		       avatar_url, is_premium, status, bio, created_at, updated_at
		FROM users WHERE id = $1
	`, id,
	).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user: failed to get by ID: %w", err)
	}
	return &user, nil
}

// SearchByTag searches for a user by their full tag (e.g., "alice#1234").
// The username comparison is case-insensitive (handled by CITEXT column).
func (s *UserService) SearchByTag(ctx context.Context, tag string) (*models.User, error) {
	parts := strings.SplitN(tag, "#", 2)
	if len(parts) != 2 || len(parts[1]) != 4 {
		return nil, errors.New("invalid tag format, expected username#0000")
	}

	username := parts[0]
	discriminator := parts[1]

	user := models.User{}
	err := s.db.QueryRow(ctx, `
		SELECT id, username, discriminator, email, password_hash, display_name,
		       avatar_url, is_premium, status, bio, created_at, updated_at
		FROM users WHERE username = $1 AND discriminator = $2
	`, username, discriminator,
	).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user: failed to search by tag: %w", err)
	}
	return &user, nil
}

// UpdateProfile updates the authenticated user's profile fields.
// Only non-nil fields in the request are updated.
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error) {
	// Build a dynamic UPDATE query based on which fields are provided.
	setClauses := make([]string, 0, 4)
	args := make([]interface{}, 0, 5)
	argIdx := 1

	if req.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, *req.DisplayName)
		argIdx++
	}
	if req.AvatarURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar_url = $%d", argIdx))
		args = append(args, *req.AvatarURL)
		argIdx++
	}
	if req.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, *req.Bio)
		argIdx++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
	}

	if len(setClauses) == 0 {
		// Nothing to update; return current user.
		return s.GetByID(ctx, userID)
	}

	query := fmt.Sprintf(`
		UPDATE users SET %s
		WHERE id = $%d
		RETURNING id, username, discriminator, email, password_hash, display_name,
		          avatar_url, is_premium, status, bio, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIdx)

	args = append(args, userID)

	user := models.User{}
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("user: failed to update profile: %w", err)
	}
	return &user, nil
}
