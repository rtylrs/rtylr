package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos/floors/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/floors/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID, branchID int64) ([]domain.Floor, error) {
	rows, err := r.queries.ListFloorsByBranch(ctx, sqlc.ListFloorsByBranchParams{
		BusinessID: businessID,
		BranchID:   branchID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list floors", logger.Error(err))
		return nil, err
	}

	floors := make([]domain.Floor, 0, len(rows))
	for _, row := range rows {
		floors = append(floors, domain.Floor{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			Name:       row.Name,
			SortOrder:  int(row.SortOrder.Int32),
			IsActive:   row.IsActive.Bool,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return floors, nil
}

func (r *Repository) Create(ctx context.Context, input domain.FloorInput) (int64, error) {
	result, err := r.queries.CreateFloor(ctx, sqlc.CreateFloorParams{
		BusinessID: input.BusinessID,
		BranchID:   input.BranchID,
		Name:       input.Name,
		SortOrder:  sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create floor", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.FloorUpdateInput) error {
	err := r.queries.UpdateFloor(ctx, sqlc.UpdateFloorParams{
		Name:       input.Name,
		SortOrder:  sql.NullInt32{Int32: int32(input.SortOrder), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		ID:         input.ID,
		BusinessID: input.BusinessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update floor", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, businessID, id int64) error {
	err := r.queries.SoftDeleteFloor(ctx, sqlc.SoftDeleteFloorParams{
		ID:         id,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to soft delete floor", logger.Error(err))
		return err
	}
	return nil
}
