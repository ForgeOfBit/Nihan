package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations reads all .sql files from the given directory and executes
// them in alphabetical order. It uses a simple migrations tracking table to
// ensure each migration is applied at most once.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	// Create the migrations tracking table if it does not exist.
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename    TEXT PRIMARY KEY,
			applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("migrations: failed to create tracking table: %w", err)
	}

	// Discover migration files.
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("migrations: failed to read directory %s: %w", migrationsDir, err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, filename := range files {
		// Check if migration was already applied.
		var exists bool
		err := pool.QueryRow(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)",
			filename,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("migrations: failed to check status of %s: %w", filename, err)
		}
		if exists {
			log.Printf("migrations: skipping %s (already applied)", filename)
			continue
		}

		// Read and execute the migration file inside a transaction.
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("migrations: failed to read %s: %w", filename, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("migrations: failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrations: failed to execute %s: %w", filename, err)
		}

		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (filename) VALUES ($1)",
			filename,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrations: failed to record %s: %w", filename, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("migrations: failed to commit %s: %w", filename, err)
		}

		log.Printf("migrations: applied %s", filename)
	}

	return nil
}
