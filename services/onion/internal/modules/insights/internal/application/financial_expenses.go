package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// ExpenseByCategoryPayload is one category bucket in the expenses donut.
type ExpenseByCategoryPayload struct {
	CategoryID       int64  `json:"category_id"`
	CategoryName     string `json:"category_name"`
	AmountMinorUnits int64  `json:"amount_minor_units"`
}

// GetExpensesByCategory returns expense spend grouped by category over
// the requested window.
func (s *Service) GetExpensesByCategory(ctx context.Context, businessID int64, from, to time.Time) ([]ExpenseByCategoryPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("financial:expenses:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return nil, err
	}
	type result struct {
		Rows []ExpenseByCategoryPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetExpensesByCategory(ctx, businessID, from, to)
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]ExpenseByCategoryPayload, 0, len(rows))
		for _, row := range rows {
			payloads = append(payloads, ExpenseByCategoryPayload{
				CategoryID:       row.CategoryID,
				CategoryName:     row.CategoryName,
				AmountMinorUnits: row.AmountMinorUnits,
			})
		}
		return result{Rows: payloads}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Rows, nil
}
