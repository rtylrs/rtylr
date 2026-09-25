package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// OperationsKPIPayload is the JSON shape for /insights/operations.
type OperationsKPIPayload struct {
	ActiveStaff       int64 `json:"active_staff"`
	ShiftsThisWeek    int64 `json:"shifts_this_week"`
	PendingTimeOff    int64 `json:"pending_time_off"`
	OpenLeaveRequests int64 `json:"open_leave_requests"`
}

// GetOperationsKPIs returns the headline HR ops counters.
func (s *Service) GetOperationsKPIs(ctx context.Context, businessID int64) (OperationsKPIPayload, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "operations:kpis")
	if err != nil {
		return OperationsKPIPayload{}, err
	}
	out, err := cacher.Of[OperationsKPIPayload]().GetOrComputeSingle(ctx, key, func(ctx context.Context) (OperationsKPIPayload, error) {
		k, rerr := s.repo.GetOperationsKPIs(ctx, businessID)
		if rerr != nil {
			return OperationsKPIPayload{}, rerr
		}
		return OperationsKPIPayload{
			ActiveStaff:       k.ActiveStaff,
			ShiftsThisWeek:    k.ShiftsThisWeek,
			PendingTimeOff:    k.PendingTimeOff,
			OpenLeaveRequests: k.OpenLeaveRequests,
		}, nil
	}, cacher.WithExpiry(ctx, 5*time.Minute))
	if err != nil {
		return OperationsKPIPayload{}, err
	}
	return out, nil
}
