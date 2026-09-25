package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure/persistence/sqlc"
)

const mysqlErrDuplicateEntry = 1062

func parseDecimalString(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func ifaceToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case []byte:
		n, err := strconv.ParseInt(string(val), 10, 64)
		if err != nil {
			return 0
		}
		return n
	case string:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}

func adminCreatedAt(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func sqlNullDecimal(f float64) sql.NullString {
	if f == 0 {
		return sql.NullString{}
	}
	return sql.NullString{String: strconv.FormatFloat(f, 'f', 2, 64), Valid: true}
}

func sqlNullInt64ID(id int64) sql.NullInt64 {
	if id <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: id, Valid: true}
}

func sqlNullTimePtr(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func openingFromCareerRow(row sqlc.GetCareerOpeningByIDRow) *domain.Opening {
	return &domain.Opening{
		ID: row.ID, Slug: row.Slug, Title: row.Title, Department: row.Department,
		Location: row.Location, EmploymentType: row.EmploymentType, WorkMode: row.WorkMode,
		Seniority: row.Seniority, Details: row.Details,
		SalaryMin: parseDecimalString(row.SalaryMin), SalaryMax: parseDecimalString(row.SalaryMax),
		SalaryCurrency: row.SalaryCurrency, IsActive: row.IsActive,
		NewCount: ifaceToInt64(row.NewCount), TotalCount: ifaceToInt64(row.TotalCount),
	}
}

func taskFromRow(id, applicationID, jobID int64, openingTitle, candidateName, candidateEmail, title, status string,
	dueAt, completedAt, createdAt sql.NullTime,
) domain.Task {
	task := domain.Task{
		ID: id, ApplicationID: applicationID, JobID: jobID, OpeningTitle: openingTitle,
		CandidateName: candidateName, CandidateEmail: candidateEmail, Title: title, Status: status,
		CreatedAt: adminCreatedAt(createdAt),
	}
	if dueAt.Valid {
		t := dueAt.Time
		task.DueAt = &t
	}
	if completedAt.Valid {
		t := completedAt.Time
		task.CompletedAt = &t
	}
	return task
}

func shortlistItemFromRow(row sqlc.ListShortlistItemsRow) domain.ShortlistItem {
	return domain.ShortlistItem{
		ID: row.ID, ShortlistID: row.ShortlistID, ApplicationID: row.ApplicationID,
		CandidateName: row.CandidateName, CandidateEmail: row.CandidateEmail, Status: row.Status,
		Rank: int(row.SortOrder), Notes: row.Notes, CreatedAt: adminCreatedAt(row.CreatedAt),
		CVProfile: parseCVProfileJSON(row.CvProfileJson),
	}
}

func analyticsBucketsFromRows(bucket string, count interface{}) domain.AnalyticsBucket {
	return domain.AnalyticsBucket{Key: bucket, Count: ifaceToInt64(count)}
}

func ifaceToTimePtr(v interface{}) *time.Time {
	switch val := v.(type) {
	case time.Time:
		return &val
	case []byte:
		if t, err := time.Parse("2006-01-02 15:04:05", string(val)); err == nil {
			return &t
		}
	case sql.NullTime:
		if val.Valid {
			t := val.Time
			return &t
		}
	}
	return nil
}

func ownerIDFromContext(ctx context.Context) (int64, error) {
	ownerID, ok := domain.OwnerIDFromContext(ctx)
	if !ok {
		return 0, domain.ErrOwnerScopeMissing
	}
	return ownerID, nil
}

const (
	updateApplicationStatusScopedSQL = `
UPDATE voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
SET a.status = ?
WHERE a.id = ?`
	updateApplicationReviewScopedSQL = `
UPDATE voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
SET a.status = CASE WHEN ? <> '' THEN ? ELSE a.status END,
    a.review_rating = ?,
    a.review_notes = ?,
    a.next_step = ?,
    a.reviewed_at = CURRENT_TIMESTAMP
WHERE a.id = ?`
	getApplicationCVKeyScopedSQL = `
SELECT COALESCE(a.cv_key, '')
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.id = ?`
	getApplicationTimelineMetaScopedSQL = `
SELECT a.created_at,
       COALESCE(a.cv_processing_status, 'pending'),
       COALESCE(a.cv_processing_error, ''),
       a.cv_processed_at,
       a.reviewed_at
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.id = ?`
	listApplicationEventsScopedSQL = `
SELECT e.kind, e.title, COALESCE(e.detail, ''), e.created_at
FROM voxhire_application_events e
JOIN voxire_career_applications a ON a.id = e.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE e.application_id = ?
ORDER BY e.created_at DESC`
	getTaskByIDScopedSQL = `
SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, ''),
       a.name, a.email, t.title, t.status, t.due_at, t.completed_at, t.created_at
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE t.id = ?`
	updateFollowUpTaskScopedSQL = `
UPDATE voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
SET t.title = CASE WHEN ? <> '' THEN ? ELSE t.title END,
    t.status = CASE WHEN ? <> '' THEN ? ELSE t.status END,
    t.due_at = CASE WHEN ? THEN ? ELSE t.due_at END,
    t.completed_at = CASE
      WHEN ? = 'done' THEN COALESCE(t.completed_at, CURRENT_TIMESTAMP)
      WHEN ? = 'open' THEN NULL
      ELSE t.completed_at
    END
WHERE t.id = ?`
	applicationExistsScopedSQL = `
SELECT COUNT(*)
FROM voxire_career_applications a
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE a.id = ?`
	careerExistsScopedSQL = `
SELECT COUNT(*) FROM careers WHERE id = ? AND owner_id = ?`
	listShortlistsScopedSQL = `
SELECT s.id, s.job_id, COALESCE(c.title, ''), s.name, s.status,
       COALESCE(s.notes, ''), s.created_at, s.updated_at
FROM voxhire_shortlists s
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
WHERE (? IS NULL OR s.job_id = ?)
ORDER BY s.updated_at DESC`
	listShortlistItemsScopedSQL = `
SELECT i.id, i.shortlist_id, i.application_id, a.name, a.email,
       COALESCE(a.status, 'new'), i.sort_order, COALESCE(i.notes, ''),
       COALESCE(a.cv_profile_json, ''), i.created_at
FROM voxhire_shortlist_items i
JOIN voxhire_shortlists s ON s.id = i.shortlist_id
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
JOIN voxire_career_applications a ON a.id = i.application_id
WHERE i.shortlist_id = ?
ORDER BY i.sort_order ASC, i.created_at ASC`
	getShortlistItemScopedSQL = `
SELECT i.id, i.shortlist_id, i.application_id, a.name, a.email,
       COALESCE(a.status, 'new'), i.sort_order, COALESCE(i.notes, ''),
       COALESCE(a.cv_profile_json, ''), i.created_at
FROM voxhire_shortlist_items i
JOIN voxhire_shortlists s ON s.id = i.shortlist_id
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
JOIN voxire_career_applications a ON a.id = i.application_id
WHERE i.id = ?`
	listShortlistTimelineEventsScopedSQL = `
SELECT s.name, i.sort_order, i.created_at
FROM voxhire_shortlist_items i
JOIN voxhire_shortlists s ON s.id = i.shortlist_id
JOIN voxire_career_applications a ON a.id = i.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE i.application_id = ?`
	shortlistExistsScopedSQL = `
SELECT COUNT(*)
FROM voxhire_shortlists s
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
WHERE s.id = ?`
	analyticsTaskStatusMixScopedSQL = `
SELECT LOWER(COALESCE(NULLIF(t.status, ''), 'open')), COUNT(*)
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
GROUP BY LOWER(COALESCE(NULLIF(t.status, ''), 'open'))
ORDER BY count DESC, 1 ASC`
	analyticsOpenTaskTotalsScopedSQL = `
SELECT COUNT(*), COALESCE(SUM(CASE WHEN t.status = 'open' AND t.due_at IS NOT NULL
    AND t.due_at < UTC_TIMESTAMP() THEN 1 ELSE 0 END), 0)
FROM voxhire_follow_up_tasks t
JOIN voxire_career_applications a ON a.id = t.application_id
JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
WHERE t.status <> 'done'`
	analyticsShortlistTotalsScopedSQL = `
SELECT COUNT(DISTINCT s.id), COUNT(i.id)
FROM voxhire_shortlists s
JOIN careers c ON c.id = s.job_id AND c.owner_id = ?
LEFT JOIN voxhire_shortlist_items i ON i.shortlist_id = s.id`
)

// ── Admin accounts ───────────────────────────────────────────────────
