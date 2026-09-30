package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

func (r *Repository) ApplyDiscount(ctx context.Context, businessID, staffID int64, input domain.DiscountInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)

	paymentStatus, err := qtx.GetOrderForDiscount(ctx, sqlc.GetOrderForDiscountParams{
		BusinessID: businessID,
		ID:         input.OrderID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}
	if paymentStatus.PosOrdersPaymentStatus == sqlc.PosOrdersPaymentStatusPaid || paymentStatus.PosOrdersPaymentStatus == sqlc.PosOrdersPaymentStatusRefunded {
		return domain.ErrOrderCannotBeDiscounted
	}

	locationID, err := qtx.GetOrderLocationID(ctx, sqlc.GetOrderLocationIDParams{
		BusinessID: businessID,
		ID:         input.OrderID,
	})
	if err != nil {
		return err
	}
	capability, err := qtx.GetPOSUserCapabilities(ctx, sqlc.GetPOSUserCapabilitiesParams{
		BusinessID:   businessID,
		UserID:       staffID,
		LocationID:   locationID,
		Column4:      locationID,
		LocationID_2: locationID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrDiscountNotAllowed
		}
		return err
	}
	if !capability.CanApplyDiscount.Valid || !capability.CanApplyDiscount.Bool {
		return domain.ErrDiscountNotAllowed
	}
	maxDiscountPct := sqlutil.DecimalToFloat(capability.MaxDiscountPct.String)
	if maxDiscountPct > 100 {
		maxDiscountPct = 100
	}

	financials, err := qtx.GetOrderSubtotal(ctx, sqlc.GetOrderSubtotalParams{
		BusinessID: businessID,
		ID:         input.OrderID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		return err
	}

	subtotal := interfaceToFloat(financials.Subtotal)
	taxAmount := interfaceToFloat(financials.TaxAmount)
	var discountAmount float64
	if input.DiscountType == "percentage" || input.DiscountType == "percent" || input.DiscountType == "pct" {
		if input.DiscountValue > maxDiscountPct {
			return domain.ErrDiscountExceedsLimit
		}
		discountAmount = subtotal * input.DiscountValue / 100
	} else {
		discountAmount = input.DiscountValue
		if subtotal <= 0 || discountAmount > subtotal*maxDiscountPct/100 {
			return domain.ErrDiscountExceedsLimit
		}
	}
	// Accumulate with any existing order-level discount so a second
	// application neither erases the first from the money nor the audit.
	totalDiscount := sqlutil.DecimalToFloat(financials.ExistingDiscount) + discountAmount

	newTotal := subtotal + taxAmount - totalDiscount
	if newTotal < 0 {
		newTotal = 0
	}

	err = qtx.ApplyDiscount(ctx, sqlc.ApplyDiscountParams{
		DiscountAmount: sql.NullString{String: sqlutil.FloatToDecimal(totalDiscount), Valid: true},
		Total:          sql.NullString{String: sqlutil.FloatToDecimal(newTotal), Valid: true},
		CONCAT:         fmt.Sprintf(" [Discount: %s]", input.Reason),
		ID:             input.OrderID,
		BusinessID:     businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to apply discount", logger.Error(err))
		return err
	}
	return tx.Commit()
}

// VoidOrder voids an order that has neither taken nor returned money. The row
// is locked and both of its state columns are checked before the write, the way
// ResumeOrder does: an unconditional status write relabels a paid sale as
// voided, erasing the sale from the record while the cash stays in the drawer
// and the stock movement stays deducted, and it leaves the order refundable so
// the same amount can be paid back out a second time.
func (r *Repository) VoidOrder(ctx context.Context, businessID, orderID int64, reason string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin void order transaction", logger.Error(err))
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var status string
	var paymentStatus sql.NullString
	if err = tx.QueryRowContext(ctx, `
		SELECT status, payment_status
		FROM pos_orders
		WHERE business_id = ? AND id = ?
		FOR UPDATE`, businessID, orderID).Scan(&status, &paymentStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to lock order for void", logger.Error(err))
		return err
	}

	guardErr := checkVoidable(domain.OrderStatus(status), domain.PaymentStatus(paymentStatus.String))
	if errors.Is(guardErr, errOrderAlreadyVoided) {
		// Already voided: answer success without a second write so a re-sent
		// void cannot append a second reason to the audit trail.
		if err := tx.Commit(); err != nil {
			logger.Ctx(ctx).Error("Failed to commit no-op void order transaction", logger.Error(err))
			return err
		}
		committed = true
		return nil
	}
	if guardErr != nil {
		logger.Ctx(ctx).Warn("VoidOrder rejected: order is not in a voidable state",
			logger.Any("order_id", orderID),
			logger.Any("status", status),
			logger.Any("payment_status", paymentStatus.String),
		)
		return guardErr
	}

	if _, err = tx.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'voided', notes = CONCAT(COALESCE(notes,''), ?)
		WHERE business_id = ? AND id = ? AND status <> 'voided'`,
		fmt.Sprintf(" [Voided: %s]", reason), businessID, orderID); err != nil {
		logger.Ctx(ctx).Error("Failed to void order", logger.Error(err))
		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit void order transaction", logger.Error(err))
		return err
	}
	committed = true
	return nil
}

func (r *Repository) HoldOrder(ctx context.Context, businessID, orderID int64, reason string) error {
	result, err := r.queries.HoldOrder(ctx, sqlc.HoldOrderParams{
		CONCAT:     fmt.Sprintf(" [Held: %s]", reason),
		ID:         orderID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to hold order", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *Repository) ListHeldOrders(ctx context.Context, businessID int64) ([]domain.Order, error) {
	orders, _, err := r.listOrdersRows(ctx, businessID, 0, "held", "", "", "", 200, 0)
	return orders, err
}

// generateOrderNumber builds the next order number from the rows the given
// queries can see. It is only ever called with the transaction-scoped queries:
// a pooled count cannot see rows the transaction has already inserted, so two
// orders in one transaction would be offered the same number and the second
// would fail on uq_pos_orders_number. Counting through the transaction does
// not by itself prevent two *concurrent* transactions from picking the same
// number, so the create paths also retry a duplicate.
func generateOrderNumber(ctx context.Context, q *sqlc.Queries, businessID int64) (string, error) {
	today := time.Now().Format("20060102")

	count, err := q.CountTodaysOrders(ctx, businessID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ORD-%s-%04d", today, count+1), nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func floatToDecimal3(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}

func interfaceToFloat(v any) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	var str string
	switch val := v.(type) {
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		str = fmt.Sprintf("%v", val)
	}
	if str == "" {
		return 0
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Global().Warn("failed to parse order decimal from database",
			logger.String("value", str), logger.Error(err))
		return 0
	}
	return f
}

func nullableDecimal(v any) sql.NullString {
	switch value := v.(type) {
	case nil:
		return sql.NullString{String: "0", Valid: true}
	case []byte:
		return sql.NullString{String: string(value), Valid: true}
	case string:
		return sql.NullString{String: value, Valid: true}
	default:
		return sql.NullString{String: fmt.Sprint(value), Valid: true}
	}
}

func scanOrderRow(rows *sql.Rows) (domain.Order, error) {
	var o domain.Order
	var paidAt sql.NullTime
	var paymentMethod, paymentStatus sql.NullString

	if err := rows.Scan(
		&o.ID, &o.OwnerID, &o.SessionID, &o.TableID, &o.LocationID,
		&o.OrderNumber, &o.Type, &o.Status, &o.CustomerID, &o.CustomerName,
		&o.StaffID, &o.Subtotal, &o.DiscountAmount, &o.TaxAmount, &o.Total,
		&paymentMethod, &paymentStatus, &o.AmountPaid, &o.ChangeDue,
		&o.Notes, &o.KitchenNotes, &o.CreatedAt, &o.UpdatedAt, &paidAt,
	); err != nil {
		return domain.Order{}, err
	}

	o.PaymentMethod = domain.PaymentMethod(paymentMethod.String)
	o.PaymentStatus = domain.PaymentStatus(paymentStatus.String)

	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}

	return o, nil
}

// ─── Paddle integration ─────────────────────────────────────────────────────

func (r *Repository) SetPaddleTransactionID(ctx context.Context, businessID, orderID int64, txID string) error {
	_, err := r.queries.SetOrderPaddleTransactionID(ctx, sqlc.SetOrderPaddleTransactionIDParams{
		PaddleTransactionID: sql.NullString{String: txID, Valid: txID != ""},
		ID:                  orderID,
		BusinessID:          businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to set paddle transaction id", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) MarkOrderPaidByPaddleTx(ctx context.Context, txID string) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	var orderID, businessID int64
	var paymentStatus string
	if err := tx.QueryRowContext(ctx, `
		SELECT id, business_id, payment_status
		FROM pos_orders
		WHERE paddle_transaction_id = ?
		LIMIT 1
		FOR UPDATE`, txID).Scan(&orderID, &businessID, &paymentStatus); err != nil {
		logger.Ctx(ctx).Error("Failed to lock order for paddle payment", logger.Error(err))
		return 0, err
	}
	if paymentStatus == "paid" {
		if err := tx.Commit(); err != nil {
			return 0, err
		}
		committed = true
		return 0, nil
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'paid', payment_status = 'paid', paid_at = NOW()
		WHERE paddle_transaction_id = ? AND payment_status <> 'paid'`, txID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark order paid by paddle tx", logger.Error(err))
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 1 {
		if err := r.deductSoldStock(ctx, tx, businessID, orderID); err != nil {
			logger.Ctx(ctx).Error("Failed to deduct stock for paddle-paid order", logger.Error(err), logger.Any("order_id", orderID))
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	committed = true
	return affected, nil
}

func (r *Repository) IsPaddleEventProcessed(ctx context.Context, txID string) (bool, error) {
	count, err := r.queries.IsPaddleEventProcessed(ctx, txID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) RecordPaddleEventProcessed(ctx context.Context, txID, eventType string, ownerID int64) error {
	_, err := r.queries.RecordPaddleEventProcessed(ctx, sqlc.RecordPaddleEventProcessedParams{
		TransactionID: txID,
		EventType:     eventType,
		OwnerID:       ownerID,
	})
	return err
}

func (r *Repository) GetOrderIDByPaddleTx(ctx context.Context, txID string) (int64, int64, error) {
	row, err := r.queries.GetOrderIDByPaddleTransaction(ctx, sql.NullString{String: txID, Valid: true})
	if err != nil {
		return 0, 0, err
	}
	return row.ID, row.OwnerID, nil
}

// ─── Public Online Ordering ──────────────────────────────────────────────────

func (r *Repository) IsOnlineOrderingEnabled(ctx context.Context, businessID int64) (bool, error) {
	count, err := r.queries.IsOnlineOrderingEnabled(ctx, businessID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) GetPublicMenuItem(ctx context.Context, businessID, menuItemID int64) (domain.OrderItemInput, error) {
	// cross-schema read from erp — public menu item lookup
	row, err := r.queries.GetPublicMenuItem(ctx, sqlc.GetPublicMenuItemParams{
		BusinessID: businessID,
		ID:         menuItemID,
	})
	if err != nil {
		return domain.OrderItemInput{}, err
	}

	item := domain.OrderItemInput{
		MenuItemID: row.ID,
		Name:       row.Name,
		SKU:        row.Sku.String,
		UnitPrice:  sqlutil.DecimalToFloat(row.SalePrice),
		TaxRate:    sqlutil.DecimalToFloat(row.TaxRate),
	}
	return item, nil
}

// CreatePublicOrder inserts an online order, generating its order number
// inside the insert transaction and retrying a lost race on
// uq_pos_orders_number. Same reasoning as CreateOrder: counting on a separate
// connection let two customers checking out at the same moment pick the same
// number, and the loser of the unique key lost the whole order.
func (r *Repository) CreatePublicOrder(ctx context.Context, businessID int64, orderNumber, customerName, customerPhone, notes string, items []domain.OrderItemInput) (int64, string, error) {
	// Only a generated number is worth retrying; see CreateOrder.
	if orderNumber != "" {
		return r.createPublicOrderOnce(ctx, businessID, orderNumber, customerName, customerPhone, notes, items)
	}
	for attempt := 1; ; attempt++ {
		orderID, generated, err := r.createPublicOrderOnce(ctx, businessID, "", customerName, customerPhone, notes, items)
		if err == nil {
			return orderID, generated, nil
		}
		if attempt >= createOrderNumberAttempts || !isDuplicateEntry(err) {
			return 0, "", err
		}
		logger.Ctx(ctx).Warn("Public order number collided, retrying create",
			logger.Any("business_id", businessID), logger.Any("attempt", attempt))
	}
}

func (r *Repository) createPublicOrderOnce(ctx context.Context, businessID int64, orderNumber, customerName, customerPhone, notes string, items []domain.OrderItemInput) (int64, string, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin public order transaction", logger.Error(err))
		return 0, "", err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
				logger.Ctx(ctx).Error("failed to rollback public order tx", logger.Error(rollbackErr))
			}
		}
	}()

	// Mirror the in-store ComputeTotals math: online orders must charge the
	// same per-item tax as the POS path instead of silently dropping it.
	totals := domain.ComputeTotals(items)

	qtx := r.queries.WithTx(tx)

	// Counted through the transaction so the number reflects every order this
	// business has, not just the ones a pooled connection can see.
	if orderNumber == "" {
		orderNumber, err = generateOrderNumber(ctx, qtx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to generate public order number", logger.Error(err))
			return 0, "", err
		}
	}

	res, err := qtx.CreatePublicOrder(ctx, sqlc.CreatePublicOrderParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		OrderNumber:   orderNumber,
		CustomerID:    sql.NullInt64{},
		CustomerName:  sqlutil.NullString(customerName),
		CustomerPhone: sqlutil.NullString(customerPhone),
		Subtotal:      sql.NullString{String: sqlutil.FloatToDecimal(totals.Subtotal), Valid: true},
		TaxAmount:     sql.NullString{String: sqlutil.FloatToDecimal(totals.TaxAmount), Valid: true},
		Total:         sql.NullString{String: sqlutil.FloatToDecimal(totals.Total), Valid: true},
		Notes:         sqlutil.NullString(notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert public order", logger.Error(err))
		return 0, "", err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, "", err
	}

	for _, item := range items {
		itemTotal := item.UnitPrice * item.Quantity
		if err := qtx.CreatePublicOrderItem(ctx, sqlc.CreatePublicOrderItemParams{
			OrderID:    orderID,
			MenuItemID: sql.NullInt64{Int64: item.MenuItemID, Valid: item.MenuItemID != 0},
			Name:       item.Name,
			Quantity:   floatToDecimal3(item.Quantity),
			UnitPrice:  sqlutil.FloatToDecimal(item.UnitPrice),
			Total:      sqlutil.FloatToDecimal(itemTotal),
			Notes:      sqlutil.NullString(item.Notes),
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to insert public order item", logger.Error(err))
			return 0, "", err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return orderID, orderNumber, nil
}

func (r *Repository) GetPublicOrderStatus(ctx context.Context, businessID, orderID int64) (orderNumber, status string, err error) {
	row, err := r.queries.GetPublicOrderNumberStatus(ctx, sqlc.GetPublicOrderNumberStatusParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		return "", "", err
	}
	return row.OrderNumber, string(row.Status.PosOrdersStatus), nil
}
