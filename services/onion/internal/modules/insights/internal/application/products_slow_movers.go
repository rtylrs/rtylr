package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// SlowMoverPayload is one slow-moving product entry.
type SlowMoverPayload struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	StockQuantity float64 `json:"stock_quantity"`
}

// GetSlowMovers returns products with zero sales in the trailing window.
func (s *Service) GetSlowMovers(ctx context.Context, businessID int64, days int32) ([]SlowMoverPayload, error) {
	if days <= 0 {
		days = 60
	}
	if days > 365 {
		days = 365
	}
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("products:slow:%d", days))
	if err != nil {
		return nil, err
	}
	type result struct {
		Rows []SlowMoverPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetSlowMovers(ctx, businessID, days)
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]SlowMoverPayload, 0, len(rows))
		for _, row := range rows {
			payloads = append(payloads, SlowMoverPayload{
				ProductID:     row.ProductID,
				Name:          row.Name,
				StockQuantity: row.StockQuantity,
			})
		}
		return result{Rows: payloads}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Rows, nil
}
