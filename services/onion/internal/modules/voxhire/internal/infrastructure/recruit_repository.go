package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	gosqlmysql "github.com/go-sql-driver/mysql"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

const mysqlErrDuplicateEntry = 1062

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
	// cross-schema read from careers — active openings + application counts
	const query = `
		SELECT c.id,
		       c.title,
		       COALESCE(c.department, ''),
		       COALESCE(c.location, ''),
		       COALESCE(c.employment_type, ''),
		       COALESCE(c.work_mode, ''),
		       COALESCE(c.seniority, ''),
		       COALESCE(SUM(CASE WHEN a.status = 'new' THEN 1 ELSE 0 END), 0) AS new_count,
		       COALESCE(COUNT(a.id), 0) AS total_count
		  FROM careers c
		  LEFT JOIN voxire_career_applications a ON a.job_id = c.id
		 WHERE c.is_active = true
		 GROUP BY c.id, c.title, c.department, c.location, c.employment_type, c.work_mode, c.seniority
		 ORDER BY c.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	openings := make([]domain.Opening, 0)
	for rows.Next() {
		var o domain.Opening
		if err := rows.Scan(&o.ID, &o.Title, &o.Department, &o.Location,
			&o.EmploymentType, &o.WorkMode, &o.Seniority, &o.NewCount, &o.TotalCount); err != nil {
			return nil, err
		}
		openings = append(openings, o)
	}
	return openings, rows.Err()
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
	         COALESCE(cover_letter,''), (cv_key IS NOT NULL AND cv_key <> '') AS has_cv, created_at`

	base := sq.Select().From("voxire_career_applications").Where(sq.Eq{"job_id": f.JobID})

	if loc := strings.TrimSpace(f.Location); loc != "" {
		base = base.Where(sq.Like{"location": "%" + loc + "%"})
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := "%" + q + "%"
		// Raw substring search across the experience-bearing columns.
		base = base.Where(sq.Or{
			sq.Like{"work_history": like},
			sq.Like{"experience_summary": like},
			sq.Like{"cover_letter": like},
			sq.Like{"name": like},
		})
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
		if err := rows.Scan(&a.ID, &a.JobID, &a.Name, &a.Email, &a.Phone, &a.Location,
			&a.SalaryExpectation, &a.SalaryCurrency, &a.Status, &a.WorkHistory, &a.Education,
			&a.ExperienceSummary, &a.CoverLetter, &a.HasCV, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		apps = append(apps, a)
	}
	return apps, total, rows.Err()
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
