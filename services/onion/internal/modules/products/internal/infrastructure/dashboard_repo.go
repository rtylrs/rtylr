package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// GetDashboard aggregates the counts + recent records used by the ERP
// dashboard. Counts run as three small queries rather than one UNION
// because MySQL optimises simple COUNT(*) scans extremely well.
func (r *Repository) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats
	db := r.pool

	// Total active products.
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_products WHERE owner_id = ? AND is_active = TRUE`,
		ownerID,
	).Scan(&stats.TotalProducts); err != nil {
		logger.Ctx(ctx).Error("Failed to count products", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	// Low stock: products at or below their low_stock_threshold.
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_products
		 WHERE owner_id = ? AND is_active = TRUE
		   AND track_inventory = TRUE
		   AND low_stock_threshold IS NOT NULL
		   AND stock_quantity <= low_stock_threshold`,
		ownerID,
	).Scan(&stats.LowStockCount); err != nil {
		logger.Ctx(ctx).Error("Failed to count low stock", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	// Pending purchase orders (draft or sent, not yet received).
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_purchase_orders
		 WHERE owner_id = ? AND status IN ('draft','sent','partial')`,
		ownerID,
	).Scan(&stats.PendingOrders); err != nil {
		logger.Ctx(ctx).Error("Failed to count pending POs", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	// Active suppliers.
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM erp_suppliers WHERE owner_id = ? AND is_active = TRUE`,
		ownerID,
	).Scan(&stats.TotalSuppliers); err != nil {
		logger.Ctx(ctx).Error("Failed to count suppliers", logger.Error(err))
		return domain.DashboardStats{}, err
	}

	// Inventory value = sum(stock_quantity * cost_price) for tracked products.
	var invValue sql.NullFloat64
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(stock_quantity * cost_price), 0)
		 FROM erp_products
		 WHERE owner_id = ? AND is_active = TRUE AND track_inventory = TRUE`,
		ownerID,
	).Scan(&invValue); err != nil {
		logger.Ctx(ctx).Error("Failed to sum inventory value", logger.Error(err))
		return domain.DashboardStats{}, err
	}
	if invValue.Valid {
		stats.InventoryValue = invValue.Float64
	}

	// Low stock items (top 5) for the dashboard sidebar card.
	items, err := r.topLowStockItems(ctx, ownerID, 5)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	stats.LowStockItems = items

	// Recent purchase orders (top 5) with supplier name.
	recent, err := r.recentPurchaseOrders(ctx, ownerID, 5)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	stats.RecentOrders = recent

	// Top 5 categories by product count.
	cats, err := r.topCategories(ctx, ownerID, 5)
	if err != nil {
		return domain.DashboardStats{}, err
	}
	stats.TopCategories = cats

	return stats, nil
}

func (r *Repository) topLowStockItems(ctx context.Context, ownerID int64, limit int) ([]domain.LowStockItem, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT p.id, p.name, COALESCE(p.sku, ''),
		       p.stock_quantity, COALESCE(p.low_stock_threshold, 0),
		       ''
		FROM erp_products p
		WHERE p.owner_id = ? AND p.is_active = TRUE AND p.track_inventory = TRUE
		  AND p.low_stock_threshold IS NOT NULL
		  AND p.stock_quantity <= p.low_stock_threshold
		ORDER BY (p.stock_quantity - p.low_stock_threshold) ASC, p.stock_quantity ASC
		LIMIT ?`, ownerID, limit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list low stock items", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	items := make([]domain.LowStockItem, 0, limit)
	for rows.Next() {
		var it domain.LowStockItem
		if err := rows.Scan(&it.ProductID, &it.Name, &it.SKU,
			&it.CurrentStock, &it.ReorderPoint, &it.WarehouseName); err != nil {
			logger.Ctx(ctx).Error("Failed to scan low stock item", logger.Error(err))
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) recentPurchaseOrders(ctx context.Context, ownerID int64, limit int) ([]domain.RecentPO, error) {
	// cross-schema read from erp_suppliers — join for display name
	rows, err := r.pool.QueryContext(ctx, `
		SELECT po.id, po.po_number, COALESCE(s.name, 'Unknown'),
		       po.status, po.total, po.created_at
		FROM erp_purchase_orders po
		LEFT JOIN erp_suppliers s ON s.id = po.supplier_id
		WHERE po.owner_id = ?
		ORDER BY po.created_at DESC
		LIMIT ?`, ownerID, limit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recent POs", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	pos := make([]domain.RecentPO, 0, limit)
	for rows.Next() {
		var p domain.RecentPO
		var createdAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.PONumber, &p.SupplierName,
			&p.Status, &p.Total, &createdAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan recent PO", logger.Error(err))
			return nil, err
		}
		if createdAt.Valid {
			p.CreatedAt = createdAt.Time
		}
		pos = append(pos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pos, nil
}

func (r *Repository) topCategories(ctx context.Context, ownerID int64, limit int) ([]domain.TopCategory, error) {
	// cross-schema read from erp_categories — join for display name
	rows, err := r.pool.QueryContext(ctx, `
		SELECT COALESCE(c.name, 'Uncategorized') AS name,
		       COUNT(p.id) AS product_count
		FROM erp_products p
		LEFT JOIN erp_categories c ON c.id = p.category_id
		WHERE p.owner_id = ? AND p.is_active = TRUE
		GROUP BY c.id, c.name
		ORDER BY product_count DESC
		LIMIT ?`, ownerID, limit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list top categories", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	cats := make([]domain.TopCategory, 0, limit)
	for rows.Next() {
		var c domain.TopCategory
		if err := rows.Scan(&c.Name, &c.ProductCount); err != nil {
			logger.Ctx(ctx).Error("Failed to scan top category", logger.Error(err))
			return nil, err
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cats, nil
}

// ListReorderSuggestions returns products that have fallen to or below
// their reorder threshold, with a suggested order quantity. Without
// sales-history integration the suggested qty is just the recorded
// reorder_quantity; confidence is "low" to flag this to the UI.
func (r *Repository) ListReorderSuggestions(ctx context.Context, ownerID int64) ([]domain.ReorderSuggestion, error) {
	rows, err := r.pool.QueryContext(ctx, `
		SELECT id, name, COALESCE(sku, ''),
		       stock_quantity,
		       COALESCE(low_stock_threshold, 0),
		       COALESCE(reorder_quantity, 0)
		FROM erp_products
		WHERE owner_id = ? AND is_active = TRUE AND track_inventory = TRUE
		  AND low_stock_threshold IS NOT NULL
		  AND stock_quantity <= low_stock_threshold
		ORDER BY (stock_quantity - low_stock_threshold) ASC
		LIMIT 100`, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reorder suggestions", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	out := make([]domain.ReorderSuggestion, 0, 100)
	for rows.Next() {
		var s domain.ReorderSuggestion
		if err := rows.Scan(&s.ProductID, &s.Name, &s.SKU,
			&s.CurrentStock, &s.ReorderPoint, &s.SuggestedQty); err != nil {
			logger.Ctx(ctx).Error("Failed to scan reorder suggestion", logger.Error(err))
			return nil, err
		}
		if s.SuggestedQty == 0 {
			// Fall back to doubling the reorder point if no reorder_quantity is set.
			s.SuggestedQty = s.ReorderPoint * 2
		}
		s.Confidence = "low"
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListExpiringItems is stubbed to return an empty slice. Batch-level
// expiry tracking isn't modelled yet — once an erp_product_batches table
// exists, this method should join and filter by expiry_date <= NOW() + N days.
func (r *Repository) ListExpiringItems(_ context.Context, _ int64, _ int) ([]domain.ExpiryItem, error) {
	return []domain.ExpiryItem{}, nil
}

func closeRows(ctx context.Context, rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logger.Ctx(ctx).Warn("Failed to close rows", logger.Error(err))
	}
}
