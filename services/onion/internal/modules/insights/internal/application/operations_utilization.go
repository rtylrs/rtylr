package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type UtilizationPayload struct {
	ScheduledHours float64 `json:"scheduled_hours"`
	AvailableHours float64 `json:"available_hours"`
	UtilizationPct float64 `json:"utilization_pct"`
}

func (s *Service) GetUtilization(ctx context.Context, businessID int64, from, to time.Time) (UtilizationPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("operations:utilization:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return UtilizationPayload{}, err
	}
	out, err := cacher.Of[UtilizationPayload]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (UtilizationPayload, error) {
		row, rerr := s.repo.GetUtilization(ctx, businessID, from, to)
		if rerr != nil {
			return UtilizationPayload{}, rerr
		}
		var pct float64
		if row.AvailableHours > 0 {
			pct = row.ScheduledHours / row.AvailableHours * 100
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
