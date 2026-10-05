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
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Address    string
	Phone      string
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
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
