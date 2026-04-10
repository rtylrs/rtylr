package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Review, error) {
	query := `
		SELECT id, owner_id, staff_id, reviewer_id, period,
		       rating, goals_rating, values_rating, skills_rating,
		       notes, status, created_at, updated_at
		FROM hr_reviews
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reviews", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.Review
	for rows.Next() {
		var item domain.Review
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ReviewInput) (int64, error) {
	query := `
		INSERT INTO hr_reviews
		(owner_id, staff_id, reviewer_id, period, rating, goals_rating, values_rating, skills_rating, notes, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	status := input.Status
	if status == "" {
		status = "draft"
	}

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.StaffID,
		sql.NullInt64{Int64: input.ReviewerID, Valid: input.ReviewerID != 0},
		sql.NullString{String: input.Period, Valid: input.Period != ""},
		input.Rating, input.GoalsRating, input.ValuesRating, input.SkillsRating,
		sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		status,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create review", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}
