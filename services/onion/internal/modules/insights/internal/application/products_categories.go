package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// CategorySalesPayload is one category row in the sales-by-category chart.
type CategorySalesPayload struct {
	CategoryID        int64   `json:"category_id"`
	CategoryName      string  `json:"category_name"`
	UnitsSold         float64 `json:"units_sold"`
	RevenueMinorUnits int64   `json:"revenue_minor_units"`
}

// GetSalesByCategory returns MTD sales grouped by product category.
func (s *Service) GetSalesByCategory(ctx context.Context, businessID int64) ([]CategorySalesPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "products:categories")
	if err != nil {
		return nil, err
	}
	type result struct {
		Rows []CategorySalesPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetSalesByCategory(ctx, businessID)
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]CategorySalesPayload, 0, len(rows))
		for _, row := range rows {
			payloads = append(payloads, CategorySalesPayload{
				CategoryID:        row.CategoryID,
				CategoryName:      row.CategoryName,
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
