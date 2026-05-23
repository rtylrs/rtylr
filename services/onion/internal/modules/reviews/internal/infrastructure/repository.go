package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
	"github.com/voxire/onion/internal/modules/reviews/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.Review, error) {
	rows, err := r.queries.ListReviews(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reviews", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Review, 0, len(rows))
	for _, row := range rows {
		var item domain.Review
		item.ID = row.ID
		item.OwnerID = row.OwnerID
		item.StaffID = row.StaffID
		if row.ReviewerID.Valid {
			item.ReviewerID = row.ReviewerID.Int64
		}
		if row.Period.Valid {
			item.Period = row.Period.String
		}
		if row.Rating.Valid {
			item.Rating, _ = strconv.ParseFloat(row.Rating.String, 64)
		}
		if row.GoalsRating.Valid {
			item.GoalsRating, _ = strconv.ParseFloat(row.GoalsRating.String, 64)
		}
		if row.ValuesRating.Valid {
			item.ValuesRating, _ = strconv.ParseFloat(row.ValuesRating.String, 64)
		}
		if row.SkillsRating.Valid {
			item.SkillsRating, _ = strconv.ParseFloat(row.SkillsRating.String, 64)
		}
		if row.Notes.Valid {
			item.Notes = row.Notes.String
		}
		if row.Status.Valid {
			item.Status = string(row.Status.HrReviewsStatus)
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			item.UpdatedAt = row.UpdatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.ReviewInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "draft"
	}

	res, err := r.queries.CreateReview(ctx, sqlc.CreateReviewParams{
		OwnerID:      ownerID,
		StaffID:      input.StaffID,
		ReviewerID:   sql.NullInt64{Int64: input.ReviewerID, Valid: input.ReviewerID != 0},
		Period:       sql.NullString{String: input.Period, Valid: input.Period != ""},
		Rating:       sql.NullString{String: fmt.Sprintf("%g", input.Rating), Valid: true},
		GoalsRating:  sql.NullString{String: fmt.Sprintf("%g", input.GoalsRating), Valid: true},
		ValuesRating: sql.NullString{String: fmt.Sprintf("%g", input.ValuesRating), Valid: true},
		SkillsRating: sql.NullString{String: fmt.Sprintf("%g", input.SkillsRating), Valid: true},
		Notes:        sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		Status:       sqlc.NullHrReviewsStatus{HrReviewsStatus: sqlc.HrReviewsStatus(status), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create review", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}
