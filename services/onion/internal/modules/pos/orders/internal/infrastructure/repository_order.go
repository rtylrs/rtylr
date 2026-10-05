package infrastructure

import (
 "github.com/voxire/onion/internal/modules/pos/internal/outbox"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	poscontract "github.com/voxire/onion/internal/contracts/pos"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// ─── Order state machine ──────────────────────────────────────────────────────
//
// pos_orders.status (workflow) and pos_orders.payment_status (money) are
// independent columns, and each of the payment/refund/void writes below sets
// only one of them. Without a shared guard the same order can be paid, then
// relabelled voided, then paid out again: a void erases the sale from the
// record while the cash stays in the drawer and the stock movement stays
// deducted, and a second charge on a refunded order is money the customer pays
// twice while the shift close still subtracts the earlier refund. The three
// helpers hold the legal transitions in one place; every repository calls them
// with the row's locked state before it writes anything.

// errOrderAlreadyVoided marks the idempotent re-send of a void. Reporting
// success without writing keeps a second reason out of the audit trail, the
// same way UpdateOrderStatus treats a same-value status write.
var errOrderAlreadyVoided = errors.New("order is already voided")

// moneyMoved reports whether money has been taken or returned against the
// order. 'partial' is in the column's enum even though no write path sets it
// yet, so a deposit recorded there is still cash in the drawer.
func moneyMoved(paymentStatus domain.PaymentStatus) bool {
	switch paymentStatus {
	case domain.PaymentStatusPartial, domain.PaymentStatusPaid, domain.PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

// checkVoidable allows a void only while the order is still editable and no
// money has moved: a paid or refunded order must be settled by refunding or
// cashing out, not by relabelling the sale.
func checkVoidable(status domain.OrderStatus, paymentStatus domain.PaymentStatus) error {
	if moneyMoved(paymentStatus) || status == domain.OrderStatusPaid || status == domain.OrderStatusRefunded {
		return domain.ErrOrderAlreadyPaid
	}
	if status == domain.OrderStatusCancelled {
		return domain.ErrOrderNotEditable
	}
	if status == domain.OrderStatusVoided {
		return errOrderAlreadyVoided
	}
	return nil
}

// checkPayable allows a charge only while the order is an open, unpaid sale.
// Voiding or cancelling an order ends it, and a settled one is never chargeable
// again.
func checkPayable(status domain.OrderStatus, paymentStatus domain.PaymentStatus) error {
	if paymentStatus == domain.PaymentStatusPaid || status == domain.OrderStatusPaid {
		return domain.ErrOrderAlreadyPaid
	}
	switch status {
	case domain.OrderStatusVoided, domain.OrderStatusCancelled, domain.OrderStatusRefunded:
		return domain.ErrOrderNotEditable
	}
	if paymentStatus == domain.PaymentStatusRefunded {
		return domain.ErrOrderNotEditable
	}
	return nil
}

// checkRefundable allows a refund only for an order that took money and is not
// already closed. A voided or cancelled order is refused even when its
// payment_status still reads paid: rows voided before the void guard existed
// carry exactly that shape and must not be paid out a second time.
func checkRefundable(status domain.OrderStatus, paymentStatus domain.PaymentStatus) error {
	switch status {
	case domain.OrderStatusVoided, domain.OrderStatusCancelled:
		return domain.ErrOrderNotEditable
	}
	if paymentStatus != domain.PaymentStatusPaid && paymentStatus != domain.PaymentStatusRefunded {
		return domain.ErrOrderNotPaid
	}
	return nil
}

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

	// MarkOrderPaid keys its own guard off payment_status alone, so a voided,
	// cancelled, or refunded order would be resurrected as a live sale here.
	// The row is already locked by the read above.
	var orderStatus string
	if err = tx.QueryRowContext(ctx,
		`SELECT status FROM pos_orders WHERE business_id = ? AND id = ?`,
		businessID, orderID).Scan(&orderStatus); err != nil {
		logger.Ctx(ctx).Error("Failed to read order status for payment", logger.Error(err))
		return err
	}
	if err := checkPayable(domain.OrderStatus(orderStatus), domain.PaymentStatus(paymentRow.PaymentStatus.PosOrdersPaymentStatus)); err != nil {
		logger.Ctx(ctx).Warn("ProcessPayment rejected: order is not payable",
			logger.Any("order_id", orderID), logger.Any("status", orderStatus))
		return err
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

	if err = r.publishOrderPaid(ctx, tx, businessID, orderID); err != nil {
		logger.Ctx(ctx).Error("Failed to record paid order fact", logger.Error(err), logger.Any("order_id", orderID))
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit payment transaction", logger.Error(err))
		return err
	}
	committed = true

	return nil
}

func (r *Repository) publishOrderPaid(ctx context.Context, tx *sql.Tx, businessID, orderID int64) error {
	qtx := r.queries.WithTx(tx)

	locationID, err := qtx.GetOrderLocationID(ctx, sqlc.GetOrderLocationIDParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		return err
	}

	rows, err := qtx.ListPaidOrderProductLines(ctx, sqlc.ListPaidOrderProductLinesParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		return err
	}

	payload := poscontract.PaidOrder{BranchID: locationID, Items: make([]poscontract.OrderItem, 0, len(rows))}
	for _, row := range rows {
		if !row.MenuItemID.Valid {
			continue
		}
		q, perr := strconv.ParseFloat(row.Quantity, 64)
		if perr != nil {
			return perr
		}
		payload.Items = append(payload.Items, poscontract.OrderItem{ProductID: row.MenuItemID.Int64, Quantity: q})
	}
	return outbox.Publish(ctx, tx, eventbus.Event{
		BusinessID: businessID, Type: poscontract.OrderPaidType,
		SubjectType: "pos_order", SubjectID: orderID, Payload: payload,
	})
}
