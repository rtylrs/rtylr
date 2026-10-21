package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
)

// SalesActivity reports the caller's last `days` days, today included (7, 30
// or 90; 30 otherwise), for one rep or everyone. Days follow the caller's
// current UTC offset.
func (s *Service) SalesActivity(ctx context.Context, businessID int64, rep *int64, days int, day domain.CallerDay) (domain.SalesActivity, error) {
	days = domain.ActivityWindow(days)
	first := day.Date.AddDate(0, 0, -(days - 1))
	offset := int64(day.Date.Sub(day.StartUTC) / time.Minute)
	byDay, reps, err := s.repo.SalesActivity(ctx, businessID, rep, day.StartOf(first), offset)
	if err != nil {
		return domain.SalesActivity{}, err
	}
	return domain.NewSalesActivity(first, days, byDay, reps), nil
}
