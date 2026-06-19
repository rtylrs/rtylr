package domain

import (
	"errors"
	"time"
)

// ErrThemeNotFound is returned when a storefront references a theme that does
// not exist or belongs to a different owner.
var ErrThemeNotFound = errors.New("theme not found for this owner")

type Storefront struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	LogoURL     string    `json:"logo_url"`
	ThemeID     int64     `json:"theme_id"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StorefrontInput struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	LogoURL     string `json:"logo_url"`
	ThemeID     int64  `json:"theme_id"`
	IsPublished bool   `json:"is_published"`
}
