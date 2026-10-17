package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

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
		CvKey:              sqlNullString(in.CVKey),
		CvProcessingStatus: domain.SubmissionCVStatus(in.CVUploadFailed),
		CvProcessingError:  sql.NullString{String: in.CVUploadError, Valid: in.CVUploadFailed && in.CVUploadError != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert voxhire career application", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

// FindRecentCareerApplicationID returns the newest application from email to
// jobID in the last 10 minutes, or 0 when there is none.
func (r *Repository) FindRecentCareerApplicationID(ctx context.Context, jobID int64, email string) (int64, error) {
	id, err := r.queries.FindRecentCareerApplicationID(ctx, sqlc.FindRecentCareerApplicationIDParams{JobID: jobID, Email: email})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
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
