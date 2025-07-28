package database

import (
	"fmt"

	"github.com/company/config-service/internal/config"
	"github.com/company/config-service/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	migrate *migrate.Migrate
	logger  *logger.Logger
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(conn *Connection, cfg config.DatabaseConfig, log *logger.Logger) (*MigrationRunner, error) {
	driver, err := postgres.WithInstance(conn.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(cfg.MigrationsPath, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &MigrationRunner{
		migrate: m,
		logger:  log,
	}, nil
}

// Up runs all pending migrations
func (mr *MigrationRunner) Up() error {
	mr.logger.Info().Msg("Running database migrations...")

	if err := mr.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			mr.logger.Info().Msg("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	mr.logger.Info().Msg("Successfully applied database migrations")
	return nil
}

// Down rolls back one migration
func (mr *MigrationRunner) Down() error {
	mr.logger.Info().Msg("Rolling back one migration...")

	if err := mr.migrate.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	mr.logger.Info().Msg("Successfully rolled back one migration")
	return nil
}

// Version returns current migration version
func (mr *MigrationRunner) Version() (uint, bool, error) {
	version, dirty, err := mr.migrate.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Close closes the migration runner
func (mr *MigrationRunner) Close() error {
	sourceErr, dbErr := mr.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	if dbErr != nil {
		return dbErr
	}
	return nil
}

// ForceVersion forces the migration version (use with caution)
func (mr *MigrationRunner) ForceVersion(version int) error {
	mr.logger.Warn().
		Int("version", version).
		Msg("Forcing migration version (USE WITH CAUTION)")

	if err := mr.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}

	mr.logger.Info().
		Int("version", version).
		Msg("Successfully forced migration version")
	return nil
}
