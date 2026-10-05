package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// ProductKPIPayload is the JSON shape for /insights/products.
type ProductKPIPayload struct {
	TotalSKUs    int64   `json:"total_skus"`
	ActiveSKUs   int64   `json:"active_skus"`
	OutOfStock   int64   `json:"out_of_stock"`
	AvgMarginPct float64 `json:"avg_margin_pct"`
}

// GetProductKPIs returns aggregate product counters for the dashboard.
func (s *Service) GetProductKPIs(ctx context.Context, businessID int64) (ProductKPIPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "products:kpis")
	if err != nil {
		return ProductKPIPayload{}, err
	}
	out, err := cacher.Of[ProductKPIPayload]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (ProductKPIPayload, error) {
		k, rerr := s.repo.GetProductKPIs(ctx, businessID)
		if rerr != nil {
			return ProductKPIPayload{}, rerr
		}
		return ProductKPIPayload{
			TotalSKUs:    k.TotalSKUs,
			ActiveSKUs:   k.ActiveSKUs,
			OutOfStock:   k.OutOfStock,
			AvgMarginPct: k.AvgMarginPct,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return ProductKPIPayload{}, err
	}
	return out, nil
}
