package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

// The recruit product extends the careers module with tenant-scoped
// queries and two sidecar tables (customizations, submissions). We
// deliberately write these as raw SQL rather than regenerating sqlc
// for every column change — the queries are small, readable, and
// sqlc's codegen has been brittle enough in this repo that keeping
// new DDL out of it avoids regen friction.

// ── careers (tenant-scoped) ──────────────────────────────────────────

func (r *Repository) ListCareers(ctx context.Context, ownerID int64, publishedOnly bool) ([]domain.RecruitCareer, error) {
	q := `SELECT id, owner_id, COALESCE(slug,''), title,
	            COALESCE(department,''), COALESCE(location,''),
	            COALESCE(employment_type,''), COALESCE(work_mode,''),
	            COALESCE(seniority,''), details,
	            COALESCE(requirements,''),
	            COALESCE(salary_min, 0), COALESCE(salary_max, 0),
	            COALESCE(salary_currency, ''),
	            COALESCE(is_published, false), is_active,
	            created_at, updated_at
	        FROM careers
	        WHERE owner_id = ?`
	if publishedOnly {
		q += " AND is_published = true AND is_active = true"
	}
	q += " ORDER BY created_at DESC"

	rows, err := r.pool.QueryContext(ctx, q, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.RecruitCareer, 0)
	for rows.Next() {
		var c domain.RecruitCareer
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.Slug, &c.Title,
			&c.Department, &c.Location, &c.EmploymentType, &c.WorkMode,
			&c.Seniority, &c.Details, &c.Requirements,
			&c.SalaryMin, &c.SalaryMax, &c.SalaryCurrency,
			&c.IsPublished, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *Repository) GetCareerBySlug(ctx context.Context, ownerID int64, slug string) (domain.RecruitCareer, error) {
	q := `SELECT id, owner_id, COALESCE(slug,''), title,
	            COALESCE(department,''), COALESCE(location,''),
	            COALESCE(employment_type,''), COALESCE(work_mode,''),
	            COALESCE(seniority,''), details,
	            COALESCE(requirements,''),
	            COALESCE(salary_min, 0), COALESCE(salary_max, 0),
	            COALESCE(salary_currency, ''),
	            COALESCE(is_published, false), is_active,
	            created_at, updated_at
	        FROM careers
	        WHERE owner_id = ? AND slug = ? LIMIT 1`
	row := r.pool.QueryRowContext(ctx, q, ownerID, slug)
	var c domain.RecruitCareer
	if err := row.Scan(&c.ID, &c.OwnerID, &c.Slug, &c.Title,
		&c.Department, &c.Location, &c.EmploymentType, &c.WorkMode,
		&c.Seniority, &c.Details, &c.Requirements,
		&c.SalaryMin, &c.SalaryMax, &c.SalaryCurrency,
		&c.IsPublished, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.RecruitCareer{}, ErrCareerNotFound
		}
		return domain.RecruitCareer{}, err
	}
	return c, nil
}

func (r *Repository) CreateCareer(ctx context.Context, in domain.RecruitCareerInput) (int64, error) {
	q := `INSERT INTO careers
	        (owner_id, slug, title, department, location, employment_type, work_mode,
	         seniority, details, requirements, salary_min, salary_max, salary_currency,
	         is_published, is_active)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, true)`
	res, err := r.pool.ExecContext(ctx, q,
		in.OwnerID, in.Slug, in.Title, nullStr(in.Department), nullStr(in.Location),
		nullStr(in.EmploymentType), nullStr(in.WorkMode), nullStr(in.Seniority),
		in.Details, nullStr(in.Requirements),
		nullFloat(in.SalaryMin), nullFloat(in.SalaryMax), nullStr(in.SalaryCurrency),
		in.IsPublished)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateCareer(ctx context.Context, in domain.RecruitCareerInput) error {
	q := `UPDATE careers SET
	        slug = ?, title = ?, department = ?, location = ?, employment_type = ?,
	        work_mode = ?, seniority = ?, details = ?, requirements = ?,
	        salary_min = ?, salary_max = ?, salary_currency = ?,
	        is_published = ?
	      WHERE id = ? AND owner_id = ?`
	_, err := r.pool.ExecContext(ctx, q,
		in.Slug, in.Title, nullStr(in.Department), nullStr(in.Location),
		nullStr(in.EmploymentType), nullStr(in.WorkMode), nullStr(in.Seniority),
		in.Details, nullStr(in.Requirements),
		nullFloat(in.SalaryMin), nullFloat(in.SalaryMax), nullStr(in.SalaryCurrency),
		in.IsPublished, in.ID, in.OwnerID)
	return err
}

func (r *Repository) DeleteCareer(ctx context.Context, ownerID, careerID int64) error {
	// Soft delete — flip is_active to false so historical submissions
	// still reference a readable row. Hard delete cascades would orphan
	// the submissions table.
	_, err := r.pool.ExecContext(ctx, `UPDATE careers SET is_active = false, is_published = false WHERE id = ? AND owner_id = ?`, careerID, ownerID)
	return err
}

// ── careers_customizations ───────────────────────────────────────────

// customizationSelectColumns is shared by both get-by-owner and
// get-by-slug so the column order stays locked with Scan.
const customizationSelectColumns = `
    id, owner_id, slug, brand_name,
    COALESCE(tagline,''), COALESCE(about_blurb,''),
    COALESCE(logo_url,''), COALESCE(primary_color,''),
    COALESCE(accent_color,''), COALESCE(bg_color,''),
    COALESCE(text_color,''), COALESCE(muted_text_color,''),
    COALESCE(font_pairing,''), COALESCE(hero_image_url,''),
    COALESCE(favicon_url,''),
    template, COALESCE(header_html,''),
    created_at, updated_at`

func scanCustomization(row interface{ Scan(...interface{}) error }) (domain.Customization, error) {
	var c domain.Customization
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.Slug, &c.BrandName,
		&c.Tagline, &c.AboutBlurb, &c.LogoURL, &c.PrimaryColor,
		&c.AccentColor, &c.BgColor, &c.TextColor, &c.MutedTextColor,
		&c.FontPairing, &c.HeroImageURL, &c.FaviconURL,
		&c.Template, &c.HeaderHTML, &c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *Repository) GetCustomizationByOwner(ctx context.Context, ownerID int64) (domain.Customization, error) {
	q := `SELECT ` + customizationSelectColumns + ` FROM careers_customizations WHERE owner_id = ? LIMIT 1`
	row := r.pool.QueryRowContext(ctx, q, ownerID)
	c, err := scanCustomization(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customization{}, ErrCustomizationNotFound
		}
		return domain.Customization{}, err
	}
	return c, nil
}

func (r *Repository) GetCustomizationBySlug(ctx context.Context, slug string) (domain.Customization, error) {
	q := `SELECT ` + customizationSelectColumns + ` FROM careers_customizations WHERE slug = ? LIMIT 1`
	row := r.pool.QueryRowContext(ctx, q, slug)
	c, err := scanCustomization(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Customization{}, ErrCustomizationNotFound
		}
		return domain.Customization{}, err
	}
	return c, nil
}

func (r *Repository) UpsertCustomization(ctx context.Context, in domain.CustomizationInput) error {
	// Upsert keyed by owner_id. The unique index on (owner_id) keeps
	// this to one row per business; the slug index is the backstop
	// against collisions across tenants.
	q := `INSERT INTO careers_customizations
	        (owner_id, slug, brand_name, tagline, about_blurb, logo_url,
	         primary_color, accent_color, bg_color, text_color,
	         muted_text_color, font_pairing, hero_image_url, favicon_url,
	         template, header_html)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	      ON DUPLICATE KEY UPDATE
	        slug = VALUES(slug),
	        brand_name = VALUES(brand_name),
	        tagline = VALUES(tagline),
	        about_blurb = VALUES(about_blurb),
	        logo_url = VALUES(logo_url),
	        primary_color = VALUES(primary_color),
	        accent_color = VALUES(accent_color),
	        bg_color = VALUES(bg_color),
	        text_color = VALUES(text_color),
	        muted_text_color = VALUES(muted_text_color),
	        font_pairing = VALUES(font_pairing),
	        hero_image_url = VALUES(hero_image_url),
	        favicon_url = VALUES(favicon_url),
	        template = VALUES(template),
	        header_html = VALUES(header_html)`
	_, err := r.pool.ExecContext(ctx, q,
		in.OwnerID, in.Slug, in.BrandName, nullStr(in.Tagline), nullStr(in.AboutBlurb),
		nullStr(in.LogoURL), nullStr(in.PrimaryColor),
		nullStr(in.AccentColor), nullStr(in.BgColor),
		nullStr(in.TextColor), nullStr(in.MutedTextColor),
		nullStr(in.FontPairing), nullStr(in.HeroImageURL), nullStr(in.FaviconURL),
		in.Template, nullStr(in.HeaderHTML))
	return err
}

// ── careers_submissions ──────────────────────────────────────────────

func (r *Repository) CreateSubmission(ctx context.Context, in domain.SubmissionInput) (int64, error) {
	q := `INSERT INTO careers_submissions
	        (career_id, owner_id, candidate_name, email, phone, resume_url, cover_letter, answers_json)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.pool.ExecContext(ctx, q,
		in.CareerID, in.OwnerID, in.CandidateName, in.Email,
		nullStr(in.Phone), nullStr(in.ResumeURL), nullStr(in.CoverLetter),
		nullStr(in.AnswersJSON))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) ListSubmissions(ctx context.Context, ownerID int64, status string) ([]domain.Submission, error) {
	q := `SELECT s.id, s.career_id, s.owner_id, s.candidate_name, s.email,
	            COALESCE(s.phone,''), COALESCE(s.resume_url,''),
	            COALESCE(s.cover_letter,''), s.status,
	            COALESCE(s.answers_json, '') AS answers_json,
	            s.created_at, s.updated_at,
	            COALESCE(c.title, '') AS career_title
	        FROM careers_submissions s
	        LEFT JOIN careers c ON c.id = s.career_id
	        WHERE s.owner_id = ?`
	args := []interface{}{ownerID}
	if status != "" {
		q += " AND s.status = ?"
		args = append(args, status)
	}
	q += " ORDER BY s.created_at DESC"

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Submission, 0)
	for rows.Next() {
		var s domain.Submission
		if err := rows.Scan(&s.ID, &s.CareerID, &s.OwnerID, &s.CandidateName,
			&s.Email, &s.Phone, &s.ResumeURL, &s.CoverLetter, &s.Status,
			&s.AnswersJSON,
			&s.CreatedAt, &s.UpdatedAt, &s.CareerTitle); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateSubmissionStatus(ctx context.Context, ownerID, submissionID int64, status string) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE careers_submissions SET status = ? WHERE id = ? AND owner_id = ?`,
		status, submissionID, ownerID)
	return err
}

// ── helpers ──────────────────────────────────────────────────────────

// ErrCareerNotFound / ErrCustomizationNotFound are sentinel errors the
// service layer translates into HTTP 404s. Keeping them in the repo
// package means callers don't need to inspect sql.ErrNoRows directly.
var (
	ErrCareerNotFound        = errors.New("career not found")
	ErrCustomizationNotFound = errors.New("customization not found")
)

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullFloat(f float64) interface{} {
	if f == 0 {
		return nil
	}
	return f
}
