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

const voxhireCareerOwnerID int64 = 1

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
	// cross-schema read from careers — admin openings + application counts
	rows, err := r.queries.ListOpenings(ctx, voxhireCareerOwnerID)
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

// GetAnalytics returns aggregate counts and activity for the recruiter overview.
func (r *Repository) GetAnalytics(ctx context.Context) (*domain.Analytics, error) {
	analytics := &domain.Analytics{}

	if err := r.loadAnalyticsTotals(ctx, &analytics.Totals); err != nil {
		return nil, err
	}

	var err error
	analytics.ApplicationsByStatus, err = r.analyticsBuckets(ctx, `
		SELECT LOWER(COALESCE(NULLIF(a.status, ''), 'new')) AS bucket, COUNT(*) AS count
		  FROM voxire_career_applications a
		  JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
		 GROUP BY bucket
		 ORDER BY count DESC, bucket ASC`, voxhireCareerOwnerID)
	if err != nil {
		return nil, err
	}

	analytics.CVStatusMix, err = r.optionalAnalyticsBuckets(ctx, `
		SELECT status AS bucket, COUNT(*) AS count
		  FROM (
		        SELECT LOWER(COALESCE(a.cv_processing_status, 'pending')) AS status
		          FROM voxire_career_applications a
		          JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
		         WHERE a.cv_key IS NOT NULL AND a.cv_key <> ''
		        UNION ALL
		        SELECT LOWER(COALESCE(cv_processing_status, 'pending')) AS status
		          FROM voxhire_talent_applications
		         WHERE cv_key IS NOT NULL AND cv_key <> ''
		       ) statuses
		 GROUP BY status
		 ORDER BY count DESC, status ASC`, voxhireCareerOwnerID)
	if err != nil {
		return nil, err
	}

	analytics.TaskStatusMix, err = r.optionalAnalyticsBuckets(ctx, `
		SELECT LOWER(COALESCE(NULLIF(status, ''), 'open')) AS bucket, COUNT(*) AS count
		  FROM voxhire_follow_up_tasks
		 GROUP BY bucket
		 ORDER BY count DESC, bucket ASC`)
	if err != nil {
		return nil, err
	}

	analytics.ProfileFunctions, err = r.optionalAnalyticsProfileFunctions(ctx)
	if err != nil {
		return nil, err
	}

	analytics.LeadServiceMix, err = r.analyticsBuckets(ctx, `
		SELECT LOWER(COALESCE(NULLIF(service, ''), 'unsure')) AS bucket, COUNT(*) AS count
		  FROM voxhire_company_leads
		 GROUP BY bucket
		 ORDER BY count DESC, bucket ASC`)
	if err != nil {
		return nil, err
	}

	analytics.LeadTimelineMix, err = r.analyticsBuckets(ctx, `
		SELECT LOWER(COALESCE(NULLIF(timeline, ''), 'unknown')) AS bucket, COUNT(*) AS count
		  FROM voxhire_company_leads
		 GROUP BY bucket
		 ORDER BY count DESC, bucket ASC`)
	if err != nil {
		return nil, err
	}

	analytics.RecentDays, err = r.analyticsRecentDays(ctx)
	if err != nil {
		return nil, err
	}

	analytics.TopOpenings, err = r.analyticsTopOpenings(ctx)
	if err != nil {
		return nil, err
	}

	analytics.RecentActivity, err = r.analyticsRecentActivity(ctx)
	if err != nil {
		return nil, err
	}

	return analytics, nil
}

func (r *Repository) loadAnalyticsTotals(ctx context.Context, totals *domain.AnalyticsTotals) error {
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN is_active = 0 THEN 1 ELSE 0 END), 0)
		  FROM careers
		 WHERE owner_id = ?`, voxhireCareerOwnerID).Scan(&totals.OpenJobs, &totals.ClosedJobs)
	if err != nil {
		return err
	}

	var cvCount int64
	var avgSalary sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN LOWER(COALESCE(a.status, 'new')) = 'new' THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN a.cv_key IS NOT NULL AND a.cv_key <> '' THEN 1 ELSE 0 END), 0),
		       AVG(NULLIF(a.salary_expectation, 0))
		  FROM voxire_career_applications a
		  JOIN careers c ON c.id = a.job_id AND c.owner_id = ?`, voxhireCareerOwnerID).Scan(
		&totals.ApplicationsTotal,
		&totals.Applications7d,
		&totals.Applications30d,
		&totals.NewApplications,
		&cvCount,
		&avgSalary,
	)
	if err != nil {
		return err
	}
	if totals.ApplicationsTotal > 0 {
		totals.CVRate = float64(cvCount) / float64(totals.ApplicationsTotal)
	}
	if avgSalary.Valid {
		totals.AvgSalaryExpectation = avgSalary.Float64
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0)
		  FROM voxhire_talent_applications`).Scan(&totals.TalentPoolTotal, &totals.TalentPool7d, &totals.TalentPool30d)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE WHEN status = 'processed' THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN status NOT IN ('processed', 'error') THEN 1 ELSE 0 END), 0)
		  FROM (
		        SELECT COALESCE(a.cv_processing_status, 'pending') AS status
		          FROM voxire_career_applications a
		          JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
		         WHERE a.cv_key IS NOT NULL AND a.cv_key <> ''
		        UNION ALL
		        SELECT COALESCE(cv_processing_status, 'pending') AS status
		          FROM voxhire_talent_applications
		         WHERE cv_key IS NOT NULL AND cv_key <> ''
		       ) cv_statuses`, voxhireCareerOwnerID).Scan(
		&totals.CVProcessedTotal,
		&totals.CVErrorTotal,
		&totals.CVPendingTotal,
	)
	if err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0),
		       COALESCE(SUM(CASE WHEN created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0)
		  FROM voxhire_company_leads`).Scan(&totals.LeadsTotal, &totals.Leads7d, &totals.Leads30d)
	if err != nil {
		return err
	}

	admins, err := r.CountAdmins(ctx)
	if err != nil {
		return err
	}
	totals.AdminsTotal = admins

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*),
		       COALESCE(SUM(CASE WHEN status = 'open' AND due_at IS NOT NULL AND due_at < UTC_TIMESTAMP() THEN 1 ELSE 0 END), 0)
		  FROM voxhire_follow_up_tasks
		 WHERE status <> 'done'`).Scan(&totals.OpenTasks, &totals.OverdueTasks); err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT
		  (SELECT COUNT(*) FROM voxhire_shortlists),
		  (SELECT COUNT(*) FROM voxhire_shortlist_items)
	`).Scan(&totals.ShortlistsTotal, &totals.ShortlistedCandidates); err != nil {
		if !isOptionalAnalyticsSchemaError(err) {
			return err
		}
	}
	return nil
}

func (r *Repository) analyticsBuckets(ctx context.Context, query string, args ...any) ([]domain.AnalyticsBucket, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	buckets := make([]domain.AnalyticsBucket, 0)
	for rows.Next() {
		var bucket domain.AnalyticsBucket
		if err := rows.Scan(&bucket.Key, &bucket.Count); err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}
	return buckets, rows.Err()
}

func (r *Repository) optionalAnalyticsBuckets(ctx context.Context, query string, args ...any) ([]domain.AnalyticsBucket, error) {
	buckets, err := r.analyticsBuckets(ctx, query, args...)
	if err != nil && isOptionalAnalyticsSchemaError(err) {
		return []domain.AnalyticsBucket{}, nil
	}
	return buckets, err
}

func (r *Repository) optionalAnalyticsProfileFunctions(ctx context.Context) ([]domain.AnalyticsBucket, error) {
	buckets, err := r.analyticsProfileFunctions(ctx)
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

func (r *Repository) analyticsProfileFunctions(ctx context.Context) ([]domain.AnalyticsBucket, error) {
	functions := []string{"Operations", "Sales", "Customer Support", "Finance", "HR", "Marketing", "Product", "Design", "Engineering", "Data", "Healthcare", "Retail", "Logistics"}
	buckets := make([]domain.AnalyticsBucket, 0, len(functions))
	for _, fn := range functions {
		pattern := `%"` + fn + `"%`
		var count int64
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
			 WHERE profile_json LIKE ?`, voxhireCareerOwnerID, pattern).Scan(&count); err != nil {
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

func (r *Repository) analyticsRecentDays(ctx context.Context) ([]domain.AnalyticsDay, error) {
	apps, err := r.analyticsDailyCounts(ctx, `
		SELECT DATE_FORMAT(a.created_at, '%Y-%m-%d') AS day, COUNT(*) AS count
		  FROM voxire_career_applications a
		  JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
		 WHERE a.created_at >= DATE_SUB(UTC_DATE(), INTERVAL 13 DAY)
		 GROUP BY day`, voxhireCareerOwnerID)
	if err != nil {
		return nil, err
	}

	leads, err := r.analyticsDailyCounts(ctx, `
		SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS count
		  FROM voxhire_company_leads
		 WHERE created_at >= DATE_SUB(UTC_DATE(), INTERVAL 13 DAY)
		 GROUP BY day`)
	if err != nil {
		return nil, err
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

func (r *Repository) analyticsDailyCounts(ctx context.Context, query string, args ...any) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int64)
	for rows.Next() {
		var day string
		var count int64
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		counts[day] = count
	}
	return counts, rows.Err()
}

func (r *Repository) analyticsTopOpenings(ctx context.Context) ([]domain.AnalyticsOpening, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id,
		       c.title,
		       c.is_active,
		       COALESCE(COUNT(a.id), 0) AS total_count,
		       COALESCE(SUM(CASE WHEN LOWER(COALESCE(a.status, 'new')) = 'new' THEN 1 ELSE 0 END), 0) AS new_count,
		       COALESCE(SUM(CASE WHEN a.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL 30 DAY) THEN 1 ELSE 0 END), 0) AS recent_count,
		       MAX(a.created_at) AS last_application_at
		  FROM careers c
		  LEFT JOIN voxire_career_applications a ON a.job_id = c.id
		 WHERE c.owner_id = ?
		 GROUP BY c.id, c.title, c.is_active, c.created_at
		 ORDER BY recent_count DESC, total_count DESC, c.created_at DESC
		 LIMIT 6`, voxhireCareerOwnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	openings := make([]domain.AnalyticsOpening, 0)
	for rows.Next() {
		var opening domain.AnalyticsOpening
		var last sql.NullTime
		if err := rows.Scan(&opening.ID, &opening.Title, &opening.IsActive, &opening.TotalCount,
			&opening.NewCount, &opening.RecentCount, &last); err != nil {
			return nil, err
		}
		if last.Valid {
			t := last.Time
			opening.LastApplicationAt = &t
		}
		openings = append(openings, opening)
	}
	return openings, rows.Err()
}

func (r *Repository) analyticsRecentActivity(ctx context.Context) ([]domain.AnalyticsActivity, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT kind, id, primary_text, secondary_text, meta, created_at
		  FROM (
		        SELECT 'application' AS kind,
		               a.id AS id,
		               c.title AS primary_text,
		               a.name AS secondary_text,
		               COALESCE(a.status, 'new') AS meta,
		               a.created_at AS created_at
		          FROM voxire_career_applications a
		          JOIN careers c ON c.id = a.job_id AND c.owner_id = ?
		        UNION ALL
		        SELECT 'lead' AS kind,
		               id,
		               company_name AS primary_text,
		               COALESCE(role, '') AS secondary_text,
		               COALESCE(timeline, '') AS meta,
		               created_at
		          FROM voxhire_company_leads
		       ) recent
		 ORDER BY created_at DESC
		 LIMIT 10`, voxhireCareerOwnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	activity := make([]domain.AnalyticsActivity, 0)
	for rows.Next() {
		var item domain.AnalyticsActivity
		if err := rows.Scan(&item.Kind, &item.ID, &item.Primary, &item.Secondary, &item.Meta, &item.CreatedAt); err != nil {
			return nil, err
		}
		activity = append(activity, item)
	}
	return activity, rows.Err()
}

// CreateCareer writes a new job posting to the shared careers table.
func (r *Repository) CreateCareer(ctx context.Context, in domain.CareerInput) (*domain.Opening, error) {
	slug, err := r.uniqueCareerSlug(ctx, in.Title, 0)
	if err != nil {
		return nil, err
	}

	// cross-schema write to careers
	query, args, err := sq.
		Insert("careers").
		Columns("owner_id", "slug", "title", "department", "location", "employment_type",
			"work_mode", "seniority", "details", "salary_min", "salary_max", "salary_currency",
			"is_active", "is_published").
		Values(voxhireCareerOwnerID, slug, in.Title, nullStr(in.Department), nullStr(in.Location),
			nullStr(in.EmploymentType), nullStr(in.WorkMode), nullStr(in.Seniority),
			in.Details, nullFloat(in.SalaryMin), nullFloat(in.SalaryMax), nullStr(in.SalaryCurrency),
			true, true).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
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
	slug, err := r.uniqueCareerSlug(ctx, in.Title, in.ID)
	if err != nil {
		return nil, err
	}

	// cross-schema write to careers
	query, args, err := sq.
		Update("careers").
		Set("slug", slug).
		Set("title", in.Title).
		Set("department", nullStr(in.Department)).
		Set("location", nullStr(in.Location)).
		Set("employment_type", nullStr(in.EmploymentType)).
		Set("work_mode", nullStr(in.WorkMode)).
		Set("seniority", nullStr(in.Seniority)).
		Set("details", in.Details).
		Set("salary_min", nullFloat(in.SalaryMin)).
		Set("salary_max", nullFloat(in.SalaryMax)).
		Set("salary_currency", nullStr(in.SalaryCurrency)).
		Where(sq.Eq{"id": in.ID}).
		Where(sq.Eq{"owner_id": voxhireCareerOwnerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		logger.Ctx(ctx).Error("Failed to update voxhire career", logger.Error(err))
		return nil, err
	}
	return r.getCareerOpeningByID(ctx, in.ID)
}

// SetCareerActive closes or reopens a job posting in the shared careers table.
func (r *Repository) SetCareerActive(ctx context.Context, id int64, active bool) (*domain.Opening, error) {
	// cross-schema write to careers
	query, args, err := sq.
		Update("careers").
		Set("is_active", active).
		Set("is_published", active).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"owner_id": voxhireCareerOwnerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		logger.Ctx(ctx).Error("Failed to update voxhire career active state", logger.Error(err))
		return nil, err
	}
	return r.getCareerOpeningByID(ctx, id)
}

func (r *Repository) getCareerOpeningByID(ctx context.Context, id int64) (*domain.Opening, error) {
	// cross-schema read from careers — one admin opening + application counts
	const query = `
		SELECT c.id,
		       COALESCE(c.slug, ''),
		       c.title,
		       COALESCE(c.department, ''),
		       COALESCE(c.location, ''),
		       COALESCE(c.employment_type, ''),
		       COALESCE(c.work_mode, ''),
		       COALESCE(c.seniority, ''),
		       COALESCE(c.details, ''),
		       COALESCE(c.salary_min, 0),
		       COALESCE(c.salary_max, 0),
		       COALESCE(c.salary_currency, 'USD'),
		       c.is_active,
		       COALESCE(SUM(CASE WHEN a.status = 'new' THEN 1 ELSE 0 END), 0) AS new_count,
		       COALESCE(COUNT(a.id), 0) AS total_count
		  FROM careers c
		  LEFT JOIN voxire_career_applications a ON a.job_id = c.id
		 WHERE c.owner_id = ? AND c.id = ?
		 GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,
		          c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,
		          c.salary_currency, c.is_active`

	var o domain.Opening
	err := r.db.QueryRowContext(ctx, query, voxhireCareerOwnerID, id).Scan(
		&o.ID, &o.Slug, &o.Title, &o.Department, &o.Location,
		&o.EmploymentType, &o.WorkMode, &o.Seniority, &o.Details, &o.SalaryMin,
		&o.SalaryMax, &o.SalaryCurrency, &o.IsActive, &o.NewCount, &o.TotalCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrCareerNotFound
		}
		return nil, err
	}
	return &o, nil
}

func (r *Repository) uniqueCareerSlug(ctx context.Context, title string, excludeID int64) (string, error) {
	base := slugifyCareerTitle(title)
	for i := 0; i < 1000; i++ {
		slug := base
		if i > 0 {
			slug = fmt.Sprintf("%s-%d", base, i+1)
		}

		q := sq.
			Select("COUNT(*)").
			From("careers").
			Where(sq.Eq{"owner_id": voxhireCareerOwnerID}).
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
func (r *Repository) ListApplications(ctx context.Context, f domain.ApplicationFilter) ([]domain.Application, int64, error) {
	cols := `id, job_id, name, email, COALESCE(phone,''), COALESCE(location,''),
	         COALESCE(salary_expectation,0), COALESCE(salary_currency,'USD'), status,
	         COALESCE(work_history,''), COALESCE(education,''), COALESCE(experience_summary,''),
	         COALESCE(cover_letter,''), (cv_key IS NOT NULL AND cv_key <> '') AS has_cv,
	         COALESCE(cv_processing_status,'pending'), COALESCE(cv_processing_error,''), COALESCE(cv_summary,''),
	         COALESCE(cv_extracted_text,''), COALESCE(cv_profile_json,''), cv_processed_at, COALESCE(review_rating,0),
	         COALESCE(review_notes,''), COALESCE(next_step,''), reviewed_at, created_at`

	base := sq.Select().From("voxire_career_applications").Where(sq.Eq{"job_id": f.JobID})

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
	listQ = listQ.Limit(uint64(limit)).Offset(uint64(f.Offset))

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
	if err := r.queries.UpdateApplicationStatus(ctx, sqlc.UpdateApplicationStatusParams{Status: status, ID: id}); err != nil {
		return err
	}
	n, err := r.queries.CountCareerApplications(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return domain.ErrApplicationNotFound
	}
	_ = r.recordApplicationEvent(ctx, id, "status", "Status changed", status)
	return nil
}

// UpdateApplicationReview stores recruiter-owned review state for a
// candidate application. Status is optional; an empty status preserves the
// current pipeline state.
func (r *Repository) UpdateApplicationReview(ctx context.Context, id int64, in domain.ApplicationReviewInput) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE voxire_career_applications
		   SET status = CASE WHEN ? <> '' THEN ? ELSE status END,
		       review_rating = ?,
		       review_notes = ?,
		       next_step = ?,
		       reviewed_at = CURRENT_TIMESTAMP
		 WHERE id = ?
	`, in.Status, in.Status, in.ReviewRating, nullStr(in.ReviewNotes), nullStr(in.NextStep), id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voxire_career_applications WHERE id = ?`, id).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			return nil
		}
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
	key, err := r.queries.GetApplicationCVKey(ctx, id)
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
	builder := sq.
		Insert("voxire_career_applications").
		Columns("job_id", "name", "email", "phone", "location", "salary_expectation",
			"work_history", "education", "experience_summary", "cover_letter", "cv_key", "status")

	var salary any
	if in.SalaryExpectation > 0 {
		salary = in.SalaryExpectation
	}

	query, args, err := builder.Values(
		in.JobID, in.Name, in.Email, in.Phone, in.Location, salary,
		in.WorkHistory, in.Education, in.ExperienceSummary, in.CoverLetter, in.CVKey, "new",
	).PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		return 0, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
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

func (r *Repository) ListCandidates(ctx context.Context, f domain.CandidateFilter) ([]domain.Candidate, int64, error) {
	where, args := candidateFilterWhere(f)
	countArgs := append([]any{voxhireCareerOwnerID}, args...)
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) candidates WHERE %s", candidateUnionSQL, strings.Join(where, " AND "))

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

	listArgs := append([]any{voxhireCareerOwnerID}, args...)
	listArgs = append(listArgs, limit, offset)
	listSQL := fmt.Sprintf(`
		SELECT source, id, application_id, job_id, opening_title, name, email, phone, location,
		       track, current_role, desired_roles, status, cv_processing_status, cv_processing_error,
		       cv_summary, cv_extracted_text, cv_profile_json, created_at, cv_processed_at
		  FROM (%s) candidates
		 WHERE %s
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`, candidateUnionSQL, strings.Join(where, " AND "))

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

func (r *Repository) ListCVJobs(ctx context.Context, f domain.CandidateFilter) ([]domain.CVJob, error) {
	where, args := candidateFilterWhere(f)
	where = append(where, "cv_key <> ''")

	limit := int(f.Limit)
	if limit <= 0 {
		limit = 100
	}
	if limit > 250 {
		limit = 250
	}

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

	queryArgs := append([]any{voxhireCareerOwnerID}, args...)
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
	table := ""
	switch source {
	case "talent":
		table = "voxhire_talent_applications"
	case "career":
		table = "voxire_career_applications"
	default:
		return domain.ErrInvalidInput
	}
	res, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		   SET cv_processing_status = 'pending',
		       cv_processing_error = NULL,
		       cv_extracted_text = NULL,
		       cv_summary = NULL,
		       cv_profile_json = NULL,
		       cv_processed_at = NULL
		 WHERE id = ?
		   AND cv_key IS NOT NULL
		   AND cv_key <> ''`, table), id)
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
	where := ""
	args := make([]any, 0)
	if status != "" && status != "all" {
		where = "WHERE t.status = ?"
		args = append(args, status)
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, ''), a.name, a.email,
		       t.title, t.status, t.due_at, t.completed_at, t.created_at
		  FROM voxhire_follow_up_tasks t
		  JOIN voxire_career_applications a ON a.id = t.application_id
		  LEFT JOIN careers c ON c.id = a.job_id
		  %s
		 ORDER BY
		   CASE t.status WHEN 'open' THEN 0 WHEN 'done' THEN 2 ELSE 1 END,
		   t.due_at IS NULL,
		   t.due_at ASC,
		   t.created_at DESC`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
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
	query := `
		INSERT INTO voxhire_follow_up_tasks (application_id, title, status, due_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			status = VALUES(status),
			due_at = VALUES(due_at),
			completed_at = CASE
				WHEN VALUES(status) = 'done' THEN COALESCE(completed_at, CURRENT_TIMESTAMP)
				ELSE NULL
			END,
			id = LAST_INSERT_ID(id)`
	res, err := r.db.ExecContext(ctx, query, in.ApplicationID, in.Title, status, in.DueAt)
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
	res, err := r.db.ExecContext(ctx, `
		UPDATE voxhire_follow_up_tasks
		   SET title = CASE WHEN ? <> '' THEN ? ELSE title END,
		       status = CASE WHEN ? <> '' THEN ? ELSE status END,
		       due_at = ?,
		       completed_at = CASE
		         WHEN ? = 'done' THEN COALESCE(completed_at, CURRENT_TIMESTAMP)
		         WHEN ? = 'open' THEN NULL
		         ELSE completed_at
		       END
		 WHERE id = ?`,
		in.Title, in.Title, in.Status, in.Status, in.DueAt, in.Status, in.Status, in.ID)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		var exists int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voxhire_follow_up_tasks WHERE id = ?`, in.ID).Scan(&exists); err != nil {
			return nil, err
		}
		if exists == 0 {
			return nil, domain.ErrApplicationNotFound
		}
	}
	task, err := r.getTaskByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	_ = r.recordApplicationEvent(ctx, task.ApplicationID, "task", "Follow-up task updated", task.Title)
	return task, nil
}

func (r *Repository) getTaskByID(ctx context.Context, id int64) (*domain.Task, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.application_id, a.job_id, COALESCE(c.title, ''), a.name, a.email,
		       t.title, t.status, t.due_at, t.completed_at, t.created_at
		  FROM voxhire_follow_up_tasks t
		  JOIN voxire_career_applications a ON a.id = t.application_id
		  LEFT JOIN careers c ON c.id = a.job_id
		 WHERE t.id = ?`, id)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, err
	}
	return &task, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (domain.Task, error) {
	var task domain.Task
	var dueAt sql.NullTime
	var completedAt sql.NullTime
	if err := scanner.Scan(&task.ID, &task.ApplicationID, &task.JobID, &task.OpeningTitle,
		&task.CandidateName, &task.CandidateEmail, &task.Title, &task.Status,
		&dueAt, &completedAt, &task.CreatedAt); err != nil {
		return task, err
	}
	if dueAt.Valid {
		task.DueAt = &dueAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}
	return task, nil
}

func (r *Repository) ListShortlists(ctx context.Context, jobID int64) ([]domain.Shortlist, error) {
	where := ""
	args := make([]any, 0)
	if jobID > 0 {
		where = "WHERE s.job_id = ?"
		args = append(args, jobID)
	}
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT s.id, s.job_id, COALESCE(c.title, ''), s.name, s.status,
		       COALESCE(s.notes, ''), s.created_at, s.updated_at
		  FROM voxhire_shortlists s
		  LEFT JOIN careers c ON c.id = s.job_id
		  %s
		 ORDER BY s.updated_at DESC`, where), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shortlists := make([]domain.Shortlist, 0)
	for rows.Next() {
		var shortlist domain.Shortlist
		if err := rows.Scan(&shortlist.ID, &shortlist.JobID, &shortlist.JobTitle, &shortlist.Name,
			&shortlist.Status, &shortlist.Notes, &shortlist.CreatedAt, &shortlist.UpdatedAt); err != nil {
			return nil, err
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO voxhire_shortlists (job_id, name, status, notes)
		VALUES (?, ?, ?, ?)`, in.JobID, in.Name, in.Status, nullStr(in.Notes))
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO voxhire_shortlist_items (shortlist_id, application_id, sort_order, notes)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			sort_order = VALUES(sort_order),
			notes = VALUES(notes),
			id = LAST_INSERT_ID(id)`, in.ShortlistID, in.ApplicationID, in.Rank, nullStr(in.Notes))
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT i.id, i.shortlist_id, i.application_id, a.name, a.email, COALESCE(a.status, 'new'),
		       i.sort_order, COALESCE(i.notes, ''), COALESCE(a.cv_profile_json, ''), i.created_at
		  FROM voxhire_shortlist_items i
		  JOIN voxire_career_applications a ON a.id = i.application_id
		 WHERE i.shortlist_id = ?
		 ORDER BY i.sort_order ASC, i.created_at ASC`, shortlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]domain.ShortlistItem, 0)
	for rows.Next() {
		item, err := scanShortlistItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) getShortlistItemByID(ctx context.Context, id int64) (*domain.ShortlistItem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT i.id, i.shortlist_id, i.application_id, a.name, a.email, COALESCE(a.status, 'new'),
		       i.sort_order, COALESCE(i.notes, ''), COALESCE(a.cv_profile_json, ''), i.created_at
		  FROM voxhire_shortlist_items i
		  JOIN voxire_career_applications a ON a.id = i.application_id
		 WHERE i.id = ?`, id)
	item, err := scanShortlistItem(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

type shortlistItemScanner interface {
	Scan(dest ...any) error
}

func scanShortlistItem(scanner shortlistItemScanner) (domain.ShortlistItem, error) {
	var item domain.ShortlistItem
	var profileJSON string
	if err := scanner.Scan(&item.ID, &item.ShortlistID, &item.ApplicationID, &item.CandidateName,
		&item.CandidateEmail, &item.Status, &item.Rank, &item.Notes, &profileJSON, &item.CreatedAt); err != nil {
		return item, err
	}
	item.CVProfile = parseCVProfileJSON(profileJSON)
	return item, nil
}

func (r *Repository) ListApplicationTimeline(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	var createdAt time.Time
	var cvStatus string
	var cvError string
	var cvProcessedAt sql.NullTime
	var reviewedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT created_at, COALESCE(cv_processing_status, 'pending'), COALESCE(cv_processing_error, ''),
		       cv_processed_at, reviewed_at
		  FROM voxire_career_applications
		 WHERE id = ?`, id).Scan(&createdAt, &cvStatus, &cvError, &cvProcessedAt, &reviewedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, err
	}

	events := []domain.TimelineEvent{{
		Kind:      "application",
		Title:     "Application submitted",
		CreatedAt: createdAt,
	}}
	if cvProcessedAt.Valid {
		title := "CV processed"
		if cvStatus == "error" {
			title = "CV extraction failed"
		}
		events = append(events, domain.TimelineEvent{
			Kind:      "cv",
			Title:     title,
			Detail:    cvError,
			CreatedAt: cvProcessedAt.Time,
		})
	}
	if reviewedAt.Valid {
		events = append(events, domain.TimelineEvent{
			Kind:      "review",
			Title:     "Recruiter review saved",
			CreatedAt: reviewedAt.Time,
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
	rows, err := r.db.QueryContext(ctx, `
		SELECT kind, title, COALESCE(detail, ''), created_at
		  FROM voxhire_application_events
		 WHERE application_id = ?
		 ORDER BY created_at DESC`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var event domain.TimelineEvent
		if err := rows.Scan(&event.Kind, &event.Title, &event.Detail, &event.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *Repository) listTaskTimelineEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT title, status, created_at, completed_at
		  FROM voxhire_follow_up_tasks
		 WHERE application_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var title, status string
		var createdAt time.Time
		var completedAt sql.NullTime
		if err := rows.Scan(&title, &status, &createdAt, &completedAt); err != nil {
			return nil, err
		}
		events = append(events, domain.TimelineEvent{Kind: "task", Title: "Task opened", Detail: title, CreatedAt: createdAt})
		if completedAt.Valid {
			events = append(events, domain.TimelineEvent{Kind: "task", Title: "Task completed", Detail: title, CreatedAt: completedAt.Time})
		} else if status != "open" {
			events[len(events)-1].Title = "Task " + status
		}
	}
	return events, rows.Err()
}

func (r *Repository) listShortlistTimelineEvents(ctx context.Context, id int64) ([]domain.TimelineEvent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT s.name, i.sort_order, i.created_at
		  FROM voxhire_shortlist_items i
		  JOIN voxhire_shortlists s ON s.id = i.shortlist_id
		 WHERE i.application_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]domain.TimelineEvent, 0)
	for rows.Next() {
		var name string
		var rank int
		var createdAt time.Time
		if err := rows.Scan(&name, &rank, &createdAt); err != nil {
			return nil, err
		}
		events = append(events, domain.TimelineEvent{
			Kind:      "shortlist",
			Title:     "Added to shortlist",
			Detail:    fmt.Sprintf("%s · rank %d", name, rank),
			CreatedAt: createdAt,
		})
	}
	return events, rows.Err()
}

func (r *Repository) applicationExists(ctx context.Context, id int64) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voxire_career_applications WHERE id = ?`, id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) shortlistExists(ctx context.Context, id int64) (bool, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voxhire_shortlists WHERE id = ?`, id).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) recordApplicationEvent(ctx context.Context, applicationID int64, kind, title, detail string) error {
	if applicationID <= 0 || kind == "" || title == "" {
		return nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO voxhire_application_events (application_id, kind, title, detail)
		VALUES (?, ?, ?, ?)`, applicationID, kind, title, nullStr(detail))
	return err
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
	table := ""
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

func nullFloat(f float64) any {
	if f == 0 {
		return nil
	}
	return f
}
