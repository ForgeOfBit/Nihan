package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/google/uuid"
)

var (
	ErrDiscriminatorTaken    = errors.New("discriminator is already taken for this username")
	ErrNotPremium            = errors.New("only premium users can change their discriminator")
	ErrNoAvailableDiscriminator = errors.New("all discriminators for this username are taken")
)

// DiscriminatorService handles discriminator generation and changes.
type DiscriminatorService struct {
	db *pgxpool.Pool
}

// NewDiscriminatorService creates a new DiscriminatorService.
func NewDiscriminatorService(db *pgxpool.Pool) *DiscriminatorService {
	return &DiscriminatorService{db: db}
}

// GenerateRandom generates a random available 4-digit discriminator for the
// given username. It queries PostgreSQL for all taken discriminators and
// picks a random one from the remaining pool.
func (s *DiscriminatorService) GenerateRandom(ctx context.Context, username string) (string, error) {
	// Find all discriminators already taken for this username.
	rows, err := s.db.Query(ctx,
		"SELECT discriminator FROM users WHERE username = $1",
		username,
	)
	if err != nil {
		return "", fmt.Errorf("discriminator: failed to query taken: %w", err)
	}
	defer rows.Close()

	taken := make(map[string]struct{})
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return "", fmt.Errorf("discriminator: failed to scan: %w", err)
		}
		taken[d] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("discriminator: row error: %w", err)
	}

	// Build list of available discriminators (0001 - 9999).
	available := make([]string, 0, 9999-len(taken))
	for i := 1; i <= 9999; i++ {
		d := fmt.Sprintf("%04d", i)
		if _, ok := taken[d]; !ok {
			available = append(available, d)
		}
	}

	if len(available) == 0 {
		return "", ErrNoAvailableDiscriminator
	}

	// Pick a random discriminator from the available pool.
	return available[rand.Intn(len(available))], nil
}

// Change allows a premium user to change their discriminator to a specific value.
func (s *DiscriminatorService) Change(ctx context.Context, userID uuid.UUID, newDiscriminator string) error {
	// Verify user is premium.
	var isPremium bool
	var username string
	err := s.db.QueryRow(ctx,
		"SELECT username, is_premium FROM users WHERE id = $1",
		userID,
	).Scan(&username, &isPremium)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return fmt.Errorf("discriminator: failed to get user: %w", err)
	}

	if !isPremium {
		return ErrNotPremium
	}

	// Check if the desired discriminator is available for this username.
	var exists bool
	err = s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND discriminator = $2 AND id <> $3)",
		username, newDiscriminator, userID,
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("discriminator: failed to check availability: %w", err)
	}
	if exists {
		return ErrDiscriminatorTaken
	}

	// Update the discriminator.
	_, err = s.db.Exec(ctx,
		"UPDATE users SET discriminator = $1 WHERE id = $2",
		newDiscriminator, userID,
	)
	if err != nil {
		return fmt.Errorf("discriminator: failed to update: %w", err)
	}

	return nil
}
