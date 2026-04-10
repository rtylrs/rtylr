package domain

import (
	"errors"
	"time"
)

var ErrComboNotFound = errors.New("combo not found")

type Combo struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Items       string    `json:"items"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ComboInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Items       string  `json:"items"`
	IsActive    bool    `json:"isActive"`
}
