package domain

import (
	"errors"
	"time"
)

var ErrModifierNotFound = errors.New("modifier not found")

type Modifier struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"ownerId"`
	Name       string    `json:"name"`
	PriceDelta float64   `json:"priceDelta"`
	GroupName  string    `json:"groupName"`
	IsRequired bool      `json:"isRequired"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type ModifierInput struct {
	Name       string  `json:"name"`
	PriceDelta float64 `json:"priceDelta"`
	GroupName  string  `json:"groupName"`
	IsRequired bool    `json:"isRequired"`
}
