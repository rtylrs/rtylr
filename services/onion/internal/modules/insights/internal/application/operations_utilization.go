package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// UtilizationPayload is the response for /insights/operations/utilization.
type UtilizationPayload struct {
	ScheduledHours float64 `json:"scheduled_hours"`
	AvailableHours float64 `json:"available_hours"`
	UtilizationPct float64 `json:"utilization_pct"`
}

// GetUtilization returns the 7-day scheduled-vs-available utilization
// ratio as a percentage. A fresh tenant with no active staff returns 0%
// rather than divide-by-zero.
func (s *Service) GetUtilization(ctx context.Context, businessID int64) (UtilizationPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "operations:utilization")
	if err != nil {
		return UtilizationPayload{}, err
	}
	out, err := cacher.Of[UtilizationPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (UtilizationPayload, error) {
		row, rerr := s.repo.GetUtilization(ctx, businessID)
		if rerr != nil {
			return UtilizationPayload{}, rerr
		}
		var pct float64
		if row.AvailableHours > 0 {
			pct = (row.ScheduledHours / row.AvailableHours) * 100
		}
		return UtilizationPayload{
			ScheduledHours: row.ScheduledHours,
			AvailableHours: row.AvailableHours,
			UtilizationPct: pct,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return UtilizationPayload{}, err
	}
	return out, nil
}
