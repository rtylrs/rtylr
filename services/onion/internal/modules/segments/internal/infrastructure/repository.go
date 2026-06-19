package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	"github.com/voxire/onion/internal/modules/segments/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.Segment, error) {
	rows, err := r.queries.ListSegments(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list segments", logger.Error(err))

		return nil, err
	}

	segments := make([]domain.Segment, 0, len(rows))

	for _, row := range rows {
		segments = append(segments, segmentFromRow(row))
	}

	return segments, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.Segment, error) {
	row, err := r.queries.GetSegmentByID(ctx, sqlc.GetSegmentByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Segment{}, domain.ErrSegmentNotFound
		}

		return domain.Segment{}, err
	}

	return segmentFromRow(row), nil
}

func (r *Repository) Create(ctx context.Context, input domain.SegmentInput) (int64, error) {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}

	result, err := r.queries.CreateSegment(ctx, sqlc.CreateSegmentParams{
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Rules:       json.RawMessage(filters),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create segment", logger.Error(err))

		return 0, err
	}

	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.SegmentUpdateInput) error {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}

	result, err := r.queries.UpdateSegment(ctx, sqlc.UpdateSegmentParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Rules:       json.RawMessage(filters),
		ID:          input.ID,
		BusinessID:  input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update segment", logger.Error(err))

		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrSegmentNotFound
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	_, err := r.queries.DeleteSegment(ctx, sqlc.DeleteSegmentParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete segment", logger.Error(err))

		return err
	}

	return nil
}

func segmentFromRow(row sqlc.CrmSegment) domain.Segment {
	seg := domain.Segment{
		ID:            row.ID,
		OwnerID:       row.OwnerID,
		BusinessID:    row.BusinessID,
		Name:          row.Name,
		Description:   row.Description.String,
		CustomerCount: int(row.CustomerCount.Int32),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}

	if row.Rules != nil {
		seg.Filters = []byte(row.Rules)
	}

	return seg
}
