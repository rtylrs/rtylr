package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// ChurnWindowPayload is the JSON shape for a single rolling window.
type ChurnWindowPayload struct {
	WindowDays   int32   `json:"window_days"`
	ChurnedCount int64   `json:"churned_count"`
	ActiveCount  int64   `json:"active_count"`
	ChurnPct     float64 `json:"churn_pct"`
}

// GetChurnRates returns all four standard windows (30/60/90/180d).
func (s *Service) GetChurnRates(ctx context.Context, businessID int64) ([]ChurnWindowPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "customers:churn")
	if err != nil {
		return nil, err
	}
	type result struct {
		Windows []ChurnWindowPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetChurnRates(ctx, businessID)
		if rerr != nil {
			return result{}, rerr
		}
		windows := make([]ChurnWindowPayload, 0, len(rows))
		for _, row := range rows {
			windows = append(windows, ChurnWindowPayload{
				WindowDays:   row.WindowDays,
				ChurnedCount: row.ChurnedCount,
				ActiveCount:  row.ActiveCount,
				ChurnPct:     row.ChurnPct,
			})
		}
		return result{Windows: windows}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Windows, nil
}
