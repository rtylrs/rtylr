package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/voxhire/internal/domain"
)

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
