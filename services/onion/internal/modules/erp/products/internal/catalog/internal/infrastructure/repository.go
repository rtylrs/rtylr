package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.MenuItem, error) {
	rows, err := r.queries.ListMenuItemsWithDetails(ctx, sqlc.ListMenuItemsWithDetailsParams{
		BusinessID:   businessID,
		BusinessID_2: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list menu items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.MenuItem, 0, len(rows))
	for _, row := range rows {
		item := domain.MenuItem{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			SalePrice:  sqlutil.DecimalToFloat(row.SalePrice),
			CostPrice:  sqlutil.DecimalToFloat(row.CostPrice),
			TaxRate:    sqlutil.DecimalToFloat(row.TaxRate),
			Stock:      stockToFloat(row.StockQuantity),
			IsActive:   row.IsActive.Valid && row.IsActive.Bool,
			Category:   row.CategoryName,
		}
		if row.CategoryID.Valid {
			item.CategoryID = row.CategoryID.Int64
		}
		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}
		if row.Barcode.Valid {
			item.Barcode = row.Barcode.String
		}
		if row.Description.Valid {
			item.Description = row.Description.String
		}
		if row.ImageUrl != "" {
			item.ImageURL = row.ImageUrl
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			item.UpdatedAt = row.UpdatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

// QuickCreate writes the business ID into owner_id, as every ERP product row
// does, and optionally stores the primary storefront image atomically.
func (r *Repository) QuickCreate(ctx context.Context, businessID int64, input domain.MenuItemInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	result, err := qtx.QuickCreateMenuItem(ctx, sqlc.QuickCreateMenuItemParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		Name:        input.Name,
		Sku:         sql.NullString{String: input.SKU, Valid: input.SKU != ""},
		SalePrice:   sqlutil.FloatToDecimal(input.SalePrice),
		Description: input.Description,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to quick-create menu item", logger.Error(err))
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if input.ImageURL != "" {
		if err := qtx.InsertPrimaryMenuItemImage(ctx, sqlc.InsertPrimaryMenuItemImageParams{
			OwnerID:    businessID,
			BusinessID: businessID,
			ProductID:  id,
			Url:        input.ImageURL,
			Alt:        sql.NullString{String: input.Name, Valid: input.Name != ""},
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to save menu item image", logger.Error(err))
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	locationID, err := stock.DefaultLocationID(ctx, tx, businessID)
	if err != nil {
		return err
	}
	if _, _, err := stock.SetTotalTx(ctx, tx, stock.Movement{
		BusinessID:    businessID,
		LocationID:    locationID,
		ProductID:     id,
		EntryType:     stock.EntryAdjustment,
		ReferenceType: "pos_menu",
		Channel:       "pos",
		Note:          "POS menu stock edit",
	}, input.Stock); err != nil {
		return err
	}

	var setDesc int32
	descVal := ""
	if input.Description != nil {
		setDesc = 1
		descVal = *input.Description
	}

	qtx := r.queries.WithTx(tx)
	result, err := qtx.UpdateMenuItem(ctx, sqlc.UpdateMenuItemParams{
		Name:           input.Name,
		SalePrice:      sqlutil.FloatToDecimal(input.SalePrice),
		SetDescription: setDesc,
		Description:    descVal,
		IsActive:       sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:     businessID,
		ID:             id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update menu item", logger.Error(err))
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	if input.ImageURL != nil {
		if err := qtx.ClearPrimaryMenuItemImage(ctx, sqlc.ClearPrimaryMenuItemImageParams{
			BusinessID: businessID,
			ProductID:  id,
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to clear previous menu item image", logger.Error(err))
			return err
		}
		if *input.ImageURL != "" {
			if err := qtx.InsertPrimaryMenuItemImage(ctx, sqlc.InsertPrimaryMenuItemImageParams{
				OwnerID:    businessID,
				BusinessID: businessID,
				ProductID:  id,
				Url:        *input.ImageURL,
				Alt:        sql.NullString{String: input.Name, Valid: input.Name != ""},
			}); err != nil {
				logger.Ctx(ctx).Error("Failed to insert updated menu item image", logger.Error(err))
				return err
			}
		}
	}
	return tx.Commit()
}

// stockToFloat converts the value returned by the COALESCE(pst.total_quantity, p.stock_quantity)
// aggregate (sqlc emits interface{} for a DECIMAL sourced from the product_stock_total view) into
// a float64, handling the driver representations MySQL may return for a DECIMAL column.
func stockToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return sqlutil.DecimalToFloat(string(val))
	case string:
		return sqlutil.DecimalToFloat(val)
	default:
		return sqlutil.DecimalToFloat(fmt.Sprintf("%v", val))
	}
}
