package domain

import (
	"errors"
	"time"
)

var ErrModifierNotFound = errors.New("modifier not found")

type Modifier struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	PriceDelta float64
	GroupName  string
	IsRequired bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type ModifierInput struct {
	Name       string
	PriceDelta float64
	GroupName  string
	IsRequired bool
}
