package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT
			p.id,
			p.owner_id,
			p.category_id,
			COALESCE(c.name, 'Other') AS category_name,
			p.name,
			p.sku,
			p.sale_price,
			p.description,
			p.stock_quantity,
			p.is_active,
			p.created_at,
			p.updated_at
		FROM erp_products p
		LEFT JOIN erp_categories c
		  ON c.id = p.category_id
		 AND c.owner_id = p.owner_id
		WHERE p.owner_id = ?
		  AND p.is_active = TRUE
		ORDER BY c.name, p.name
	`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list menu items", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.MenuItem, 0)
	for rows.Next() {
		var (
			categoryID sql.NullInt64
			category   sql.NullString
			sku        sql.NullString
			salePrice  string
			desc       sql.NullString
			stock      string
			active     sql.NullBool
			createdAt  sql.NullTime
			updatedAt  sql.NullTime
			item       domain.MenuItem
		)
		if err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&categoryID,
			&category,
			&item.Name,
			&sku,
			&salePrice,
			&desc,
			&stock,
			&active,
			&createdAt,
			&updatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan menu item", logger.Error(err))
			return nil, err
		}
		if categoryID.Valid {
			item.CategoryID = categoryID.Int64
		}
		if category.Valid && category.String != "" {
			item.Category = category.String
		} else {
			item.Category = "Other"
		}
		if sku.Valid {
			item.SKU = sku.String
		}
		item.SalePrice = decimalToFloat(salePrice)
		if desc.Valid {
			item.Description = desc.String
		}
		item.Stock = decimalToFloat(stock)
		item.IsActive = active.Valid && active.Bool
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time
		}
		items = append(items, domain.MenuItem{
			ID:          item.ID,
			OwnerID:     item.OwnerID,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Name:        item.Name,
			SKU:         item.SKU,
			SalePrice:   item.SalePrice,
			Description: item.Description,
			Stock:       item.Stock,
			IsActive:    item.IsActive,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		logger.Ctx(ctx).Error("Failed to iterate menu items", logger.Error(err))
		return nil, err
	}
	return items, nil
}

func (r *Repository) QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error) {
	result, err := r.queries.QuickCreateMenuItem(ctx, sqlc.QuickCreateMenuItemParams{
		OwnerID:   ownerID,
		Name:      input.Name,
		Sku:       sql.NullString{String: input.SKU, Valid: input.SKU != ""},
		SalePrice: floatToDecimal(input.SalePrice),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to quick-create menu item", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) BusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	var ownerID int64
	err := r.pool.QueryRowContext(ctx, `SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE`, businessID).Scan(&ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve menu business owner", logger.Error(err))
		return 0, err
	}
	return ownerID, nil
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse menu decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
