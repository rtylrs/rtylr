package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

// The recruit product extends the careers module with tenant-scoped
// queries and two sidecar tables (customizations, submissions). The
// careers table carries recruit-only columns (owner_id, slug,
// requirements, salary_*, is_published) that live in the recruit
// migration directory and are intentionally absent from the schema
// the careers sqlc config loads — so these queries use squirrel
// rather than sqlc, which would fail to resolve the columns.

// ── careers (tenant-scoped) ──────────────────────────────────────────

// recruitCareerSelectColumns is the projection shared by the
// tenant-scoped career reads so the order stays locked with the Scan
// calls below.
var recruitCareerSelectColumns = []string{
	"id", "owner_id", "business_id", "COALESCE(slug,'')", "title",
	"COALESCE(department,'')", "COALESCE(location,'')",
	"COALESCE(employment_type,'')", "COALESCE(work_mode,'')",
	"COALESCE(seniority,'')", "details",
	"COALESCE(requirements,'')",
	"COALESCE(salary_min, 0)", "COALESCE(salary_max, 0)",
	"COALESCE(salary_currency, '')",
	"COALESCE(is_published, false)", "is_active",
	"created_at", "updated_at",
}

func (r *Repository) ListCareers(ctx context.Context, businessID int64, publishedOnly bool) ([]domain.RecruitCareer, error) {
	b := sq.
		Select(recruitCareerSelectColumns...).
		From("careers").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.Eq{"is_active": true})
	if publishedOnly {
		b = b.Where(sq.Eq{"is_published": true})
	}
	q, args, err := b.
		OrderBy("created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.RecruitCareer, 0)
	for rows.Next() {
		var c domain.RecruitCareer
		if err := rows.Scan(&c.ID, &c.OwnerID, &c.BusinessID, &c.Slug, &c.Title,
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

func (r *Repository) GetCareerBySlug(ctx context.Context, businessID int64, slug string) (domain.RecruitCareer, error) {
	q, args, err := sq.
		Select(recruitCareerSelectColumns...).
		From("careers").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.Eq{"slug": slug}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.RecruitCareer{}, err
	}
	row := r.pool.QueryRowContext(ctx, q, args...)
	var c domain.RecruitCareer
	if err := row.Scan(&c.ID, &c.OwnerID, &c.BusinessID, &c.Slug, &c.Title,
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
	q, args, err := sq.
		Insert("careers").
		Columns("owner_id", "business_id", "slug", "title", "department", "location", "employment_type", "work_mode",
			"seniority", "details", "requirements", "salary_min", "salary_max", "salary_currency",
			"is_published", "is_active").
		Values(in.OwnerID, in.BusinessID, in.Slug, in.Title, nullStr(in.Department), nullStr(in.Location),
			nullStr(in.EmploymentType), nullStr(in.WorkMode), nullStr(in.Seniority),
			in.Details, nullStr(in.Requirements),
			nullFloat(in.SalaryMin), nullFloat(in.SalaryMax), nullStr(in.SalaryCurrency),
			in.IsPublished, true).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateCareer(ctx context.Context, in domain.RecruitCareerInput) error {
	q, args, err := sq.
		Update("careers").
		Set("slug", in.Slug).
		Set("title", in.Title).
		Set("department", nullStr(in.Department)).
		Set("location", nullStr(in.Location)).
		Set("employment_type", nullStr(in.EmploymentType)).
		Set("work_mode", nullStr(in.WorkMode)).
		Set("seniority", nullStr(in.Seniority)).
		Set("details", in.Details).
		Set("requirements", nullStr(in.Requirements)).
		Set("salary_min", nullFloat(in.SalaryMin)).
		Set("salary_max", nullFloat(in.SalaryMax)).
		Set("salary_currency", nullStr(in.SalaryCurrency)).
		Set("is_published", in.IsPublished).
		Where(sq.Eq{"id": in.ID}).
		Where(sq.Eq{"business_id": in.BusinessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, q, args...)
	return err
}

func (r *Repository) DeleteCareer(ctx context.Context, businessID, careerID int64) error {
	// Soft delete — flip is_active to false so historical submissions
	// still reference a readable row. Hard delete cascades would orphan
	// the submissions table.
	q, args, err := sq.
		Update("careers").
		Set("is_active", false).
		Set("is_published", false).
		Where(sq.Eq{"id": careerID}).
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, q, args...)
	return err
}

// ── careers_customizations ───────────────────────────────────────────

// customizationSelectColumns is shared by both get-by-owner and
// get-by-slug so the column order stays locked with Scan.
var customizationSelectColumns = []string{
	"id", "owner_id", "business_id", "slug", "brand_name",
	"COALESCE(tagline,'')", "COALESCE(about_blurb,'')",
	"COALESCE(logo_url,'')", "COALESCE(primary_color,'')",
	"COALESCE(accent_color,'')", "COALESCE(bg_color,'')",
	"COALESCE(text_color,'')", "COALESCE(muted_text_color,'')",
	"COALESCE(font_pairing,'')", "COALESCE(hero_image_url,'')",
	"COALESCE(favicon_url,'')",
	"template", "COALESCE(header_html,'')",
	"COALESCE(meta_title,'')", "COALESCE(meta_description,'')",
	"COALESCE(privacy_url,'')", "COALESCE(terms_url,'')",
	"COALESCE(ga_measurement_id,'')", "COALESCE(meta_pixel_id,'')",
	"require_consent",
	"COALESCE(notification_email,'')", "email_notifications",
	"created_at", "updated_at",
}

func scanCustomization(row interface{ Scan(...interface{}) error }) (domain.Customization, error) {
	var c domain.Customization
	err := row.Scan(
		&c.ID, &c.OwnerID, &c.BusinessID, &c.Slug, &c.BrandName,
		&c.Tagline, &c.AboutBlurb, &c.LogoURL, &c.PrimaryColor,
		&c.AccentColor, &c.BgColor, &c.TextColor, &c.MutedTextColor,
		&c.FontPairing, &c.HeroImageURL, &c.FaviconURL,
		&c.Template, &c.HeaderHTML,
		&c.MetaTitle, &c.MetaDescription,
		&c.PrivacyURL, &c.TermsURL,
		&c.GAMeasurementID, &c.MetaPixelID,
		&c.RequireConsent,
		&c.NotificationEmail, &c.EmailNotifications,
		&c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *Repository) GetCustomizationByOwner(ctx context.Context, businessID int64) (domain.Customization, error) {
	q, args, err := sq.
		Select(customizationSelectColumns...).
		From("careers_customizations").
		Where(sq.Eq{"business_id": businessID}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Customization{}, err
	}
	row := r.pool.QueryRowContext(ctx, q, args...)
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
	q, args, err := sq.
		Select(customizationSelectColumns...).
		From("careers_customizations").
		Where(sq.Eq{"slug": slug}).
		Limit(1).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return domain.Customization{}, err
	}
	row := r.pool.QueryRowContext(ctx, q, args...)
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
	// Upsert keyed by business_id. The unique index on (business_id) keeps
	// this to one row per business; the slug index is the backstop
	// against collisions across tenants.
	const onDuplicate = `ON DUPLICATE KEY UPDATE
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
        header_html = VALUES(header_html),
        meta_title = VALUES(meta_title),
        meta_description = VALUES(meta_description),
        privacy_url = VALUES(privacy_url),
        terms_url = VALUES(terms_url),
        ga_measurement_id = VALUES(ga_measurement_id),
        meta_pixel_id = VALUES(meta_pixel_id),
        require_consent = VALUES(require_consent),
        notification_email = VALUES(notification_email),
        email_notifications = VALUES(email_notifications)`

	q, args, err := sq.
		Insert("careers_customizations").
		Columns("owner_id", "business_id", "slug", "brand_name", "tagline", "about_blurb", "logo_url",
			"primary_color", "accent_color", "bg_color", "text_color",
			"muted_text_color", "font_pairing", "hero_image_url", "favicon_url",
			"template", "header_html",
			"meta_title", "meta_description", "privacy_url", "terms_url",
			"ga_measurement_id", "meta_pixel_id", "require_consent",
			"notification_email", "email_notifications").
		Values(in.OwnerID, in.BusinessID, in.Slug, in.BrandName, nullStr(in.Tagline), nullStr(in.AboutBlurb),
			nullStr(in.LogoURL), nullStr(in.PrimaryColor),
			nullStr(in.AccentColor), nullStr(in.BgColor),
			nullStr(in.TextColor), nullStr(in.MutedTextColor),
			nullStr(in.FontPairing), nullStr(in.HeroImageURL), nullStr(in.FaviconURL),
			in.Template, nullStr(in.HeaderHTML),
			nullStr(in.MetaTitle), nullStr(in.MetaDescription),
			nullStr(in.PrivacyURL), nullStr(in.TermsURL),
			nullStr(in.GAMeasurementID), nullStr(in.MetaPixelID),
			in.RequireConsent,
			nullStr(in.NotificationEmail), in.EmailNotifications).
		Suffix(onDuplicate).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, q, args...)
	return err
}

// ── careers_submissions ──────────────────────────────────────────────

func (r *Repository) CreateSubmission(ctx context.Context, in domain.SubmissionInput) (int64, error) {
	q, args, err := sq.
		Insert("careers_submissions").
		Columns("career_id", "owner_id", "business_id", "candidate_name", "email", "phone", "resume_url", "cover_letter", "answers_json").
		Values(in.CareerID, in.OwnerID, in.BusinessID, in.CandidateName, in.Email,
			nullStr(in.Phone), nullStr(in.ResumeURL), nullStr(in.CoverLetter),
			nullStr(in.AnswersJSON)).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return 0, err
	}
	res, err := r.pool.ExecContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) ListSubmissions(ctx context.Context, businessID int64, status string) ([]domain.Submission, error) {
	b := sq.
		Select("s.id", "s.career_id", "s.owner_id", "s.business_id", "s.candidate_name", "s.email",
			"COALESCE(s.phone,'')", "COALESCE(s.resume_url,'')",
			"COALESCE(s.cover_letter,'')", "s.status",
			"COALESCE(s.answers_json, '') AS answers_json",
			"s.created_at", "s.updated_at",
			"COALESCE(c.title, '') AS career_title").
		From("careers_submissions s").
		LeftJoin("careers c ON c.id = s.career_id").
		Where(sq.Eq{"s.business_id": businessID})
	if status != "" {
		b = b.Where(sq.Eq{"s.status": status})
	}
	q, args, err := b.
		OrderBy("s.created_at DESC").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Submission, 0)
	for rows.Next() {
		var s domain.Submission
		if err := rows.Scan(&s.ID, &s.CareerID, &s.OwnerID, &s.BusinessID, &s.CandidateName,
			&s.Email, &s.Phone, &s.ResumeURL, &s.CoverLetter, &s.Status,
			&s.AnswersJSON,
			&s.CreatedAt, &s.UpdatedAt, &s.CareerTitle); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateSubmissionStatus(ctx context.Context, businessID, submissionID int64, status string) error {
	q, args, err := sq.
		Update("careers_submissions").
		Set("status", status).
		Where(sq.Eq{"id": submissionID}).
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}
	_, err = r.pool.ExecContext(ctx, q, args...)
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
