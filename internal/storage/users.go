package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CreateOrGetUser creates a new user or returns existing one by Telegram ID
func (s *Storage) CreateOrGetUser(ctx context.Context, telegramID int64, username string) (*User, error) {
	// Try to get existing user first
	user, err := s.GetUserByTelegramID(ctx, telegramID)
	if err == nil {
		// User exists, update username if changed
		if user.Username != username {
			user.Username = username
			if err := s.UpdateUser(ctx, user); err != nil {
				s.log.Warnf("Failed to update username for user %d: %v", telegramID, err)
			}
		}
		return user, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// User doesn't exist, create new one
	query := `
		INSERT INTO users (telegram_id, username, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, telegram_id, username, created_at, updated_at
	`

	now := time.Now()
	user = &User{}
	err = s.db.QueryRowContext(
		ctx, query,
		telegramID, username, now, now,
	).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.log.Infof("Created new user: telegram_id=%d, username=%s", telegramID, username)
	return user, nil
}

// GetUserByTelegramID retrieves a user by their Telegram ID
func (s *Storage) GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	query := `SELECT id, telegram_id, username, created_at, updated_at FROM users WHERE telegram_id = $1`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, telegramID).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by their internal ID
func (s *Storage) GetUserByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, telegram_id, username, created_at, updated_at FROM users WHERE id = $1`

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.TelegramID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user's information
func (s *Storage) UpdateUser(ctx context.Context, user *User) error {
	query := `
		UPDATE users
		SET username = $1, updated_at = $2
		WHERE id = $3
	`

	user.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx, query, user.Username, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
