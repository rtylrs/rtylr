package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/domain"
	"github.com/voxire/onion/internal/modules/crm/crm_sales/internal/infrastructure/persistence/sqlc"
)

// insertTask stores a normalized task after checking its lead or deal is in
// the business, then refreshes their next follow-up.
func insertTask(ctx context.Context, q *sqlc.Queries, businessID, actorID int64, in domain.TaskInput) (int64, error) {
	if err := checkTaskSubject(ctx, q, businessID, in.LeadID, in.DealID); err != nil {
		return 0, err
	}
	result, err := q.CreateTask(ctx, sqlc.CreateTaskParams{
		BusinessID: businessID,
		OwnerID:    actorID,
		LeadID:     nullInt64(in.LeadID),
		DealID:     nullInt64(in.DealID),
		Title:      in.Title,
		Type:       sqlc.CrmTasksType(in.Type),
		DueDate:    nullDate(in.DueDate).Time,
		AssignedTo: nullInt64(in.AssignedTo),
		CreatedBy:  actorID,
	})
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, refreshFollowUps(ctx, q, businessID, in.LeadID, in.DealID)
}

func checkTaskSubject(ctx context.Context, q *sqlc.Queries, businessID int64, leadID, dealID *int64) error {
	if leadID != nil {
		if _, err := q.GetLead(ctx, sqlc.GetLeadParams{ID: *leadID, BusinessID: businessID}); err != nil {
			return notFound(err, domain.ErrLeadNotFound)
		}
	}
	if dealID != nil {
		if _, err := q.GetDeal(ctx, sqlc.GetDealParams{ID: *dealID, BusinessID: businessID}); err != nil {
			return notFound(err, domain.ErrDealNotFound)
		}
	}
	return nil
}

func refreshFollowUps(ctx context.Context, q *sqlc.Queries, businessID int64, leadID, dealID *int64) error {
	if leadID != nil {
		if err := q.SetLeadNextFollowUp(ctx, sqlc.SetLeadNextFollowUpParams{ID: *leadID, BusinessID: businessID}); err != nil {
			return err
		}
	}
	if dealID != nil {
		return q.SetDealNextFollowUp(ctx, sqlc.SetDealNextFollowUpParams{ID: *dealID, BusinessID: businessID})
	}
	return nil
}

func (r *Repository) CreateTask(ctx context.Context, businessID, actorID int64, in domain.TaskInput) (domain.Task, error) {
	var id int64
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) (err error) {
		id, err = insertTask(ctx, q, businessID, actorID, in)
		return err
	})
	if err != nil {
		return domain.Task{}, err
	}
	return r.GetTask(ctx, businessID, id)
}

func (r *Repository) GetTask(ctx context.Context, businessID, id int64) (domain.Task, error) {
	row, err := r.queries.GetTask(ctx, sqlc.GetTaskParams{ID: id, BusinessID: businessID})
	if err != nil {
		return domain.Task{}, notFound(err, domain.ErrTaskNotFound)
	}
	return taskFromRow(row), nil
}

// UpdateTask edits a task's title, type, due date and assignee.
func (r *Repository) UpdateTask(ctx context.Context, businessID, id int64, in domain.TaskInput) (domain.Task, error) {
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		row, err := q.GetTask(ctx, sqlc.GetTaskParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrTaskNotFound)
		}
		if err := q.UpdateTask(ctx, sqlc.UpdateTaskParams{
			Title: in.Title, Type: sqlc.CrmTasksType(in.Type), DueDate: nullDate(in.DueDate).Time,
			AssignedTo: nullInt64(in.AssignedTo), ID: id, BusinessID: businessID,
		}); err != nil {
			return err
		}
		return refreshFollowUps(ctx, q, businessID, int64Ptr(row.LeadID), int64Ptr(row.DealID))
	})
	if err != nil {
		return domain.Task{}, err
	}
	return r.GetTask(ctx, businessID, id)
}

// CompleteTask marks a task done at now, or open again.
func (r *Repository) CompleteTask(ctx context.Context, businessID, id int64, done bool, now time.Time) (domain.Task, error) {
	err := r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		row, err := q.GetTask(ctx, sqlc.GetTaskParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrTaskNotFound)
		}
		completedAt := sql.NullTime{}
		if done {
			completedAt = someTime(now)
		}
		if err := q.SetTaskCompletedAt(ctx, sqlc.SetTaskCompletedAtParams{CompletedAt: completedAt, ID: id, BusinessID: businessID}); err != nil {
			return err
		}
		return refreshFollowUps(ctx, q, businessID, int64Ptr(row.LeadID), int64Ptr(row.DealID))
	})
	if err != nil {
		return domain.Task{}, err
	}
	return r.GetTask(ctx, businessID, id)
}

func (r *Repository) DeleteTask(ctx context.Context, businessID, id int64) error {
	return r.inTx(ctx, func(q *sqlc.Queries, _ *sql.Tx) error {
		row, err := q.GetTask(ctx, sqlc.GetTaskParams{ID: id, BusinessID: businessID})
		if err != nil {
			return notFound(err, domain.ErrTaskNotFound)
		}
		if _, err := q.DeleteTask(ctx, sqlc.DeleteTaskParams{ID: id, BusinessID: businessID}); err != nil {
			return err
		}
		return refreshFollowUps(ctx, q, businessID, int64Ptr(row.LeadID), int64Ptr(row.DealID))
	})
}

// ListTasks returns a page of open tasks (earliest due first) or completed
// ones (latest first). today is the caller's date for due windows.
func (r *Repository) ListTasks(ctx context.Context, businessID int64, f domain.TaskFilter, today time.Time) (domain.TaskPage, error) {
	cursor, err := decodeCursor(f.Cursor, string(f.Scope))
	if err != nil {
		return domain.TaskPage{}, err
	}
	limit := int32(f.Limit + 1) // #nosec G115 -- capped by the service.
	var rows []taskListRow
	if f.Scope == domain.TasksCompleted {
		at, err := cursor.time()
		if err != nil {
			return domain.TaskPage{}, err
		}
		found, err := r.queries.ListCompletedTasks(ctx, sqlc.ListCompletedTasksParams{
			BusinessID: businessID, AssignedTo: nullInt64(f.AssignedTo), LeadID: nullInt64(f.LeadID),
			DealID: nullInt64(f.DealID), CursorID: cursor.id(), CursorTime: at, Limit: limit,
		})
		if err != nil {
			return domain.TaskPage{}, err
		}
		for _, row := range found {
			rows = append(rows, taskListRow{
				task: sqlc.CrmTask{
					ID: row.ID, LeadID: row.LeadID, DealID: row.DealID, Title: row.Title, Type: row.Type,
					DueDate: row.DueDate, AssignedTo: row.AssignedTo, CompletedAt: row.CompletedAt,
					CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt,
				},
				leadFirst: row.LeadFirstName, leadLast: row.LeadLastName, leadCompany: row.LeadCompanyName,
				leadPhone: row.LeadPhone, dealTitle: row.DealTitle,
			})
		}
	} else {
		at, err := cursor.date()
		if err != nil {
			return domain.TaskPage{}, err
		}
		from, to := f.Due.Range(today)
		found, err := r.queries.ListOpenTasks(ctx, sqlc.ListOpenTasksParams{
			BusinessID: businessID, AssignedTo: nullInt64(f.AssignedTo), DueFrom: nullTime(from), DueTo: nullTime(to),
			LeadID: nullInt64(f.LeadID), DealID: nullInt64(f.DealID), CursorID: cursor.id(), CursorDate: at, Limit: limit,
		})
		if err != nil {
			return domain.TaskPage{}, err
		}
		for _, row := range found {
			rows = append(rows, taskListRow{
				task: sqlc.CrmTask{
					ID: row.ID, LeadID: row.LeadID, DealID: row.DealID, Title: row.Title, Type: row.Type,
					DueDate: row.DueDate, AssignedTo: row.AssignedTo, CompletedAt: row.CompletedAt,
					CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt,
				},
				leadFirst: row.LeadFirstName, leadLast: row.LeadLastName, leadCompany: row.LeadCompanyName,
				leadPhone: row.LeadPhone, dealTitle: row.DealTitle,
			})
		}
	}
	page := domain.TaskPage{}
	if len(rows) > f.Limit {
		rows = rows[:f.Limit]
		last := rows[len(rows)-1].task
		value := last.DueDate.Format(time.DateOnly)
		if f.Scope == domain.TasksCompleted {
			value = timeValue(last.CompletedAt.Time)
		}
		page.NextCursor = encodeCursor(string(f.Scope), value, last.ID)
	}
	page.Tasks = make([]domain.Task, 0, len(rows))
	for _, row := range rows {
		page.Tasks = append(page.Tasks, taskFromListRow(row))
	}
	return page, nil
}

func (r *Repository) CountTasks(ctx context.Context, businessID int64, assignedTo *int64, today time.Time) (domain.TaskCounts, error) {
	row, err := r.queries.CountOpenTasks(ctx, sqlc.CountOpenTasksParams{
		Today: dayStart(today), BusinessID: businessID, AssignedTo: nullInt64(assignedTo),
	})
	if err != nil {
		return domain.TaskCounts{}, err
	}
	return domain.TaskCounts{Overdue: row.Overdue, DueToday: row.DueToday, Upcoming: row.Upcoming}, nil
}
