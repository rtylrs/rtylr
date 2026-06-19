package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/stock_core"
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

func (r *Repository) ListPurchaseOrders(ctx context.Context, businessID int64, status string, limit, offset int32) ([]domain.PurchaseOrder, error) {
	var rows []sqlc.ErpPurchaseOrder
	var err error

	if status != "" {
		rows, err = r.queries.ListPurchaseOrdersByStatus(ctx, sqlc.ListPurchaseOrdersByStatusParams{
			BusinessID: businessID,
			Status:     sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(status), Valid: true},
			Limit:      limit,
			Offset:     offset,
		})
	} else {
		rows, err = r.queries.ListPurchaseOrders(ctx, sqlc.ListPurchaseOrdersParams{
			BusinessID: businessID,
			Limit:      limit,
			Offset:     offset,
		})
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to list purchase orders", logger.Error(err))
		return nil, err
	}

	orders := make([]domain.PurchaseOrder, 0, len(rows))
	for _, row := range rows {
		orders = append(orders, purchaseOrderFromRow(row))
	}

	return orders, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.PurchaseOrder, error) {
	row, err := r.queries.GetPurchaseOrderByID(ctx, sqlc.GetPurchaseOrderByIDParams{
		ID:         id,
		BusinessID: businessID,
	})
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

	items := make([]domain.PurchaseOrderItem, 0, len(rows))
	for _, row := range rows {
		var item domain.PurchaseOrderItem
		item.ID = row.ID
		item.PurchaseOrderID = row.PurchaseOrderID
		item.ProductID = row.ProductID
		item.QuantityOrdered = parseDecimal(row.QuantityOrdered)
		item.UnitCost = parseDecimal(row.UnitCost)
		item.TotalCost = parseDecimal(row.TotalCost)

		if row.QuantityReceived.Valid {
			item.QuantityReceived = parseDecimal(row.QuantityReceived.String)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *Repository) CreatePurchaseOrder(ctx context.Context, businessID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error) {
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
		OwnerID:          createdBy,
		BusinessID:       businessID,
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

func (r *Repository) GetPurchaseOrderStatus(ctx context.Context, businessID, id int64) (string, error) {
	status, err := r.queries.GetPurchaseOrderStatus(ctx, sqlc.GetPurchaseOrderStatusParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrPurchaseOrderNotFound
		}

		logger.Ctx(ctx).Error("Failed to get purchase order status",
			logger.Any("business_id", businessID),
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

func (r *Repository) ApprovePurchaseOrder(ctx context.Context, businessID, id int64) error {
	err := r.queries.ApprovePurchaseOrder(ctx, sqlc.ApprovePurchaseOrderParams{
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to approve purchase order",
			logger.Any("business_id", businessID),
			logger.Any("purchase_order_id", id),
			logger.Error(err),
		)

		return err
	}

	return nil
}

func (r *Repository) UpdatePurchaseOrderStatus(ctx context.Context, businessID, id int64, status string) error {
	var err error

	statusVal := sqlc.NullErpPurchaseOrdersStatus{
		ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(status),
		Valid:                   true,
	}

	if status == string(domain.POStatusReceived) {
		err = r.queries.UpdatePurchaseOrderStatusReceived(ctx, sqlc.UpdatePurchaseOrderStatusReceivedParams{
			Status:     statusVal,
			BusinessID: businessID,
			ID:         id,
		})
	} else {
		err = r.queries.UpdatePurchaseOrderStatus(ctx, sqlc.UpdatePurchaseOrderStatusParams{
			Status:     statusVal,
			BusinessID: businessID,
			ID:         id,
		})
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to update purchase order status", logger.Error(err))
		return err
	}

	return nil
}

// ReceivePurchaseOrder atomically receives a purchase order: it verifies the PO
// belongs to the caller and isn't already received (row-locked), then for every
// line item increments the product's on-hand stock (row-locked) and writes a
// ledgered inventory transaction referencing this PO, marks the items received,
// and flips the PO to 'received'. The whole thing is one transaction, so a
// failure on any item rolls back every stock change — and the
// already-received guard makes a repeated call a safe no-op (no double-add).
func (r *Repository) ReceivePurchaseOrder(ctx context.Context, businessID, poID, createdBy int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin receive transaction", logger.Error(err))
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	// Lock the PO row and confirm ownership + current status.
	status, err := qtx.GetPurchaseOrderForUpdate(ctx, sqlc.GetPurchaseOrderForUpdateParams{
		BusinessID: businessID,
		ID:         poID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrPurchaseOrderNotFound
		return err
	}
	if err != nil {
		return err
	}
	if status.Valid && status.ErpPurchaseOrdersStatus == sqlc.ErpPurchaseOrdersStatus(domain.POStatusReceived) {
		err = domain.ErrPurchaseOrderAlreadyReceived
		return err
	}

	items, err := qtx.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		return err
	}

	// Lock product rows in ascending product order to avoid deadlocks with other
	// concurrent multi-item stock operations.
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })

	// PO receiving lands stock in the tenant's default warehouse location
	// (per-warehouse routing via destination_location_id is a later slice).
	locID, err := stock_core.DefaultLocationID(ctx, tx, businessID)
	if err != nil {
		return err
	}

	for _, it := range items {
		qty, perr := strconv.ParseFloat(it.QuantityOrdered, 64)
		if perr != nil {
			err = perr
			return err
		}

		before, after, merr := stock_core.ApplyMovementTx(ctx, tx, stock_core.Movement{
			OwnerID:       businessID,
			LocationID:    locID,
			ProductID:     it.ProductID,
			EntryType:     stock_core.EntryPurchase,
			Change:        qty,
			ReferenceID:   &poID,
			ReferenceType: "purchase_order",
			Channel:       "erp",
			Note:          fmt.Sprintf("received from purchase order %d", poID),
			CreatedBy:     &createdBy,
		})
		if merr != nil {
			err = merr
			return err
		}

		// Legacy ledger row (read-compat).
		if _, err = qtx.InsertInventoryTransactionWithRef(ctx, sqlc.InsertInventoryTransactionWithRefParams{
			OwnerID:        createdBy,
			BusinessID:     businessID,
			ProductID:      it.ProductID,
			Type:           sqlc.ErpInventoryTransactionsType(domain.InventoryTypePurchase),
			QuantityChange: strconv.FormatFloat(qty, 'f', -1, 64),
			QuantityBefore: strconv.FormatFloat(before, 'f', -1, 64),
			QuantityAfter:  strconv.FormatFloat(after, 'f', -1, 64),
			ReferenceID:    nullInt64(poID),
			ReferenceType:  nullString("purchase_order"),
			CreatedBy:      nullInt64(createdBy),
			Notes:          nullString(fmt.Sprintf("received from purchase order %d", poID)),
		}); err != nil {
			return err
		}
	}

	if err = qtx.MarkPurchaseOrderItemsReceived(ctx, poID); err != nil {
		return err
	}

	if err = qtx.UpdatePurchaseOrderStatusReceived(ctx, sqlc.UpdatePurchaseOrderStatusReceivedParams{
		Status:     sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(domain.POStatusReceived), Valid: true},
		BusinessID: businessID,
		ID:         poID,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

func purchaseOrderFromRow(row sqlc.ErpPurchaseOrder) domain.PurchaseOrder {
	var po domain.PurchaseOrder
	po.ID = row.ID
	po.OwnerID = row.OwnerID
	po.BusinessID = row.BusinessID
	po.SupplierID = row.SupplierID
	po.PONumber = row.PoNumber

	if row.Status.Valid {
		po.Status = domain.PurchaseOrderStatus(row.Status.ErpPurchaseOrdersStatus)
	}

	if row.Subtotal.Valid {
		po.Subtotal = parseDecimal(row.Subtotal.String)
	}

	if row.TaxAmount.Valid {
		po.TaxAmount = parseDecimal(row.TaxAmount.String)
	}

	if row.Total.Valid {
		po.Total = parseDecimal(row.Total.String)
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

// parseDecimal converts a DB decimal string to float64, logging (rather than
// silently swallowing) corrupt values so bad money/quantity data surfaces.
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse purchase-order decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
