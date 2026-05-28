package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	"github.com/voxire/onion/internal/modules/warehouses/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListWarehouses(ctx context.Context, ownerID int64) ([]domain.Warehouse, error) {
	rows, err := r.queries.ListWarehouses(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list warehouses", logger.Error(err))

		return nil, err
	}

	var warehouses []domain.Warehouse

	for _, row := range rows {
		var w domain.Warehouse
		w.ID = row.ID
		w.OwnerID = row.OwnerID
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

func (r *Repository) CreateWarehouse(ctx context.Context, ownerID int64, input domain.WarehouseInput) (int64, error) {
	result, err := r.queries.CreateWarehouse(ctx, sqlc.CreateWarehouseParams{
		OwnerID:     ownerID,
		Name:        input.Name,
		Address:     nullString(input.Address),
		City:        nullString(input.City),
		Country:     nullString(input.Country),
		ManagerName: nullString(input.ManagerName),
		Capacity:    nullInt64(input.Capacity),
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

	return id, nil
}

func (r *Repository) UpdateWarehouse(ctx context.Context, ownerID, id int64, input domain.WarehouseInput) error {
	err := r.queries.UpdateWarehouse(ctx, sqlc.UpdateWarehouseParams{
		Name:        input.Name,
		Address:     nullString(input.Address),
		City:        nullString(input.City),
		Country:     nullString(input.Country),
		ManagerName: nullString(input.ManagerName),
		Capacity:    nullInt64(input.Capacity),
		IsDefault:   sql.NullBool{Bool: input.IsDefault, Valid: true},
		IsActive:    sql.NullBool{Bool: input.IsActive, Valid: true},
		OwnerID:     ownerID,
		ID:          id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update warehouse", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteWarehouse(ctx context.Context, ownerID, id int64) error {
	err := r.queries.DeleteWarehouse(ctx, sqlc.DeleteWarehouseParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete warehouse", logger.Error(err))

		return err
	}

	return nil
}

// CountWarehouses returns the total number of warehouses for the given owner.
// Pass activeOnly=true to count only active warehouses.
func (r *Repository) CountWarehouses(ctx context.Context, ownerID int64, activeOnly bool) (int64, error) {
	var count int64
	var err error

	if activeOnly {
		const q = `SELECT COUNT(*) FROM erp_warehouses WHERE owner_id = ? AND is_active = 1`
		err = r.pool.QueryRowContext(ctx, q, ownerID).Scan(&count)
	} else {
		const q = `SELECT COUNT(*) FROM erp_warehouses WHERE owner_id = ?`
		err = r.pool.QueryRowContext(ctx, q, ownerID).Scan(&count)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to count warehouses", logger.Error(err))
		return 0, err
	}

	return count, nil
}

// GetWarehouseByID fetches a single warehouse by primary key for the given owner.
// Returns ErrWarehouseNotFound when no matching row exists.
func (r *Repository) GetWarehouseByID(ctx context.Context, ownerID, id int64) (domain.Warehouse, error) {
	const q = `
		SELECT id, owner_id, name, address, city, country,
		       manager_name, capacity, is_default, is_active,
		       created_at, updated_at
		FROM erp_warehouses
		WHERE owner_id = ? AND id = ?
		LIMIT 1`

	row := r.pool.QueryRowContext(ctx, q, ownerID, id)

	var w domain.Warehouse
	var address, city, country, managerName sql.NullString
	var capacity sql.NullInt64
	var isDefault, isActive sql.NullBool
	var createdAt, updatedAt sql.NullTime

	err := row.Scan(
		&w.ID, &w.OwnerID, &w.Name,
		&address, &city, &country, &managerName, &capacity,
		&isDefault, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Warehouse{}, domain.ErrWarehouseNotFound
		}
		logger.Ctx(ctx).Error("Failed to get warehouse by ID", logger.Error(err))
		return domain.Warehouse{}, err
	}

	if address.Valid {
		w.Address = address.String
	}
	if city.Valid {
		w.City = city.String
	}
	if country.Valid {
		w.Country = country.String
	}
	if managerName.Valid {
		w.ManagerName = managerName.String
	}
	if capacity.Valid {
		w.Capacity = capacity.Int64
	}
	if isDefault.Valid {
		w.IsDefault = isDefault.Bool
	}
	if isActive.Valid {
		w.IsActive = isActive.Bool
	}
	if createdAt.Valid {
		w.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		w.UpdatedAt = updatedAt.Time
	}

	return w, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
