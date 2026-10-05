package domain

import "time"

type PaymentMethod struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Provider   string
	Config     string
	IsActive   bool
	CreatedAt  time.Time
}

type PaymentMethodInput struct {
	Provider string
	Config   string
	IsActive bool
}
