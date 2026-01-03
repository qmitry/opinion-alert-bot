package storage

import (
	"context"
	"fmt"
	"time"
)

// CreateAlertHistory creates a new alert history record
func (s *Storage) CreateAlertHistory(ctx context.Context, alertID int64, marketID string, previousPrice, currentPrice, changePct float64) (*AlertHistory, error) {
	query := `
		INSERT INTO alert_history (alert_id, market_id, triggered_at, previous_price, current_price, change_pct, message_sent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, alert_id, market_id, triggered_at, previous_price, current_price, change_pct, message_sent
	`

	history := &AlertHistory{}
	err := s.db.QueryRowContext(
		ctx, query,
		alertID, marketID, time.Now(), previousPrice, currentPrice, changePct, false,
	).Scan(
		&history.ID,
		&history.AlertID,
		&history.MarketID,
		&history.TriggeredAt,
		&history.PreviousPrice,
		&history.CurrentPrice,
		&history.ChangePct,
		&history.MessageSent,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create alert history: %w", err)
	}

	s.log.Debugf("Created alert history: alert_id=%d, market=%s, change=%.2f%%", alertID, marketID, changePct)
	return history, nil
}

// MarkMessageSent marks an alert history record as having its message sent
func (s *Storage) MarkMessageSent(ctx context.Context, historyID int64) error {
	query := `UPDATE alert_history SET message_sent = true WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, historyID)
	if err != nil {
		return fmt.Errorf("failed to mark message as sent: %w", err)
	}

	return nil
}

// GetAlertHistoryByAlertID retrieves all history records for a specific alert
func (s *Storage) GetAlertHistoryByAlertID(ctx context.Context, alertID int64, limit int) ([]AlertHistory, error) {
	query := `
		SELECT id, alert_id, market_id, triggered_at, previous_price, current_price, change_pct, message_sent
		FROM alert_history
		WHERE alert_id = $1
		ORDER BY triggered_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, alertID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert history: %w", err)
	}
	defer rows.Close()

	var history []AlertHistory
	for rows.Next() {
		var h AlertHistory
		if err := rows.Scan(&h.ID, &h.AlertID, &h.MarketID, &h.TriggeredAt, &h.PreviousPrice, &h.CurrentPrice, &h.ChangePct, &h.MessageSent); err != nil {
			return nil, fmt.Errorf("failed to scan alert history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// GetRecentAlertHistory retrieves recent alert history across all alerts
func (s *Storage) GetRecentAlertHistory(ctx context.Context, since time.Time, limit int) ([]AlertHistory, error) {
	query := `
		SELECT id, alert_id, market_id, triggered_at, previous_price, current_price, change_pct, message_sent
		FROM alert_history
		WHERE triggered_at >= $1
		ORDER BY triggered_at DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, since, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent alert history: %w", err)
	}
	defer rows.Close()

	var history []AlertHistory
	for rows.Next() {
		var h AlertHistory
		if err := rows.Scan(&h.ID, &h.AlertID, &h.MarketID, &h.TriggeredAt, &h.PreviousPrice, &h.CurrentPrice, &h.ChangePct, &h.MessageSent); err != nil {
			return nil, fmt.Errorf("failed to scan alert history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}
