package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ForgeOfBit/Nihan/server/internal/config"
	"github.com/ForgeOfBit/Nihan/server/internal/middleware"
	"github.com/ForgeOfBit/Nihan/server/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrUsernameFull       = errors.New("all discriminators for this username are taken")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

// AuthService handles user registration, login, and JWT token management.
type AuthService struct {
	db  *pgxpool.Pool
	cfg *config.JWTConfig
	disc *DiscriminatorService
}

// NewAuthService creates a new AuthService.
func NewAuthService(db *pgxpool.Pool, cfg *config.JWTConfig, disc *DiscriminatorService) *AuthService {
	return &AuthService{db: db, cfg: cfg, disc: disc}
}

// Register creates a new user account with a randomly assigned discriminator.
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.AuthResponse, error) {
	// Hash the password.
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to hash password: %w", err)
	}

	// Check if email is already taken.
	var exists bool
	err = s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)",
		req.Email,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailTaken
	}

	// Generate a random available discriminator for this username.
	discriminator, err := s.disc.GenerateRandom(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	// Insert the new user.
	user := models.User{}
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (username, discriminator, email, password_hash, display_name, status)
		VALUES ($1, $2, $3, $4, $5, 'offline')
		RETURNING id, username, discriminator, email, password_hash, display_name,
		          avatar_url, is_premium, status, bio, created_at, updated_at
	`, req.Username, discriminator, req.Email, string(hash), req.Username,
	).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to create user: %w", err)
	}

	// Generate tokens.
	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToPublicProfile(),
	}, nil
}

// Login authenticates a user with email and password and returns JWT tokens.
func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	user := models.User{}
	err := s.db.QueryRow(ctx, `
		SELECT id, username, discriminator, email, password_hash, display_name,
		       avatar_url, is_premium, status, bio, created_at, updated_at
		FROM users WHERE email = $1
	`, req.Email,
	).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("auth: failed to query user: %w", err)
	}

	// Compare password.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update status to online.
	_, _ = s.db.Exec(ctx, "UPDATE users SET status = 'online' WHERE id = $1", user.ID)
	user.Status = "online"

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToPublicProfile(),
	}, nil
}

// RefreshTokens validates a refresh token and issues a new access/refresh token pair.
func (s *AuthService) RefreshTokens(ctx context.Context, refreshTokenStr string) (*models.AuthResponse, error) {
	claims := &middleware.JWTClaims{}
	token, err := jwt.ParseWithClaims(refreshTokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.cfg.RefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidRefreshToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Fetch the user to ensure they still exist.
	user := models.User{}
	err = s.db.QueryRow(ctx, `
		SELECT id, username, discriminator, email, password_hash, display_name,
		       avatar_url, is_premium, status, bio, created_at, updated_at
		FROM users WHERE id = $1
	`, userID,
	).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("auth: failed to query user for refresh: %w", err)
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user.ToPublicProfile(),
	}, nil
}

// GetCurrentUser returns the user for the given user ID.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	user := models.User{}
	err := s.db.QueryRow(ctx, `
		SELECT id, username, discriminator, email, password_hash, display_name,
		       avatar_url, is_premium, status, bio, created_at, updated_at
		FROM users WHERE id = $1
	`, userID,
	).Scan(
		&user.ID, &user.Username, &user.Discriminator, &user.Email,
		&user.PasswordHash, &user.DisplayName, &user.AvatarURL,
		&user.IsPremium, &user.Status, &user.Bio,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("auth: failed to query user: %w", err)
	}
	return &user, nil
}

// generateAccessToken creates a short-lived JWT access token.
func (s *AuthService) generateAccessToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := middleware.JWTClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nihan-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.AccessSecret))
	if err != nil {
		return "", fmt.Errorf("auth: failed to sign access token: %w", err)
	}
	return signed, nil
}

// generateRefreshToken creates a long-lived JWT refresh token.
func (s *AuthService) generateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := middleware.JWTClaims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nihan-server",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.RefreshSecret))
	if err != nil {
		return "", fmt.Errorf("auth: failed to sign refresh token: %w", err)
	}
	return signed, nil
}
