package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// AppRevenueShare is the shape returned to the frontend: each entry carries
// the raw amount in minor units plus its share of the total as a percent.
type AppRevenueShare struct {
	App              string  `json:"app"`
	AmountMinorUnits int64   `json:"amount_minor_units"`
	Pct              float64 `json:"pct"`
}

// GetRevenueByApp splits MTD revenue across app surfaces (POS, Finance,
// ERP, CRM) and computes each surface's share of the total.
func (s *Service) GetRevenueByApp(ctx context.Context, businessID int64) ([]AppRevenueShare, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "revenue:by-app")
	if err != nil {
		return nil, err
	}
	type result struct {
		Shares []AppRevenueShare
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetRevenueByApp(ctx, businessID)
		if rerr != nil {
			return result{}, rerr
		}
		var total int64
		for _, row := range rows {
			total += row.AmountMinorUnits
		}
		shares := make([]AppRevenueShare, 0, len(rows))
		for _, row := range rows {
			var pct float64
			if total > 0 {
				pct = (float64(row.AmountMinorUnits) / float64(total)) * 100
			}
			shares = append(shares, AppRevenueShare{
				App:              row.App,
				AmountMinorUnits: row.AmountMinorUnits,
				Pct:              pct,
			})
		}
		return result{Shares: shares}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Shares, nil
}
