package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/inventory/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) AdjustInventory(ctx context.Context, ownerID, productID int64, change float64, txType domain.InventoryTransactionType, notes string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin inventory transaction", logger.Error(err))

		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	stockQtyStr, err := qtx.GetProductStockForUpdate(ctx, sqlc.GetProductStockForUpdateParams{
		OwnerID: ownerID,
		ID:      productID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get current stock quantity", logger.Error(err))

		return err
	}

	currentQty, _ := strconv.ParseFloat(stockQtyStr, 64)
	newQty := currentQty + change

	err = qtx.UpdateProductStock(ctx, sqlc.UpdateProductStockParams{
		StockQuantity: strconv.FormatFloat(newQty, 'f', -1, 64),
		OwnerID:       ownerID,
		ID:            productID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock quantity", logger.Error(err))

		return err
	}

	_, err = qtx.InsertInventoryTransaction(ctx, sqlc.InsertInventoryTransactionParams{
		OwnerID:        ownerID,
		ProductID:      productID,
		Type:           sqlc.ErpInventoryTransactionsType(txType),
		QuantityChange: strconv.FormatFloat(change, 'f', -1, 64),
		QuantityBefore: strconv.FormatFloat(currentQty, 'f', -1, 64),
		QuantityAfter:  strconv.FormatFloat(newQty, 'f', -1, 64),
		Notes:          nullString(notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert inventory transaction", logger.Error(err))

		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit inventory transaction", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) GetInventoryReport(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	rows, err := r.queries.GetInventoryReport(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory report", logger.Error(err))

		return nil, err
	}

	report := make([]domain.InventoryReportItem, 0, len(rows))

	for _, row := range rows {
		var item domain.InventoryReportItem
		item.ProductID = row.ID
		item.ProductName = row.Name

		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}

		if row.Unit.Valid {
			item.Unit = row.Unit.String
		}

		item.StockQuantity, _ = strconv.ParseFloat(row.StockQuantity, 64)
		item.CostPrice, _ = strconv.ParseFloat(row.CostPrice, 64)
		item.SalePrice, _ = strconv.ParseFloat(row.SalePrice, 64)
		item.StockValue = float64(row.StockValue)
		item.IsLowStock = row.IsLowStock

		if row.LowStockThreshold.Valid {
			item.LowStockThreshold, _ = strconv.ParseFloat(row.LowStockThreshold.String, 64)
		}

		if row.ReorderQuantity.Valid {
			item.ReorderQuantity, _ = strconv.ParseFloat(row.ReorderQuantity.String, 64)
		}

		report = append(report, item)
	}

	return report, nil
}

func (r *Repository) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	rows, err := r.queries.GetLowStockProducts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get low stock products", logger.Error(err))

		return nil, err
	}

	report := make([]domain.InventoryReportItem, 0, len(rows))

	for _, row := range rows {
		var item domain.InventoryReportItem
		item.ProductID = row.ID
		item.ProductName = row.Name

		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}

		if row.Unit.Valid {
			item.Unit = row.Unit.String
		}

		item.StockQuantity, _ = strconv.ParseFloat(row.StockQuantity, 64)
		item.CostPrice, _ = strconv.ParseFloat(row.CostPrice, 64)
		item.SalePrice, _ = strconv.ParseFloat(row.SalePrice, 64)
		item.StockValue = float64(row.StockValue)
		item.IsLowStock = row.IsLowStock != 0

		if row.LowStockThreshold.Valid {
			item.LowStockThreshold, _ = strconv.ParseFloat(row.LowStockThreshold.String, 64)
		}

		if row.ReorderQuantity.Valid {
			item.ReorderQuantity, _ = strconv.ParseFloat(row.ReorderQuantity.String, 64)
		}

		report = append(report, item)
	}

	return report, nil
}

func (r *Repository) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	rows, err := r.queries.GetValuation(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory valuation", logger.Error(err))

		return nil, err
	}

	items := make([]domain.ValuationItem, 0, len(rows))

	for _, row := range rows {
		var item domain.ValuationItem
		item.CategoryName = row.CategoryName
		item.WarehouseName = row.WarehouseName
		item.TotalItems = row.TotalItems

		if row.TotalValue != nil {
			switch v := row.TotalValue.(type) {
			case float64:
				item.TotalValue = v
			case []byte:
				item.TotalValue, _ = strconv.ParseFloat(string(v), 64)
			}
		}

		if row.TotalCost != nil {
			switch v := row.TotalCost.(type) {
			case float64:
				item.TotalCost = v
			case []byte:
				item.TotalCost, _ = strconv.ParseFloat(string(v), 64)
			}
		}

		items = append(items, item)
	}

	return items, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
