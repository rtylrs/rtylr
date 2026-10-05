package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type ChurnWindowPayload struct {
	WindowDays   int32   `json:"window_days"`
	ChurnedCount int64   `json:"churned_count"`
	ActiveCount  int64   `json:"active_count"`
	ChurnPct     float64 `json:"churn_pct"`
}

func (s *Service) GetChurnRates(ctx context.Context, businessID int64, from, to time.Time) ([]ChurnWindowPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("customers:churn:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return nil, err
	}
	type result struct {
		Windows []ChurnWindowPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetChurnRates(ctx, businessID, from, to)
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
