package domain

import (
	"context"
	"time"
)

type PriceList struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Currency   string    `json:"currency"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PriceListInput struct {
	Name      string `json:"name"`
	Currency  string `json:"currency"`
	IsDefault bool   `json:"is_default"`
}

type Repository interface {
	ListPriceLists(ctx context.Context, ownerID int64) ([]PriceList, error)
	CreatePriceList(ctx context.Context, ownerID int64, input PriceListInput) (int64, error)
}
