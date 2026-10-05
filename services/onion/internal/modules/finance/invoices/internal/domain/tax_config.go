package domain

import (
	"errors"
	"time"
)

var ErrTaxConfigNotFound = errors.New("tax config not found")

type TaxConfig struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Rate       float64
	Type       string
	IsDefault  bool
	IsActive   bool
	CreatedAt  time.Time
}
type TaxConfigInput struct {
	Name      string
	Rate      float64
	Type      string
	IsDefault bool
}
