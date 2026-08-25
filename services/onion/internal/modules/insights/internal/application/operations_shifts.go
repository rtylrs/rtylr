package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// ShiftMetricsPayload is the /insights/operations/shifts response.
type ShiftMetricsPayload struct {
	Scheduled      int64   `json:"scheduled"`
	Completed      int64   `json:"completed"`
	NoShows        int64   `json:"no_shows"`
	CompletionRate float64 `json:"completion_rate_pct"`
}

// GetShiftMetrics returns scheduled/completed counts for the window.
func (s *Service) GetShiftMetrics(ctx context.Context, businessID int64, from, to time.Time) (ShiftMetricsPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("operations:shifts:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return ShiftMetricsPayload{}, err
	}
	out, err := cacher.Of[ShiftMetricsPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (ShiftMetricsPayload, error) {
		row, rerr := s.repo.GetShiftMetrics(ctx, businessID, from, to)
		if rerr != nil {
			return ShiftMetricsPayload{}, rerr
		}
		var rate float64
		if row.Scheduled > 0 {
			rate = (float64(row.Completed) / float64(row.Scheduled)) * 100
		}
		return ShiftMetricsPayload{
			Scheduled:      row.Scheduled,
			Completed:      row.Completed,
			NoShows:        row.NoShows,
			CompletionRate: rate,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return ShiftMetricsPayload{}, err
	}
	return out, nil
}
