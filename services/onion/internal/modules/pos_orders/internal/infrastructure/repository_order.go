package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
	"github.com/voxire/onion/internal/platform/stock_core"
)

func (r *Repository) UpdateOrderStatus(ctx context.Context, businessID, orderID int64, status domain.OrderStatus) error {
	result, err := r.queries.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		Status:     sqlc.NullPosOrdersStatus{PosOrdersStatus: sqlc.PosOrdersStatus(status), Valid: true},
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update order status", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// MySQL reports 0 changed rows both when no row matched and when the
		// requested status equals the stored one. Confirm the order exists
		// before answering 404: a stale/foreign id must not succeed, but an
		// idempotent re-send must not look like a missing order either.
		existing, existsErr := r.queries.OrderExists(ctx, sqlc.OrderExistsParams{
			BusinessID: businessID,
			ID:         orderID,
		})
		if existsErr != nil {
			return existsErr
		}
		if existing == 0 {
			return domain.ErrOrderNotFound
		}
	}
	return nil
}

func (r *Repository) UpdateOrderItemStatus(ctx context.Context, orderID, itemID int64, status domain.ItemStatus) error {
	result, err := r.queries.UpdateOrderItemStatus(ctx, sqlc.UpdateOrderItemStatusParams{
		Status:  sqlc.NullPosOrderItemsStatus{PosOrderItemsStatus: sqlc.PosOrderItemsStatus(status), Valid: true},
		OrderID: orderID,
		ID:      itemID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update order item status", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		// Same-value updates also report 0 changed rows; only a missing item
		// is an error. The caller already verified the order's business.
		existing, existsErr := r.queries.OrderItemExists(ctx, sqlc.OrderItemExistsParams{
			OrderID: orderID,
			ID:      itemID,
		})
		if existsErr != nil {
			return existsErr
		}
		if existing == 0 {
			return domain.ErrOrderNotFound
		}
	}
	return nil
}

func (r *Repository) GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	rows, err := r.queries.GetOrderItems(ctx, orderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get order items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.OrderItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.OrderItem{
			ID:             row.ID,
			OrderID:        row.OrderID,
			MenuItemID:     row.MenuItemID.Int64,
			Name:           row.Name,
			SKU:            row.Sku.String,
			Quantity:       sqlutil.DecimalToFloat(row.Quantity),
			UnitPrice:      sqlutil.DecimalToFloat(row.UnitPrice),
			DiscountPct:    sqlutil.DecimalToFloat(row.DiscountPct.String),
			DiscountAmount: sqlutil.DecimalToFloat(row.DiscountAmount.String),
			TaxRate:        sqlutil.DecimalToFloat(row.TaxRate.String),
			Total:          sqlutil.DecimalToFloat(row.Total),
			Notes:          row.Notes.String,
			Status:         domain.ItemStatus(row.Status.PosOrderItemsStatus),
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return items, nil
}

func (r *Repository) ProcessPayment(ctx context.Context, businessID, orderID int64, input domain.PaymentInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin payment transaction", logger.Error(err))
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	paymentRow, err := qtx.GetOrderForPayment(ctx, sqlc.GetOrderForPaymentParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to get order for payment", logger.Error(err))
		return err
	}
	if paymentRow.PaymentStatus.Valid && paymentRow.PaymentStatus.PosOrdersPaymentStatus == sqlc.PosOrdersPaymentStatusPaid {
		return domain.ErrOrderAlreadyPaid
	}

	total := sqlutil.DecimalToFloat(paymentRow.Total.String)
	change, err := domain.ComputeChange(total, input.AmountPaid)
	if err != nil {
		return err
	}
	changeDue := change.Change

	now := time.Now().UTC()

	markResult, err := qtx.MarkOrderPaid(ctx, sqlc.MarkOrderPaidParams{
		PaymentMethod: nullPaymentMethod(input.PaymentMethod),
		AmountPaid:    sql.NullString{String: sqlutil.FloatToDecimal(input.AmountPaid), Valid: true},
		ChangeDue:     sql.NullString{String: sqlutil.FloatToDecimal(changeDue), Valid: true},
		PaidAt:        sql.NullTime{Time: now, Valid: true},
		BusinessID:    businessID,
		ID:            orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark order paid", logger.Error(err))
		return err
	}
	affected, err := markResult.RowsAffected()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to inspect payment update", logger.Error(err))
		return err
	}
	if affected == 0 {
		return domain.ErrOrderAlreadyPaid
	}

	if paymentRow.TableID.Valid && paymentRow.TableID.Int64 > 0 && string(paymentRow.Type.PosOrdersType) == string(domain.OrderTypeDineIn) {
		// Business-scoped: an order that was (wrongly) pointed at another
		// tenant's table must never free that table.
		err = qtx.FreeTable(ctx, sqlc.FreeTableParams{
			ID:         paymentRow.TableID.Int64,
			BusinessID: businessID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to free table after payment", logger.Error(err))
			return err
		}
	}

	// Decrement on-hand stock for inventory-tracked line items so the sale is
	// reflected in the shared stock (the same stock ERP reads). Atomic with the
	// payment; idempotent so a re-fired payment can't double-deduct.
	if err = r.deductSoldStock(ctx, tx, businessID, orderID); err != nil {
		logger.Ctx(ctx).Error("Failed to deduct stock for paid order", logger.Error(err), logger.Any("order_id", orderID))
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit payment transaction", logger.Error(err))
		return err
	}
	committed = true

	return nil
}

// deductSoldStock decrements stock_levels (and the synced scalar) for every
// inventory-tracked line item of a paid order, through the shared stock_core
// primitive at the order's POS branch location. It is idempotent on the
// (pos_order, sale) ledger key, so re-firing a payment is a no-op. Oversell is
// permitted (POS should not block a sale on a lagging count) — this should
// become a per-tenant policy flag.
func (r *Repository) deductSoldStock(ctx context.Context, tx *sql.Tx, businessID, orderID int64) error {
	qtx := r.queries.WithTx(tx)

	// cross-schema read from stock_core — idempotency guard on the shared stock ledger
	if _, err := qtx.HasSaleLedgerEntry(ctx, sqlc.HasSaleLedgerEntryParams{
		BusinessID:  businessID,
		ReferenceID: sql.NullInt64{Int64: orderID, Valid: true},
	}); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	locationID, err := qtx.GetOrderLocationID(ctx, sqlc.GetOrderLocationIDParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		return err
	}

	stockLoc, err := stock_core.POSBranchLocationID(ctx, tx, businessID, locationID)
	if err != nil {
		return err
	}

	// cross-schema read from erp — join POS line items against ERP products to find inventory-tracked SKUs
	rows, err := qtx.ListInventoryTrackedOrderLines(ctx, sqlc.ListInventoryTrackedOrderLinesParams{
		BusinessID: businessID,
		OrderID:    orderID,
	})
	if err != nil {
		return err
	}

	type line struct {
		productID int64
		qty       float64
	}
	lines := make([]line, 0, len(rows))
	for _, row := range rows {
		if !row.MenuItemID.Valid {
			continue
		}
		q, perr := strconv.ParseFloat(row.Quantity, 64)
		if perr != nil {
			return perr
		}
		lines = append(lines, line{productID: row.MenuItemID.Int64, qty: q})
	}

	sort.Slice(lines, func(i, j int) bool { return lines[i].productID < lines[j].productID })

	for _, l := range lines {
		oid := orderID
		if _, _, err := stock_core.ApplyMovementTx(ctx, tx, stock_core.Movement{
			OwnerID:       businessID,
			BusinessID:    businessID,
			LocationID:    stockLoc,
			ProductID:     l.productID,
			EntryType:     stock_core.EntrySale,
			Change:        -l.qty,
			ReferenceID:   &oid,
			ReferenceType: "pos_order",
			Channel:       "pos",
			Note:          "pos sale",
			AllowNegative: true,
		}); err != nil {
			return err
		}
	}

	return nil
}
