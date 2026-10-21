package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func validatePurchaseOrderReferences(ctx context.Context, qtx *sqlc.Queries, businessID int64, input domain.PurchaseOrderInput) error {
	if _, err := qtx.GetSupplierReference(ctx, sqlc.GetSupplierReferenceParams{BusinessID: businessID, ID: input.SupplierID}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPurchaseOrderReferenceNotFound
		}
		return err
	}
	for _, item := range input.Items {
		if _, err := qtx.GetProductReference(ctx, sqlc.GetProductReferenceParams{BusinessID: businessID, ID: item.ProductID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrPurchaseOrderReferenceNotFound
			}
			return err
		}
	}
	return nil
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
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	old, err := qtx.GetPurchaseOrderForUpdate(ctx, sqlc.GetPurchaseOrderForUpdateParams{BusinessID: businessID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrPurchaseOrderNotFound
	}
	if err != nil {
		return err
	}
	if old.Valid && old.ErpPurchaseOrdersStatus == sqlc.ErpPurchaseOrdersStatusReceived {
		return domain.ErrPurchaseOrderAlreadyReceived
	}
	if old.Valid && old.ErpPurchaseOrdersStatus == sqlc.ErpPurchaseOrdersStatusApproved {
		return domain.ErrPurchaseOrderAlreadyApproved
	}
	result, err := qtx.ApprovePurchaseOrder(ctx, sqlc.ApprovePurchaseOrderParams{BusinessID: businessID, ID: id})
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrPurchaseOrderInvalidTransition
	}
	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "purchase_order", EntityID: id, Action: "approved", OldValues: map[string]string{"status": string(old.ErpPurchaseOrdersStatus)}, NewValues: map[string]string{"status": "approved"}}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) UpdatePurchaseOrderStatus(ctx context.Context, businessID, id int64, status string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	old, err := qtx.GetPurchaseOrderForUpdate(ctx, sqlc.GetPurchaseOrderForUpdateParams{BusinessID: businessID, ID: id})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrPurchaseOrderNotFound
	}
	if err != nil {
		return err
	}
	currentStatus := domain.PurchaseOrderStatus(old.ErpPurchaseOrdersStatus)
	nextStatus := domain.PurchaseOrderStatus(status)
	if err := domain.ValidatePOStatusTransition(currentStatus, nextStatus); err != nil {
		return err
	}
	if currentStatus == nextStatus {
		return nil
	}
	if nextStatus == domain.POStatusReceived || nextStatus == domain.POStatusPartial {
		// These statuses record stock arriving; only ReceivePurchaseOrder
		// moves the stock that goes with them.
		return domain.ErrPurchaseOrderInvalidTransition
	}
	value := sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(status), Valid: true}
	result, err := qtx.UpdatePurchaseOrderStatus(ctx, sqlc.UpdatePurchaseOrderStatusParams{Status: value, BusinessID: businessID, ID: id})
	if err != nil {
		return err
	}
	changed, err := statusUpdateChanged(result)
	if err != nil {
		return err
	}
	if !changed {
		return domain.ErrPurchaseOrderInvalidTransition
	}
	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, EntityType: "purchase_order", EntityID: id, Action: "status_changed", OldValues: map[string]string{"status": string(old.ErpPurchaseOrdersStatus)}, NewValues: map[string]string{"status": status}}); err != nil {
		return err
	}
	return tx.Commit()
}

// ReceivePurchaseOrder atomically receives a purchase order: it verifies the PO
// belongs to the caller and isn't already received (row-locked), then for every
// line item increments the product's on-hand stock (row-locked) and writes a
// ledgered inventory transaction referencing this PO, marks the items received,
// and flips the PO to 'received'. The whole thing is one transaction, so a
// failure on any item rolls back every stock change — and the
// already-received guard makes a repeated call a safe no-op (no double-add).
func (r *Repository) ReceivePurchaseOrder(ctx context.Context, businessID, poID, createdBy int64, submitted []domain.PurchaseOrderItem) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin receive transaction", logger.Error(err))
		return err
	}

	// Always roll back when the function exits; after a successful commit the
	// rollback is a no-op (sql.ErrTxDone), and every early return is covered
	// even if it does not assign the outer err.
	defer func() { _ = tx.Rollback() }()

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
	if status.Valid && status.ErpPurchaseOrdersStatus == sqlc.ErpPurchaseOrdersStatus(domain.POStatusCancelled) {
		err = domain.ErrPurchaseOrderInvalidTransition
		return err
	}

	oldRow, err := qtx.GetPurchaseOrderByID(ctx, sqlc.GetPurchaseOrderByIDParams{BusinessID: businessID, ID: poID})
	if err != nil {
		return err
	}
	old := purchaseOrderFromRow(oldRow)
	items, err := qtx.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		return err
	}

	old.Items = make([]domain.PurchaseOrderItem, 0, len(items))
	for _, item := range items {
		old.Items = append(old.Items, purchaseOrderItemFromRow(item))
	}
	// Lock product rows in ascending product order to avoid deadlocks with other
	// concurrent multi-item stock operations.
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })

	// PO receiving lands stock in the tenant's default warehouse location
	// (per-warehouse routing via destination_location_id is a later slice).
	locID, err := stock.DefaultLocationID(ctx, tx, businessID)
	if err != nil {
		return err
	}

	// Index the submitted quantities by line id. They are range-checked below
	// against each line's stored quantity_ordered: the request's own
	// quantity_ordered is client-supplied and must never widen the bound.
	receivedByID := make(map[int64]float64, len(submitted))
	for _, item := range submitted {
		if item.QuantityReceived < 0 {
			err = fmt.Errorf("%w: item %d must not be negative", domain.ErrInvalidReceivedQuantity, item.ID)
			return err
		}
		receivedByID[item.ID] = item.QuantityReceived
	}
	fullyReceived := true
	receiptLines := make([]purchaseOrderReceiptLine, 0, len(items))
	receivedUpdates := make([]purchaseOrderItemReceivedUpdate, 0, len(items))
	for _, it := range items {
		ordered, perr := strconv.ParseFloat(it.QuantityOrdered, 64)
		if perr != nil {
			err = perr
			return err
		}
		previous := 0.0
		if it.QuantityReceived.Valid {
			parsed, parseErr := strconv.ParseFloat(it.QuantityReceived.String, 64)
			if parseErr != nil {
				err = parseErr
				return err
			}
			previous = parsed
		}
		qty := ordered
		if len(submitted) > 0 {
			var ok bool
			qty, ok = receivedByID[it.ID]
			if !ok {
				err = fmt.Errorf("%w: missing quantity for item %d", domain.ErrInvalidReceivedQuantity, it.ID)
				return err
			}
			// Bound the received quantity by the stored ordered quantity, not
			// by the request's quantity_ordered, so a submitted value cannot
			// inflate stock beyond what was actually ordered.
			if qty > ordered {
				err = fmt.Errorf("%w: item %d cannot exceed the ordered quantity %v", domain.ErrInvalidReceivedQuantity, it.ID, ordered)
				return err
			}
		}
		if qty < ordered {
			fullyReceived = false
		}
		delta := qty - previous
		if delta < 0 {
			err = fmt.Errorf("%w: item %d cannot decrease from %v", domain.ErrInvalidReceivedQuantity, it.ID, previous)
			return err
		}
		receiptLines = append(receiptLines, purchaseOrderReceiptLine{item: it, delta: delta})
		receivedUpdates = append(receivedUpdates, purchaseOrderItemReceivedUpdate{
			ID:       it.ID,
			Quantity: strconv.FormatFloat(qty, 'f', -1, 64),
		})
	}
	if err = updatePurchaseOrderItemsReceived(ctx, tx, poID, receivedUpdates); err != nil {
		return err
	}

	receipts, err := groupReceiptLinesByProduct(receiptLines)
	if err != nil {
		return err
	}
	legacyEntries := make([]purchaseOrderInventoryLedgerEntry, 0, len(receipts))
	if len(receipts) > 0 {
		// Every delivery is its own receipt so the ledger, which allows one
		// movement per reference and product, records a second partial
		// delivery instead of discarding it as a duplicate of the first.
		result, insertErr := tx.ExecContext(ctx, `
			INSERT INTO erp_purchase_order_receipts (owner_id, business_id, purchase_order_id, received_by)
			VALUES (?, ?, ?, ?)`, businessID, businessID, poID, createdBy)
		if insertErr != nil {
			return insertErr
		}
		receiptID, idErr := result.LastInsertId()
		if idErr != nil {
			return idErr
		}
		note := fmt.Sprintf("received from purchase order %d", poID)
		for _, receipt := range receipts {
			unitCost := receipt.unitCost
			before, after, merr := stock.ApplyMovementTx(ctx, tx, stock.Movement{
				OwnerID:       businessID,
				LocationID:    locID,
				ProductID:     receipt.productID,
				EntryType:     stock.EntryPurchase,
				Change:        receipt.quantity,
				ReferenceID:   &receiptID,
				ReferenceType: "purchase_order_receipt",
				Channel:       "erp",
				Note:          note,
				CreatedBy:     &createdBy,
				UnitCost:      &unitCost,
			})
			if merr != nil {
				return merr
			}
			legacyEntries = append(legacyEntries, purchaseOrderInventoryLedgerEntry{
				OwnerID:        createdBy,
				BusinessID:     businessID,
				ProductID:      receipt.productID,
				QuantityChange: strconv.FormatFloat(receipt.quantity, 'f', -1, 64),
				QuantityBefore: strconv.FormatFloat(before, 'f', -1, 64),
				QuantityAfter:  strconv.FormatFloat(after, 'f', -1, 64),
				ReferenceID:    poID,
				CreatedBy:      createdBy,
				Notes:          note,
			})
		}
	}
	if err = insertPurchaseOrderInventoryLedger(ctx, tx, legacyEntries); err != nil {
		return err
	}

	nextStatus := domain.POStatusPartial
	if fullyReceived {
		nextStatus = domain.POStatusReceived
	}
	if nextStatus == domain.POStatusReceived {
		_, err = qtx.UpdatePurchaseOrderStatusReceived(ctx, sqlc.UpdatePurchaseOrderStatusReceivedParams{Status: sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(nextStatus), Valid: true}, BusinessID: businessID, ID: poID})
	} else {
		err = qtx.UpdatePurchaseOrderStatusPartial(ctx, sqlc.UpdatePurchaseOrderStatusPartialParams{Status: sqlc.NullErpPurchaseOrdersStatus{ErpPurchaseOrdersStatus: sqlc.ErpPurchaseOrdersStatus(nextStatus), Valid: true}, BusinessID: businessID, ID: poID})
	}
	if err != nil {
		return err
	}

	if err := outbox.AppendAuditTx(ctx, tx, audit.Event{BusinessID: businessID, ActorID: createdBy, EntityType: "purchase_order", EntityID: poID, Action: "received", OldValues: purchaseOrderSnapshotOf(old), NewValues: map[string]string{"status": string(nextStatus)}}); err != nil {
		return err
	}
	return tx.Commit()
}

type productReceipt struct {
	productID int64
	quantity  float64
	unitCost  float64
}

// groupReceiptLinesByProduct folds the received deltas into one movement per
// product, in ascending product order so stock rows are always locked in the
// same order. Lines for the same product are costed at their weighted
// average.
func groupReceiptLinesByProduct(lines []purchaseOrderReceiptLine) ([]productReceipt, error) {
	byProduct := make(map[int64]*productReceipt, len(lines))
	order := make([]int64, 0, len(lines))
	for _, line := range lines {
		if line.delta <= 0 {
			continue
		}
		unitCost, err := strconv.ParseFloat(line.item.UnitCost, 64)
		if err != nil {
			return nil, fmt.Errorf("unit cost for item %d: %w", line.item.ID, err)
		}
		receipt, ok := byProduct[line.item.ProductID]
		if !ok {
			receipt = &productReceipt{productID: line.item.ProductID}
			byProduct[line.item.ProductID] = receipt
			order = append(order, line.item.ProductID)
		}
		total := receipt.unitCost*receipt.quantity + unitCost*line.delta
		receipt.quantity += line.delta
		receipt.unitCost = total / receipt.quantity
	}
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	receipts := make([]productReceipt, 0, len(order))
	for _, productID := range order {
		receipts = append(receipts, *byProduct[productID])
	}
	return receipts, nil
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
		po.Subtotal = sqlutil.ParseDecimal(row.Subtotal.String)
	}

	if row.TaxAmount.Valid {
		po.TaxAmount = sqlutil.ParseDecimal(row.TaxAmount.String)
	}

	if row.Total.Valid {
		po.Total = sqlutil.ParseDecimal(row.Total.String)
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

// parseDecimal converts a DB decimal string to float64, logging (rather than
// silently swallowing) corrupt values so bad money/quantity data surfaces.

func purchaseOrderItemFromRow(row sqlc.ErpPurchaseOrderItem) domain.PurchaseOrderItem {
	return domain.PurchaseOrderItem{ID: row.ID, PurchaseOrderID: row.PurchaseOrderID, ProductID: row.ProductID, QuantityOrdered: sqlutil.ParseDecimal(row.QuantityOrdered), QuantityReceived: sqlutil.ParseDecimal(row.QuantityReceived.String), UnitCost: sqlutil.ParseDecimal(row.UnitCost), TotalCost: sqlutil.ParseDecimal(row.TotalCost)}
}
