package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		pool: pool,
	}
}

func (r *Repository) ListWarehouses(ctx context.Context, ownerID int64) ([]domain.Warehouse, error) {
	query := `
		SELECT id, owner_id, name, address, city, country, manager_name,
		       capacity, is_default, is_active, created_at, updated_at
		FROM erp_warehouses
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list warehouses", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var warehouses []domain.Warehouse

	for rows.Next() {
		var w domain.Warehouse
		if err := w.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan warehouse row", logger.Error(err))

			return nil, err
		}

		warehouses = append(warehouses, w)
	}

	return warehouses, rows.Err()
}

func (r *Repository) CreateWarehouse(ctx context.Context, ownerID int64, input domain.WarehouseInput) (int64, error) {
	query := `
		INSERT INTO erp_warehouses (owner_id, name, address, city, country, manager_name, capacity, is_default, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		input.Name,
		nullString(input.Address),
		nullString(input.City),
		nullString(input.Country),
		nullString(input.ManagerName),
		nullInt64(input.Capacity),
		input.IsDefault,
		input.IsActive,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create warehouse", logger.Error(err))

		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) UpdateWarehouse(ctx context.Context, ownerID, id int64, input domain.WarehouseInput) error {
	query := `
		UPDATE erp_warehouses
		SET name = ?, address = ?, city = ?, country = ?, manager_name = ?,
		    capacity = ?, is_default = ?, is_active = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		input.Name,
		nullString(input.Address),
		nullString(input.City),
		nullString(input.Country),
		nullString(input.ManagerName),
		nullInt64(input.Capacity),
		input.IsDefault,
		input.IsActive,
		ownerID,
		id,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update warehouse", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteWarehouse(ctx context.Context, ownerID, id int64) error {
	query := `DELETE FROM erp_warehouses WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete warehouse", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
