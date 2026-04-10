package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/views/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID, workspaceID int64) ([]domain.View, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, workspace_id, name, filters, sort_by, group_by, is_default, created_at, updated_at
		 FROM flow_views WHERE owner_id = ? AND workspace_id = ? ORDER BY is_default DESC, name ASC`,
		ownerID, workspaceID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list views", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	views := make([]domain.View, 0)
	for rows.Next() {
		v, scanErr := scanView(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		views = append(views, v)
	}
	return views, rows.Err()
}

func (r *Repository) Create(ctx context.Context, input domain.ViewInput) (int64, error) {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO flow_views (owner_id, workspace_id, name, filters, sort_by, group_by, is_default)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.OwnerID, input.WorkspaceID, input.Name, filters,
		nullableString(input.SortBy), nullableString(input.GroupBy), input.IsDefault,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create view", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.ViewUpdateInput) error {
	filters := input.Filters
	if filters == nil {
		filters = []byte("{}")
	}
	result, err := r.pool.ExecContext(ctx,
		`UPDATE flow_views SET name = ?, filters = ?, sort_by = ?, group_by = ?, is_default = ?
		 WHERE id = ? AND owner_id = ?`,
		input.Name, filters, nullableString(input.SortBy), nullableString(input.GroupBy),
		input.IsDefault, input.ID, input.OwnerID,
	)
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
	result, err := r.pool.ExecContext(ctx,
		`DELETE FROM flow_views WHERE id = ? AND owner_id = ?`, id, ownerID,
	)
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

type viewScanner interface {
	Scan(dest ...any) error
}

func scanView(row viewScanner) (domain.View, error) {
	var v domain.View
	var filters []byte
	var sortBy, groupBy sql.NullString

	err := row.Scan(&v.ID, &v.OwnerID, &v.WorkspaceID, &v.Name, &filters, &sortBy, &groupBy, &v.IsDefault, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return domain.View{}, err
	}
	v.Filters = filters
	v.SortBy = sortBy.String
	v.GroupBy = groupBy.String
	return v, nil
}

func nullableString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
