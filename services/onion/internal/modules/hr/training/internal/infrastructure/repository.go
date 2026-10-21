package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr/training/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/training/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.TrainingCourse, error) {
	rows, err := r.queries.ListTrainingCourses(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list training courses", logger.Error(err))
		return nil, err
	}

	items := make([]domain.TrainingCourse, 0, len(rows))
	for _, row := range rows {
		var item domain.TrainingCourse
		item.ID = row.ID
		item.OwnerID = row.OwnerID
		item.BusinessID = row.BusinessID
		item.Name = row.Name
		if row.Description.Valid {
			item.Description = row.Description.String
		}
		if row.DurationHours.Valid {
			item.DurationHours = sqlutil.ParseDecimal(row.DurationHours.String)
		}
		if row.IsRequired.Valid {
			item.IsRequired = row.IsRequired.Bool
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.TrainingCourseInput) (int64, error) {
	res, err := r.queries.CreateTrainingCourse(ctx, sqlc.CreateTrainingCourseParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		Name:          input.Name,
		Description:   sql.NullString{String: input.Description, Valid: input.Description != ""},
		DurationHours: sql.NullString{String: strconv.FormatFloat(input.DurationHours, 'f', 1, 64), Valid: true},
		IsRequired:    sql.NullBool{Bool: input.IsRequired, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create training course", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
