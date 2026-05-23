package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
	"github.com/voxire/onion/internal/modules/cycles/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64, workspaceID *int64) ([]domain.Cycle, error) {
	var rows []sqlc.FlowCycle
	var err error

	if workspaceID != nil {
		rows, err = r.queries.ListCyclesByWorkspace(ctx, sqlc.ListCyclesByWorkspaceParams{
			OwnerID:     ownerID,
			WorkspaceID: *workspaceID,
		})
	} else {
		rows, err = r.queries.ListCyclesByOwner(ctx, ownerID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list cycles", logger.Error(err))
		return nil, err
	}

	cycles := make([]domain.Cycle, 0, len(rows))
	for _, row := range rows {
		c := domain.Cycle{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Status:      string(row.Status.FlowCyclesStatus),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.StartDate.Valid {
			c.StartDate = &row.StartDate.Time
		}
		if row.EndDate.Valid {
			c.EndDate = &row.EndDate.Time
		}
		cycles = append(cycles, c)
	}
	return cycles, nil
}

func (r *Repository) Create(ctx context.Context, input domain.CycleInput) (int64, error) {
	result, err := r.queries.CreateCycle(ctx, sqlc.CreateCycleParams{
		OwnerID:     input.OwnerID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		StartDate:   nullTimePtrToSql(input.StartDate),
		EndDate:     nullTimePtrToSql(input.EndDate),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create cycle", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.CycleUpdateInput) error {
	status := input.Status
	if status == "" {
		status = "upcoming"
	}
	result, err := r.queries.UpdateCycle(ctx, sqlc.UpdateCycleParams{
		Name:      input.Name,
		StartDate: nullTimePtrToSql(input.StartDate),
		EndDate:   nullTimePtrToSql(input.EndDate),
		Status:    sqlc.NullFlowCyclesStatus{FlowCyclesStatus: sqlc.FlowCyclesStatus(status), Valid: true},
		ID:        input.ID,
		OwnerID:   input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update cycle", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrCycleNotFound
	}
	return nil
}

func nullTimePtrToSql(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}
