package domain

import (
	"errors"
	"time"
)

var ErrModifierNotFound = errors.New("modifier not found")

type Modifier struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	PriceDelta float64   `json:"price_delta"`
	GroupName  string    `json:"group_name"`
	IsRequired bool      `json:"is_required"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ModifierInput struct {
	Name       string  `json:"name"`
	PriceDelta float64 `json:"price_delta"`
	GroupName  string  `json:"group_name"`
	IsRequired bool    `json:"is_required"`
}
