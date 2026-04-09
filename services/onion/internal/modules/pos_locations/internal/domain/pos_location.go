package domain

import (
	"errors"
	"time"
)

var ErrLocationNotFound = errors.New("location not found")

type Location struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Phone     string    `json:"phone"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LocationInput struct {
	OwnerID  int64
	Name     string
	Address  string
	Phone    string
	IsActive bool
}

type LocationUpdateInput struct {
	ID       int64
	OwnerID  int64
	Name     string
	Address  string
	Phone    string
	IsActive bool
}
