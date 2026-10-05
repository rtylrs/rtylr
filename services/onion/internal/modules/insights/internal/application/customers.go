package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

type CustomerKPIPayload struct {
	Total         int64 `json:"total"`
	NewThisPeriod int64 `json:"new_this_period"`
	Active        int64 `json:"active"`
	Churned       int64 `json:"churned"`
}

func (s *Service) GetCustomerKPIs(ctx context.Context, businessID int64, from, to time.Time) (CustomerKPIPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, fmt.Sprintf("customers:kpis:%s:%s", from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		return CustomerKPIPayload{}, err
	}
	out, err := cacher.Of[CustomerKPIPayload]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (CustomerKPIPayload, error) {
		k, rerr := s.repo.GetCustomerKPIs(ctx, businessID, from, to)
		if rerr != nil {
			return CustomerKPIPayload{}, rerr
		}
		return CustomerKPIPayload{
			Total:         k.Total,
			NewThisPeriod: k.NewThisPeriod,
			Active:        k.Active,
			Churned:       k.Churned,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return CustomerKPIPayload{}, err
	}
	return out, nil
}
