package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const MaxMarketsPerUser = 10

// CreateAlert creates a new price alert for a user or updates existing one
func (s *Storage) CreateAlert(ctx context.Context, userID int64, marketID, marketName string, tokenID *string, thresholdPct float64) (*Alert, error) {
	// Check if alert already exists for this user and market
	existingAlert, err := s.GetAlertByUserAndMarket(ctx, userID, marketID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing alert: %w", err)
	}

	// If alert exists, update it
	if existingAlert != nil {
		query := `
			UPDATE alerts
			SET threshold_pct = $1, market_name = $2, token_id = $3, updated_at = $4, is_active = true
			WHERE user_id = $5 AND market_id = $6
			RETURNING id, user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at
		`
		alert := &Alert{}
		err = s.db.QueryRowContext(
			ctx, query,
			thresholdPct, marketName, tokenID, time.Now(), userID, marketID,
		).Scan(
			&alert.ID, &alert.UserID, &alert.MarketID, &alert.MarketName, &alert.TokenID, &alert.ThresholdPct, &alert.IsActive, &alert.CreatedAt, &alert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update alert: %w", err)
		}
		tokenIDStr := "nil"
		if tokenID != nil {
			tokenIDStr = *tokenID
		}
		s.log.Infof("Updated alert: user_id=%d, market_id=%s, token_id=%s, threshold=%.1f%%", userID, marketID, tokenIDStr, thresholdPct)
		return alert, nil
	}

	// Check if user is already tracking 10 unique markets
	trackedMarkets, err := s.GetTrackedMarketsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tracked markets: %w", err)
	}

	// If user already has 10 markets, reject
	if len(trackedMarkets) >= MaxMarketsPerUser {
		return nil, fmt.Errorf("cannot track more than %d markets", MaxMarketsPerUser)
	}

	// Create new alert
	query := `
		INSERT INTO alerts (user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at
	`

	now := time.Now()
	alert := &Alert{}
	err = s.db.QueryRowContext(
		ctx, query,
		userID, marketID, marketName, tokenID, thresholdPct, true, now, now,
	).Scan(
		&alert.ID, &alert.UserID, &alert.MarketID, &alert.MarketName, &alert.TokenID, &alert.ThresholdPct, &alert.IsActive, &alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	tokenIDStr := "nil"
	if tokenID != nil {
		tokenIDStr = *tokenID
	}
	s.log.Infof("Created alert: user_id=%d, market_id=%s, token_id=%s, threshold=%.1f%%", userID, marketID, tokenIDStr, thresholdPct)
	return alert, nil
}

// GetAlertsByUserID retrieves all alerts for a user
func (s *Storage) GetAlertsByUserID(ctx context.Context, userID int64) ([]Alert, error) {
	query := `
		SELECT id, user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at
		FROM alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		if err := rows.Scan(&alert.ID, &alert.UserID, &alert.MarketID, &alert.MarketName, &alert.TokenID, &alert.ThresholdPct, &alert.IsActive, &alert.CreatedAt, &alert.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetActiveAlerts retrieves all active alerts
func (s *Storage) GetActiveAlerts(ctx context.Context) ([]Alert, error) {
	query := `
		SELECT id, user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at
		FROM alerts
		WHERE is_active = true
		ORDER BY market_id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		if err := rows.Scan(&alert.ID, &alert.UserID, &alert.MarketID, &alert.MarketName, &alert.TokenID, &alert.ThresholdPct, &alert.IsActive, &alert.CreatedAt, &alert.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// GetTrackedMarketsByUserID returns the list of unique market IDs tracked by a user
func (s *Storage) GetTrackedMarketsByUserID(ctx context.Context, userID int64) ([]string, error) {
	query := `
		SELECT DISTINCT market_id
		FROM alerts
		WHERE user_id = $1 AND is_active = true
		ORDER BY market_id
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tracked markets: %w", err)
	}
	defer rows.Close()

	var markets []string
	for rows.Next() {
		var marketID string
		if err := rows.Scan(&marketID); err != nil {
			return nil, fmt.Errorf("failed to scan market_id: %w", err)
		}
		markets = append(markets, marketID)
	}

	return markets, nil
}

// GetUniqueTrackedMarkets returns all unique market IDs being tracked by any user
func (s *Storage) GetUniqueTrackedMarkets(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT market_id
		FROM alerts
		WHERE is_active = true
		ORDER BY market_id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get unique tracked markets: %w", err)
	}
	defer rows.Close()

	var markets []string
	for rows.Next() {
		var marketID string
		if err := rows.Scan(&marketID); err != nil {
			return nil, fmt.Errorf("failed to scan market_id: %w", err)
		}
		markets = append(markets, marketID)
	}

	return markets, nil
}

// MarketWithName holds market ID and name
type MarketWithName struct {
	MarketID   string
	MarketName string
}

// GetRandomTrackedMarkets returns random tracked markets with names (limit 5)
func (s *Storage) GetRandomTrackedMarkets(ctx context.Context, limit int) ([]MarketWithName, error) {
	query := `
		SELECT market_id, market_name
		FROM (
			SELECT DISTINCT ON (market_id)
				market_id,
				COALESCE(NULLIF(market_name, ''), 'Market #' || market_id) as market_name
			FROM alerts
			WHERE is_active = true
		) AS unique_markets
		ORDER BY RANDOM()
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get random tracked markets: %w", err)
	}
	defer rows.Close()

	var markets []MarketWithName
	for rows.Next() {
		var market MarketWithName
		if err := rows.Scan(&market.MarketID, &market.MarketName); err != nil {
			return nil, fmt.Errorf("failed to scan market: %w", err)
		}
		markets = append(markets, market)
	}

	return markets, nil
}

// DeleteAlert deletes an alert by ID
func (s *Storage) DeleteAlert(ctx context.Context, alertID, userID int64) error {
	query := `DELETE FROM alerts WHERE id = $1 AND user_id = $2`

	result, err := s.db.ExecContext(ctx, query, alertID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	s.log.Infof("Deleted alert: id=%d, user_id=%d", alertID, userID)
	return nil
}

// GetAlert retrieves an alert by ID
func (s *Storage) GetAlert(ctx context.Context, alertID int64) (*Alert, error) {
	query := `
		SELECT id, user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at
		FROM alerts
		WHERE id = $1
	`

	alert := &Alert{}
	err := s.db.QueryRowContext(ctx, query, alertID).Scan(
		&alert.ID, &alert.UserID, &alert.MarketID, &alert.MarketName, &alert.TokenID, &alert.ThresholdPct, &alert.IsActive, &alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return alert, nil
}

// GetAlertByUserAndMarket retrieves an alert by user ID and market ID
func (s *Storage) GetAlertByUserAndMarket(ctx context.Context, userID int64, marketID string) (*Alert, error) {
	query := `
		SELECT id, user_id, market_id, market_name, token_id, threshold_pct, is_active, created_at, updated_at
		FROM alerts
		WHERE user_id = $1 AND market_id = $2 AND is_active = true
	`

	alert := &Alert{}
	err := s.db.QueryRowContext(ctx, query, userID, marketID).Scan(
		&alert.ID, &alert.UserID, &alert.MarketID, &alert.MarketName, &alert.TokenID, &alert.ThresholdPct, &alert.IsActive, &alert.CreatedAt, &alert.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	return alert, nil
}

// UpdateAlertStatus updates the active status of an alert
func (s *Storage) UpdateAlertStatus(ctx context.Context, alertID int64, isActive bool) error {
	query := `UPDATE alerts SET is_active = $1, updated_at = $2 WHERE id = $3`

	_, err := s.db.ExecContext(ctx, query, isActive, time.Now(), alertID)
	if err != nil {
		return fmt.Errorf("failed to update alert status: %w", err)
	}

	return nil
}
