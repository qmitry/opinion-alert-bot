package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// StoreTokenPrice stores a new token price snapshot
func (s *Storage) StoreTokenPrice(ctx context.Context, tokenID, marketID string, price float64, side string, size float64) error {
	query := `
		INSERT INTO token_prices (token_id, market_id, price, side, size, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.ExecContext(ctx, query, tokenID, marketID, price, side, size, time.Now())
	if err != nil {
		return fmt.Errorf("failed to store token price: %w", err)
	}

	return nil
}

// GetPriceOneMinuteAgo retrieves the price from approximately 1 minute ago for a given market
func (s *Storage) GetPriceOneMinuteAgo(ctx context.Context, marketID string) (*TokenPrice, error) {
	// Get the price from 1 minute ago (allowing a small window)
	query := `
		SELECT id, token_id, market_id, price, side, size, recorded_at
		FROM token_prices
		WHERE market_id = $1
		  AND recorded_at >= NOW() - INTERVAL '70 seconds'
		  AND recorded_at <= NOW() - INTERVAL '50 seconds'
		ORDER BY recorded_at ASC
		LIMIT 1
	`

	tokenPrice := &TokenPrice{}
	err := s.db.QueryRowContext(ctx, query, marketID).Scan(
		&tokenPrice.ID,
		&tokenPrice.TokenID,
		&tokenPrice.MarketID,
		&tokenPrice.Price,
		&tokenPrice.Side,
		&tokenPrice.Size,
		&tokenPrice.RecordedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get price from 1 minute ago: %w", err)
	}

	return tokenPrice, nil
}

// GetLatestPrice retrieves the most recent price for a given market
func (s *Storage) GetLatestPrice(ctx context.Context, marketID string) (*TokenPrice, error) {
	query := `
		SELECT id, token_id, market_id, price, side, size, recorded_at
		FROM token_prices
		WHERE market_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`

	tokenPrice := &TokenPrice{}
	err := s.db.QueryRowContext(ctx, query, marketID).Scan(
		&tokenPrice.ID,
		&tokenPrice.TokenID,
		&tokenPrice.MarketID,
		&tokenPrice.Price,
		&tokenPrice.Side,
		&tokenPrice.Size,
		&tokenPrice.RecordedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("failed to get latest price: %w", err)
	}

	return tokenPrice, nil
}

// CleanupOldPrices deletes token prices older than the specified duration
func (s *Storage) CleanupOldPrices(ctx context.Context, olderThan time.Duration) error {
	query := `DELETE FROM token_prices WHERE recorded_at < NOW() - $1::interval`

	result, err := s.db.ExecContext(ctx, query, olderThan.String())
	if err != nil {
		return fmt.Errorf("failed to cleanup old prices: %w", err)
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		s.log.Warnf("Failed to get rows affected during cleanup: %v", err)
		return nil
	}

	if rowsDeleted > 0 {
		s.log.Debugf("Cleaned up %d old price records", rowsDeleted)
	}

	return nil
}

// GetPriceHistory retrieves price history for a market within a time range
func (s *Storage) GetPriceHistory(ctx context.Context, marketID string, since time.Time) ([]TokenPrice, error) {
	query := `
		SELECT id, token_id, market_id, price, side, size, recorded_at
		FROM token_prices
		WHERE market_id = $1 AND recorded_at >= $2
		ORDER BY recorded_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, marketID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}
	defer rows.Close()

	var prices []TokenPrice
	for rows.Next() {
		var price TokenPrice
		if err := rows.Scan(&price.ID, &price.TokenID, &price.MarketID, &price.Price, &price.Side, &price.Size, &price.RecordedAt); err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}
		prices = append(prices, price)
	}

	return prices, nil
}
