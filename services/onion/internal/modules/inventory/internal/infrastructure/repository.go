package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
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

	var currentQty float64

	err = tx.QueryRowContext(ctx,
		`SELECT stock_quantity FROM erp_products WHERE owner_id = ? AND id = ? FOR UPDATE`,
		ownerID, productID,
	).Scan(&currentQty)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get current stock quantity", logger.Error(err))

		return err
	}

	newQty := currentQty + change

	_, err = tx.ExecContext(ctx,
		`UPDATE erp_products SET stock_quantity = ? WHERE owner_id = ? AND id = ?`,
		newQty, ownerID, productID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock quantity", logger.Error(err))

		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO erp_inventory_transactions
		(owner_id, product_id, type, quantity_change, quantity_before, quantity_after, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ownerID,
		productID,
		string(txType),
		change,
		currentQty,
		newQty,
		nullString(notes),
	)
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
	query := `
		SELECT id, name, sku, unit, stock_quantity, low_stock_threshold,
		       reorder_quantity, cost_price, sale_price,
		       (stock_quantity * cost_price) AS stock_value,
		       (stock_quantity <= low_stock_threshold) AS is_low_stock
		FROM erp_products
		WHERE owner_id = ? AND track_inventory = TRUE AND is_active = TRUE
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory report", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	return scanInventoryReportRows(ctx, rows)
}

func (r *Repository) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	query := `
		SELECT id, name, sku, unit, stock_quantity, low_stock_threshold,
		       reorder_quantity, cost_price, sale_price,
		       (stock_quantity * cost_price) AS stock_value,
		       1 AS is_low_stock
		FROM erp_products
		WHERE owner_id = ? AND track_inventory = TRUE AND is_active = TRUE
		  AND stock_quantity <= low_stock_threshold
		ORDER BY stock_quantity ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get low stock products", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	return scanInventoryReportRows(ctx, rows)
}

func (r *Repository) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	query := `
		SELECT
			COALESCE(c.name, 'Uncategorized') AS category_name,
			'' AS warehouse_name,
			COUNT(*) AS total_items,
			SUM(p.stock_quantity * p.sale_price) AS total_value,
			SUM(p.stock_quantity * p.cost_price) AS total_cost
		FROM erp_products p
		LEFT JOIN erp_categories c ON c.id = p.category_id AND c.owner_id = p.owner_id
		WHERE p.owner_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE
		GROUP BY c.name
		ORDER BY total_value DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory valuation", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var items []domain.ValuationItem

	for rows.Next() {
		var item domain.ValuationItem
		var totalValue sql.NullFloat64
		var totalCost sql.NullFloat64

		if err := rows.Scan(
			&item.CategoryName,
			&item.WarehouseName,
			&item.TotalItems,
			&totalValue,
			&totalCost,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan valuation row", logger.Error(err))

			return nil, err
		}

		if totalValue.Valid {
			item.TotalValue = totalValue.Float64
		}

		if totalCost.Valid {
			item.TotalCost = totalCost.Float64
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func scanInventoryReportRows(ctx context.Context, rows *sql.Rows) ([]domain.InventoryReportItem, error) {
	var report []domain.InventoryReportItem

	for rows.Next() {
		var item domain.InventoryReportItem
		var sku sql.NullString
		var unit sql.NullString
		var lowStockThreshold sql.NullFloat64
		var reorderQty sql.NullFloat64
		var isLowStock int

		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&sku,
			&unit,
			&item.StockQuantity,
			&lowStockThreshold,
			&reorderQty,
			&item.CostPrice,
			&item.SalePrice,
			&item.StockValue,
			&isLowStock,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan inventory report row", logger.Error(err))

			return nil, err
		}

		if sku.Valid {
			item.SKU = sku.String
		}

		if unit.Valid {
			item.Unit = unit.String
		}

		if lowStockThreshold.Valid {
			item.LowStockThreshold = lowStockThreshold.Float64
		}

		if reorderQty.Valid {
			item.ReorderQuantity = reorderQty.Float64
		}

		item.IsLowStock = isLowStock == 1

		report = append(report, item)
	}

	return report, rows.Err()
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
