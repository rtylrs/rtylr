package domain

import (
	"errors"
	"time"
)

var (
	ErrBannerInvalidInput = errors.New("invalid menu banner input")
	ErrBannerNotFound     = errors.New("menu banner not found")
)

type Banner struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Title      string
	Body       string
	ImageURL   string
	LinkURL    string
	Position   int32
	IsActive   bool
	StartsAt   time.Time
	EndsAt     time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type BannerInput struct {
	Title    string
	Body     string
	ImageURL string
	LinkURL  string
	Position int32
	IsActive bool
	StartsAt time.Time
	EndsAt   time.Time
}
