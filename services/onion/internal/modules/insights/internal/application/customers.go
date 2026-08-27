package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// CustomerKPIPayload is the JSON-shaped response served at /insights/customers.
type CustomerKPIPayload struct {
	Total        int64 `json:"total"`
	NewThisMonth int64 `json:"new_this_month"`
	Active       int64 `json:"active"`
	Churned      int64 `json:"churned"`
}

// GetCustomerKPIs returns aggregate customer counters for the owner.
func (s *Service) GetCustomerKPIs(ctx context.Context, businessID int64) (CustomerKPIPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "customers:kpis")
	if err != nil {
		return CustomerKPIPayload{}, err
	}
	out, err := cacher.Of[CustomerKPIPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (CustomerKPIPayload, error) {
		k, rerr := s.repo.GetCustomerKPIs(ctx, businessID)
		if rerr != nil {
			return CustomerKPIPayload{}, rerr
		}
		return CustomerKPIPayload{
			Total:        k.Total,
			NewThisMonth: k.NewThisMonth,
			Active:       k.Active,
			Churned:      k.Churned,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return CustomerKPIPayload{}, err
	}
	return out, nil
}
