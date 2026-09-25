package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure/persistence/sqlc"
)

func (r *Repository) ListTasks(ctx context.Context, status string) ([]domain.Task, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var filterStatus sql.NullString
	if status != "" && status != "all" {
		filterStatus = sql.NullString{String: status, Valid: true}
	}
	query := `
SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, ''),
       a.name, a.email, t.title, t.status, t.due_at, t.completed_at, t.created_at
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE (? IS NULL OR t.status = ?)
ORDER BY CASE t.status WHEN 'open' THEN 0 WHEN 'done' THEN 2 ELSE 1 END,
         t.due_at IS NULL, t.due_at ASC, t.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, ownerID, filterStatus, filterStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := make([]domain.Task, 0)
	for rows.Next() {
		var row sqlc.ListTasksRow
		if err := rows.Scan(&row.ID, &row.ApplicationID, &row.JobID, &row.OpeningTitle,
			&row.CandidateName, &row.CandidateEmail, &row.Title, &row.Status,
			&row.DueAt, &row.CompletedAt, &row.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, taskFromRow(row.ID, row.ApplicationID, row.JobID, row.OpeningTitle,
			row.CandidateName, row.CandidateEmail, row.Title, row.Status,
			row.DueAt, row.CompletedAt, row.CreatedAt))
	}
	return tasks, rows.Err()
}

func (r *Repository) UpsertTask(ctx context.Context, in domain.TaskInput) (*domain.Task, error) {
	if ok, err := r.applicationExists(ctx, in.ApplicationID); err != nil {
		return nil, err
	} else if !ok {
		return nil, domain.ErrApplicationNotFound
	}

	status := in.Status
	if status == "" {
		status = "open"
	}
	res, err := r.queries.UpsertFollowUpTask(ctx, sqlc.UpsertFollowUpTaskParams{
		ApplicationID: in.ApplicationID, Title: in.Title, Status: status, DueAt: sqlNullTimePtr(in.DueAt),
	})
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	_ = r.recordApplicationEvent(ctx, in.ApplicationID, "task", "Follow-up task set", in.Title)
	return r.getTaskByID(ctx, id)
}

func (r *Repository) UpdateTask(ctx context.Context, in domain.TaskInput) (*domain.Task, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := r.db.ExecContext(ctx, updateFollowUpTaskScopedSQL,
		ownerID, in.Title, in.Title, in.Status, in.Status,
		in.DueAtSet, sqlNullTimePtr(in.DueAt),
		in.Status, in.Status, in.ID,
	); err != nil {
		return nil, err
	}
	var count int64
	err = r.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE t.id = ?`, ownerID, in.ID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, domain.ErrApplicationNotFound
	}
	task, err := r.getTaskByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	_ = r.recordApplicationEvent(ctx, task.ApplicationID, "task", "Follow-up task updated", task.Title)
	return task, nil
}

func (r *Repository) getTaskByID(ctx context.Context, id int64) (*domain.Task, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var row sqlc.GetTaskByIDRow
	err = r.db.QueryRowContext(ctx, getTaskByIDScopedSQL, ownerID, id).Scan(
		&row.ID, &row.ApplicationID, &row.JobID, &row.OpeningTitle, &row.CandidateName,
		&row.CandidateEmail, &row.Title, &row.Status, &row.DueAt, &row.CompletedAt, &row.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, err
	}
	task := taskFromRow(row.ID, row.ApplicationID, row.JobID, row.OpeningTitle,
		row.CandidateName, row.CandidateEmail, row.Title, row.Status,
		row.DueAt, row.CompletedAt, row.CreatedAt)
	return &task, nil
}

func (r *Repository) ListShortlists(ctx context.Context, jobID int64) ([]domain.Shortlist, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	jobFilter := sqlNullInt64ID(jobID)
	rows, err := r.db.QueryContext(ctx, listShortlistsScopedSQL, ownerID, jobFilter, jobFilter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shortlists := make([]domain.Shortlist, 0)
	for rows.Next() {
		var row sqlc.ListShortlistsRow
		if err := rows.Scan(&row.ID, &row.JobID, &row.JobTitle, &row.Name, &row.Status,
			&row.Notes, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		shortlist := domain.Shortlist{
			ID: row.ID, JobID: row.JobID, JobTitle: row.JobTitle, Name: row.Name,
			Status: row.Status, Notes: row.Notes,
			CreatedAt: adminCreatedAt(row.CreatedAt), UpdatedAt: adminCreatedAt(row.UpdatedAt),
		}
		items, err := r.listShortlistItems(ctx, shortlist.ID)
		if err != nil {
			return nil, err
		}
		shortlist.Items = items
		shortlists = append(shortlists, shortlist)
	}
	return shortlists, rows.Err()
}

func (r *Repository) CreateShortlist(ctx context.Context, in domain.ShortlistInput) (*domain.Shortlist, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var careerCount int64
	if err := r.db.QueryRowContext(ctx, careerExistsScopedSQL, in.JobID, ownerID).Scan(&careerCount); err != nil {
		return nil, err
	}
	if careerCount == 0 {
		return nil, domain.ErrCareerNotFound
	}
	res, err := r.queries.CreateShortlist(ctx, sqlc.CreateShortlistParams{
		JobID: in.JobID, Name: in.Name, Status: in.Status, Notes: sqlNullString(in.Notes),
	})
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	lists, err := r.ListShortlists(ctx, in.JobID)
	if err != nil {
		return nil, err
	}
	for _, shortlist := range lists {
		if shortlist.ID == id {
			return &shortlist, nil
		}
	}
	return nil, domain.ErrApplicationNotFound
}

func (r *Repository) AddShortlistItem(ctx context.Context, in domain.ShortlistItemInput) (*domain.ShortlistItem, error) {
	if ok, err := r.shortlistExists(ctx, in.ShortlistID); err != nil {
		return nil, err
	} else if !ok {
		return nil, domain.ErrApplicationNotFound
	}
	if ok, err := r.applicationExists(ctx, in.ApplicationID); err != nil {
		return nil, err
	} else if !ok {
		return nil, domain.ErrApplicationNotFound
	}
	res, err := r.queries.UpsertShortlistItem(ctx, sqlc.UpsertShortlistItemParams{
		ShortlistID: in.ShortlistID, ApplicationID: in.ApplicationID,
		SortOrder: int32(in.Rank), Notes: sqlNullString(in.Notes), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	item, err := r.getShortlistItemByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_ = r.recordApplicationEvent(ctx, in.ApplicationID, "shortlist", "Added to shortlist", fmt.Sprintf("Rank %d", item.Rank))
	return item, nil
}
