package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, owner_id, COALESCE(category_id,0), name, COALESCE(sku,''),
		        sale_price, is_active, created_at, updated_at
		 FROM erp_products
		 WHERE owner_id = ? AND is_active = TRUE
		 ORDER BY name`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list menu items", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MenuItem, 0)
	for rows.Next() {
		var item domain.MenuItem
		if err := rows.Scan(
			&item.ID, &item.OwnerID, &item.CategoryID, &item.Name, &item.SKU,
			&item.SalePrice, &item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan menu item", logger.Error(err))
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error) {
	result, err := r.pool.ExecContext(ctx,
		`INSERT INTO erp_products (owner_id, name, sku, sale_price, is_active)
		 VALUES (?, ?, ?, ?, TRUE)`,
		ownerID, input.Name, input.SKU, input.SalePrice,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to quick-create menu item", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
