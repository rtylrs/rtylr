package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/careers/internal/domain"
	"github.com/voxire/onion/internal/modules/careers/internal/infrastructure/persistence/sqlc"
)

// SaveApplication inserts a Voxire career application into
// voxire_career_applications.
func (r *Repository) SaveApplication(ctx context.Context, in domain.ApplicationInput) (int64, error) {
	result, err := r.queries.SaveCareerApplication(ctx, sqlc.SaveCareerApplicationParams{
		JobID:             in.JobID,
		Name:              in.Name,
		Email:             in.Email,
		Phone:             sql.NullString{String: in.Phone, Valid: true},
		WorkHistory:       sql.NullString{String: in.WorkHistory, Valid: true},
		Education:         sql.NullString{String: in.Education, Valid: true},
		ExperienceSummary: sql.NullString{String: in.ExperienceSummary, Valid: true},
		CoverLetter:       sql.NullString{String: in.CoverLetter, Valid: true},
		CvKey:             sql.NullString{String: in.CVKey, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
