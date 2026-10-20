package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

// LogActivity records a touch and rolls it up onto its lead or deal: a lead's
// status follows the outcome and calls count as attempts. A follow-up becomes
// a task in the same transaction.
func (r *Repository) LogActivity(ctx context.Context, businessID, actorID int64, in domain.ActivityInput) (domain.Activity, *domain.Task, error) {
	var activityID, taskID int64
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		params := sqlc.CreateActivityParams{
			OwnerID:    actorID,
			BusinessID: businessID,
			Type:       string(in.Type),
			Outcome:    string(in.Outcome),
			Notes:      nullString(in.Notes),
			CreatedBy:  actorID,
			OccurredAt: in.OccurredAt.UTC(),
		}
		if in.DurationSeconds != nil {
			params.DurationSeconds = sql.NullInt32{Int32: int32(*in.DurationSeconds), Valid: true} // #nosec G115 -- capped at a day.
		}
		var assignee *int64
		switch {
		case in.LeadID != nil:
			row, err := q.GetLeadForUpdate(ctx, sqlc.GetLeadForUpdateParams{ID: *in.LeadID, BusinessID: businessID})
			if err != nil {
				return notFound(err, domain.ErrLeadNotFound)
			}
			lead := leadFromRow(row)
			params.LeadID, params.CustomerID = someInt64(lead.ID), nullInt64(lead.CustomerID)
			attempts := lead.AttemptCount
			if in.Type == domain.ActivityCall {
				attempts++
			}
			last := *in.OccurredAt
			if lead.LastActivityAt != nil && lead.LastActivityAt.After(last) {
				last = *lead.LastActivityAt
			}
			if err := q.RecordLeadWork(ctx, sqlc.RecordLeadWorkParams{
				Status:         sqlc.CrmLeadsStatus(domain.NextLeadStatus(lead.Status, in.Type, in.Outcome)),
				LastActivityAt: someTime(last),
				AttemptCount:   int32(attempts), // #nosec G115 -- one per logged call.
				ID:             lead.ID,
				BusinessID:     businessID,
			}); err != nil {
				return err
			}
			assignee = lead.AssignedTo
		default:
			row, err := q.GetDealForUpdate(ctx, sqlc.GetDealForUpdateParams{ID: *in.DealID, BusinessID: businessID})
			if err != nil {
				return notFound(err, domain.ErrDealNotFound)
			}
			deal := dealFromRow(row)
			params.DealID, params.CustomerID = someInt64(deal.ID), nullInt64(deal.CustomerID)
			last := *in.OccurredAt
			if deal.LastActivityAt != nil && deal.LastActivityAt.After(last) {
				last = *deal.LastActivityAt
			}
			if err := q.RecordDealWork(ctx, sqlc.RecordDealWorkParams{LastActivityAt: someTime(last), ID: deal.ID, BusinessID: businessID}); err != nil {
				return err
			}
			assignee = deal.AssignedTo
		}
		result, err := q.CreateActivity(ctx, params)
		if err != nil {
			return err
		}
		if activityID, err = result.LastInsertId(); err != nil {
			return err
		}
		if in.FollowUp == nil {
			return nil
		}
		if in.FollowUp.AssignedTo == nil {
			in.FollowUp.AssignedTo = assignee
		}
		taskID, err = insertTask(ctx, q, businessID, actorID, *in.FollowUp)
		return err
	})
	if err != nil {
		return domain.Activity{}, nil, err
	}
	row, err := r.queries.GetActivity(ctx, sqlc.GetActivityParams{ID: activityID, BusinessID: businessID})
	if err != nil {
		return domain.Activity{}, nil, err
	}
	activity := activityFromRow(row)
	if taskID == 0 {
		return activity, nil, nil
	}
	task, err := r.GetTask(ctx, businessID, taskID)
	if err != nil {
		return activity, nil, err
	}
	return activity, &task, nil
}

// ListLeadActivities returns a page of a lead's timeline, newest first.
func (r *Repository) ListLeadActivities(ctx context.Context, businessID, leadID int64, token string, limit int) (domain.ActivityPage, error) {
	const sort = "activity"
	cursor, err := decodeCursor(token, sort)
	if err != nil {
		return domain.ActivityPage{}, err
	}
	at, err := cursor.time()
	if err != nil {
		return domain.ActivityPage{}, err
	}
	rows, err := r.queries.ListLeadActivities(ctx, sqlc.ListLeadActivitiesParams{
		BusinessID: businessID, LeadID: someInt64(leadID), CursorID: cursor.id(), CursorTime: at,
		Limit: int32(limit + 1), // #nosec G115 -- capped by the service.
	})
	if err != nil {
		return domain.ActivityPage{}, err
	}
	return activityPage(rows, limit), nil
}

// ListDealActivities returns a page of a deal's timeline, which includes the
// work logged on the lead it came from.
func (r *Repository) ListDealActivities(ctx context.Context, businessID int64, deal domain.Deal, token string, limit int) (domain.ActivityPage, error) {
	const sort = "activity"
	cursor, err := decodeCursor(token, sort)
	if err != nil {
		return domain.ActivityPage{}, err
	}
	at, err := cursor.time()
	if err != nil {
		return domain.ActivityPage{}, err
	}
	rows, err := r.queries.ListDealActivities(ctx, sqlc.ListDealActivitiesParams{
		BusinessID: businessID, DealID: someInt64(deal.ID), LeadID: nullInt64(deal.LeadID),
		CursorID: cursor.id(), CursorTime: at, Limit: int32(limit + 1), // #nosec G115 -- capped by the service.
	})
	if err != nil {
		return domain.ActivityPage{}, err
	}
	return activityPage(rows, limit), nil
}

func activityPage(rows []sqlc.CrmTouchpoint, limit int) domain.ActivityPage {
	page := domain.ActivityPage{}
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		page.NextCursor = encodeCursor("activity", timeValue(last.OccurredAt), last.ID)
	}
	page.Activities = make([]domain.Activity, 0, len(rows))
	for _, row := range rows {
		page.Activities = append(page.Activities, activityFromRow(row))
	}
	return page
}

func (r *Repository) DeleteActivity(ctx context.Context, businessID, id int64) error {
	result, err := r.queries.DeleteActivity(ctx, sqlc.DeleteActivityParams{ID: id, BusinessID: businessID})
	if err != nil {
		return err
	}
	if rows, err := result.RowsAffected(); err != nil || rows == 0 {
		if err != nil {
			return err
		}
		return domain.ErrActivityNotFound
	}
	return nil
}
