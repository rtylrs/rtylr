package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// HeadcountPayload is the /insights/operations/headcount response shape.
type HeadcountPayload struct {
	ActiveStaff int64   `json:"active_staff"`
	MomDelta    int64   `json:"mom_delta"`
	MomDeltaPct float64 `json:"mom_delta_pct"`
}

// GetHeadcount returns current active staff + month-over-month change.
func (s *Service) GetHeadcount(ctx context.Context, businessID int64) (HeadcountPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "operations:headcount")
	if err != nil {
		return HeadcountPayload{}, err
	}
	out, err := cacher.Of[HeadcountPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (HeadcountPayload, error) {
		snap, rerr := s.repo.GetHeadcount(ctx, businessID)
		if rerr != nil {
			return HeadcountPayload{}, rerr
		}
		delta := snap.ActiveHeadcount - snap.HeadcountOneMonthAgo
		var pct float64
		if snap.HeadcountOneMonthAgo > 0 {
			pct = (float64(delta) / float64(snap.HeadcountOneMonthAgo)) * 100
		}
		return HeadcountPayload{
			ActiveStaff: snap.ActiveHeadcount,
			MomDelta:    delta,
			MomDeltaPct: pct,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return HeadcountPayload{}, err
	}
	return out, nil
}
