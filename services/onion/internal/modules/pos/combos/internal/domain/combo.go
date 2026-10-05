package domain

import (
	"errors"
	"time"
)

var ErrComboNotFound = errors.New("combo not found")

type Combo struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Description string
	Price       float64
	Items       string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ComboInput struct {
	Name        string
	Description string
	Price       float64
	Items       string
	IsActive    bool
}
