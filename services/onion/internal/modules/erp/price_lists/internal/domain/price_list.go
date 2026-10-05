package domain

import (
	"context"
	"time"
)

type PriceList struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Currency   string
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PriceListInput struct {
	Name      string
	Currency  string
	IsDefault bool
}

type Repository interface {
	ListPriceLists(ctx context.Context, ownerID int64) ([]PriceList, error)
	CreatePriceList(ctx context.Context, ownerID int64, input PriceListInput) (int64, error)
}
