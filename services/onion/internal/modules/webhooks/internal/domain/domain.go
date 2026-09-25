package domain

import (
	"errors"
	"time"
)

var ErrWebhookNotFound = errors.New("webhook not found")

type Webhook struct {
	ID              int64      `json:"id"`
	OwnerID         int64      `json:"owner_id"`
	BusinessID      int64      `json:"business_id"`
	URL             string     `json:"url"`
	Events          []string   `json:"events"`
	Secret          string     `json:"-"`
	IsActive        bool       `json:"is_active"`
	LastTriggeredAt *time.Time `json:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
