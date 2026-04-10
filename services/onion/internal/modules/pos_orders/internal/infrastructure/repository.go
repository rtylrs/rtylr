package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

// ─── Locations ───────────────────────────────────────────────────────────────

func (r *Repository) ListLocations(ctx context.Context, ownerID int64) ([]domain.Location, error) {
	rows, err := r.queries.ListLocationsByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos locations", logger.Error(err))
		return nil, err
	}

	var locations []domain.Location
	for _, row := range rows {
		locations = append(locations, domain.Location{
			ID:        row.ID,
			OwnerID:   row.OwnerID,
			Name:      row.Name,
			Address:   row.Address.String,
			Phone:     row.Phone.String,
			IsActive:  row.IsActive.Bool,
			CreatedAt: row.CreatedAt.Time,
		})
	}
	return locations, nil
}

func (r *Repository) CreateLocation(ctx context.Context, ownerID int64, input domain.LocationInput) (int64, error) {
	res, err := r.queries.CreateLocation(ctx, sqlc.CreateLocationParams{
		OwnerID: ownerID,
		Name:    input.Name,
		Address: nullString(input.Address),
		Phone:   nullString(input.Phone),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos location", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// ─── Tables ──────────────────────────────────────────────────────────────────

func (r *Repository) ListTables(ctx context.Context, ownerID int64, locationID int64, status string) ([]domain.Table, error) {
	// ListTablesByOwner returns all tables; we filter in-memory for locationID/status
	// since the sqlc query doesn't support optional filters.
	rows, err := r.queries.ListTablesByOwner(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos tables", logger.Error(err))
		return nil, err
	}

	var tables []domain.Table
	for _, row := range rows {
		if locationID > 0 && row.LocationID.Int64 != locationID {
			continue
		}
		if status != "" && string(row.Status.PosTablesStatus) != status {
			continue
		}
		tables = append(tables, domain.Table{
			ID:                    row.ID,
			OwnerID:               row.OwnerID,
			LocationID:            row.LocationID.Int64,
			Name:                  row.Name,
			Capacity:              int(row.Capacity.Int32),
			Status:                domain.TableStatus(row.Status.PosTablesStatus),
			CurrentSessionOrderID: row.CurrentSessionOrderID.Int64,
			QRCodeURL:             row.QrCodeUrl.String,
			CreatedAt:             row.CreatedAt.Time,
			UpdatedAt:             row.UpdatedAt.Time,
		})
	}
	return tables, nil
}

func (r *Repository) CreateTable(ctx context.Context, ownerID int64, input domain.TableInput) (int64, error) {
	res, err := r.queries.CreatePosTable(ctx, sqlc.CreatePosTableParams{
		OwnerID:    ownerID,
		LocationID: nullInt64(input.LocationID),
		Name:       input.Name,
		Capacity:   sql.NullInt32{Int32: int32(input.Capacity), Valid: true},
		QrCodeUrl:  nullString(input.QRCodeURL),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos table", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateTable(ctx context.Context, ownerID int64, input domain.TableInput) error {
	err := r.queries.UpdatePosTable(ctx, sqlc.UpdatePosTableParams{
		LocationID: nullInt64(input.LocationID),
		Name:       input.Name,
		Capacity:   sql.NullInt32{Int32: int32(input.Capacity), Valid: true},
		QrCodeUrl:  nullString(input.QRCodeURL),
		OwnerID:    ownerID,
		ID:         input.ID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update pos table", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) UpdateTableStatus(ctx context.Context, ownerID, tableID int64, status domain.TableStatus) error {
	err := r.queries.UpdateTableStatus(ctx, sqlc.UpdateTableStatusParams{
		Status:  sqlc.NullPosTablesStatus{PosTablesStatus: sqlc.PosTablesStatus(status), Valid: true},
		OwnerID: ownerID,
		ID:      tableID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update table status", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteTable(ctx context.Context, ownerID, tableID int64) error {
	err := r.queries.DeletePosTable(ctx, sqlc.DeletePosTableParams{
		OwnerID: ownerID,
		ID:      tableID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete pos table", logger.Error(err))
		return err
	}
	return nil
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (r *Repository) GetCurrentOpenSession(ctx context.Context, ownerID, staffID int64) (*domain.Session, error) {
	row, err := r.queries.GetCurrentOpenSession(ctx, sqlc.GetCurrentOpenSessionParams{
		OwnerID: ownerID,
		StaffID: staffID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Ctx(ctx).Error("Failed to get current open session", logger.Error(err))
		return nil, err
	}
	return sessionFromRow(row), nil
}

func (r *Repository) ListSessions(ctx context.Context, ownerID, locationID int64, dateFrom, dateTo string) ([]domain.Session, error) {
	// ListOrdersByOwner returns all sessions; we filter in-memory for optional params
	// as sqlc doesn't support dynamic WHERE clauses.
	// For this complex query, we keep using raw SQL.
	query := `
		SELECT id, owner_id, COALESCE(location_id,0), staff_id, status,
		       opening_cash, COALESCE(closing_cash,0), COALESCE(expected_cash,0),
		       COALESCE(cash_difference,0), total_sales, total_orders,
		       total_cash_sales, total_card_sales, total_online_sales, total_refunds,
		       opened_at, closed_at, COALESCE(notes,'')
		FROM pos_sessions
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if locationID > 0 {
		query += ` AND location_id = ?`
		args = append(args, locationID)
	}

	if dateFrom != "" {
		query += ` AND opened_at >= ?`
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		query += ` AND opened_at <= ?`
		args = append(args, dateTo)
	}

	query += ` ORDER BY opened_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos sessions", logger.Error(err))
		return nil, err
	}

	defer rows.Close()

	var sessions []domain.Session

	for rows.Next() {
		var s domain.Session
		var closedAt sql.NullTime

		if err := rows.Scan(
			&s.ID, &s.OwnerID, &s.LocationID, &s.StaffID, &s.Status,
			&s.OpeningCash, &s.ClosingCash, &s.ExpectedCash, &s.CashDifference,
			&s.TotalSales, &s.TotalOrders,
			&s.TotalCashSales, &s.TotalCardSales, &s.TotalOnlineSales, &s.TotalRefunds,
			&s.OpenedAt, &closedAt, &s.Notes,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan session row", logger.Error(err))
			return nil, err
		}

		if closedAt.Valid {
			s.ClosedAt = &closedAt.Time
		}

		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (r *Repository) OpenSession(ctx context.Context, ownerID, staffID, locationID int64, openingCash float64) (int64, error) {
	res, err := r.queries.OpenSession(ctx, sqlc.OpenSessionParams{
		OwnerID:     ownerID,
		LocationID:  nullInt64(locationID),
		StaffID:     staffID,
		OpeningCash: sql.NullString{String: floatToDecimal(openingCash), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to open pos session", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) CloseSession(ctx context.Context, ownerID, sessionID int64, closingCash, expectedCash, cashDiff float64, notes string) error {
	err := r.queries.CloseSession(ctx, sqlc.CloseSessionParams{
		ClosingCash:    sql.NullString{String: floatToDecimal(closingCash), Valid: true},
		ExpectedCash:   sql.NullString{String: floatToDecimal(expectedCash), Valid: true},
		CashDifference: sql.NullString{String: floatToDecimal(cashDiff), Valid: true},
		Notes:          nullString(notes),
		OwnerID:        ownerID,
		ID:             sessionID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to close pos session", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetSessionTotals(ctx context.Context, sessionID int64) (totalSales, cashSales, cardSales, onlineSales float64, totalOrders int, err error) {
	row, err := r.queries.GetSessionTotals(ctx, sql.NullInt64{Int64: sessionID, Valid: true})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session totals", logger.Error(err))
		return
	}
	totalSales = interfaceToFloat(row.TotalSales)
	cashSales = interfaceToFloat(row.CashSales)
	cardSales = interfaceToFloat(row.CardSales)
	onlineSales = interfaceToFloat(row.OnlineSales)
	totalOrders = int(row.TotalOrders)
	return
}

func (r *Repository) UpdateSessionTotals(ctx context.Context, sessionID int64, totalSales, cashSales, cardSales, onlineSales float64, totalOrders int) error {
	err := r.queries.UpdateSessionTotals(ctx, sqlc.UpdateSessionTotalsParams{
		TotalSales:       sql.NullString{String: floatToDecimal(totalSales), Valid: true},
		TotalCashSales:   sql.NullString{String: floatToDecimal(cashSales), Valid: true},
		TotalCardSales:   sql.NullString{String: floatToDecimal(cardSales), Valid: true},
		TotalOnlineSales: sql.NullString{String: floatToDecimal(onlineSales), Valid: true},
		TotalOrders:      sql.NullInt32{Int32: int32(totalOrders), Valid: true},
		ID:               sessionID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update session totals", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetSessionByID(ctx context.Context, ownerID, sessionID int64) (*domain.Session, error) {
	row, err := r.queries.GetSessionByID(ctx, sqlc.GetSessionByIDParams{
		OwnerID: ownerID,
		ID:      sessionID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}
		logger.Ctx(ctx).Error("Failed to get session by id", logger.Error(err))
		return nil, err
	}
	return sessionFromRow(row), nil
}

func (r *Repository) UpdateSessionRefundTotals(ctx context.Context, sessionID int64, refundAmount float64) error {
	err := r.queries.UpdateSessionRefundTotals(ctx, sqlc.UpdateSessionRefundTotalsParams{
		TotalRefunds: sql.NullString{String: floatToDecimal(refundAmount), Valid: true},
		ID:           sessionID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update session refund totals", logger.Error(err))
		return err
	}
	return nil
}

// ─── Orders ──────────────────────────────────────────────────────────────────

func (r *Repository) ListOrders(ctx context.Context, ownerID, sessionID int64, status, dateFrom, dateTo string) ([]domain.Order, error) {
	// Dynamic WHERE clause: use raw SQL for optional filters
	query := `
		SELECT id, owner_id, COALESCE(session_id,0), COALESCE(table_id,0), COALESCE(location_id,0),
		       order_number, type, status, COALESCE(customer_id,0), COALESCE(customer_name,''),
		       COALESCE(staff_id,0), subtotal, discount_amount, tax_amount, total,
		       payment_method, payment_status, amount_paid, change_due,
		       COALESCE(notes,''), COALESCE(kitchen_notes,''), created_at, updated_at, paid_at
		FROM pos_orders
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if sessionID > 0 {
		query += ` AND session_id = ?`
		args = append(args, sessionID)
	}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}

	if dateFrom != "" {
		query += ` AND created_at >= ?`
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		query += ` AND created_at <= ?`
		args = append(args, dateTo)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos orders", logger.Error(err))
		return nil, err
	}

	defer rows.Close()

	var orders []domain.Order

	for rows.Next() {
		o, scanErr := scanOrderRow(rows)
		if scanErr != nil {
			logger.Ctx(ctx).Error("Failed to scan order row", logger.Error(scanErr))
			return nil, scanErr
		}
		orders = append(orders, o)
	}

	return orders, rows.Err()
}

func (r *Repository) GetOrder(ctx context.Context, ownerID, orderID int64) (*domain.Order, error) {
	row, err := r.queries.GetOrder(ctx, sqlc.GetOrderParams{
		OwnerID: ownerID,
		ID:      orderID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to get pos order", logger.Error(err))
		return nil, err
	}

	o := orderFromRow(row)

	items, err := r.GetOrderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return &o, nil
}

func (r *Repository) CreateOrder(ctx context.Context, ownerID, staffID int64, orderNumber string, input domain.OrderInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin order transaction", logger.Error(err))
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
		discAmt := item.UnitPrice * item.Quantity * item.DiscountPct / 100
		itemTotal := (item.UnitPrice*item.Quantity - discAmt) * (1 + item.TaxRate/100)
		subtotal += itemTotal
	}

	total := subtotal

	res, err := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
		OwnerID:       ownerID,
		SessionID:     nullInt64(input.SessionID),
		TableID:       nullInt64(input.TableID),
		LocationID:    nullInt64(input.LocationID),
		OrderNumber:   orderNumber,
		Type:          sqlc.NullPosOrdersType{PosOrdersType: sqlc.PosOrdersType(input.Type), Valid: true},
		CustomerID:    nullInt64(input.CustomerID),
		CustomerName:  nullString(input.CustomerName),
		StaffID:       nullInt64(staffID),
		Subtotal:      sql.NullString{String: floatToDecimal(subtotal), Valid: true},
		Total:         sql.NullString{String: floatToDecimal(total), Valid: true},
		PaymentMethod: sqlc.NullPosOrdersPaymentMethod{PosOrdersPaymentMethod: sqlc.PosOrdersPaymentMethod(input.PaymentMethod), Valid: true},
		Notes:         nullString(input.Notes),
		KitchenNotes:  nullString(input.KitchenNotes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert pos order", logger.Error(err))
		return 0, err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range input.Items {
		discAmt := item.UnitPrice * item.Quantity * item.DiscountPct / 100
		itemTotal := (item.UnitPrice*item.Quantity - discAmt) * (1 + item.TaxRate/100)

		_, err = qtx.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID:        orderID,
			MenuItemID:     nullInt64(item.MenuItemID),
			Name:           item.Name,
			Sku:            nullString(item.SKU),
			Quantity:       floatToDecimal3(item.Quantity),
			UnitPrice:      floatToDecimal(item.UnitPrice),
			DiscountPct:    sql.NullString{String: floatToDecimal(item.DiscountPct), Valid: true},
			DiscountAmount: sql.NullString{String: floatToDecimal(discAmt), Valid: true},
			TaxRate:        sql.NullString{String: floatToDecimal(item.TaxRate), Valid: true},
			Total:          floatToDecimal(itemTotal),
			Notes:          nullString(item.Notes),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to insert order item", logger.Error(err))
			return 0, err
		}
	}

	if input.TableID > 0 {
		err = qtx.SetTableOccupied(ctx, sqlc.SetTableOccupiedParams{
			CurrentSessionOrderID: sql.NullInt64{Int64: orderID, Valid: true},
			ID:                    input.TableID,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to set table occupied", logger.Error(err))
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit create order transaction", logger.Error(err))
		return 0, err
	}

	return orderID, nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, ownerID, orderID int64, status domain.OrderStatus) error {
	err := r.queries.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		Status:  sqlc.NullPosOrdersStatus{PosOrdersStatus: sqlc.PosOrdersStatus(status), Valid: true},
		OwnerID: ownerID,
		ID:      orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update order status", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) UpdateOrderItemStatus(ctx context.Context, orderID, itemID int64, status domain.ItemStatus) error {
	err := r.queries.UpdateOrderItemStatus(ctx, sqlc.UpdateOrderItemStatusParams{
		Status:  sqlc.NullPosOrderItemsStatus{PosOrderItemsStatus: sqlc.PosOrderItemsStatus(status), Valid: true},
		OrderID: orderID,
		ID:      itemID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update order item status", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	rows, err := r.queries.GetOrderItems(ctx, orderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get order items", logger.Error(err))
		return nil, err
	}

	var items []domain.OrderItem
	for _, row := range rows {
		items = append(items, domain.OrderItem{
			ID:             row.ID,
			OrderID:        row.OrderID,
			MenuItemID:     row.MenuItemID.Int64,
			Name:           row.Name,
			SKU:            row.Sku.String,
			Quantity:       decimalToFloat(row.Quantity),
			UnitPrice:      decimalToFloat(row.UnitPrice),
			DiscountPct:    decimalToFloat(row.DiscountPct.String),
			DiscountAmount: decimalToFloat(row.DiscountAmount.String),
			TaxRate:        decimalToFloat(row.TaxRate.String),
			Total:          decimalToFloat(row.Total),
			Notes:          row.Notes.String,
			Status:         domain.ItemStatus(row.Status.PosOrderItemsStatus),
			CreatedAt:      row.CreatedAt.Time,
		})
	}
	return items, nil
}

func (r *Repository) ProcessPayment(ctx context.Context, ownerID, orderID int64, input domain.PaymentInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin payment transaction", logger.Error(err))
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	paymentRow, err := qtx.GetOrderForPayment(ctx, sqlc.GetOrderForPaymentParams{
		OwnerID: ownerID,
		ID:      orderID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrOrderNotFound
		}
		logger.Ctx(ctx).Error("Failed to get order for payment", logger.Error(err))
		return err
	}

	total := decimalToFloat(paymentRow.Total.String)
	changeDue := input.AmountPaid - total
	if changeDue < 0 {
		changeDue = 0
	}

	now := time.Now()

	err = qtx.MarkOrderPaid(ctx, sqlc.MarkOrderPaidParams{
		PaymentMethod: sqlc.NullPosOrdersPaymentMethod{PosOrdersPaymentMethod: sqlc.PosOrdersPaymentMethod(input.PaymentMethod), Valid: true},
		AmountPaid:    sql.NullString{String: floatToDecimal(input.AmountPaid), Valid: true},
		ChangeDue:     sql.NullString{String: floatToDecimal(changeDue), Valid: true},
		PaidAt:        sql.NullTime{Time: now, Valid: true},
		OwnerID:       ownerID,
		ID:            orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark order paid", logger.Error(err))
		return err
	}

	if paymentRow.TableID.Valid && paymentRow.TableID.Int64 > 0 && string(paymentRow.Type.PosOrdersType) == string(domain.OrderTypeDineIn) {
		err = qtx.FreeTable(ctx, paymentRow.TableID.Int64)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to free table after payment", logger.Error(err))
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit payment transaction", logger.Error(err))
		return err
	}

	return nil
}

func (r *Repository) CreateRefund(ctx context.Context, ownerID, orderID, refundedBy int64, amount float64, reason string) (int64, error) {
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

	res, err := qtx.CreateRefund(ctx, sqlc.CreateRefundParams{
		OwnerID:    ownerID,
		OrderID:    orderID,
		Amount:     floatToDecimal(amount),
		Reason:     nullString(reason),
		RefundedBy: nullInt64(refundedBy),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert refund", logger.Error(err))
		return 0, err
	}

	refundID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	err = qtx.MarkOrderRefunded(ctx, sqlc.MarkOrderRefundedParams{
		OwnerID: ownerID,
		ID:      orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark order refunded", logger.Error(err))
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit refund transaction", logger.Error(err))
		return 0, err
	}

	return refundID, nil
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	salesRow, err := r.queries.GetDashboardSalesToday(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get dashboard stats", logger.Error(err))
		return stats, err
	}

	stats.TotalSales = interfaceToFloat(salesRow.TotalSales)
	stats.OrderCount = int(salesRow.OrderCount)
	stats.AvgOrder = interfaceToFloat(salesRow.AvgOrder)

	turnover, err := r.queries.GetTableTurnover(ctx, ownerID)
	if err == nil && turnover.TableCount > 0 {
		stats.TableTurnover = interfaceToFloat(turnover.TurnedTables) / float64(turnover.TableCount)
	}

	return stats, nil
}

// ─── Order extras ───────────────────────────────────────────────────────────

func (r *Repository) GetOrderSessionID(ctx context.Context, orderID int64) (int64, error) {
	sessionID, err := r.queries.GetOrderSessionID(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if sessionID.Valid {
		return sessionID.Int64, nil
	}
	return 0, nil
}

func (r *Repository) SplitOrder(ctx context.Context, ownerID, orderID int64, splits [][]int64) ([]int64, error) {
	order, err := r.GetOrder(ctx, ownerID, orderID)
	if err != nil {
		return nil, err
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
	var newOrderIDs []int64

	for _, splitItemIDs := range splits {
		orderNum, genErr := r.GenerateOrderNumber(ctx, ownerID)
		if genErr != nil {
			txErr = genErr
			return nil, genErr
		}

		res, insertErr := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
			OwnerID:       ownerID,
			SessionID:     nullInt64(order.SessionID),
			TableID:       nullInt64(order.TableID),
			LocationID:    nullInt64(order.LocationID),
			OrderNumber:   orderNum,
			Type:          sqlc.NullPosOrdersType{PosOrdersType: sqlc.PosOrdersType(order.Type), Valid: true},
			CustomerID:    nullInt64(order.CustomerID),
			CustomerName:  nullString(order.CustomerName),
			StaffID:       nullInt64(order.StaffID),
			Subtotal:      sql.NullString{String: "0", Valid: true},
			Total:         sql.NullString{String: "0", Valid: true},
			PaymentMethod: sqlc.NullPosOrdersPaymentMethod{PosOrdersPaymentMethod: sqlc.PosOrdersPaymentMethod(order.PaymentMethod), Valid: true},
			Notes:         nullString(order.Notes),
			KitchenNotes:  nullString(order.KitchenNotes),
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

		recalcErr := qtx.RecalcOrderTotals(ctx, newID)
		if recalcErr != nil {
			txErr = recalcErr
			return nil, recalcErr
		}
	}

	cancelErr := qtx.CancelOrder(ctx, sqlc.CancelOrderParams{
		ID:      orderID,
		OwnerID: ownerID,
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

func (r *Repository) ApplyDiscount(ctx context.Context, ownerID int64, input domain.DiscountInput) error {
	subtotal, err := r.queries.GetOrderSubtotal(ctx, sqlc.GetOrderSubtotalParams{
		OwnerID: ownerID,
		ID:      input.OrderID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrOrderNotFound
		}
		return err
	}

	total := decimalToFloat(subtotal.String)
	var discountAmount float64
	if input.DiscountType == "percentage" {
		discountAmount = total * input.DiscountValue / 100
	} else {
		discountAmount = input.DiscountValue
	}

	newTotal := total - discountAmount
	if newTotal < 0 {
		newTotal = 0
	}

	err = r.queries.ApplyDiscount(ctx, sqlc.ApplyDiscountParams{
		DiscountAmount: sql.NullString{String: floatToDecimal(discountAmount), Valid: true},
		Total:          sql.NullString{String: floatToDecimal(newTotal), Valid: true},
		CONCAT:         fmt.Sprintf(" [Discount: %s]", input.Reason),
		ID:             input.OrderID,
		OwnerID:        ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to apply discount", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) VoidOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	result, err := r.queries.VoidOrder(ctx, sqlc.VoidOrderParams{
		CONCAT:  fmt.Sprintf(" [Voided: %s]", reason),
		ID:      orderID,
		OwnerID: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to void order", logger.Error(err))
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

func (r *Repository) HoldOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	result, err := r.queries.HoldOrder(ctx, sqlc.HoldOrderParams{
		CONCAT:  fmt.Sprintf(" [Held: %s]", reason),
		ID:      orderID,
		OwnerID: ownerID,
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

func (r *Repository) ListHeldOrders(ctx context.Context, ownerID int64) ([]domain.Order, error) {
	return r.ListOrders(ctx, ownerID, 0, "held", "", "")
}

func (r *Repository) GenerateOrderNumber(ctx context.Context, ownerID int64) (string, error) {
	today := time.Now().Format("20060102")

	count, err := r.queries.CountTodaysOrders(ctx, ownerID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ORD-%s-%04d", today, count+1), nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func floatToDecimal(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func floatToDecimal3(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}

func interfaceToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case []byte:
		f, _ := strconv.ParseFloat(string(val), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	case float64:
		return val
	default:
		f, _ := strconv.ParseFloat(fmt.Sprintf("%v", val), 64)
		return f
	}
}

func sessionFromRow(row sqlc.PosSession) *domain.Session {
	s := &domain.Session{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		LocationID:       row.LocationID.Int64,
		StaffID:          row.StaffID,
		Status:           domain.SessionStatus(row.Status.PosSessionsStatus),
		OpeningCash:      decimalToFloat(row.OpeningCash.String),
		ClosingCash:      decimalToFloat(row.ClosingCash.String),
		ExpectedCash:     decimalToFloat(row.ExpectedCash.String),
		CashDifference:   decimalToFloat(row.CashDifference.String),
		TotalSales:       decimalToFloat(row.TotalSales.String),
		TotalOrders:      int(row.TotalOrders.Int32),
		TotalCashSales:   decimalToFloat(row.TotalCashSales.String),
		TotalCardSales:   decimalToFloat(row.TotalCardSales.String),
		TotalOnlineSales: decimalToFloat(row.TotalOnlineSales.String),
		TotalRefunds:     decimalToFloat(row.TotalRefunds.String),
		OpenedAt:         row.OpenedAt.Time,
		Notes:            row.Notes.String,
	}
	if row.ClosedAt.Valid {
		s.ClosedAt = &row.ClosedAt.Time
	}
	return s
}

func orderFromRow(row sqlc.PosOrder) domain.Order {
	o := domain.Order{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		SessionID:      row.SessionID.Int64,
		TableID:        row.TableID.Int64,
		LocationID:     row.LocationID.Int64,
		OrderNumber:    row.OrderNumber,
		Type:           domain.OrderType(row.Type.PosOrdersType),
		Status:         domain.OrderStatus(row.Status.PosOrdersStatus),
		CustomerID:     row.CustomerID.Int64,
		CustomerName:   row.CustomerName.String,
		StaffID:        row.StaffID.Int64,
		Subtotal:       decimalToFloat(row.Subtotal.String),
		DiscountAmount: decimalToFloat(row.DiscountAmount.String),
		TaxAmount:      decimalToFloat(row.TaxAmount.String),
		Total:          decimalToFloat(row.Total.String),
		PaymentMethod:  domain.PaymentMethod(row.PaymentMethod.PosOrdersPaymentMethod),
		PaymentStatus:  domain.PaymentStatus(row.PaymentStatus.PosOrdersPaymentStatus),
		AmountPaid:     decimalToFloat(row.AmountPaid.String),
		ChangeDue:      decimalToFloat(row.ChangeDue.String),
		Notes:          row.Notes.String,
		KitchenNotes:   row.KitchenNotes.String,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
	if row.PaidAt.Valid {
		o.PaidAt = &row.PaidAt.Time
	}
	return o
}

func scanOrderRow(rows *sql.Rows) (domain.Order, error) {
	var o domain.Order
	var paidAt sql.NullTime

	if err := rows.Scan(
		&o.ID, &o.OwnerID, &o.SessionID, &o.TableID, &o.LocationID,
		&o.OrderNumber, &o.Type, &o.Status, &o.CustomerID, &o.CustomerName,
		&o.StaffID, &o.Subtotal, &o.DiscountAmount, &o.TaxAmount, &o.Total,
		&o.PaymentMethod, &o.PaymentStatus, &o.AmountPaid, &o.ChangeDue,
		&o.Notes, &o.KitchenNotes, &o.CreatedAt, &o.UpdatedAt, &paidAt,
	); err != nil {
		return domain.Order{}, err
	}

	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}

	return o, nil
}
