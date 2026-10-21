package infrastructure

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) SalesActivity(ctx context.Context, businessID int64, rep *int64, since time.Time, offsetMinutes int64) (map[string]domain.ActivityDay, []domain.RepActivity, error) {
	daily := sqlc.CallsByDayParams{OffsetMinutes: offsetMinutes, BusinessID: businessID, Since: since, Rep: nullInt64(rep)}

	byDay := map[string]domain.ActivityDay{}
	calls, err := r.queries.CallsByDay(ctx, daily)
	if err != nil {
		return nil, nil, err
	}
	for _, row := range calls {
		day := byDay[row.LocalDay]
		day.Calls, day.Reached = row.Calls, row.Reached
		byDay[row.LocalDay] = day
	}
	leads, err := r.queries.LeadsAddedByDay(ctx, sqlc.LeadsAddedByDayParams(daily))
	if err != nil {
		return nil, nil, err
	}
	for _, row := range leads {
		day := byDay[row.LocalDay]
		day.LeadsAdded = row.Leads
		byDay[row.LocalDay] = day
	}
	won, err := r.queries.DealsWonByDay(ctx, sqlc.DealsWonByDayParams{
		OffsetMinutes: offsetMinutes, BusinessID: businessID, Since: someTime(since), Rep: nullInt64(rep),
	})
	if err != nil {
		return nil, nil, err
	}
	for _, row := range won {
		day := byDay[row.LocalDay]
		day.DealsWon = row.Won
		byDay[row.LocalDay] = day
	}

	reps := map[int64]*domain.RepActivity{}
	share := func(id int64) *domain.RepActivity {
		if reps[id] == nil {
			reps[id] = &domain.RepActivity{RepID: id}
		}
		return reps[id]
	}
	repCalls, err := r.queries.CallsByRep(ctx, sqlc.CallsByRepParams{BusinessID: businessID, Since: since, Rep: nullInt64(rep)})
	if err != nil {
		return nil, nil, err
	}
	for _, row := range repCalls {
		share(row.RepID).Calls, share(row.RepID).Reached = row.Calls, row.Reached
	}
	qualified, err := r.queries.LeadsQualifiedByRep(ctx, sqlc.LeadsQualifiedByRepParams{
		BusinessID: businessID, Since: someTime(since), Rep: nullInt64(rep),
	})
	if err != nil {
		return nil, nil, err
	}
	for _, row := range qualified {
		share(row.RepID).LeadsQualified = row.Qualified
	}
	outcomes, err := r.queries.DealOutcomesByRep(ctx, sqlc.DealOutcomesByRepParams{
		BusinessID: businessID, Since: someTime(since), Rep: nullInt64(rep),
	})
	if err != nil {
		return nil, nil, err
	}
	for _, row := range outcomes {
		rep := share(row.RepID)
		rep.DealsWon, rep.DealsLost, rep.WonAmount = row.Won, row.Lost, amount(row.WonAmount)
	}

	shares := make([]domain.RepActivity, 0, len(reps))
	for _, rep := range reps {
		shares = append(shares, *rep)
	}
	return byDay, shares, nil
}
