package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/views/internal/domain"
	"github.com/voxire/onion/internal/modules/views/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID, workspaceID int64) ([]domain.View, error) {
	rows, err := r.queries.ListViewsByWorkspace(ctx, sqlc.ListViewsByWorkspaceParams{
		OwnerID:     ownerID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list views", logger.Error(err))
		return nil, err
	}

	views := make([]domain.View, 0, len(rows))
	for _, row := range rows {
		views = append(views, domain.View{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Filters:     row.Filters,
			SortBy:      row.SortBy.String,
			GroupBy:     row.GroupBy.String,
			IsDefault:   row.IsDefault.Bool,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		})
	}
	return views, nil
}

func (r *Repository) Create(ctx context.Context, input domain.ViewInput) (int64, error) {
	filters := json.RawMessage(input.Filters)
	if filters == nil {
		filters = json.RawMessage("{}")
	}
	result, err := r.queries.CreateView(ctx, sqlc.CreateViewParams{
		OwnerID:     input.OwnerID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Filters:     filters,
		SortBy:      sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		GroupBy:     sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		IsDefault:   sql.NullBool{Bool: input.IsDefault, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create view", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ViewUpdateInput) error {
	filters := json.RawMessage(input.Filters)
	if filters == nil {
		filters = json.RawMessage("{}")
	}
	result, err := r.queries.UpdateView(ctx, sqlc.UpdateViewParams{
		Name:      input.Name,
		Filters:   filters,
		SortBy:    sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		GroupBy:   sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		IsDefault: sql.NullBool{Bool: input.IsDefault, Valid: true},
		ID:        input.ID,
		OwnerID:   input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update view", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrViewNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id, ownerID int64) error {
	result, err := r.queries.DeleteView(ctx, sqlc.DeleteViewParams{
		ID:      id,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete view", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrViewNotFound
	}
	return nil
}
