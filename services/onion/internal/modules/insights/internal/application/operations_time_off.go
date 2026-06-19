package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// TimeOffPayload is the response shape for /insights/operations/time-off.
type TimeOffPayload struct {
	Pending  int64 `json:"pending"`
	Approved int64 `json:"approved"`
}

// GetTimeOffSummary returns pending + approved leave request counts.
func (s *Service) GetTimeOffSummary(ctx context.Context, businessID int64) (TimeOffPayload, error) {
	key := fmt.Sprintf("insights:%d:operations:time-off", businessID)
	out, err := cacher.Of[TimeOffPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (TimeOffPayload, error) {
		row, rerr := s.repo.GetTimeOffSummary(ctx, businessID)
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
