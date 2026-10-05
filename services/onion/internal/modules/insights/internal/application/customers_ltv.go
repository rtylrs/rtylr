package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// LTVBucketPayload is one bucket in the LTV distribution histogram.
type LTVBucketPayload struct {
	Label           string `json:"label"`
	LowerMinorUnits int64  `json:"lower_minor_units"`
	UpperMinorUnits int64  `json:"upper_minor_units"`
	CustomerCount   int64  `json:"customer_count"`
}

// GetLTVDistribution returns the bucketed LTV histogram for the owner.
func (s *Service) GetLTVDistribution(ctx context.Context, businessID int64) ([]LTVBucketPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "customers:ltv")
	if err != nil {
		return nil, err
	}
	type result struct {
		Buckets []LTVBucketPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetLTVDistribution(ctx, businessID)
		if rerr != nil {
			return result{}, rerr
		}
		buckets := make([]LTVBucketPayload, 0, len(rows))
		for _, row := range rows {
			buckets = append(buckets, LTVBucketPayload{
				Label:           row.Label,
				LowerMinorUnits: row.LowerMinorUnits,
				UpperMinorUnits: row.UpperMinorUnits,
				CustomerCount:   row.CustomerCount,
			})
		}
		return result{Buckets: buckets}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Buckets, nil
}
