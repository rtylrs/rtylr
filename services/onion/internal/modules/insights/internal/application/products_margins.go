package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// MarginBucketPayload is one bucket in the margin distribution histogram.
type MarginBucketPayload struct {
	Label        string  `json:"label"`
	LowerPct     float64 `json:"lower_pct"`
	UpperPct     float64 `json:"upper_pct"`
	ProductCount int64   `json:"product_count"`
}

// GetMarginDistribution returns the bucketed margin histogram.
func (s *Service) GetMarginDistribution(ctx context.Context, businessID int64) ([]MarginBucketPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "products:margins")
	if err != nil {
		return nil, err
	}
	type result struct {
		Buckets []MarginBucketPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetMarginDistribution(ctx, businessID)
		if rerr != nil {
			return result{}, rerr
		}
		buckets := make([]MarginBucketPayload, 0, len(rows))
		for _, row := range rows {
			buckets = append(buckets, MarginBucketPayload{
				Label:        row.Label,
				LowerPct:     row.LowerPct,
				UpperPct:     row.UpperPct,
				ProductCount: row.ProductCount,
			})
		}
		return result{Buckets: buckets}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Buckets, nil
}
