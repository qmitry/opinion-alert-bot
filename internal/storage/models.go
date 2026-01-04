package storage

import "time"

// User represents a Telegram user
type User struct {
	ID         int64     `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// Alert represents a user-configured price alert
type Alert struct {
	ID           int64     `db:"id"`
	UserID       int64     `db:"user_id"`
	MarketID     string    `db:"market_id"`
	MarketName   string    `db:"market_name"`
	ThresholdPct float64   `db:"threshold_pct"`
	IsActive     bool      `db:"is_active"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// TokenPrice represents a YES token price snapshot
type TokenPrice struct {
	ID         int64     `db:"id"`
	TokenID    string    `db:"token_id"`
	MarketID   string    `db:"market_id"`
	Price      float64   `db:"price"`
	Side       string    `db:"side"`
	Size       float64   `db:"size"`
	RecordedAt time.Time `db:"recorded_at"`
}

// AlertHistory represents a triggered alert record
type AlertHistory struct {
	ID            int64     `db:"id"`
	AlertID       int64     `db:"alert_id"`
	MarketID      string    `db:"market_id"`
	TriggeredAt   time.Time `db:"triggered_at"`
	PreviousPrice float64   `db:"previous_price"`
	CurrentPrice  float64   `db:"current_price"`
	ChangePct     float64   `db:"change_pct"`
	MessageSent   bool      `db:"message_sent"`
}
