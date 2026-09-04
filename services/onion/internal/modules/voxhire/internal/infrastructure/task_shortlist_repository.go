package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

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
		ownerID, in.Title, in.Title, in.Status, in.Status, sqlNullTimePtr(in.DueAt),
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

func (r *Repository) listShortlistItems(ctx context.Context, shortlistID int64) ([]domain.ShortlistItem, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, listShortlistItemsScopedSQL, ownerID, shortlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.ShortlistItem, 0)
	for rows.Next() {
		var row sqlc.ListShortlistItemsRow
		if err := rows.Scan(&row.ID, &row.ShortlistID, &row.ApplicationID, &row.CandidateName,
			&row.CandidateEmail, &row.Status, &row.SortOrder, &row.Notes, &row.CvProfileJson,
			&row.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, shortlistItemFromRow(row))
	}
	return items, rows.Err()
}

func (r *Repository) getShortlistItemByID(ctx context.Context, id int64) (*domain.ShortlistItem, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var row sqlc.GetShortlistItemByIDRow
	err = r.db.QueryRowContext(ctx, getShortlistItemScopedSQL, ownerID, id).Scan(
		&row.ID, &row.ShortlistID, &row.ApplicationID, &row.CandidateName, &row.CandidateEmail,
		&row.Status, &row.SortOrder, &row.Notes, &row.CvProfileJson, &row.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	item := domain.ShortlistItem{
		ID: row.ID, ShortlistID: row.ShortlistID, ApplicationID: row.ApplicationID,
		CandidateName: row.CandidateName, CandidateEmail: row.CandidateEmail, Status: row.Status,
		Rank: int(row.SortOrder), Notes: row.Notes, CreatedAt: adminCreatedAt(row.CreatedAt),
		CVProfile: parseCVProfileJSON(row.CvProfileJson),
	}
	return &item, nil
}

func (r *Repository) ListApplicationTimeline(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var meta sqlc.GetApplicationTimelineMetaRow
	err = r.db.QueryRowContext(ctx, getApplicationTimelineMetaScopedSQL, ownerID, id).Scan(
		&meta.CreatedAt, &meta.CvProcessingStatus, &meta.CvProcessingError,
		&meta.CvProcessedAt, &meta.ReviewedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, err
	}

	events := []domain.TimelineEvent{{
		Kind:      "application",
		Title:     "Application submitted",
		CreatedAt: adminCreatedAt(meta.CreatedAt),
	}}
	if meta.CvProcessedAt.Valid {
		title := "CV processed"
		if meta.CvProcessingStatus == "error" {
			title = "CV extraction failed"
		}
		events = append(events, domain.TimelineEvent{
			Kind:      "cv",
			Title:     title,
			Detail:    meta.CvProcessingError,
			CreatedAt: meta.CvProcessedAt.Time,
		})
	}
	if meta.ReviewedAt.Valid {
		events = append(events, domain.TimelineEvent{
			Kind:      "review",
			Title:     "Recruiter review saved",
			CreatedAt: meta.ReviewedAt.Time,
		})
	}
	extra, err := r.listRecordedApplicationEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	events = append(events, extra...)
	taskEvents, err := r.listTaskTimelineEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	events = append(events, taskEvents...)
	shortlistEvents, err := r.listShortlistTimelineEvents(ctx, id)
	if err != nil {
		return nil, err
	}
	events = append(events, shortlistEvents...)
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].CreatedAt.After(events[j].CreatedAt)
	})
	return events, nil
}

func (r *Repository) listRecordedApplicationEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, listApplicationEventsScopedSQL, ownerID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var row sqlc.ListApplicationEventsRow
		if err := rows.Scan(&row.Kind, &row.Title, &row.Detail, &row.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, domain.TimelineEvent{
			Kind: row.Kind, Title: row.Title, Detail: row.Detail, CreatedAt: adminCreatedAt(row.CreatedAt),
		})
	}
	return events, rows.Err()
}

func (r *Repository) listTaskTimelineEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListTaskTimelineEvents(ctx, sqlc.ListTaskTimelineEventsParams{
		OwnerID: ownerID, ApplicationID: id,
	})
	if err != nil {
		return nil, err
	}
	events := make([]domain.TimelineEvent, 0)
	for _, row := range rows {
		events = append(events, domain.TimelineEvent{
			Kind: "task", Title: "Task opened", Detail: row.Title, CreatedAt: adminCreatedAt(row.CreatedAt),
		})
		if row.CompletedAt.Valid {
			events = append(events, domain.TimelineEvent{
				Kind: "task", Title: "Task completed", Detail: row.Title, CreatedAt: row.CompletedAt.Time,
			})
		} else if row.Status != "open" {
			events[len(events)-1].Title = "Task " + row.Status
		}
	}
	return events, nil
}

func (r *Repository) listShortlistTimelineEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, listShortlistTimelineEventsScopedSQL, ownerID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var row sqlc.ListShortlistTimelineEventsRow
		if err := rows.Scan(&row.Name, &row.SortOrder, &row.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, domain.TimelineEvent{
			Kind:      "shortlist",
			Title:     "Added to shortlist",
			Detail:    fmt.Sprintf("%s · rank %d", row.Name, row.SortOrder),
			CreatedAt: adminCreatedAt(row.CreatedAt),
		})
	}
	return events, rows.Err()
}

func (r *Repository) applicationExists(ctx context.Context, id int64) (bool, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.QueryRowContext(ctx, applicationExistsScopedSQL, ownerID, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) shortlistExists(ctx context.Context, id int64) (bool, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return false, err
	}
	var count int64
	err = r.db.QueryRowContext(ctx, shortlistExistsScopedSQL, ownerID, id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) recordApplicationEvent(ctx context.Context, applicationID int64, kind, title, detail string) error {
	if applicationID <= 0 || kind == "" || title == "" {
		return nil
	}
	return r.queries.InsertApplicationEvent(ctx, sqlc.InsertApplicationEventParams{
		ApplicationID: applicationID, Kind: kind, Title: title, Detail: sqlNullString(detail),
	})
}

// ListPendingCVProcessingTargets returns stored CVs that have not been
// extracted into searchable text yet. It covers both the talent pool and
// job-specific application tables.
func (r *Repository) ListPendingCVProcessingTargets(ctx context.Context, limit int) ([]domain.CVProcessingTarget, error) {
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}

	targets := make([]domain.CVProcessingTarget, 0, limit)
	perTable := limit / 2
	if perTable < 1 {
		perTable = 1
	}

	talentRows, err := r.db.QueryContext(ctx, `
		SELECT id, cv_key
		  FROM voxhire_talent_applications
		 WHERE cv_key IS NOT NULL
		   AND cv_key <> ''
		   AND COALESCE(cv_processing_status, 'pending') IN ('pending', 'error')
		 ORDER BY created_at ASC
		 LIMIT ?
	`, perTable)
	if err != nil {
		return nil, err
	}
	defer talentRows.Close()
	for talentRows.Next() {
		var t domain.CVProcessingTarget
		t.Source = "talent"
		if err := talentRows.Scan(&t.ID, &t.CVKey); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	if err := talentRows.Err(); err != nil {
		return nil, err
	}

	remaining := limit - len(targets)
	if remaining <= 0 {
		return targets, nil
	}
	careerRows, err := r.db.QueryContext(ctx, `
		SELECT id, cv_key
		  FROM voxire_career_applications
		 WHERE cv_key IS NOT NULL
		   AND cv_key <> ''
		   AND COALESCE(cv_processing_status, 'pending') IN ('pending', 'error')
		 ORDER BY created_at ASC
		 LIMIT ?
	`, remaining)
	if err != nil {
		return nil, err
	}
	defer careerRows.Close()
	for careerRows.Next() {
		var t domain.CVProcessingTarget
		t.Source = "career"
		if err := careerRows.Scan(&t.ID, &t.CVKey); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, careerRows.Err()
}

// UpdateCVProcessingResult stores extracted CV text and summary for search.
func (r *Repository) UpdateCVProcessingResult(ctx context.Context, result domain.CVProcessingResult) error {
	var table string
	switch result.Source {
	case "talent":
		table = "voxhire_talent_applications"
	case "career":
		table = "voxire_career_applications"
	default:
		return fmt.Errorf("unknown cv processing source %q", result.Source)
	}
	status := strings.TrimSpace(result.Status)
	if status == "" {
		status = "processed"
	}
	profileJSON := marshalCVProfileJSON(result.Profile)

	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		   SET cv_processing_status = ?,
		       cv_processing_error = ?,
		       cv_extracted_text = ?,
		       cv_summary = ?,
		       cv_profile_json = ?,
		       cv_processed_at = CURRENT_TIMESTAMP
		 WHERE id = ?
	`, table), status, nullStr(result.Error), nullStr(result.ExtractedText), nullStr(result.Summary), nullStr(profileJSON), result.ID)
	return err
}

func marshalCVProfileJSON(profile domain.CVProfile) string {
	raw, err := json.Marshal(profile)
	if err != nil || string(raw) == "{}" {
		return ""
	}
	return string(raw)
}

func parseCVProfileJSON(raw string) domain.CVProfile {
	var profile domain.CVProfile
	if strings.TrimSpace(raw) == "" {
		return profile
	}
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		return domain.CVProfile{}
	}
	return profile
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
