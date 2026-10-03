package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) CreateRefund(ctx context.Context, businessID, orderID, refundedBy int64, amount float64, reason string) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin refund transaction", logger.Error(err))
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	// Lock the order row and validate before writing: only paid orders can
	// be refunded, and cumulative refunds may never exceed the order total.
	order, err := qtx.GetOrderForRefund(ctx, sqlc.GetOrderForRefundParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to lock order for refund", logger.Error(err))
		return 0, err
	}
	status := ""
	if order.PaymentStatus.Valid {
		status = string(order.PaymentStatus.PosOrdersPaymentStatus)
	}
	// The same locked read covers the order's workflow status: payment_status
	// alone cannot tell a live sale from one that was voided or cancelled,
	// because those writes leave payment_status untouched.
	var orderStatus string
	if err = tx.QueryRowContext(ctx,
		`SELECT status FROM pos_orders WHERE business_id = ? AND id = ?`,
		businessID, orderID).Scan(&orderStatus); err != nil {
		logger.Ctx(ctx).Error("Failed to read order status for refund", logger.Error(err))
		return 0, err
	}
	if err = checkRefundable(domain.OrderStatus(orderStatus), domain.PaymentStatus(status)); err != nil {
		logger.Ctx(ctx).Warn("CreateRefund rejected: order is not refundable",
			logger.Any("order_id", orderID), logger.Any("status", orderStatus))
		return 0, err
	}
	// Read inside the same transaction that holds the order's row lock, so two
	// concurrent refunds cannot both see the same prior total and each believe
	// it fits under the cap.
	prevRefunds, err := qtx.SumRefundsForOrder(ctx, sqlc.SumRefundsForOrderParams{
		BusinessID: businessID,
		OrderID:    orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum prior refunds", logger.Error(err))
		return 0, err
	}

	// Every comparison below is integer cents. The float64 form refused a
	// legitimate full refund: 0.03 + 0.27 is 0.30000000000000004 in binary
	// floating point, which is greater than 0.30, so the last cent of a
	// perfectly valid refund was answered with a 400.
	capCents, capErr := refundCapCents(order)
	if capErr != nil {
		logger.Ctx(ctx).Error("Failed to compute refundable cap",
			logger.Error(capErr), logger.Any("order_id", orderID))
		return 0, capErr
	}

	priorCents, err := sumRefundCents(prevRefunds)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read prior refund total", logger.Error(err))
		return 0, err
	}

	// Round once, here, and reuse the rounded value for both the cap check and
	// the stored DECIMAL, so the amount guarded against and the amount written
	// are the same number rather than two roundings that may disagree.
	refundCents := money.Amount(money.DollarsToCents(amount))
	if refundCents <= 0 {
		// A sub-cent request would otherwise be stored as a 0.00 pos_refunds
		// row and bump the session's total_refunds by nothing.
		return 0, errors.New("refund amount must be at least one cent")
	}

	totalRefunded := priorCents + refundCents
	if totalRefunded > capCents {
		err = domain.ErrRefundExceedsTotal
		return 0, err
	}

	res, err := qtx.CreateRefund(ctx, sqlc.CreateRefundParams{
		OwnerID:    businessID,
		BusinessID: businessID,
		OrderID:    orderID,
		Amount:     refundCents.String(),
		Reason:     sqlutil.NullString(reason),
		RefundedBy: sqlutil.NullInt64(refundedBy),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert refund", logger.Error(err))
		return 0, err
	}

	refundID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Only a refund that reaches the cap retires the sale. A smaller one
	// leaves the order exactly as it was, so the session and dashboard totals
	// still count the goods the customer took. Note the enum's 'partial' value
	// is deliberately not written here: every revenue query keys on
	// payment_status = 'paid', so 'partial' would drop the sale from those
	// reports exactly as 'refunded' does. The part refund is recorded in
	// pos_refunds and added to the session's total_refunds.
	if refundPaymentStatusFor(totalRefunded, capCents) == domain.PaymentStatusRefunded {
		if err = qtx.MarkOrderRefunded(ctx, sqlc.MarkOrderRefundedParams{
			BusinessID: businessID,
			ID:         orderID,
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to mark order refunded", logger.Error(err))
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit refund transaction", logger.Error(err))
		return 0, err
	}

	return refundID, nil
}

func (r *Repository) VoidOrderItem(ctx context.Context, businessID, orderID, itemID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin void-order-item transaction", logger.Error(err))
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Ctx(ctx).Error("failed to rollback void-order-item panic", logger.Error(rollbackErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	if _, err := qtx.GetOrderItemStatus(ctx, sqlc.GetOrderItemStatusParams{
		ID:      itemID,
		OrderID: orderID,
	}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item lookup", logger.Error(rollbackErr))
		}
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to fetch order item for void", logger.Error(err))
		return err
	}

	voidResult, err := qtx.VoidOrderItem(ctx, sqlc.VoidOrderItemParams{
		ID:      itemID,
		OrderID: orderID,
	})
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item update", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to cancel order item", logger.Error(err))
		return err
	}
	voided, err := voidResult.RowsAffected()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item update", logger.Error(rollbackErr))
		}
		return err
	}
	if voided == 0 {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item update", logger.Error(rollbackErr))
		}
		// The item exists (checked above) but is not in a voidable state, e.g.
		// it is already ready/served/cancelled. Report that instead of
		// answering success for a no-op.
		return domain.ErrItemNotVoidable
	}

	itemsTotal, err := qtx.GetActiveOrderItemsTotal(ctx, orderID)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item total lookup", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to calculate order total after item void", logger.Error(err))
		return err
	}

	itemsSubtotal := nullableDecimal(itemsTotal.Subtotal)
	itemsTaxAmount := nullableDecimal(itemsTotal.TaxAmount)
	if err := qtx.UpdateOrderTotal(ctx, sqlc.UpdateOrderTotalParams{
		Subtotal:   itemsSubtotal,
		TaxAmount:  itemsTaxAmount,
		ID:         orderID,
		BusinessID: businessID,
	}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item recalc", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to recalculate order total after item void", logger.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit void-order-item transaction", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) TransferTable(ctx context.Context, businessID, orderID, oldTableID, newTableID int64) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transfer-table transaction", logger.Error(err))
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Ctx(ctx).Error("failed to rollback transfer-table panic", logger.Error(rollbackErr))
			}
			panic(p)
		}
	}()

	qtx := r.queries.WithTx(tx)

	if oldTableID > 0 {
		if err := qtx.FreeTableByID(ctx, sqlc.FreeTableByIDParams{
			ID:         oldTableID,
			BusinessID: businessID,
		}); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Ctx(ctx).Error("failed to rollback free old table", logger.Error(rollbackErr))
			}
			logger.Ctx(ctx).Error("Failed to free old table during transfer", logger.Error(err))
			return err
		}
	}

	if err := qtx.MoveOrderToTable(ctx, sqlc.MoveOrderToTableParams{
		TableID:    sql.NullInt64{Int64: newTableID, Valid: true},
		ID:         orderID,
		BusinessID: businessID,
	}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback move order", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to update order table_id during transfer", logger.Error(err))
		return err
	}

	occupyResult, err := qtx.OccupyTableByID(ctx, sqlc.OccupyTableByIDParams{
		ID:         newTableID,
		BusinessID: businessID,
	})
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback occupy table", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to mark new table occupied during transfer", logger.Error(err))
		return err
	}
	affected, err := occupyResult.RowsAffected()
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback occupy table", logger.Error(rollbackErr))
		}
		return err
	}
	if affected == 0 {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback occupy table", logger.Error(rollbackErr))
		}
		// The target table either doesn't exist or belongs to another
		// business. Rolling back leaves the order on its current table.
		logger.Ctx(ctx).Warn("TransferTable rejected: table not in business scope",
			logger.Any("table_id", newTableID), logger.Any("business_id", businessID))
		return domain.ErrTableNotFound
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit transfer-table transaction", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) SplitOrder(ctx context.Context, businessID, orderID int64, splits [][]int64) ([]int64, error) {
	order, err := r.GetOrder(ctx, businessID, orderID)
	if err != nil {
		return nil, err
	}

	// Splitting moves the items into new unpaid orders and cancels the
	// original. A paid/refunded order would keep its paid totals while the
	// children could be charged again, so refuse those up front.
	if order.PaymentStatus == domain.PaymentStatusPaid || order.PaymentStatus == domain.PaymentStatusRefunded {
		return nil, domain.ErrOrderCannotBeSplit
	}
	switch order.Status {
	case domain.OrderStatusPaid, domain.OrderStatusVoided, domain.OrderStatusCancelled:
		return nil, domain.ErrOrderCannotBeSplit
	}

	tx, txErr := r.pool.BeginTx(ctx, nil)
	if txErr != nil {
		logger.Ctx(ctx).Error("Failed to begin split order tx", logger.Error(txErr))
		return nil, txErr
	}

	defer func() {
		if txErr != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	newOrderIDs := make([]int64, 0, len(splits))

	for _, splitItemIDs := range splits {
		// The allocator sees earlier children inserted in this transaction.
		orderNum, genErr := generateOrderNumber(ctx, tx, qtx, businessID)
		if genErr != nil {
			txErr = genErr
			return nil, genErr
		}

		res, insertErr := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
			OwnerID:       businessID,
			BusinessID:    businessID,
			SessionID:     sqlutil.NullInt64(order.SessionID),
			TableID:       sqlutil.NullInt64(order.TableID),
			LocationID:    sqlutil.NullInt64(order.LocationID),
			OrderNumber:   orderNum,
			Type:          sqlc.NullPosOrdersType{PosOrdersType: sqlc.PosOrdersType(order.Type), Valid: true},
			CustomerID:    sqlutil.NullInt64(order.CustomerID),
			CustomerName:  sqlutil.NullString(order.CustomerName),
			StaffID:       sqlutil.NullInt64(order.StaffID),
			Subtotal:      sql.NullString{String: "0", Valid: true},
			TaxAmount:     sql.NullString{String: "0", Valid: true},
			Total:         sql.NullString{String: "0", Valid: true},
			PaymentMethod: nullPaymentMethod(order.PaymentMethod),
			Notes:         sqlutil.NullString(order.Notes),
			KitchenNotes:  sqlutil.NullString(order.KitchenNotes),
		})
		if insertErr != nil {
			txErr = insertErr
			logger.Ctx(ctx).Error("Failed to insert split order", logger.Error(insertErr))
			return nil, insertErr
		}

		newID, _ := res.LastInsertId()
		newOrderIDs = append(newOrderIDs, newID)

		for _, itemID := range splitItemIDs {
			moveErr := qtx.MoveOrderItems(ctx, sqlc.MoveOrderItemsParams{
				OrderID:   newID,
				ID:        itemID,
				OrderID_2: orderID,
			})
			if moveErr != nil {
				txErr = moveErr
				return nil, moveErr
			}
		}

		itemsTotal, totalErr := qtx.GetActiveOrderItemsTotal(ctx, newID)
		if totalErr != nil {
			txErr = totalErr
			return nil, totalErr
		}

		itemsSubtotal := nullableDecimal(itemsTotal.Subtotal)
		itemsTaxAmount := nullableDecimal(itemsTotal.TaxAmount)
		recalcErr := qtx.UpdateOrderTotal(ctx, sqlc.UpdateOrderTotalParams{
			Subtotal:   itemsSubtotal,
			TaxAmount:  itemsTaxAmount,
			ID:         newID,
			BusinessID: businessID,
		})
		if recalcErr != nil {
			txErr = recalcErr
			return nil, recalcErr
		}
	}

	cancelErr := qtx.CancelOrder(ctx, sqlc.CancelOrderParams{
		ID:         orderID,
		BusinessID: businessID,
	})
	if cancelErr != nil {
		txErr = cancelErr
		return nil, cancelErr
	}

	if commitErr := tx.Commit(); commitErr != nil {
		txErr = commitErr
		logger.Ctx(ctx).Error("Failed to commit split order tx", logger.Error(commitErr))
		return nil, commitErr
	}

	return newOrderIDs, nil
}
