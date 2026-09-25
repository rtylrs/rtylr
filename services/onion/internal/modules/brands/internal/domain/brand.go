package domain

import (
	"errors"
	"time"
)

var ErrBrandNotFound = errors.New("brand not found")

type Brand struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	LogoURL      string    `json:"logo_url"`
	Description  string    `json:"description"`
	ProductCount int64     `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
type BrandInput struct {
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	Description string `json:"description"`
}
