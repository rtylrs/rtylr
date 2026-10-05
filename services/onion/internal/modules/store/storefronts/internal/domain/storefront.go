package domain

import (
	"errors"
	"time"
)

// ErrThemeNotFound is returned when a storefront references a theme that does
// not exist or belongs to a different owner.
var ErrThemeNotFound = errors.New("theme not found for this owner")

type Storefront struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	Domain      string
	LogoURL     string
	ThemeID     int64
	IsPublished bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type StorefrontInput struct {
	Name        string
	Domain      string
	LogoURL     string
	ThemeID     int64
	IsPublished bool
}
