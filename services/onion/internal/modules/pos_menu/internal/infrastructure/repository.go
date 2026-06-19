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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.MenuItem, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT
			p.id,
			p.owner_id,
			p.business_id,
			p.category_id,
			COALESCE(c.name, 'Other') AS category_name,
			p.name,
			p.sku,
			p.barcode,
			p.sale_price,
			p.tax_rate,
			p.description,
			COALESCE(img.url, '') AS image_url,
			COALESCE(pst.total_quantity, p.stock_quantity) AS stock_quantity,
			p.is_active,
			p.created_at,
			p.updated_at
		FROM erp_products p
		LEFT JOIN erp_categories c
		  ON c.id = p.category_id
		 AND c.business_id = p.business_id
		LEFT JOIN product_stock_total pst
		  ON pst.business_id = p.business_id
		 AND pst.product_id = p.id
		LEFT JOIN erp_product_images img
		  ON img.id = (
				SELECT i.id
				FROM erp_product_images i
				WHERE i.business_id = p.business_id
				  AND i.product_id = p.id
				ORDER BY i.is_primary DESC, i.position ASC, i.id ASC
				LIMIT 1
		  )
		WHERE p.business_id = ?
		  AND p.is_active = TRUE
		ORDER BY c.name, p.name
	`, businessID)
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
			barcode    sql.NullString
			salePrice  string
			taxRate    string
			desc       sql.NullString
			imageURL   sql.NullString
			stock      string
			active     sql.NullBool
			createdAt  sql.NullTime
			updatedAt  sql.NullTime
			item       domain.MenuItem
		)
		if err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.BusinessID,
			&categoryID,
			&category,
			&item.Name,
			&sku,
			&barcode,
			&salePrice,
			&taxRate,
			&desc,
			&imageURL,
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
		if barcode.Valid {
			item.Barcode = barcode.String
		}
		item.SalePrice = decimalToFloat(salePrice)
		item.TaxRate = decimalToFloat(taxRate)
		if desc.Valid {
			item.Description = desc.String
		}
		if imageURL.Valid {
			item.ImageURL = imageURL.String
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
			BusinessID:  item.BusinessID,
			CategoryID:  item.CategoryID,
			Category:    item.Category,
			Name:        item.Name,
			SKU:         item.SKU,
			Barcode:     item.Barcode,
			SalePrice:   item.SalePrice,
			TaxRate:     item.TaxRate,
			Description: item.Description,
			ImageURL:    item.ImageURL,
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

func (r *Repository) QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error) {
	result, err := r.queries.QuickCreateMenuItem(ctx, sqlc.QuickCreateMenuItemParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Name:       input.Name,
		Sku:        sql.NullString{String: input.SKU, Valid: input.SKU != ""},
		SalePrice:  floatToDecimal(input.SalePrice),
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
