package domain

import (
	"errors"
	"time"
)

var (
	ErrLocationNotFound      = errors.New("location not found")
	ErrLocationLimitExceeded = errors.New("subscription location limit exceeded")
)

type Location struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Phone      string    `json:"phone"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type LocationInput struct {
	OwnerID    int64
	BusinessID int64
	Name       string
	Address    string
	Phone      string
	IsActive   bool
}

type LocationUpdateInput struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Address    string
	Phone      string
	IsActive   *bool
}
