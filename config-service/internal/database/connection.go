package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/company/config-service/internal/config"
	"github.com/company/config-service/internal/logger"
	_ "github.com/lib/pq"
)

// Connection represents database connection wrapper
type Connection struct {
	DB     *sql.DB
	config config.DatabaseConfig
	logger *logger.Logger
}

// New creates a new database connection
func New(cfg config.DatabaseConfig, log *logger.Logger) (*Connection, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Str("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("Successfully connected to database")

	return &Connection{
		DB:     db,
		config: cfg,
		logger: log,
	}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	c.logger.Info().Msg("Closing database connection")
	return c.DB.Close()
}

// HealthCheck performs a health check on the database
func (c *Connection) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Stats returns database connection statistics
func (c *Connection) Stats() sql.DBStats {
	return c.DB.Stats()
}
