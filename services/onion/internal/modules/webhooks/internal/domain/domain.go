package domain

import (
	"errors"
	"time"
)

var (
	ErrWebhookNotFound               = errors.New("webhook not found")
	ErrDeliveryNotClaimed            = errors.New("webhook delivery already claimed")
	ErrDeliveryRegistrationNotFound  = errors.New("webhook registration not found")
	ErrNotificationRecipientNotFound = errors.New("webhook notification recipient not found")
)

type Webhook struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	URL             string
	Events          []string
	Secret          string
	IsActive        bool
	LastTriggeredAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
