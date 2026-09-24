package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type TimeOffPayload struct {
	Pending  int64 `json:"pending"`
	Approved int64 `json:"approved"`
}

func (s *Service) GetTimeOffSummary(ctx context.Context, businessID int64, from, to time.Time) (TimeOffPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("operations:time-off:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return TimeOffPayload{}, err
	}
	out, err := cacher.Of[TimeOffPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (TimeOffPayload, error) {
		row, rerr := s.repo.GetTimeOffSummary(ctx, businessID, from, to)
		if rerr != nil {
			return TimeOffPayload{}, rerr
		}
		return TimeOffPayload{Pending: row.Pending, Approved: row.Approved}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return TimeOffPayload{}, err
	}
	return out, nil
}
