package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type HeadcountPayload struct {
	ActiveStaff        int64   `json:"active_staff"`
	HeadcountAtStart   int64   `json:"headcount_at_start"`
	HiredThisPeriod    int64   `json:"hired_this_period"`
	DepartedThisPeriod int64   `json:"departed_this_period"`
	TurnoverRate       float64 `json:"turnover_rate"`
}

func (s *Service) GetHeadcount(ctx context.Context, businessID int64, from, to time.Time) (HeadcountPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("operations:headcount:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return HeadcountPayload{}, err
	}
	out, err := cacher.Of[HeadcountPayload]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (HeadcountPayload, error) {
		snap, rerr := s.repo.GetHeadcount(ctx, businessID, from, to)
		if rerr != nil {
			return HeadcountPayload{}, rerr
		}
		var turnover float64
		if snap.HeadcountAtStart > 0 {
			turnover = float64(snap.DepartedThisPeriod) / float64(snap.HeadcountAtStart) * 100
		}
		return HeadcountPayload{
			ActiveStaff:        snap.ActiveHeadcount,
			HeadcountAtStart:   snap.HeadcountAtStart,
			HiredThisPeriod:    snap.HiredThisPeriod,
			DepartedThisPeriod: snap.DepartedThisPeriod,
			TurnoverRate:       turnover,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return HeadcountPayload{}, err
	}
	return out, nil
}
