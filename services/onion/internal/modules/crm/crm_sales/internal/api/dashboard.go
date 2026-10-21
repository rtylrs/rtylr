package api

import (
	"context"
	"time"

	crmsalesv1 "github.com/voxire/voxie/gen/go/crm/sales/v1"
)

func (a *API) GetSalesActivity(ctx context.Context, req *crmsalesv1.GetSalesActivityRequest) (*crmsalesv1.GetSalesActivityResponse, error) {
	c, err := a.caller(ctx, permView)
	if err != nil {
		return nil, err
	}
	activity, err := a.Service.SalesActivity(ctx, c.businessID, req.AssignedTo, int(req.Days), a.callerDay(req.Today, req.UtcOffsetMinutes))
	if err != nil {
		return nil, fail(ctx, "load sales activity", err)
	}
	resp := &crmsalesv1.GetSalesActivityResponse{
		Days:           make([]*crmsalesv1.SalesActivityDay, 0, len(activity.Days)),
		Calls:          activity.Calls,
		Reached:        activity.Reached,
		LeadsAdded:     activity.LeadsAdded,
		LeadsQualified: activity.LeadsQualified,
		DealsWon:       activity.DealsWon,
		DealsLost:      activity.DealsLost,
		WonAmount:      activity.WonAmount.String(),
		Reps:           make([]*crmsalesv1.RepActivity, 0, len(activity.Reps)),
	}
	for _, day := range activity.Days {
		resp.Days = append(resp.Days, &crmsalesv1.SalesActivityDay{
			Date:       day.Date.Format(time.DateOnly),
			Calls:      day.Calls,
			Reached:    day.Reached,
			LeadsAdded: day.LeadsAdded,
			DealsWon:   day.DealsWon,
		})
	}
	for _, rep := range activity.Reps {
		resp.Reps = append(resp.Reps, &crmsalesv1.RepActivity{
			RepId:          rep.RepID,
			Calls:          rep.Calls,
			Reached:        rep.Reached,
			LeadsQualified: rep.LeadsQualified,
			DealsWon:       rep.DealsWon,
			DealsLost:      rep.DealsLost,
			WonAmount:      rep.WonAmount.String(),
		})
	}
	return resp, nil
}
