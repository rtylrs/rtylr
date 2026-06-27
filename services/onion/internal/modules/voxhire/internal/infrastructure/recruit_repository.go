package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	gosqlmysql "github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

const mysqlErrDuplicateEntry = 1062

const voxhireCareerOwnerID int64 = 1

// ── Admin accounts ───────────────────────────────────────────────────

// CreateAdmin inserts a new recruiter-admin account. Returns
// domain.ErrAdminExists if the email is already taken.
func (r *Repository) CreateAdmin(ctx context.Context, email, passwordHash, name string) (int64, error) {
	query, args, err := sq.
		Insert("voxhire_admins").
		Columns("email", "password_hash", "name").
		Values(email, passwordHash, name).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}

	result, err := r.db.ExecContext(ctx, query, args...)
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
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, COALESCE(name,''), password_hash, created_at
		   FROM voxhire_admins WHERE email = ?`, email)

	var a domain.Admin
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.PasswordHash, &a.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return &a, nil
}

// GetAdminByID returns the admin with the given id, or
// domain.ErrAdminNotFound.
func (r *Repository) GetAdminByID(ctx context.Context, id int64) (*domain.Admin, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, email, COALESCE(name,''), password_hash, created_at
		   FROM voxhire_admins WHERE id = ?`, id)

	var a domain.Admin
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.PasswordHash, &a.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAdminNotFound
		}
		return nil, err
	}
	return &a, nil
}

// ListAdmins returns all recruiter-admin accounts, newest first.
func (r *Repository) ListAdmins(ctx context.Context) ([]domain.Admin, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, email, COALESCE(name,''), created_at
		   FROM voxhire_admins ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	admins := make([]domain.Admin, 0)
	for rows.Next() {
		var a domain.Admin
		if err := rows.Scan(&a.ID, &a.Email, &a.Name, &a.CreatedAt); err != nil {
			return nil, err
		}
		admins = append(admins, a)
	}
	return admins, rows.Err()
}

// CountAdmins returns the number of admin accounts. Used to decide
// whether the default admin needs seeding.
func (r *Repository) CountAdmins(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM voxhire_admins`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// ── Openings (read from the shared careers table) ────────────────────

// ListOpenings returns the active job openings plus their application
// counts for the recruiter panel's position picker.
func (r *Repository) ListOpenings(ctx context.Context) ([]domain.Opening, error) {
	// cross-schema read from careers — admin openings + application counts
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
		 WHERE c.owner_id = ?
		 GROUP BY c.id, c.slug, c.title, c.department, c.location, c.employment_type,
		          c.work_mode, c.seniority, c.details, c.salary_min, c.salary_max,
		          c.salary_currency, c.is_active
		 ORDER BY c.is_active DESC, c.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, voxhireCareerOwnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	openings := make([]domain.Opening, 0)
	for rows.Next() {
		var o domain.Opening
		if err := rows.Scan(&o.ID, &o.Slug, &o.Title, &o.Department, &o.Location,
			&o.EmploymentType, &o.WorkMode, &o.Seniority, &o.Details, &o.SalaryMin,
			&o.SalaryMax, &o.SalaryCurrency, &o.IsActive, &o.NewCount, &o.TotalCount); err != nil {
			return nil, err
		}
		openings = append(openings, o)
	}
	return openings, rows.Err()
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
		return err
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
		return err
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
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, company_name, contact_name, COALESCE(contact_title,''), work_email,
		        COALESCE(phone,''), role, COALESCE(headcount,''), COALESCE(timeline,''),
		        COALESCE(service,''), COALESCE(message,''), agency_opt_in, status, created_at
		   FROM voxhire_company_leads
		  ORDER BY created_at DESC
		  LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leads := make([]domain.Lead, 0)
	for rows.Next() {
		var l domain.Lead
		if err := rows.Scan(&l.ID, &l.CompanyName, &l.ContactName, &l.ContactTitle, &l.WorkEmail,
			&l.Phone, &l.Role, &l.Headcount, &l.Timeline, &l.Service, &l.Message,
			&l.AgencyOptIn, &l.Status, &l.CreatedAt); err != nil {
			return nil, err
		}
		leads = append(leads, l)
	}
	return leads, rows.Err()
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
	         COALESCE(cv_extracted_text,''), cv_processed_at, COALESCE(review_rating,0),
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
		var cvProcessedAt sql.NullTime
		var reviewedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.JobID, &a.Name, &a.Email, &a.Phone, &a.Location,
			&a.SalaryExpectation, &a.SalaryCurrency, &a.Status, &a.WorkHistory, &a.Education,
			&a.ExperienceSummary, &a.CoverLetter, &a.HasCV, &a.CVProcessingStatus, &a.CVProcessingError,
			&a.CVSummary, &cvExtractedText, &cvProcessedAt, &a.ReviewRating, &a.ReviewNotes,
			&a.NextStep, &reviewedAt, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		a.CVSearchExcerpt = cvSearchExcerpt(f.Query, a.CVSummary, cvExtractedText)
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
	res, err := r.db.ExecContext(ctx,
		`UPDATE voxire_career_applications SET status = ? WHERE id = ?`, status, id)
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
	return nil
}

// GetApplicationCVKey returns the R2 object key of an application's CV,
// or domain.ErrApplicationNotFound. An empty string means no CV.
func (r *Repository) GetApplicationCVKey(ctx context.Context, id int64) (string, error) {
	var key sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT cv_key FROM voxire_career_applications WHERE id = ?`, id).Scan(&key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrApplicationNotFound
		}
		return "", err
	}
	return key.String, nil
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

	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		   SET cv_processing_status = ?,
		       cv_processing_error = ?,
		       cv_extracted_text = ?,
		       cv_summary = ?,
		       cv_processed_at = CURRENT_TIMESTAMP
		 WHERE id = ?
	`, table), status, nullStr(result.Error), nullStr(result.ExtractedText), nullStr(result.Summary), result.ID)
	return err
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
