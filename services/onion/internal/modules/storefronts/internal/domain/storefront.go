package domain

import "time"

type Storefront struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
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
