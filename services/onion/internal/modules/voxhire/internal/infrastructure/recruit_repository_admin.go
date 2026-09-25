package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	gosqlmysql "github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

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
