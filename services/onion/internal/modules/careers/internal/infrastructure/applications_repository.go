package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
)

// SaveApplication inserts a Voxire career application into
// voxire_career_applications using a plain parameterised query.
// We avoid sqlc here so we don't need to re-run code generation.
func (r *Repository) SaveApplication(ctx context.Context, in domain.ApplicationInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx, `
		INSERT INTO voxire_career_applications
			(job_id, name, email, phone, work_history, education, experience_summary, cover_letter, cv_key)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.JobID,
		in.Name,
		in.Email,
		in.Phone,
		in.WorkHistory,
		in.Education,
		in.ExperienceSummary,
		in.CoverLetter,
		in.CVKey,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
