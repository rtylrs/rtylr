package infrastructure

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/reviews/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/reviews/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Review, error) {
	rows, err := r.queries.ListReviews(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reviews", logger.Error(err))
		return nil, err
	}

	items := make([]domain.Review, 0, len(rows))
	for _, row := range rows {
		var item domain.Review
		item.ID = row.ID
		item.OwnerID = row.OwnerID
		item.BusinessID = row.BusinessID
		item.StaffID = row.StaffID
		if row.ReviewerID.Valid {
			item.ReviewerID = row.ReviewerID.Int64
		}
		if row.Period.Valid {
			item.Period = row.Period.String
		}
		if row.Rating.Valid {
			rating, err := strconv.ParseFloat(row.Rating.String, 64)
			if err != nil {
				logger.Ctx(ctx).Warn("invalid rating value",
					logger.String("value", row.Rating.String),
					logger.Error(err),
				)
				rating = 0
			}
			item.Rating = rating
		}
		if row.GoalsRating.Valid {
			goalsRating, err := strconv.ParseFloat(row.GoalsRating.String, 64)
			if err != nil {
				logger.Ctx(ctx).Warn("invalid goals rating value",
					logger.String("value", row.GoalsRating.String),
					logger.Error(err),
				)
				goalsRating = 0
			}
			item.GoalsRating = goalsRating
		}
		if row.ValuesRating.Valid {
			valuesRating, err := strconv.ParseFloat(row.ValuesRating.String, 64)
			if err != nil {
				logger.Ctx(ctx).Warn("invalid values rating value",
					logger.String("value", row.ValuesRating.String),
					logger.Error(err),
				)
				valuesRating = 0
			}
			item.ValuesRating = valuesRating
		}
		if row.SkillsRating.Valid {
			skillsRating, err := strconv.ParseFloat(row.SkillsRating.String, 64)
			if err != nil {
				logger.Ctx(ctx).Warn("invalid skills rating value",
					logger.String("value", row.SkillsRating.String),
					logger.Error(err),
				)
				skillsRating = 0
			}
			item.SkillsRating = skillsRating
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

func (r *Repository) Create(ctx context.Context, businessID int64, input domain.ReviewInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "draft"
	}

	ids := []int64{input.StaffID}
	if input.ReviewerID != 0 {
		ids = append(ids, input.ReviewerID)
	}
	if err := r.requireStaff(ctx, businessID, ids...); err != nil {
		return 0, err
	}

	res, err := r.queries.CreateReview(ctx, sqlc.CreateReviewParams{
		OwnerID:      businessID,
		BusinessID:   businessID,
		StaffID:      input.StaffID,
		ReviewerID:   sql.NullInt64{Int64: input.ReviewerID, Valid: input.ReviewerID != 0},
		Period:       sql.NullString{String: input.Period, Valid: input.Period != ""},
		Rating:       sql.NullString{String: strconv.FormatFloat(input.Rating, 'g', -1, 64), Valid: true},
		GoalsRating:  sql.NullString{String: strconv.FormatFloat(input.GoalsRating, 'g', -1, 64), Valid: true},
		ValuesRating: sql.NullString{String: strconv.FormatFloat(input.ValuesRating, 'g', -1, 64), Valid: true},
		SkillsRating: sql.NullString{String: strconv.FormatFloat(input.SkillsRating, 'g', -1, 64), Valid: true},
		Notes:        sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		Status:       sqlc.NullHrReviewsStatus{HrReviewsStatus: sqlc.HrReviewsStatus(status), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create review", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// requireStaff fails with ErrStaffNotFound unless every id is a staff member of
// the business, so a review cannot point at another tenant's staff.
func (r *Repository) requireStaff(ctx context.Context, businessID int64, ids ...int64) error {
	unique := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	args := make([]any, 0, len(unique)+1)
	args = append(args, businessID)
	placeholders := make([]string, 0, len(unique))
	for id := range unique {
		args = append(args, id)
		placeholders = append(placeholders, "?")
	}
	var found int
	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM hr_staff WHERE business_id = ? AND id IN ("+strings.Join(placeholders, ",")+")",
		args...,
	).Scan(&found)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify review staff", logger.Error(err))
		return err
	}
	if found != len(unique) {
		return domain.ErrStaffNotFound
	}
	return nil
}
