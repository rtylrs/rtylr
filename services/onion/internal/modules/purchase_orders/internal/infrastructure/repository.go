package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListPurchaseOrders(ctx context.Context, ownerID int64, status string) ([]domain.PurchaseOrder, error) {
	query := `
		SELECT id, owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
		       expected_delivery, received_at, notes, created_by, created_at, updated_at
		FROM erp_purchase_orders
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list purchase orders", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var orders []domain.PurchaseOrder
	for rows.Next() {
		var po domain.PurchaseOrder
		if err := po.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan purchase order row", logger.Error(err))
			return nil, err
		}
		orders = append(orders, po)
	}

	return orders, rows.Err()
}

func (r *Repository) GetPurchaseOrderItems(ctx context.Context, purchaseOrderID int64) ([]domain.PurchaseOrderItem, error) {
	query := `
		SELECT id, purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost
		FROM erp_purchase_order_items
		WHERE purchase_order_id = ?`

	rows, err := r.pool.QueryContext(ctx, query, purchaseOrderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order items", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	var items []domain.PurchaseOrderItem
	for rows.Next() {
		var item domain.PurchaseOrderItem
		if err := item.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan purchase order item row", logger.Error(err))
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) CreatePurchaseOrder(ctx context.Context, ownerID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", logger.Error(err))
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var subtotal float64
	for _, item := range input.Items {
		subtotal += item.QuantityOrdered * item.UnitCost
	}

	total := subtotal + input.TaxAmount

	var expectedDelivery interface{}
	if input.ExpectedDelivery != "" {
		expectedDelivery = input.ExpectedDelivery
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO erp_purchase_orders
		(owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
		 expected_delivery, notes, created_by)
		VALUES (?, ?, ?, 'draft', ?, ?, ?, ?, ?, ?)`,
		ownerID,
		input.SupplierID,
		input.PONumber,
		subtotal,
		input.TaxAmount,
		total,
		expectedDelivery,
		nullString(input.Notes),
		nullInt64(createdBy),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert purchase order", logger.Error(err))
		return 0, err
	}

	poID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range input.Items {
		itemTotal := item.QuantityOrdered * item.UnitCost

		_, err = tx.ExecContext(ctx, `
			INSERT INTO erp_purchase_order_items
			(purchase_order_id, product_id, quantity_ordered, quantity_received, unit_cost, total_cost)
			VALUES (?, ?, ?, 0, ?, ?)`,
			poID,
			item.ProductID,
			item.QuantityOrdered,
			item.UnitCost,
			itemTotal,
		)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to insert purchase order item", logger.Error(err))
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit purchase order transaction", logger.Error(err))
		return 0, err
	}

	return poID, nil
}

func (r *Repository) UpdatePurchaseOrderStatus(ctx context.Context, ownerID, id int64, status string) error {
	query := `UPDATE erp_purchase_orders SET status = ? WHERE owner_id = ? AND id = ?`

	if status == string(domain.POStatusReceived) {
		query = `UPDATE erp_purchase_orders SET status = ?, received_at = NOW() WHERE owner_id = ? AND id = ?`
	}

	_, err := r.pool.ExecContext(ctx, query, status, ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update purchase order status", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) AdjustInventory(ctx context.Context, ownerID, productID int64, change float64, txType domain.InventoryTransactionType, notes string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
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
		return err
	}

	newQty := currentQty + change

	_, err = tx.ExecContext(ctx,
		`UPDATE erp_products SET stock_quantity = ? WHERE owner_id = ? AND id = ?`,
		newQty, ownerID, productID,
	)
	if err != nil {
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
		return err
	}

	return tx.Commit()
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
