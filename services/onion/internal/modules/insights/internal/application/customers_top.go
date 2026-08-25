package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// TopCustomerPayload is one row in the top-customers response. Name is
// already joined here so the frontend doesn't need to stitch first+last.
type TopCustomerPayload struct {
	CustomerID              int64  `json:"customer_id"`
	Name                    string `json:"name"`
	Email                   string `json:"email"`
	LifetimeSpendMinorUnits int64  `json:"lifetime_spend_minor_units"`
	OrderCount              int64  `json:"order_count"`
}

// GetTopCustomers returns the leaderboard. Limit is capped at 100 upstream
// in the repository. Cache key varies on limit so two callers with
// different limits get distinct cached responses.
func (s *Service) GetTopCustomers(ctx context.Context, businessID int64, limit int32) ([]TopCustomerPayload, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("customers:top:%d", limit))
	if err != nil {
		return nil, err
	}
	type result struct {
		Rows []TopCustomerPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetTopCustomers(ctx, businessID, limit)
		if rerr != nil {
			return result{}, rerr
		}
		payloads := make([]TopCustomerPayload, 0, len(rows))
		for _, row := range rows {
			name := strings.TrimSpace(row.FirstName + " " + row.LastName)
			payloads = append(payloads, TopCustomerPayload{
				CustomerID:              row.CustomerID,
				Name:                    name,
				Email:                   row.Email,
				LifetimeSpendMinorUnits: row.LifetimeSpendMinor,
				OrderCount:              row.OrderCount,
			})
		}
		return result{Rows: payloads}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Rows, nil
}
