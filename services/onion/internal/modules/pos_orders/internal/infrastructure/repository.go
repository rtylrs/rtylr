package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/eventbus"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

// ─── Orders ──────────────────────────────────────────────────────────────────

func (r *Repository) ListOrders(ctx context.Context, businessID, sessionID int64, status, dateFrom, dateTo, search string, limit, offset int) ([]domain.Order, int64, bool, error) {
	orders, hasMore, err := r.listOrdersRows(ctx, businessID, sessionID, status, dateFrom, dateTo, search, limit, offset)
	if err != nil {
		return nil, 0, false, err
	}

	orders, err = r.loadOrderItemsForOrders(ctx, orders)
	if err != nil {
		return nil, 0, false, err
	}

	total, err := r.countOrders(ctx, businessID, sessionID, status, dateFrom, dateTo, search)
	if err != nil {
		return nil, 0, false, err
	}

	return orders, total, hasMore, nil
}

// dynamic-query: squirrel
func (r *Repository) listOrdersRows(ctx context.Context, businessID, sessionID int64, status, dateFrom, dateTo, search string, limit, offset int) ([]domain.Order, bool, error) {
	builder := sq.
		Select(
			"o.id", "o.owner_id", "COALESCE(o.session_id,0)", "COALESCE(o.table_id,0)", "COALESCE(o.location_id,0)",
			"o.order_number", "o.type", "o.status", "COALESCE(o.customer_id,0)", "COALESCE(o.customer_name,'')",
			"COALESCE(o.staff_id,0)", "o.subtotal", "o.discount_amount", "o.tax_amount", "o.total",
			"COALESCE(o.payment_method,'')", "COALESCE(o.payment_status,'')", "COALESCE(o.amount_paid,0)", "COALESCE(o.change_due,0)",
			"COALESCE(o.notes,'')", "COALESCE(o.kitchen_notes,'')", "o.created_at", "o.updated_at", "o.paid_at",
		).
		From("pos_orders o").
		Where(sq.Eq{"o.business_id": businessID})

	builder = applyPOSOrderListFilters(builder, sessionID, status, dateFrom, dateTo, search)

	query, args, err := builder.
		OrderBy("o.created_at DESC").
		Limit(uint64(limit + 1)). // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset(uint64(offset)).   // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list pos orders query", logger.Error(err))
		return nil, false, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos orders", logger.Error(err))
		return nil, false, err
	}

	defer rows.Close()

	var orders []domain.Order

	for rows.Next() {
		o, scanErr := scanOrderRow(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan order row", logger.Error(scanErr))
			return nil, false, scanErr
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(orders) > limit
	if hasMore {
		orders = orders[:limit]
	}

	return orders, hasMore, nil
}

func (r *Repository) loadOrderItemsForOrders(ctx context.Context, orders []domain.Order) ([]domain.Order, error) {
	if len(orders) == 0 {
		return orders, nil
	}

	orderIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	rows, err := r.queries.GetOrderItemsForOrders(ctx, orderIDs)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos order items", logger.Error(err))
		return nil, err
	}

	itemsByOrderID := make(map[int64][]domain.OrderItem, len(orders))
	for _, row := range rows {
		item := domain.OrderItem{
			ID:             row.ID,
			OrderID:        row.OrderID,
			MenuItemID:     row.MenuItemID,
			Name:           row.Name,
			SKU:            row.Sku,
			Quantity:       sqlutil.DecimalToFloat(row.Quantity),
			UnitPrice:      sqlutil.DecimalToFloat(row.UnitPrice),
			DiscountPct:    sqlutil.DecimalToFloat(row.DiscountPct),
			DiscountAmount: sqlutil.DecimalToFloat(row.DiscountAmount),
			TaxRate:        sqlutil.DecimalToFloat(row.TaxRate),
			Total:          sqlutil.DecimalToFloat(row.Total),
			Notes:          row.Notes,
			Status:         domain.ItemStatus(row.Status),
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		itemsByOrderID[item.OrderID] = append(itemsByOrderID[item.OrderID], item)
	}

	for i := range orders {
		orders[i].Items = itemsByOrderID[orders[i].ID]
	}

	return orders, nil
}

// dynamic-query: squirrel
func (r *Repository) countOrders(ctx context.Context, businessID, sessionID int64, status, dateFrom, dateTo, search string) (int64, error) {
	builder := sq.Select("COUNT(*)").
		From("pos_orders o").
		Where(sq.Eq{"o.business_id": businessID})
	builder = applyPOSOrderListFilters(builder, sessionID, status, dateFrom, dateTo, search)

	query, args, err := builder.PlaceholderFormat(sq.Question).ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build count pos orders query", logger.Error(err))
		return 0, err
	}

	var total int64
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("Failed to count pos orders", logger.Error(err))
		return 0, err
	}
	return total, nil
}

// dynamic-query: squirrel
func applyPOSOrderListFilters(builder sq.SelectBuilder, sessionID int64, status, dateFrom, dateTo, search string) sq.SelectBuilder {
	if sessionID > 0 {
		builder = builder.Where(sq.Eq{"o.session_id": sessionID})
	}

	if status != "" {
		builder = builder.Where(sq.Eq{"o.status": status})
	}

	if dateFrom != "" {
		builder = builder.Where(sq.GtOrEq{"o.created_at": dateFrom})
	}

	if dateTo != "" {
		builder = builder.Where("o.created_at < DATE_ADD(?, INTERVAL 1 DAY)", dateTo)
	}

	search = strings.TrimSpace(search)
	if search != "" {
		booleanQuery := searchutil.BooleanPrefixQuery(search)
		if booleanQuery != "" {
			builder = builder.Where(
				"(o.order_number = ? OR MATCH(o.order_number, o.customer_name, o.customer_phone) AGAINST (? IN BOOLEAN MODE))",
				search,
				booleanQuery,
			)
		} else {
			builder = builder.Where(sq.Eq{"o.order_number": search})
		}
	}

	return builder
}

func (r *Repository) GetOrder(ctx context.Context, businessID, orderID int64) (*domain.Order, error) {
	row, err := r.queries.GetOrder(ctx, sqlc.GetOrderParams{
		BusinessID: businessID,
		ID:         orderID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to get pos order", logger.Error(err))
		return nil, err
	}

	o := domain.Order{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		SessionID:      row.SessionID.Int64,
		TableID:        row.TableID.Int64,
		LocationID:     row.LocationID.Int64,
		OrderNumber:    row.OrderNumber,
		Type:           domain.OrderType(row.Type.PosOrdersType),
		Status:         domain.OrderStatus(row.Status.PosOrdersStatus),
		CustomerID:     row.CustomerID.Int64,
		CustomerName:   row.CustomerName.String,
		StaffID:        row.StaffID.Int64,
		Subtotal:       sqlutil.DecimalToFloat(row.Subtotal.String),
		DiscountAmount: sqlutil.DecimalToFloat(row.DiscountAmount.String),
		TaxAmount:      sqlutil.DecimalToFloat(row.TaxAmount.String),
		Total:          sqlutil.DecimalToFloat(row.Total.String),
		PaymentMethod:  domain.PaymentMethod(row.PaymentMethod.PosOrdersPaymentMethod),
		PaymentStatus:  domain.PaymentStatus(row.PaymentStatus.PosOrdersPaymentStatus),
		AmountPaid:     sqlutil.DecimalToFloat(row.AmountPaid.String),
		ChangeDue:      sqlutil.DecimalToFloat(row.ChangeDue.String),
		Notes:          row.Notes.String,
		KitchenNotes:   row.KitchenNotes.String,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		TipAmount:      sqlutil.DecimalToFloat(row.TipAmount.String),
		TerminalID:     row.TerminalID.Int64,
		BranchID:       row.BranchID.Int64,
	}
	if row.PaidAt.Valid {
		o.PaidAt = &row.PaidAt.Time
	}

	items, err := r.GetOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

// Retry generated numbers if an import races with allocation.
const createOrderNumberAttempts = 5

// CreateOrder allocates generated numbers under the business row lock.
func (r *Repository) CreateOrder(ctx context.Context, businessID, staffID int64, orderNumber string, input domain.OrderInput) (int64, error) {
	// Only a generated number is worth retrying: an explicit one is the
	// caller's assertion, and re-running the transaction with the same string
	// would collide identically every time.
	if orderNumber != "" {
		return r.createOrderOnce(ctx, businessID, staffID, orderNumber, input)
	}
	for attempt := 1; ; attempt++ {
		orderID, err := r.createOrderOnce(ctx, businessID, staffID, "", input)
		if err == nil {
			return orderID, nil
		}
		if attempt >= createOrderNumberAttempts || !isDuplicateEntry(err) {
			return 0, err
		}
		logger.Ctx(ctx).Warn("Order number collided, retrying create",
			logger.Any("business_id", businessID), logger.Any("attempt", attempt))
	}
}

func (r *Repository) createOrderOnce(ctx context.Context, businessID, staffID int64, orderNumber string, input domain.OrderInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin order transaction", logger.Error(err))
		return 0, err
	}

	// The guard keys off commit state, not the local err: several paths below
	// return a sentinel without assigning err, and the table-occupancy block
	// shadows it, so an err check would skip the rollback and leave the
	// transaction (and its row locks) open.
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	// Explicit numbers from imports are preserved.
	if orderNumber == "" {
		orderNumber, err = generateOrderNumber(ctx, tx, qtx, businessID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to generate order number", logger.Error(err))
			return 0, err
		}
	}

	input.Items, err = r.resolveOrderItems(ctx, qtx, businessID, staffID, input)
	if err != nil {
		return 0, err
	}

	// Compute per-item discount and totals once; reuse them in the insert loop
	// below instead of recomputing the same arithmetic a second time.
	totals := domain.ComputeTotals(input.Items)
	discAmts := make([]float64, len(input.Items))
	itemTotals := make([]float64, len(input.Items))
	for i, item := range input.Items {
		lineSubtotal := item.UnitPrice * item.Quantity
		discAmt := lineSubtotal * item.DiscountPct / 100
		if item.DiscountAmount > 0 {
			remaining := lineSubtotal - discAmt
			if item.DiscountAmount < remaining {
				discAmt += item.DiscountAmount
			} else {
				discAmt = lineSubtotal
			}
		}
		lineAfterDiscount := lineSubtotal - discAmt
		lineTax := lineAfterDiscount * item.TaxRate / 100
		itemTotal := lineAfterDiscount + lineTax
		discAmts[i] = discAmt
		itemTotals[i] = itemTotal
	}

	subtotal := totals.Subtotal
	taxAmount := totals.TaxAmount
	total := totals.Total

	res, err := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		SessionID:     sqlutil.NullInt64(input.SessionID),
		TableID:       sqlutil.NullInt64(input.TableID),
		LocationID:    sqlutil.NullInt64(input.LocationID),
		OrderNumber:   orderNumber,
		Type:          sqlc.NullPosOrdersType{PosOrdersType: sqlc.PosOrdersType(input.Type), Valid: true},
		CustomerID:    sqlutil.NullInt64(input.CustomerID),
		CustomerName:  sqlutil.NullString(input.CustomerName),
		StaffID:       sqlutil.NullInt64(staffID),
		Subtotal:      sql.NullString{String: sqlutil.FloatToDecimal(subtotal), Valid: true},
		TaxAmount:     sql.NullString{String: sqlutil.FloatToDecimal(taxAmount), Valid: true},
		Total:         sql.NullString{String: sqlutil.FloatToDecimal(total), Valid: true},
		PaymentMethod: nullPaymentMethod(input.PaymentMethod),
		Notes:         sqlutil.NullString(input.Notes),
		KitchenNotes:  sqlutil.NullString(input.KitchenNotes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert pos order", logger.Error(err))
		return 0, err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for i, item := range input.Items {
		modifierJSON, marshalErr := json.Marshal(item.ModifierIDs)
		if marshalErr != nil {
			return 0, marshalErr
		}
		_, err = qtx.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID:        orderID,
			MenuItemID:     sqlutil.NullInt64(item.MenuItemID),
			Name:           item.Name,
			Sku:            sqlutil.NullString(item.SKU),
			Quantity:       floatToDecimal3(item.Quantity),
			UnitPrice:      sqlutil.FloatToDecimal(item.UnitPrice),
			DiscountPct:    sql.NullString{String: sqlutil.FloatToDecimal(item.DiscountPct), Valid: true},
			DiscountAmount: sql.NullString{String: sqlutil.FloatToDecimal(discAmts[i]), Valid: true},
			TaxRate:        sql.NullString{String: sqlutil.FloatToDecimal(item.TaxRate), Valid: true},
			Total:          sqlutil.FloatToDecimal(itemTotals[i]),
			Notes:          sqlutil.NullString(item.Notes),
			Modifiers:      modifierJSON,
			ComboID:        sqlutil.NullInt64(item.ComboID),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to insert order item", logger.Error(err))
			return 0, err
		}
	}

	if input.TableID > 0 {
		res, err := qtx.SetTableOccupied(ctx, sqlc.SetTableOccupiedParams{
			CurrentSessionOrderID: sql.NullInt64{Int64: orderID, Valid: true},
			BusinessID:            businessID,
			ID:                    input.TableID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to set table occupied", logger.Error(err))
			return 0, err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		if affected == 0 {
			// The table either doesn't exist or belongs to another business.
			logger.Ctx(ctx).Warn("CreateOrder rejected: table not in business scope",
				logger.Any("table_id", input.TableID), logger.Any("business_id", businessID))
			return 0, domain.ErrTableNotFound
		}
	}

	// Recorded inside this transaction, so the order and the fact that it
	// happened commit together. Publishing after the commit would lose the
	// event on a crash between the two; publishing before would announce an
	// order that never existed.
	//
	// The payload carries identity only. Money is deliberately absent: the
	// totals in scope here are float64 (onion#1183) and putting one on the
	// wire would mint a new float64 money field into a contract that
	// subscribers depend on. A consumer that needs the value reads
	// pos_orders.total, which is DECIMAL.
	payload := map[string]any{
		"order_id":     orderID,
		"order_number": orderNumber,
		"order_type":   string(input.Type),
		"item_count":   len(input.Items),
	}
	if input.CustomerID > 0 {
		payload["customer_id"] = input.CustomerID
	}
	if err = eventbus.Publish(ctx, tx, eventbus.Event{
		BusinessID:  businessID,
		Type:        eventbus.TypeOrderCreated,
		SubjectType: eventbus.SubjectOrder,
		SubjectID:   orderID,
		Payload:     payload,
	}); err != nil {
		logger.Ctx(ctx).Error("Failed to record the order_created event", logger.Error(err))
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit create order transaction", logger.Error(err))
		return 0, err
	}
	committed = true

	return orderID, nil
}
