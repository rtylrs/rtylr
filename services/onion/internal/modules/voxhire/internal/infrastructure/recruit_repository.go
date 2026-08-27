package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	gosqlmysql "github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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
    t.due_at = ?,
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

// CreateAdmin inserts a new recruiter-admin account. Returns
// domain.ErrAdminExists if the email is already taken.
func (r *Repository) CreateAdmin(ctx context.Context, email, passwordHash, name string) (int64, error) {
	result, err := r.queries.CreateAdmin(ctx, sqlc.CreateAdminParams{
		Email: email, PasswordHash: passwordHash, Name: sql.NullString{String: name, Valid: name != ""},
	})
	if err != nil {
		var myErr *gosqlmysql.MySQLError
		if errors.As(err, &myErr) && myErr.Number == mysqlErrDuplicateEntry {
			return 0, domain.ErrAdminExists
		}
		logger.Ctx(ctx).Error("Failed to insert voxhire admin", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

// GetAdminByEmail returns the admin with the given email, or
// domain.ErrAdminNotFound.
func (r *Repository) GetAdminByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	row, err := r.queries.GetAdminByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return &domain.Admin{ID: row.ID, Email: row.Email, Name: row.Name, PasswordHash: row.PasswordHash, CreatedAt: adminCreatedAt(row.CreatedAt)}, nil
}

// GetAdminByID returns the admin with the given id, or
// domain.ErrAdminNotFound.
func (r *Repository) GetAdminByID(ctx context.Context, id int64) (*domain.Admin, error) {
	row, err := r.queries.GetAdminByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return &domain.Admin{ID: row.ID, Email: row.Email, Name: row.Name, PasswordHash: row.PasswordHash, CreatedAt: adminCreatedAt(row.CreatedAt)}, nil
}

// ListAdmins returns all recruiter-admin accounts, newest first.
func (r *Repository) ListAdmins(ctx context.Context) ([]domain.Admin, error) {
	rows, err := r.queries.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}
	admins := make([]domain.Admin, 0, len(rows))
	for _, row := range rows {
		admins = append(admins, domain.Admin{ID: row.ID, Email: row.Email, Name: row.Name, CreatedAt: adminCreatedAt(row.CreatedAt)})
	}
	return admins, nil
}

// CountAdmins returns the number of admin accounts. Used to decide
// whether the default admin needs seeding.
func (r *Repository) CountAdmins(ctx context.Context) (int64, error) {
	return r.queries.CountAdmins(ctx)
}

// ── Openings (read from the shared careers table) ────────────────────

// ListOpenings returns the active job openings plus their application
// counts for the recruiter panel's position picker.
func (r *Repository) ListOpenings(ctx context.Context) ([]domain.Opening, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// cross-schema read from careers — admin openings + application counts
	rows, err := r.queries.ListOpenings(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	openings := make([]domain.Opening, 0, len(rows))
	for _, row := range rows {
		openings = append(openings, domain.Opening{
			ID: row.ID, Slug: row.Slug, Title: row.Title, Department: row.Department,
			Location: row.Location, EmploymentType: row.EmploymentType, WorkMode: row.WorkMode,
			Seniority: row.Seniority, Details: row.Details,
			SalaryMin: parseDecimalString(row.SalaryMin), SalaryMax: parseDecimalString(row.SalaryMax),
			SalaryCurrency: row.SalaryCurrency, IsActive: row.IsActive,
			NewCount: ifaceToInt64(row.NewCount), TotalCount: ifaceToInt64(row.TotalCount),
		})
	}
	return openings, nil
}

func (r *Repository) ListOpeningsPage(ctx context.Context, limit, offset int32) (domain.OpeningPage, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return domain.OpeningPage{}, err
	}
	query := `
SELECT c.id, COALESCE(c.slug, ''), c.title, COALESCE(c.department, ''),
       COALESCE(c.location, ''), COALESCE(c.employment_type, ''),
       COALESCE(c.work_mode, ''), COALESCE(c.seniority, ''), COALESCE(c.details, ''),
       COALESCE(c.salary_min, 0), COALESCE(c.salary_max, 0),
       COALESCE(c.salary_currency, 'USD'), c.is_active,
       COALESCE(SUM(CASE WHEN a.status = 'new' THEN 1 ELSE 0 END), 0),
       COALESCE(COUNT(a.id), 0)
FROM careers c
LEFT JOIN voxire_career_applications a ON a.job_id = c.id
WHERE c.owner_id = ?
GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,
         c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,
         c.salary_currency, c.is_active, c.created_at
ORDER BY c.is_active DESC, c.created_at DESC, c.id DESC
LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return domain.OpeningPage{}, err
	}
	defer rows.Close()
	var total int64
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM careers WHERE owner_id = ?`, ownerID).Scan(&total)
	if err != nil {
		return domain.OpeningPage{}, err
	}
	openings := make([]domain.Opening, 0)
	for rows.Next() {
		var row sqlc.ListOpeningsRow
		if err := rows.Scan(&row.ID, &row.Slug, &row.Title, &row.Department, &row.Location,
			&row.EmploymentType, &row.WorkMode, &row.Seniority, &row.Details,
			&row.SalaryMin, &row.SalaryMax, &row.SalaryCurrency, &row.IsActive,
			&row.NewCount, &row.TotalCount); err != nil {
			return domain.OpeningPage{}, err
		}
		openings = append(openings, domain.Opening{
			ID: row.ID, Slug: row.Slug, Title: row.Title, Department: row.Department,
			Location: row.Location, EmploymentType: row.EmploymentType, WorkMode: row.WorkMode,
			Seniority: row.Seniority, Details: row.Details,
			SalaryMin: parseDecimalString(row.SalaryMin), SalaryMax: parseDecimalString(row.SalaryMax),
			SalaryCurrency: row.SalaryCurrency, IsActive: row.IsActive,
			NewCount: ifaceToInt64(row.NewCount), TotalCount: ifaceToInt64(row.TotalCount),
		})
	}
	return domain.OpeningPage{Openings: openings, Total: total, Limit: limit, Offset: offset}, rows.Err()
}

func (r *Repository) GetOpening(ctx context.Context, id int64) (*domain.Opening, error) {
	return r.getCareerOpeningByID(ctx, id)
}

// GetAnalytics returns aggregate counts and activity for the recruiter overview.
func (r *Repository) GetAnalytics(ctx context.Context) (*domain.Analytics, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	analytics := &domain.Analytics{}

	if err := r.loadAnalyticsTotals(ctx, ownerID, &analytics.Totals); err != nil {
		return nil, err
	}

	statusRows, err := r.queries.AnalyticsApplicationsByStatus(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	analytics.ApplicationsByStatus = make([]domain.AnalyticsBucket, 0, len(statusRows))
	for _, row := range statusRows {
		analytics.ApplicationsByStatus = append(analytics.ApplicationsByStatus, analyticsBucketsFromRows(row.Bucket, row.Count))
	}

	cvRows, err := r.queries.AnalyticsCVStatusMix(ctx, ownerID)
	if err != nil && !isOptionalAnalyticsSchemaError(err) {
		return nil, err
	}
	if err == nil {
		analytics.CVStatusMix = make([]domain.AnalyticsBucket, 0, len(cvRows))
		for _, row := range cvRows {
			analytics.CVStatusMix = append(analytics.CVStatusMix, analyticsBucketsFromRows(row.Bucket, row.Count))
		}
	}

	rows, err := r.db.QueryContext(ctx, analyticsTaskStatusMixScopedSQL, ownerID)
	if err != nil && !isOptionalAnalyticsSchemaError(err) {
		return nil, err
	}
	if err == nil {
		defer rows.Close()
		analytics.TaskStatusMix = make([]domain.AnalyticsBucket, 0)
		for rows.Next() {
			var bucket string
			var count interface{}
			if err := rows.Scan(&bucket, &count); err != nil {
				return nil, err
			}
			analytics.TaskStatusMix = append(analytics.TaskStatusMix, analyticsBucketsFromRows(bucket, count))
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	analytics.ProfileFunctions, err = r.optionalAnalyticsProfileFunctions(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	serviceRows, err := r.queries.AnalyticsLeadServiceMix(ctx)
	if err != nil {
		return nil, err
	}
	analytics.LeadServiceMix = make([]domain.AnalyticsBucket, 0, len(serviceRows))
	for _, row := range serviceRows {
		analytics.LeadServiceMix = append(analytics.LeadServiceMix, analyticsBucketsFromRows(row.Bucket, row.Count))
	}

	timelineRows, err := r.queries.AnalyticsLeadTimelineMix(ctx)
	if err != nil {
		return nil, err
	}
	analytics.LeadTimelineMix = make([]domain.AnalyticsBucket, 0, len(timelineRows))
	for _, row := range timelineRows {
		analytics.LeadTimelineMix = append(analytics.LeadTimelineMix, analyticsBucketsFromRows(row.Bucket, row.Count))
	}

	analytics.RecentDays, err = r.analyticsRecentDays(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	topRows, err := r.queries.AnalyticsTopOpenings(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	analytics.TopOpenings = make([]domain.AnalyticsOpening, 0, len(topRows))
	for _, row := range topRows {
		opening := domain.AnalyticsOpening{
			ID: row.ID, Title: row.Title, IsActive: row.IsActive,
			TotalCount: ifaceToInt64(row.TotalCount), NewCount: ifaceToInt64(row.NewCount),
			RecentCount: ifaceToInt64(row.RecentCount), LastApplicationAt: ifaceToTimePtr(row.LastApplicationAt),
		}
		analytics.TopOpenings = append(analytics.TopOpenings, opening)
	}

	activityRows, err := r.queries.AnalyticsRecentActivity(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	analytics.RecentActivity = make([]domain.AnalyticsActivity, 0, len(activityRows))
	for _, row := range activityRows {
		analytics.RecentActivity = append(analytics.RecentActivity, domain.AnalyticsActivity{
			Kind: row.Kind, ID: row.ID, Primary: row.PrimaryText, Secondary: row.SecondaryText,
			Meta: row.Meta, CreatedAt: adminCreatedAt(row.CreatedAt),
		})
	}

	return analytics, nil
}

func (r *Repository) loadAnalyticsTotals(ctx context.Context, ownerID int64, totals *domain.AnalyticsTotals) error {
	jobTotals, err := r.queries.AnalyticsCareerJobTotals(ctx, ownerID)
	if err != nil {
		return err
	}
	totals.OpenJobs = ifaceToInt64(jobTotals.OpenJobs)
	totals.ClosedJobs = ifaceToInt64(jobTotals.ClosedJobs)

	appTotals, err := r.queries.AnalyticsApplicationTotals(ctx, ownerID)
	if err != nil {
		return err
	}
	totals.ApplicationsTotal = appTotals.ApplicationsTotal
	totals.Applications7d = ifaceToInt64(appTotals.Applications7d)
	totals.Applications30d = ifaceToInt64(appTotals.Applications30d)
	totals.NewApplications = ifaceToInt64(appTotals.NewApplications)
	cvCount := ifaceToInt64(appTotals.CvCount)
	if totals.ApplicationsTotal > 0 {
		totals.CVRate = float64(cvCount) / float64(totals.ApplicationsTotal)
	}
	if avg, ok := appTotals.AvgSalaryExpectation.(float64); ok {
		totals.AvgSalaryExpectation = avg
	} else if avgBytes, ok := appTotals.AvgSalaryExpectation.([]byte); ok {
		if f, parseErr := strconv.ParseFloat(string(avgBytes), 64); parseErr == nil {
			totals.AvgSalaryExpectation = f
		}
	}

	talentTotals, err := r.queries.AnalyticsTalentPoolTotals(ctx)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.TalentPoolTotal = talentTotals.TalentPoolTotal
		totals.TalentPool7d = ifaceToInt64(talentTotals.TalentPool7d)
		totals.TalentPool30d = ifaceToInt64(talentTotals.TalentPool30d)
	}

	cvTotals, err := r.queries.AnalyticsCVProcessingTotals(ctx, ownerID)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.CVProcessedTotal = ifaceToInt64(cvTotals.CvProcessedTotal)
		totals.CVErrorTotal = ifaceToInt64(cvTotals.CvErrorTotal)
		totals.CVPendingTotal = ifaceToInt64(cvTotals.CvPendingTotal)
	}

	leadTotals, err := r.queries.AnalyticsLeadTotals(ctx)
	if err != nil {
		return err
	}
	totals.LeadsTotal = leadTotals.LeadsTotal
	totals.Leads7d = ifaceToInt64(leadTotals.Leads7d)
	totals.Leads30d = ifaceToInt64(leadTotals.Leads30d)

	admins, err := r.CountAdmins(ctx)
	if err != nil {
		return err
	}
	totals.AdminsTotal = admins

	var taskTotals sqlc.AnalyticsOpenTaskTotalsRow
	err = r.db.QueryRowContext(ctx, analyticsOpenTaskTotalsScopedSQL, ownerID).Scan(
		&taskTotals.OpenTasks, &taskTotals.OverdueTasks,
	)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.OpenTasks = taskTotals.OpenTasks
		totals.OverdueTasks = ifaceToInt64(taskTotals.OverdueTasks)
	}

	var shortlistTotals sqlc.AnalyticsShortlistTotalsRow
	err = r.db.QueryRowContext(ctx, analyticsShortlistTotalsScopedSQL, ownerID).Scan(
		&shortlistTotals.ShortlistsTotal, &shortlistTotals.ShortlistedCandidates,
	)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	} else {
		totals.ShortlistsTotal = shortlistTotals.ShortlistsTotal
		totals.ShortlistedCandidates = shortlistTotals.ShortlistedCandidates
	}
	return nil
}

func (r *Repository) optionalAnalyticsProfileFunctions(ctx context.Context, ownerID int64) ([]domain.AnalyticsBucket, error) {
	buckets, err := r.analyticsProfileFunctions(ctx, ownerID)
	if err != nil && isOptionalAnalyticsSchemaError(err) {
		return []domain.AnalyticsBucket{}, nil
	}
	return buckets, err
}

func isOptionalAnalyticsSchemaError(err error) bool {
	var myErr *gosqlmysql.MySQLError
	if !errors.As(err, &myErr) {
		return false
	}
	switch myErr.Number {
	case 1054, // unknown column, usually a just-added optional analytics column
		1146: // table does not exist, usually a just-added optional analytics table
		return true
	default:
		return false
	}
}

func (r *Repository) analyticsProfileFunctions(ctx context.Context, ownerID int64) ([]domain.AnalyticsBucket, error) {
	functions := []string{"Operations", "Sales", "Customer Support", "Finance", "HR", "Marketing", "Product", "Design", "Engineering", "Data", "Healthcare", "Retail", "Logistics"}
	buckets := make([]domain.AnalyticsBucket, 0, len(functions))
	for _, fn := range functions {
		pattern := `%"` + fn + `"%`
		var count int64
		// cross-schema read from careers — profile function counts across career + talent applications
		if err := r.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			  FROM (
			        SELECT COALESCE(a.cv_profile_json, '') AS profile_json
			          FROM voxire_career_applications a
			          JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
			        UNION ALL
			        SELECT COALESCE(cv_profile_json, '') AS profile_json
			          FROM voxhire_talent_applications
			       ) profiles
				 WHERE profile_json LIKE ?`, ownerID, pattern).Scan(&count); err != nil {
			return nil, err
		}
		if count > 0 {
			buckets = append(buckets, domain.AnalyticsBucket{Key: strings.ToLower(strings.ReplaceAll(fn, " ", "_")), Count: count})
		}
	}
	sort.SliceStable(buckets, func(i, j int) bool {
		if buckets[i].Count == buckets[j].Count {
			return buckets[i].Key < buckets[j].Key
		}
		return buckets[i].Count > buckets[j].Count
	})
	return buckets, nil
}

func (r *Repository) analyticsRecentDays(ctx context.Context, ownerID int64) ([]domain.AnalyticsDay, error) {
	appRows, err := r.queries.AnalyticsApplicationDailyCounts(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	apps := make(map[string]int64, len(appRows))
	for _, row := range appRows {
		apps[row.Day] = ifaceToInt64(row.Count)
	}

	leadRows, err := r.queries.AnalyticsLeadDailyCounts(ctx)
	if err != nil {
		return nil, err
	}
	leads := make(map[string]int64, len(leadRows))
	for _, row := range leadRows {
		leads[row.Day] = ifaceToInt64(row.Count)
	}

	days := make([]domain.AnalyticsDay, 0, 14)
	start := time.Now().UTC().AddDate(0, 0, -13)
	for i := 0; i < 14; i++ {
		key := start.AddDate(0, 0, i).Format("2006-01-02")
		days = append(days, domain.AnalyticsDay{
			Date:         key,
			Applications: apps[key],
			Leads:        leads[key],
		})
	}
	return days, nil
}

// CreateCareer writes a new job posting to the shared careers table.
func (r *Repository) CreateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	slug, err := r.uniqueCareerSlug(ctx, ownerID, in.Title, 0)
	if err != nil {
		return nil, err
	}

	// cross-schema write to careers
	result, err := r.queries.CreateCareer(ctx, sqlc.CreateCareerParams{
		OwnerID: ownerID, Slug: slug, Title: in.Title,
		Department: sqlNullString(in.Department), Location: sqlNullString(in.Location),
		EmploymentType: sqlNullString(in.EmploymentType), WorkMode: sqlNullString(in.WorkMode),
		Seniority: sqlNullString(in.Seniority), Details: in.Details,
		SalaryMin: sqlNullDecimal(in.SalaryMin), SalaryMax: sqlNullDecimal(in.SalaryMax),
		SalaryCurrency: sqlNullString(in.SalaryCurrency),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire career", logger.Error(err))
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.getCareerOpeningByID(ctx, id)
}

// UpdateCareer writes editable job posting fields to the shared careers table.
func (r *Repository) UpdateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	slug, err := r.uniqueCareerSlug(ctx, ownerID, in.Title, in.ID)
	if err != nil {
		return nil, err
	}

	// cross-schema write to careers
	if err := r.queries.UpdateCareer(ctx, sqlc.UpdateCareerParams{
		Slug: slug, Title: in.Title,
		Department: sqlNullString(in.Department), Location: sqlNullString(in.Location),
		EmploymentType: sqlNullString(in.EmploymentType), WorkMode: sqlNullString(in.WorkMode),
		Seniority: sqlNullString(in.Seniority), Details: in.Details,
		SalaryMin: sqlNullDecimal(in.SalaryMin), SalaryMax: sqlNullDecimal(in.SalaryMax),
		SalaryCurrency: sqlNullString(in.SalaryCurrency),
		ID:             in.ID, OwnerID: ownerID,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to update voxhire career", logger.Error(err))
		return nil, err
	}
	return r.getCareerOpeningByID(ctx, in.ID)
}

// SetCareerActive closes or reopens a job posting in the shared careers table.
func (r *Repository) SetCareerActive(ctx context.Context, id int64, active bool) (*domain.Opening, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// cross-schema write to careers
	if err := r.queries.SetCareerActive(ctx, sqlc.SetCareerActiveParams{
		IsActive: active, IsPublished: active, ID: id, OwnerID: ownerID,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to update voxhire career active state", logger.Error(err))
		return nil, err
	}
	return r.getCareerOpeningByID(ctx, id)
}

func (r *Repository) getCareerOpeningByID(ctx context.Context, id int64) (*domain.Opening, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// cross-schema read from careers — one admin opening + application counts
	row, err := r.queries.GetCareerOpeningByID(ctx, sqlc.GetCareerOpeningByIDParams{
		OwnerID: ownerID, ID: id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCareerNotFound
		}
		return nil, err
	}
	return openingFromCareerRow(row), nil
}

// dynamic-query: squirrel
func (r *Repository) uniqueCareerSlug(ctx context.Context, ownerID int64, title string, excludeID int64) (string, error) {
	base := slugifyCareerTitle(title)
	for i := 0; i < 1000; i++ {
		slug := base
		if i > 0 {
			slug = fmt.Sprintf("%s-%d", base, i+1)
		}

		q := sq.
			Select("COUNT(*)").
			From("careers").
			Where(sq.Eq{"owner_id": ownerID}).
			Where(sq.Eq{"slug": slug})
		if excludeID > 0 {
			q = q.Where(sq.NotEq{"id": excludeID})
		}
		query, args, err := q.PlaceholderFormat(sq.Question).ToSql()
		if err != nil {
			return "", err
		}

		var count int64
		if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
			return "", err
		}
		if count == 0 {
			return slug, nil
		}
	}
	return "", fmt.Errorf("could not generate unique career slug for %q", title)
}

func slugifyCareerTitle(title string) string {
	title = strings.ToLower(strings.TrimSpace(title))
	var b strings.Builder
	lastDash := false
	for _, r := range title {
		isAlphaNum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if isAlphaNum {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		return "job"
	}
	return slug
}

// ── Company leads (Talk to us about a role) ──────────────────────────

// ListCompanyLeads returns company hiring inquiries newest first.
func (r *Repository) ListCompanyLeads(ctx context.Context) ([]domain.Lead, error) {
	rows, err := r.queries.ListCompanyLeads(ctx)
	if err != nil {
		return nil, err
	}
	leads := make([]domain.Lead, 0, len(rows))
	for _, row := range rows {
		leads = append(leads, domain.Lead{
			ID: row.ID, CompanyName: row.CompanyName, ContactName: row.ContactName,
			ContactTitle: row.ContactTitle, WorkEmail: row.WorkEmail, Phone: row.Phone,
			Role: row.Role, Headcount: row.Headcount, Timeline: row.Timeline, Service: row.Service,
			Message: row.Message, AgencyOptIn: row.AgencyOptIn, Status: row.Status,
			CreatedAt: adminCreatedAt(row.CreatedAt),
		})
	}
	if len(leads) > 500 {
		leads = leads[:500]
	}
	return leads, nil
}

// ── Applications ─────────────────────────────────────────────────────

// ListApplications returns applications for a single opening, filtered
// and sorted per the supplied filter, plus the total matching count.
//
// dynamic-query: squirrel
func (r *Repository) ListApplications(ctx context.Context, f domain.ApplicationFilter) ([]domain.Application, int64, error) {
	// Tenant-scope like every other application read (UpdateApplicationStatus,
	// GetApplicationCVKey, ...): restrict to jobs whose career belongs to the
	// authenticated admin, otherwise any recruiter token could enumerate
	// another owner's applicants by scanning job ids.
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}

	cols := `id, job_id, name, email, COALESCE(phone,''), COALESCE(location,''),
	         COALESCE(salary_expectation,0), COALESCE(salary_currency,'USD'), status,
	         COALESCE(work_history,''), COALESCE(education,''), COALESCE(experience_summary,''),
	         COALESCE(cover_letter,''), (cv_key IS NOT NULL AND cv_key <> '') AS has_cv,
	         COALESCE(cv_processing_status,'pending'), COALESCE(cv_processing_error,''), COALESCE(cv_summary,''),
	         COALESCE(cv_extracted_text,''), COALESCE(cv_profile_json,''), cv_processed_at, COALESCE(review_rating,0),
	         COALESCE(review_notes,''), COALESCE(next_step,''), reviewed_at, created_at`

	base := sq.Select().From("voxire_career_applications").
		Where(sq.Eq{"job_id": f.JobID}).
		Where("job_id IN (SELECT id FROM careers WHERE owner_id = ?)", ownerID)

	if loc := strings.TrimSpace(f.Location); loc != "" {
		base = base.Where(sq.Like{"location": "%" + loc + "%"})
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		base = base.Where(applicationSearchPredicate(q))
	}

	// Total count for pagination.
	countSQL, countArgs, err := base.Column("COUNT(*)").PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := base.Column(cols)
	switch f.Sort {
	case "salary_asc":
		listQ = listQ.OrderBy("salary_expectation IS NULL", "salary_expectation ASC", "created_at DESC")
	case "salary_desc":
		listQ = listQ.OrderBy("salary_expectation DESC", "created_at DESC")
	default: // "recent"
		listQ = listQ.OrderBy("created_at DESC")
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	listQ = listQ.Limit(uint64(limit)).Offset(uint64(f.Offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.

	sqlStr, args, err := listQ.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	apps := make([]domain.Application, 0)
	for rows.Next() {
		var a domain.Application
		var cvExtractedText string
		var cvProfileJSON string
		var cvProcessedAt sql.NullTime
		var reviewedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.JobID, &a.Name, &a.Email, &a.Phone, &a.Location,
			&a.SalaryExpectation, &a.SalaryCurrency, &a.Status, &a.WorkHistory, &a.Education,
			&a.ExperienceSummary, &a.CoverLetter, &a.HasCV, &a.CVProcessingStatus, &a.CVProcessingError,
			&a.CVSummary, &cvExtractedText, &cvProfileJSON, &cvProcessedAt, &a.ReviewRating, &a.ReviewNotes,
			&a.NextStep, &reviewedAt, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		a.CVSearchExcerpt = cvSearchExcerpt(f.Query, a.CVSummary, cvExtractedText)
		a.CVProfile = parseCVProfileJSON(cvProfileJSON)
		if cvProcessedAt.Valid {
			a.CVProcessedAt = &cvProcessedAt.Time
		}
		if reviewedAt.Valid {
			a.ReviewedAt = &reviewedAt.Time
		}
		apps = append(apps, a)
	}
	return apps, total, rows.Err()
}

func cvSearchExcerpt(query, summary, extractedText string) string {
	if excerpt := searchExcerpt(summary, query, 90); excerpt != "" {
		return excerpt
	}
	if excerpt := searchExcerpt(extractedText, query, 90); excerpt != "" {
		return excerpt
	}
	for _, term := range searchTerms(query) {
		if excerpt := searchExcerpt(summary, term, 90); excerpt != "" {
			return excerpt
		}
		if excerpt := searchExcerpt(extractedText, term, 90); excerpt != "" {
			return excerpt
		}
	}
	return ""
}

func applicationSearchPredicate(query string) sq.Sqlizer {
	raw := strings.TrimSpace(query)
	exact := applicationSearchLike("%" + raw + "%")
	terms := searchTerms(raw)
	if len(terms) <= 1 {
		return exact
	}

	allTerms := make(sq.And, 0, len(terms))
	for _, term := range terms {
		allTerms = append(allTerms, applicationSearchLike("%"+term+"%"))
	}
	return sq.Or{exact, allTerms}
}

func applicationSearchLike(pattern string) sq.Or {
	return sq.Or{
		sq.Like{"name": pattern},
		sq.Like{"email": pattern},
		sq.Like{"phone": pattern},
		sq.Like{"location": pattern},
		sq.Like{"work_history": pattern},
		sq.Like{"education": pattern},
		sq.Like{"experience_summary": pattern},
		sq.Like{"cover_letter": pattern},
		sq.Like{"cv_summary": pattern},
		sq.Like{"cv_extracted_text": pattern},
		sq.Like{"review_notes": pattern},
		sq.Like{"next_step": pattern},
	}
}

func searchTerms(query string) []string {
	parts := strings.FieldsFunc(query, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\t' || r == ' '
	})
	terms := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		term := strings.TrimSpace(part)
		if term == "" {
			continue
		}
		key := strings.ToLower(term)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}

func searchExcerpt(text, query string, radius int) string {
	clean := strings.Join(strings.Fields(text), " ")
	q := strings.TrimSpace(query)
	if clean == "" || q == "" {
		return ""
	}

	lowerClean := strings.ToLower(clean)
	lowerQuery := strings.ToLower(q)
	idx := strings.Index(lowerClean, lowerQuery)
	if idx < 0 {
		return ""
	}

	runes := []rune(clean)
	matchStart := len([]rune(clean[:idx]))
	matchLen := len([]rune(clean[idx : idx+len(q)]))
	start := matchStart - radius
	if start < 0 {
		start = 0
	}
	end := matchStart + matchLen + radius
	if end > len(runes) {
		end = len(runes)
	}

	excerpt := string(runes[start:end])
	if start > 0 {
		excerpt = "..." + excerpt
	}
	if end < len(runes) {
		excerpt += "..."
	}
	return excerpt
}

// UpdateApplicationStatus sets the status of a single application
// (e.g. "viewed", "new"). Returns domain.ErrApplicationNotFound if no
// row matched.
func (r *Repository) UpdateApplicationStatus(ctx context.Context, id int64, status string) error {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, updateApplicationStatusScopedSQL, ownerID, status, id); err != nil {
		return err
	}
	ok, err := r.applicationExists(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrApplicationNotFound
	}
	_ = r.recordApplicationEvent(ctx, id, "status", "Status changed", status)
	return nil
}

// UpdateApplicationReview stores recruiter-owned review state for a
// candidate application. Status is optional; an empty status preserves the
// current pipeline state.
func (r *Repository) UpdateApplicationReview(ctx context.Context, id int64, in domain.ApplicationReviewInput) error {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, updateApplicationReviewScopedSQL,
		ownerID, in.Status, in.Status, in.ReviewRating,
		sqlNullString(in.ReviewNotes), sqlNullString(in.NextStep), id,
	); err != nil {
		return err
	}
	exists, err := r.applicationExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return domain.ErrApplicationNotFound
	}
	detail := strings.TrimSpace(in.ReviewNotes)
	if detail == "" {
		detail = in.NextStep
	}
	_ = r.recordApplicationEvent(ctx, id, "review", "Recruiter review saved", detail)
	return nil
}

// GetApplicationCVKey returns the R2 object key of an application's CV,
// or domain.ErrApplicationNotFound. An empty string means no CV.
func (r *Repository) GetApplicationCVKey(ctx context.Context, id int64) (string, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return "", err
	}
	var key string
	err = r.db.QueryRowContext(ctx, getApplicationCVKeyScopedSQL, ownerID, id).Scan(&key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrApplicationNotFound
		}
		return "", err
	}
	return key, nil
}

// CreateCareerApplication inserts a public application submitted on
// hire.voxire.com into voxire_career_applications.
func (r *Repository) CreateCareerApplication(ctx context.Context, in domain.CareerApplicationInput) (int64, error) {
	result, err := r.queries.CreateCareerApplication(ctx, sqlc.CreateCareerApplicationParams{
		JobID: in.JobID, Name: in.Name, Email: in.Email,
		Phone: sqlNullString(in.Phone), Location: sqlNullString(in.Location),
		SalaryExpectation: sqlNullDecimal(in.SalaryExpectation),
		WorkHistory:       sqlNullString(in.WorkHistory), Education: sqlNullString(in.Education),
		ExperienceSummary: sqlNullString(in.ExperienceSummary), CoverLetter: sqlNullString(in.CoverLetter),
		CvKey: sqlNullString(in.CVKey),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire career application", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

const candidateUnionSQL = `
	SELECT 'career' AS source,
	       a.id,
	       a.id AS application_id,
	       a.job_id,
	       COALESCE(c.title, '') AS opening_title,
	       a.name,
	       a.email,
	       COALESCE(a.phone, '') AS phone,
	       COALESCE(a.location, '') AS location,
	       '' AS track,
	       '' AS current_role,
	       '' AS desired_roles,
	       COALESCE(a.status, 'new') AS status,
	       COALESCE(a.cv_processing_status, 'pending') AS cv_processing_status,
	       COALESCE(a.cv_processing_error, '') AS cv_processing_error,
	       COALESCE(a.cv_summary, '') AS cv_summary,
	       COALESCE(a.cv_extracted_text, '') AS cv_extracted_text,
	       COALESCE(a.cv_profile_json, '') AS cv_profile_json,
	       a.created_at,
	       a.cv_processed_at
	  FROM voxire_career_applications a
	  LEFT JOIN careers c ON c.id = a.job_id
	 WHERE c.owner_id = ?
	UNION ALL
	SELECT 'talent' AS source,
	       t.id,
	       0 AS application_id,
	       0 AS job_id,
	       'Talent pool' AS opening_title,
	       t.full_name AS name,
	       t.email,
	       COALESCE(t.phone, '') AS phone,
	       COALESCE(t.country_of_residence, '') AS location,
	       COALESCE(t.track, '') AS track,
	       COALESCE(t.current_role, '') AS current_role,
	       COALESCE(t.desired_roles, '') AS desired_roles,
	       COALESCE(t.status, 'new') AS status,
	       COALESCE(t.cv_processing_status, 'pending') AS cv_processing_status,
	       COALESCE(t.cv_processing_error, '') AS cv_processing_error,
	       COALESCE(t.cv_summary, '') AS cv_summary,
	       COALESCE(t.cv_extracted_text, '') AS cv_extracted_text,
	       COALESCE(t.cv_profile_json, '') AS cv_profile_json,
	       t.created_at,
	       t.cv_processed_at
	  FROM voxhire_talent_applications t`

// dynamic-query: squirrel
func (r *Repository) ListCandidates(ctx context.Context, f domain.CandidateFilter) ([]domain.Candidate, int64, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	where, args := candidateFilterWhere(f)
	countArgs := append([]any{ownerID}, args...)
	// #nosec G201 -- candidateUnionSQL is a package constant and where is built only from fixed predicates with bound values.
	// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query
	countSQL := "SELECT COUNT(*) FROM (" + candidateUnionSQL + ") candidates WHERE " + strings.Join(where, " AND ")

	var total int64
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := int(f.Limit)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := int(f.Offset)
	if offset < 0 {
		offset = 0
	}

	listArgs := append([]any{ownerID}, args...)
	listArgs = append(listArgs, limit, offset)
	// #nosec G202 -- candidateUnionSQL is a package constant and where is built only from fixed predicates with bound values.
	listSQL := `
		SELECT source, id, application_id, job_id, opening_title, name, email, phone, location,
		       track, current_role, desired_roles, status, cv_processing_status, cv_processing_error,
		       cv_summary, cv_extracted_text, cv_profile_json, created_at, cv_processed_at
		  FROM (` + candidateUnionSQL + `) candidates
		 WHERE ` + strings.Join(where, " AND ") + `
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	candidates := make([]domain.Candidate, 0, limit)
	for rows.Next() {
		var c domain.Candidate
		var cvExtractedText string
		var cvProfileJSON string
		var cvProcessedAt sql.NullTime
		if err := rows.Scan(&c.Source, &c.ID, &c.ApplicationID, &c.JobID, &c.OpeningTitle,
			&c.Name, &c.Email, &c.Phone, &c.Location, &c.Track, &c.CurrentRole, &c.DesiredRoles,
			&c.Status, &c.CVProcessingStatus, &c.CVProcessingError, &c.CVSummary,
			&cvExtractedText, &cvProfileJSON, &c.CreatedAt, &cvProcessedAt); err != nil {
			return nil, 0, err
		}
		c.CVSearchExcerpt = cvSearchExcerpt(f.Query, c.CVSummary, cvExtractedText)
		c.CVProfile = parseCVProfileJSON(cvProfileJSON)
		if cvProcessedAt.Valid {
			c.CVProcessedAt = &cvProcessedAt.Time
		}
		candidates = append(candidates, c)
	}
	return candidates, total, rows.Err()
}

func candidateFilterWhere(f domain.CandidateFilter) ([]string, []any) {
	where := []string{"1=1"}
	args := make([]any, 0)
	if source := strings.TrimSpace(f.Source); source != "" && source != "all" {
		where = append(where, "source = ?")
		args = append(args, source)
	}
	if status := strings.TrimSpace(f.CVStatus); status != "" && status != "all" {
		where = append(where, "cv_processing_status = ?")
		args = append(args, status)
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		for _, term := range searchTerms(q) {
			where = append(where, `LOWER(CONCAT_WS(' ', name, email, phone, location, opening_title, track, current_role, desired_roles, status, cv_summary, cv_extracted_text, cv_profile_json)) LIKE ?`)
			args = append(args, "%"+strings.ToLower(term)+"%")
		}
	}
	return where, args
}

// dynamic-query: squirrel
func (r *Repository) ListCVJobs(ctx context.Context, f domain.CandidateFilter) ([]domain.CVJob, error) {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	where, args := candidateFilterWhere(f)
	where = append(where, "cv_key <> ''")

	limit := int(f.Limit)
	if limit <= 0 {
		limit = 100
	}
	if limit > 250 {
		limit = 250
	}

	// #nosec G201 -- The query uses a package-fixed projection and a predicate list built only from fixed clauses with bound values.
	sqlStr := fmt.Sprintf(`
		SELECT source, id, name, email, role, cv_key, cv_processing_status, cv_processing_error,
		       cv_summary, cv_profile_json, created_at, cv_processed_at
		  FROM (
			SELECT 'career' AS source,
			       a.id,
			       a.name,
			       a.email,
			       COALESCE(c.title, '') AS role,
			       COALESCE(a.cv_key, '') AS cv_key,
			       COALESCE(a.cv_processing_status, 'pending') AS cv_processing_status,
			       COALESCE(a.cv_processing_error, '') AS cv_processing_error,
			       COALESCE(a.cv_summary, '') AS cv_summary,
			       COALESCE(a.cv_profile_json, '') AS cv_profile_json,
			       a.created_at,
			       a.cv_processed_at,
			       COALESCE(a.cv_extracted_text, '') AS cv_extracted_text,
			       COALESCE(a.phone, '') AS phone,
			       COALESCE(a.location, '') AS location,
			       COALESCE(a.status, 'new') AS status,
			       COALESCE(c.title, '') AS opening_title,
			       '' AS track,
			       '' AS current_role,
			       '' AS desired_roles
			  FROM voxire_career_applications a
			  LEFT JOIN careers c ON c.id = a.job_id
			 WHERE c.owner_id = ?
			UNION ALL
			SELECT 'talent' AS source,
			       t.id,
			       t.full_name AS name,
			       t.email,
			       COALESCE(NULLIF(t.current_role, ''), NULLIF(t.track, ''), 'Talent pool') AS role,
			       COALESCE(t.cv_key, '') AS cv_key,
			       COALESCE(t.cv_processing_status, 'pending') AS cv_processing_status,
			       COALESCE(t.cv_processing_error, '') AS cv_processing_error,
			       COALESCE(t.cv_summary, '') AS cv_summary,
			       COALESCE(t.cv_profile_json, '') AS cv_profile_json,
			       t.created_at,
			       t.cv_processed_at,
			       COALESCE(t.cv_extracted_text, '') AS cv_extracted_text,
			       COALESCE(t.phone, '') AS phone,
			       COALESCE(t.country_of_residence, '') AS location,
			       COALESCE(t.status, 'new') AS status,
			       'Talent pool' AS opening_title,
			       COALESCE(t.track, '') AS track,
			       COALESCE(t.current_role, '') AS current_role,
			       COALESCE(t.desired_roles, '') AS desired_roles
			  FROM voxhire_talent_applications t
		  ) cv_jobs
		 WHERE %s
		 ORDER BY
		   CASE cv_processing_status WHEN 'error' THEN 0 WHEN 'pending' THEN 1 ELSE 2 END,
		   created_at ASC
		 LIMIT ?`, strings.Join(where, " AND "))

	queryArgs := append([]any{ownerID}, args...)
	queryArgs = append(queryArgs, limit)
	rows, err := r.db.QueryContext(ctx, sqlStr, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := make([]domain.CVJob, 0, limit)
	for rows.Next() {
		var job domain.CVJob
		var cvProfileJSON string
		var cvProcessedAt sql.NullTime
		if err := rows.Scan(&job.Source, &job.ID, &job.Name, &job.Email, &job.Role, &job.CVKey,
			&job.CVProcessingStatus, &job.CVProcessingError, &job.CVSummary, &cvProfileJSON,
			&job.CreatedAt, &cvProcessedAt); err != nil {
			return nil, err
		}
		job.CVProfile = parseCVProfileJSON(cvProfileJSON)
		if cvProcessedAt.Valid {
			job.CVProcessedAt = &cvProcessedAt.Time
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

func (r *Repository) RetryCVProcessing(ctx context.Context, source string, id int64) error {
	ownerID, err := ownerIDFromContext(ctx)
	if err != nil {
		return err
	}
	var table string
	var query string
	switch source {
	case "talent":
		table = "voxhire_talent_applications"
		// nosemgrep: go.lang.security.audit.database.string-formatted-query.string-formatted-query
		query = fmt.Sprintf(`
			UPDATE %s
			   SET cv_processing_status = 'pending', cv_processing_error = NULL,
			       cv_extracted_text = NULL, cv_summary = NULL, cv_profile_json = NULL,
			       cv_processed_at = NULL
			 WHERE id = ? AND cv_key IS NOT NULL AND cv_key <> ''`, table)
	case "career":
		query = `
			UPDATE voxire_career_applications a
			JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
			   SET a.cv_processing_status = 'pending', a.cv_processing_error = NULL,
			       a.cv_extracted_text = NULL, a.cv_summary = NULL, a.cv_profile_json = NULL,
			       a.cv_processed_at = NULL
			 WHERE a.id = ? AND a.cv_key IS NOT NULL AND a.cv_key <> ''`
	default:
		return domain.ErrInvalidInput
	}
	args := []any{id}
	if source == "career" {
		args = []any{ownerID, id}
	}
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrApplicationNotFound
	}
	if source == "career" {
		_ = r.recordApplicationEvent(ctx, id, "cv_retry", "CV queued for reprocessing", "")
	}
	return nil
}

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
