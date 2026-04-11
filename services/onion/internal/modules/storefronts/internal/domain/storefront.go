package domain

import "time"

type Storefront struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"ownerId"`
	Name        string    `json:"name"`
	Domain      string    `json:"domain"`
	LogoURL     string    `json:"logoUrl"`
	ThemeID     int64     `json:"themeId"`
	IsPublished bool      `json:"isPublished"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type StorefrontInput struct {
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	LogoURL     string `json:"logoUrl"`
	ThemeID     int64  `json:"themeId"`
	IsPublished bool   `json:"isPublished"`
}
