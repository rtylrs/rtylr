package domain

import (
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Rate       float64   `json:"rate"`
	Type       string    `json:"type"`
	IsDefault  bool      `json:"is_default"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}
type TaxConfigInput struct {
	Name      string  `json:"name"`
	Rate      float64 `json:"rate"`
	Type      string  `json:"type"`
	IsDefault bool    `json:"is_default"`
}
