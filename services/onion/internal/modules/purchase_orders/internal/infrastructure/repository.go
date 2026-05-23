package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListPurchaseOrders(ctx context.Context, ownerID int64, status string) ([]domain.PurchaseOrder, error) {
	var rows []sqlc.ErpPurchaseOrder
	var err error

	if status != "" {
		rows, err = r.queries.ListPurchaseOrdersByStatus(ctx, sqlc.ListPurchaseOrdersByStatusParams{
			OwnerID: ownerID,
			Status:  sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(status), Valid: true},
		})
	} else {
		rows, err = r.queries.ListPurchaseOrders(ctx, ownerID)
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to list purchase orders", logger.Error(err))
		return nil, err
	}

	var orders []domain.PurchaseOrder
	for _, row := range rows {
		orders = append(orders, purchaseOrderFromRow(row))
	}

	return orders, nil
}

func (r *Repository) GetByID(ctx context.Context, ownerID, id int64) (domain.PurchaseOrder, error) {
	const q = `SELECT id, owner_id, supplier_id, po_number, status, subtotal, tax_amount, total,
	             expected_delivery, received_at, notes, created_by, created_at, updated_at
	           FROM erp_purchase_orders WHERE id = ? AND owner_id = ?`
	var row sqlc.ErpPurchaseOrder
	err := r.pool.QueryRowContext(ctx, q, id, ownerID).Scan(
		&row.ID, &row.OwnerID, &row.SupplierID, &row.PoNumber, &row.Status,
		&row.Subtotal, &row.TaxAmount, &row.Total, &row.ExpectedDelivery,
		&row.ReceivedAt, &row.Notes, &row.CreatedBy, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PurchaseOrder{}, domain.ErrPurchaseOrderNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order by ID", logger.Error(err))
		return domain.PurchaseOrder{}, err
	}
	po := purchaseOrderFromRow(row)
	items, itemErr := r.GetPurchaseOrderItems(ctx, po.ID)
	if itemErr != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order items", logger.Error(itemErr))
		return domain.PurchaseOrder{}, itemErr
	}
	po.Items = items
	return po, nil
}

func (r *Repository) GetPurchaseOrderItems(ctx context.Context, purchaseOrderID int64) ([]domain.PurchaseOrderItem, error) {
	rows, err := r.queries.GetPurchaseOrderItems(ctx, purchaseOrderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get purchase order items", logger.Error(err))
		return nil, err
	}

	var items []domain.PurchaseOrderItem
	for _, row := range rows {
		var item domain.PurchaseOrderItem
		item.ID = row.ID
		item.PurchaseOrderID = row.PurchaseOrderID
		item.ProductID = row.ProductID
		item.QuantityOrdered, _ = strconv.ParseFloat(row.QuantityOrdered, 64)
		item.UnitCost, _ = strconv.ParseFloat(row.UnitCost, 64)
		item.TotalCost, _ = strconv.ParseFloat(row.TotalCost, 64)

		if row.QuantityReceived.Valid {
			item.QuantityReceived, _ = strconv.ParseFloat(row.QuantityReceived.String, 64)
		}

		items = append(items, item)
	}

	return items, nil
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

	qtx := r.queries.WithTx(tx)

	var subtotal float64
	for _, item := range input.Items {
		subtotal += item.QuantityOrdered * item.UnitCost
	}

	total := subtotal + input.TaxAmount

	var expectedDelivery sql.NullTime
	if input.ExpectedDelivery != "" {
		t, parseErr := time.Parse("2006-01-02", input.ExpectedDelivery)
		if parseErr == nil {
			expectedDelivery = sql.NullTime{Time: t, Valid: true}
		}
	}

	res, err := qtx.InsertPurchaseOrder(ctx, sqlc.InsertPurchaseOrderParams{
		OwnerID:          ownerID,
		SupplierID:       input.SupplierID,
		PoNumber:         input.PONumber,
		Subtotal:         sql.NullString{String: strconv.FormatFloat(subtotal, 'f', 2, 64), Valid: true},
		TaxAmount:        sql.NullString{String: strconv.FormatFloat(input.TaxAmount, 'f', 2, 64), Valid: true},
		Total:            sql.NullString{String: strconv.FormatFloat(total, 'f', 2, 64), Valid: true},
		ExpectedDelivery: expectedDelivery,
		Notes:            nullString(input.Notes),
		CreatedBy:        nullInt64(createdBy),
	})
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

		_, err = qtx.InsertPurchaseOrderItem(ctx, sqlc.InsertPurchaseOrderItemParams{
			PurchaseOrderID: poID,
			ProductID:       item.ProductID,
			QuantityOrdered: strconv.FormatFloat(item.QuantityOrdered, 'f', -1, 64),
			UnitCost:        strconv.FormatFloat(item.UnitCost, 'f', -1, 64),
			TotalCost:       strconv.FormatFloat(itemTotal, 'f', -1, 64),
		})
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

func (r *Repository) GetPurchaseOrderStatus(ctx context.Context, ownerID, id int64) (string, error) {
	status, err := r.queries.GetPurchaseOrderStatus(ctx, sqlc.GetPurchaseOrderStatusParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrPurchaseOrderNotFound
		}

		logger.Ctx(ctx).Error("Failed to get purchase order status",
			logger.Any("owner_id", ownerID),
			logger.Any("purchase_order_id", id),
			logger.Error(err),
		)

		return "", err
	}

	if !status.Valid {
		return "", nil
	}

	return string(status.ErpPurchaseOrdersStatus), nil
}

func (r *Repository) ApprovePurchaseOrder(ctx context.Context, ownerID, id int64) error {
	err := r.queries.ApprovePurchaseOrder(ctx, sqlc.ApprovePurchaseOrderParams{
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to approve purchase order",
			logger.Any("owner_id", ownerID),
			logger.Any("purchase_order_id", id),
			logger.Error(err),
		)

		return err
	}

	return nil
}

func (r *Repository) UpdatePurchaseOrderStatus(ctx context.Context, ownerID, id int64, status string) error {
	var err error

	statusVal := sqlc.NullErpPurchaseOrdersStatus{
		ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(status),
		Valid:                   true,
	}

	if status == string(domain.POStatusReceived) {
		err = r.queries.UpdatePurchaseOrderStatusReceived(ctx, sqlc.UpdatePurchaseOrderStatusReceivedParams{
			Status:  statusVal,
			OwnerID: ownerID,
			ID:      id,
		})
	} else {
		err = r.queries.UpdatePurchaseOrderStatus(ctx, sqlc.UpdatePurchaseOrderStatusParams{
			Status:  statusVal,
			OwnerID: ownerID,
			ID:      id,
		})
	}

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

	qtx := r.queries.WithTx(tx)

	stockQtyStr, err := qtx.GetProductStockForUpdate(ctx, sqlc.GetProductStockForUpdateParams{
		OwnerID: ownerID,
		ID:      productID,
	})
	if err != nil {
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
		return err
	}

	return tx.Commit()
}

func purchaseOrderFromRow(row sqlc.ErpPurchaseOrder) domain.PurchaseOrder {
	var po domain.PurchaseOrder
	po.ID = row.ID
	po.OwnerID = row.OwnerID
	po.SupplierID = row.SupplierID
	po.PONumber = row.PoNumber

	if row.Status.Valid {
		po.Status = domain.PurchaseOrderStatus(row.Status.ErpPurchaseOrdersStatus)
	}

	if row.Subtotal.Valid {
		po.Subtotal, _ = strconv.ParseFloat(row.Subtotal.String, 64)
	}

	if row.TaxAmount.Valid {
		po.TaxAmount, _ = strconv.ParseFloat(row.TaxAmount.String, 64)
	}

	if row.Total.Valid {
		po.Total, _ = strconv.ParseFloat(row.Total.String, 64)
	}

	if row.ExpectedDelivery.Valid {
		po.ExpectedDelivery = &row.ExpectedDelivery.Time
	}

	if row.ReceivedAt.Valid {
		po.ReceivedAt = &row.ReceivedAt.Time
	}

	if row.Notes.Valid {
		po.Notes = row.Notes.String
	}

	if row.CreatedBy.Valid {
		po.CreatedBy = row.CreatedBy.Int64
	}

	if row.CreatedAt.Valid {
		po.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		po.UpdatedAt = row.UpdatedAt.Time
	}

	return po
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
