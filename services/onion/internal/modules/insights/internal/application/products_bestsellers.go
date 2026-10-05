package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// BestSellerPayload is one entry in the best-sellers list.
type BestSellerPayload struct {
	ProductID         int64   `json:"product_id"`
	Name              string  `json:"name"`
	UnitsSold         float64 `json:"units_sold"`
	RevenueMinorUnits int64   `json:"revenue_minor_units"`
}

// GetBestSellers returns the 30-day top-N by units sold, cached per-limit.
func (s *Service) GetBestSellers(ctx context.Context, businessID int64, limit int32) ([]BestSellerPayload, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("products:bestsellers:%d", limit))
	if err != nil {
		return nil, err
	}
	type result struct {
		Rows []BestSellerPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetBestSellers(ctx, businessID, limit)
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]BestSellerPayload, 0, len(rows))
		for _, row := range rows {
			payloads = append(payloads, BestSellerPayload{
				ProductID:         row.ProductID,
				Name:              row.Name,
				UnitsSold:         row.UnitsSold,
				RevenueMinorUnits: row.RevenueMinorUnits,
			})
		}
		return result{Rows: payloads}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Rows, nil
}
