package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/infrastructure/persistence/sqlc"
)

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
// job-specific application tables. Targets that previously failed extraction
// are still retried, but after every pending CV: permanently failing files
// (for example legacy .doc) must not consume the whole batch and starve newer
// CVs.
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
		 ORDER BY (COALESCE(cv_processing_status, 'pending') = 'error'), created_at ASC
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
		 ORDER BY (COALESCE(cv_processing_status, 'pending') = 'error'), created_at ASC
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
	raw, err := json.Marshal(cvProfileJSON(profile))
	if err != nil || string(raw) == "{}" {
		return ""
	}
	return string(raw)
}

func parseCVProfileJSON(raw string) domain.CVProfile {
	var profile cvProfileJSON
	if strings.TrimSpace(raw) == "" {
		return domain.CVProfile(profile)
	}
	if err := json.Unmarshal([]byte(raw), &profile); err != nil {
		return domain.CVProfile{}
	}
	return domain.CVProfile(profile)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
