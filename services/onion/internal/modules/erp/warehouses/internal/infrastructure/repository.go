package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListWarehouses(ctx context.Context, businessID int64) ([]domain.Warehouse, error) {
	rows, err := r.queries.ListWarehouses(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list warehouses", logger.Error(err))

		return nil, err
	}

	warehouses := make([]domain.Warehouse, 0, len(rows))

	for _, row := range rows {
		var w domain.Warehouse
		w.ID = row.ID
		w.OwnerID = row.OwnerID
		w.BusinessID = row.BusinessID
		w.Name = row.Name

		if row.Address.Valid {
			w.Address = row.Address.String
		}

		if row.City.Valid {
			w.City = row.City.String
		}

		if row.Country.Valid {
			w.Country = row.Country.String
		}

		if row.ManagerName.Valid {
			w.ManagerName = row.ManagerName.String
		}

		if row.Capacity.Valid {
			w.Capacity = row.Capacity.Int64
		}

		if row.IsDefault.Valid {
			w.IsDefault = row.IsDefault.Bool
		}

		if row.IsActive.Valid {
			w.IsActive = row.IsActive.Bool
		}

		if row.CreatedAt.Valid {
			w.CreatedAt = row.CreatedAt.Time
		}

		if row.UpdatedAt.Valid {
			w.UpdatedAt = row.UpdatedAt.Time
		}

		warehouses = append(warehouses, w)
	}

	return warehouses, nil
}

func (r *Repository) CreateWarehouse(ctx context.Context, businessID int64, input domain.WarehouseInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	result, err := qtx.CreateWarehouse(ctx, sqlc.CreateWarehouseParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		Address:     sqlutil.NullString(input.Address),
		City:        sqlutil.NullString(input.City),
		Country:     sqlutil.NullString(input.Country),
		ManagerName: sqlutil.NullString(input.ManagerName),
		Capacity:    sqlutil.NullInt64(input.Capacity),
		IsDefault:   sql.NullBool{Bool: input.IsDefault, Valid: true},
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create warehouse", logger.Error(err))

		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "warehouse", EntityID: id, Action: "created", OldValues: nil, NewValues: warehouseInputSnapshot(input)}); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) UpdateWarehouse(ctx context.Context, businessID, id int64, input domain.WarehouseInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	locked, err := qtx.GetWarehouseForUpdate(ctx, sqlc.GetWarehouseForUpdateParams{BusinessID: businessID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrWarehouseNotFound
	}
	if err != nil {
		return err
	}
	old := warehouseSnapshot(warehouseFromRow(locked))

	err = qtx.UpdateWarehouse(ctx, sqlc.UpdateWarehouseParams{
		Name:        input.Name,
		Address:     sqlutil.NullString(input.Address),
		City:        sqlutil.NullString(input.City),
		Country:     sqlutil.NullString(input.Country),
		ManagerName: sqlutil.NullString(input.ManagerName),
		Capacity:    sqlutil.NullInt64(input.Capacity),
		IsDefault:   sql.NullBool{Bool: input.IsDefault, Valid: true},
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:  businessID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update warehouse", logger.Error(err))

		return err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "warehouse", EntityID: id, Action: "updated", OldValues: old, NewValues: warehouseInputSnapshot(input)}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteWarehouse(ctx context.Context, businessID, id int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	locked, err := qtx.GetWarehouseForUpdate(ctx, sqlc.GetWarehouseForUpdateParams{BusinessID: businessID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrWarehouseNotFound
	}
	if err != nil {
		return err
	}
	old := warehouseSnapshot(warehouseFromRow(locked))

	err = qtx.DeleteWarehouse(ctx, sqlc.DeleteWarehouseParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete warehouse", logger.Error(err))

		return err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "warehouse", EntityID: id, Action: "deleted", OldValues: old, NewValues: nil}); err != nil {
		return err
	}
	return tx.Commit()
}

// CountWarehouses returns the total number of warehouses for the given business.
// Pass activeOnly=true to count only active warehouses.
func (r *Repository) CountWarehouses(ctx context.Context, businessID int64, activeOnly bool) (int64, error) {
	var count int64
	var err error

	if activeOnly {
		count, err = r.queries.CountActiveWarehouses(ctx, businessID)
	} else {
		count, err = r.queries.CountWarehouses(ctx, businessID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to count warehouses", logger.Error(err))
		return 0, err
	}

	return count, nil
}

// GetWarehouseByID fetches a single warehouse by primary key for the given business.
// Returns ErrWarehouseNotFound when no matching row exists.
func (r *Repository) GetWarehouseByID(ctx context.Context, businessID, id int64) (domain.Warehouse, error) {
	row, err := r.queries.GetWarehouseByID(ctx, sqlc.GetWarehouseByIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Warehouse{}, domain.ErrWarehouseNotFound
		}
		logger.Ctx(ctx).Error("Failed to get warehouse by ID", logger.Error(err))
		return domain.Warehouse{}, err
	}

	return warehouseFromRow(row), nil
}

func warehouseFromRow(row sqlc.ErpWarehouse) domain.Warehouse {
	var w domain.Warehouse
	w.ID = row.ID
	w.OwnerID = row.OwnerID
	w.BusinessID = row.BusinessID
	w.Name = row.Name

	if row.Address.Valid {
		w.Address = row.Address.String
	}
	if row.City.Valid {
		w.City = row.City.String
	}
	if row.Country.Valid {
		w.Country = row.Country.String
	}
	if row.ManagerName.Valid {
		w.ManagerName = row.ManagerName.String
	}
	if row.Capacity.Valid {
		w.Capacity = row.Capacity.Int64
	}
	if row.IsDefault.Valid {
		w.IsDefault = row.IsDefault.Bool
	}
	if row.IsActive.Valid {
		w.IsActive = row.IsActive.Bool
	}
	if row.CreatedAt.Valid {
		w.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		w.UpdatedAt = row.UpdatedAt.Time
	}

	return w
}
