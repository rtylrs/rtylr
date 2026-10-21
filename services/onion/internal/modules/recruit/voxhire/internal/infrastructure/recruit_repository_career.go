package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

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
