package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Storage represents the database connection and operations
type Storage struct {
	db  *sqlx.DB
	log *logrus.Logger
}

// NewStorage creates a new storage instance and connects to PostgreSQL
func NewStorage(connString string, log *logrus.Logger) (*Storage, error) {
	// Connect to database with retry logic
	var db *sqlx.DB
	var err error
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sqlx.Connect("postgres", connString)
		if err == nil {
			break
		}
		log.Warnf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Info("Successfully connected to PostgreSQL")

	return &Storage{
		db:  db,
		log: log,
	}, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	s.log.Info("Closing database connection")
	return s.db.Close()
}

// RunMigrations runs database migrations
func (s *Storage) RunMigrations() error {
	s.log.Info("Running database migrations...")

	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			telegram_id BIGINT UNIQUE NOT NULL,
			username VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Alerts table
		`CREATE TABLE IF NOT EXISTS alerts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			market_id VARCHAR(255) NOT NULL,
			market_name TEXT,
			threshold_pct DECIMAL NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Add market_name column if it doesn't exist (for existing databases)
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS market_name TEXT`,

		// Add token_id column for tracking specific outcomes in multi-outcome markets
		`ALTER TABLE alerts ADD COLUMN IF NOT EXISTS token_id VARCHAR(255)`,

		// Update existing NULL market_name values to use market_id as fallback
		`UPDATE alerts SET market_name = 'Market #' || market_id WHERE market_name IS NULL`,

		// Add unique constraint for one alert per market per user
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_alerts_user_market_unique ON alerts(user_id, market_id) WHERE is_active = true`,

		// Token prices table
		`CREATE TABLE IF NOT EXISTS token_prices (
			id BIGSERIAL PRIMARY KEY,
			token_id VARCHAR(255) NOT NULL,
			market_id VARCHAR(255) NOT NULL,
			price DECIMAL NOT NULL,
			side VARCHAR(20),
			size DECIMAL,
			recorded_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		// Alert history table
		`CREATE TABLE IF NOT EXISTS alert_history (
			id BIGSERIAL PRIMARY KEY,
			alert_id BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
			market_id VARCHAR(255) NOT NULL,
			triggered_at TIMESTAMP NOT NULL DEFAULT NOW(),
			previous_price DECIMAL NOT NULL,
			current_price DECIMAL NOT NULL,
			change_pct DECIMAL NOT NULL,
			message_sent BOOLEAN NOT NULL DEFAULT false
		)`,

		// Create indexes
		`CREATE INDEX IF NOT EXISTS idx_alerts_user_id ON alerts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_user_active ON alerts(user_id, is_active) WHERE is_active = true`,
		`CREATE INDEX IF NOT EXISTS idx_token_prices_market_time ON token_prices(market_id, recorded_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_history_alert_id ON alert_history(alert_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alert_history_triggered_at ON alert_history(triggered_at DESC)`,
	}

	for i, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	s.log.Info("Database migrations completed successfully")
	return nil
}

// Ping checks if the database connection is alive
func (s *Storage) Ping() error {
	return s.db.Ping()
}

// DB returns the underlying database connection
func (s *Storage) DB() *sqlx.DB {
	return s.db
}

// WithTx executes a function within a database transaction
func (s *Storage) WithTx(fn func(*sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			s.log.Errorf("Failed to rollback transaction: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
