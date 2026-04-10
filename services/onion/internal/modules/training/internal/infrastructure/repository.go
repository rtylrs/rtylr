package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/training/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.TrainingCourse, error) {
	query := `
		SELECT id, owner_id, name, description, duration_hours, is_required, created_at
		FROM hr_training_courses
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list training courses", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.TrainingCourse
	for rows.Next() {
		var item domain.TrainingCourse
		if err := item.FromRow(rows); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.TrainingCourseInput) (int64, error) {
	query := `
		INSERT INTO hr_training_courses (owner_id, name, description, duration_hours, is_required)
		VALUES (?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID, input.Name,
		sql.NullString{String: input.Description, Valid: input.Description != ""},
		input.DurationHours, input.IsRequired,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create training course", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}
