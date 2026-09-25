package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// CashFlowPointPayload is one day in the cash-flow chart.
type CashFlowPointPayload struct {
	Date              string `json:"date"`
	InflowMinorUnits  int64  `json:"inflow_minor_units"`
	OutflowMinorUnits int64  `json:"outflow_minor_units"`
}

// GetCashFlowSeries returns the daily (inflow, outflow) series for the
// given window.
func (s *Service) GetCashFlowSeries(ctx context.Context, businessID int64, from, to time.Time) ([]CashFlowPointPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("financial:cashflow:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return nil, err
	}
	type result struct {
		Points []CashFlowPointPayload
	}
	out, err := cacher.Of[result]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (result, error) {
		rows, rerr := s.repo.GetCashFlowSeries(ctx, businessID, from, to)
		if rerr != nil {
			return result{}, rerr
		}
		points := make([]CashFlowPointPayload, 0, len(rows))
		for _, row := range rows {
			points = append(points, CashFlowPointPayload{
				Date:              row.Date,
				InflowMinorUnits:  row.InflowMinorUnits,
				OutflowMinorUnits: row.OutflowMinorUnits,
			})
		}
		return result{Points: points}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return nil, err
	}
	return out.Points, nil
}
