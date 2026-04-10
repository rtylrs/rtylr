package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

// ─── Locations ───────────────────────────────────────────────────────────────

func (r *Repository) ListLocations(ctx context.Context, ownerID int64) ([]domain.Location, error) {
	query := `
		SELECT id, owner_id, name, COALESCE(address,''), COALESCE(phone,''), is_active, created_at
		FROM pos_locations
		WHERE owner_id = ?
		ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos locations", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var locations []domain.Location

	for rows.Next() {
		var loc domain.Location
		if err := rows.Scan(&loc.ID, &loc.OwnerID, &loc.Name, &loc.Address, &loc.Phone, &loc.IsActive, &loc.CreatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan location row", logger.Error(err))

			return nil, err
		}

		locations = append(locations, loc)
	}

	return locations, rows.Err()
}

func (r *Repository) CreateLocation(ctx context.Context, ownerID int64, input domain.LocationInput) (int64, error) {
	query := `
		INSERT INTO pos_locations (owner_id, name, address, phone)
		VALUES (?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query, ownerID, input.Name, nullString(input.Address), nullString(input.Phone))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos location", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

// ─── Tables ──────────────────────────────────────────────────────────────────

func (r *Repository) ListTables(ctx context.Context, ownerID int64, locationID int64, status string) ([]domain.Table, error) {
	query := `
		SELECT id, owner_id, COALESCE(location_id,0), name, capacity, status,
		       COALESCE(current_session_order_id,0), COALESCE(qr_code_url,''), created_at, updated_at
		FROM pos_tables
		WHERE owner_id = ?`

	args := []interface{}{ownerID}

	if locationID > 0 {
		query += ` AND location_id = ?`
		args = append(args, locationID)
	}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}

	query += ` ORDER BY name ASC`

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos tables", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var tables []domain.Table

	for rows.Next() {
		var t domain.Table
		if err := rows.Scan(
			&t.ID, &t.OwnerID, &t.LocationID, &t.Name, &t.Capacity,
			&t.Status, &t.CurrentSessionOrderID, &t.QRCodeURL, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan table row", logger.Error(err))

			return nil, err
		}

		tables = append(tables, t)
	}

	return tables, rows.Err()
}

func (r *Repository) CreateTable(ctx context.Context, ownerID int64, input domain.TableInput) (int64, error) {
	query := `
		INSERT INTO pos_tables (owner_id, location_id, name, capacity, qr_code_url)
		VALUES (?, ?, ?, ?, ?)`

	res, err := r.pool.ExecContext(ctx, query,
		ownerID,
		nullInt64(input.LocationID),
		input.Name,
		input.Capacity,
		nullString(input.QRCodeURL),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos table", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) UpdateTable(ctx context.Context, ownerID int64, input domain.TableInput) error {
	query := `
		UPDATE pos_tables
		SET location_id = ?, name = ?, capacity = ?, qr_code_url = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query,
		nullInt64(input.LocationID),
		input.Name,
		input.Capacity,
		nullString(input.QRCodeURL),
		ownerID,
		input.ID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update pos table", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateTableStatus(ctx context.Context, ownerID, tableID int64, status domain.TableStatus) error {
	query := `UPDATE pos_tables SET status = ? WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, string(status), ownerID, tableID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update table status", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) DeleteTable(ctx context.Context, ownerID, tableID int64) error {
	query := `DELETE FROM pos_tables WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, ownerID, tableID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete pos table", logger.Error(err))

		return err
	}

	return nil
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (r *Repository) GetCurrentOpenSession(ctx context.Context, ownerID, staffID int64) (*domain.Session, error) {
	query := `
		SELECT id, owner_id, COALESCE(location_id,0), staff_id, status,
		       opening_cash, COALESCE(closing_cash,0), COALESCE(expected_cash,0),
		       COALESCE(cash_difference,0), total_sales, total_orders,
		       total_cash_sales, total_card_sales, total_online_sales, total_refunds,
		       opened_at, closed_at, COALESCE(notes,'')
		FROM pos_sessions
		WHERE owner_id = ? AND staff_id = ? AND status = 'open'
		LIMIT 1`

	row := r.pool.QueryRowContext(ctx, query, ownerID, staffID)

	sess, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		logger.Ctx(ctx).Error("Failed to get current open session", logger.Error(err))

		return nil, err
	}

	return sess, nil
}

func (r *Repository) ListSessions(ctx context.Context, ownerID, locationID int64, dateFrom, dateTo string) ([]domain.Session, error) {
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
	query := `
		INSERT INTO pos_sessions (owner_id, location_id, staff_id, status, opening_cash)
		VALUES (?, ?, ?, 'open', ?)`

	res, err := r.pool.ExecContext(ctx, query, ownerID, nullInt64(locationID), staffID, openingCash)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to open pos session", logger.Error(err))

		return 0, err
	}

	return res.LastInsertId()
}

func (r *Repository) CloseSession(ctx context.Context, ownerID, sessionID int64, closingCash, expectedCash, cashDiff float64, notes string) error {
	query := `
		UPDATE pos_sessions
		SET status = 'closed', closing_cash = ?, expected_cash = ?, cash_difference = ?,
		    closed_at = NOW(), notes = ?
		WHERE owner_id = ? AND id = ? AND status = 'open'`

	_, err := r.pool.ExecContext(ctx, query, closingCash, expectedCash, cashDiff, nullString(notes), ownerID, sessionID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to close pos session", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) GetSessionTotals(ctx context.Context, sessionID int64) (totalSales, cashSales, cardSales, onlineSales float64, totalOrders int, err error) {
	query := `
		SELECT
			COALESCE(SUM(total), 0),
			COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN payment_method = 'online' THEN total ELSE 0 END), 0),
			COUNT(*)
		FROM pos_orders
		WHERE session_id = ? AND payment_status = 'paid'`

	err = r.pool.QueryRowContext(ctx, query, sessionID).Scan(
		&totalSales, &cashSales, &cardSales, &onlineSales, &totalOrders,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session totals", logger.Error(err))
	}

	return
}

func (r *Repository) UpdateSessionTotals(ctx context.Context, sessionID int64, totalSales, cashSales, cardSales, onlineSales float64, totalOrders int) error {
	query := `
		UPDATE pos_sessions
		SET total_sales = ?, total_cash_sales = ?, total_card_sales = ?,
		    total_online_sales = ?, total_orders = ?
		WHERE id = ?`

	_, err := r.pool.ExecContext(ctx, query, totalSales, cashSales, cardSales, onlineSales, totalOrders, sessionID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update session totals", logger.Error(err))

		return err
	}

	return nil
}

// ─── Orders ──────────────────────────────────────────────────────────────────

func (r *Repository) ListOrders(ctx context.Context, ownerID, sessionID int64, status, dateFrom, dateTo string) ([]domain.Order, error) {
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
	query := `
		SELECT id, owner_id, COALESCE(session_id,0), COALESCE(table_id,0), COALESCE(location_id,0),
		       order_number, type, status, COALESCE(customer_id,0), COALESCE(customer_name,''),
		       COALESCE(staff_id,0), subtotal, discount_amount, tax_amount, total,
		       payment_method, payment_status, amount_paid, change_due,
		       COALESCE(notes,''), COALESCE(kitchen_notes,''), created_at, updated_at, paid_at
		FROM pos_orders
		WHERE owner_id = ? AND id = ?`

	row := r.pool.QueryRowContext(ctx, query, ownerID, orderID)

	var o domain.Order
	var paidAt sql.NullTime

	if err := row.Scan(
		&o.ID, &o.OwnerID, &o.SessionID, &o.TableID, &o.LocationID,
		&o.OrderNumber, &o.Type, &o.Status, &o.CustomerID, &o.CustomerName,
		&o.StaffID, &o.Subtotal, &o.DiscountAmount, &o.TaxAmount, &o.Total,
		&o.PaymentMethod, &o.PaymentStatus, &o.AmountPaid, &o.ChangeDue,
		&o.Notes, &o.KitchenNotes, &o.CreatedAt, &o.UpdatedAt, &paidAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}

		logger.Ctx(ctx).Error("Failed to get pos order", logger.Error(err))

		return nil, err
	}

	if paidAt.Valid {
		o.PaidAt = &paidAt.Time
	}

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

	var subtotal float64

	for _, item := range input.Items {
		discAmt := item.UnitPrice * item.Quantity * item.DiscountPct / 100
		itemTotal := (item.UnitPrice*item.Quantity - discAmt) * (1 + item.TaxRate/100)
		subtotal += itemTotal
	}

	total := subtotal

	res, err := tx.ExecContext(ctx, `
		INSERT INTO pos_orders
		(owner_id, session_id, table_id, location_id, order_number, type, status,
		 customer_id, customer_name, staff_id, subtotal, total, payment_method, notes, kitchen_notes)
		VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, ?, ?, ?, ?, ?)`,
		ownerID,
		nullInt64(input.SessionID),
		nullInt64(input.TableID),
		nullInt64(input.LocationID),
		orderNumber,
		string(input.Type),
		nullInt64(input.CustomerID),
		nullString(input.CustomerName),
		nullInt64(staffID),
		subtotal,
		total,
		string(input.PaymentMethod),
		nullString(input.Notes),
		nullString(input.KitchenNotes),
	)
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

		_, err = tx.ExecContext(ctx, `
			INSERT INTO pos_order_items
			(order_id, menu_item_id, name, sku, quantity, unit_price,
			 discount_pct, discount_amount, tax_rate, total, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			orderID,
			nullInt64(item.MenuItemID),
			item.Name,
			nullString(item.SKU),
			item.Quantity,
			item.UnitPrice,
			item.DiscountPct,
			discAmt,
			item.TaxRate,
			itemTotal,
			nullString(item.Notes),
		)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to insert order item", logger.Error(err))

			return 0, err
		}
	}

	if input.TableID > 0 {
		_, err = tx.ExecContext(ctx,
			`UPDATE pos_tables SET status = 'occupied', current_session_order_id = ? WHERE id = ?`,
			orderID, input.TableID,
		)
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
	query := `UPDATE pos_orders SET status = ? WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, string(status), ownerID, orderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update order status", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) UpdateOrderItemStatus(ctx context.Context, orderID, itemID int64, status domain.ItemStatus) error {
	query := `UPDATE pos_order_items SET status = ? WHERE order_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, string(status), orderID, itemID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update order item status", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	query := `
		SELECT id, order_id, COALESCE(menu_item_id,0), name, COALESCE(sku,''),
		       quantity, unit_price, discount_pct, discount_amount, tax_rate, total,
		       COALESCE(notes,''), status, created_at
		FROM pos_order_items
		WHERE order_id = ?`

	rows, err := r.pool.QueryContext(ctx, query, orderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get order items", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var items []domain.OrderItem

	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(
			&item.ID, &item.OrderID, &item.MenuItemID, &item.Name, &item.SKU,
			&item.Quantity, &item.UnitPrice, &item.DiscountPct, &item.DiscountAmount,
			&item.TaxRate, &item.Total, &item.Notes, &item.Status, &item.CreatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan order item row", logger.Error(err))

			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
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

	var total float64
	var tableID sql.NullInt64
	var orderType string

	err = tx.QueryRowContext(ctx,
		`SELECT total, table_id, type FROM pos_orders WHERE owner_id = ? AND id = ?`,
		ownerID, orderID,
	).Scan(&total, &tableID, &orderType)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrOrderNotFound
		}

		logger.Ctx(ctx).Error("Failed to get order for payment", logger.Error(err))

		return err
	}

	changeDue := input.AmountPaid - total
	if changeDue < 0 {
		changeDue = 0
	}

	now := time.Now()

	_, err = tx.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = 'paid', payment_status = 'paid', payment_method = ?,
		    amount_paid = ?, change_due = ?, paid_at = ?
		WHERE owner_id = ? AND id = ?`,
		string(input.PaymentMethod),
		input.AmountPaid,
		changeDue,
		now,
		ownerID,
		orderID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark order paid", logger.Error(err))

		return err
	}

	if tableID.Valid && tableID.Int64 > 0 && orderType == string(domain.OrderTypeDineIn) {
		_, err = tx.ExecContext(ctx,
			`UPDATE pos_tables SET status = 'available', current_session_order_id = NULL WHERE id = ?`,
			tableID.Int64,
		)
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

	res, err := tx.ExecContext(ctx, `
		INSERT INTO pos_refunds (owner_id, order_id, amount, reason, refunded_by)
		VALUES (?, ?, ?, ?, ?)`,
		ownerID, orderID, amount, nullString(reason), nullInt64(refundedBy),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert refund", logger.Error(err))

		return 0, err
	}

	refundID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE pos_orders SET status = 'refunded', payment_status = 'refunded' WHERE owner_id = ? AND id = ?`,
		ownerID, orderID,
	)
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

	err := r.pool.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(total), 0),
			COUNT(*),
			COALESCE(AVG(total), 0)
		FROM pos_orders
		WHERE owner_id = ? AND payment_status = 'paid' AND DATE(paid_at) = CURDATE()`,
		ownerID,
	).Scan(&stats.TotalSales, &stats.OrderCount, &stats.AvgOrder)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get dashboard stats", logger.Error(err))

		return stats, err
	}

	var tableCount, turnedTables int

	err = r.pool.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN status != 'available' THEN 1 ELSE 0 END), 0)
		FROM pos_tables WHERE owner_id = ?`,
		ownerID,
	).Scan(&tableCount, &turnedTables)
	if err == nil && tableCount > 0 {
		stats.TableTurnover = float64(turnedTables) / float64(tableCount)
	}

	return stats, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

type sessionScanner interface {
	Scan(dest ...any) error
}

func scanSession(row sessionScanner) (*domain.Session, error) {
	var s domain.Session
	var closedAt sql.NullTime

	if err := row.Scan(
		&s.ID, &s.OwnerID, &s.LocationID, &s.StaffID, &s.Status,
		&s.OpeningCash, &s.ClosingCash, &s.ExpectedCash, &s.CashDifference,
		&s.TotalSales, &s.TotalOrders,
		&s.TotalCashSales, &s.TotalCardSales, &s.TotalOnlineSales, &s.TotalRefunds,
		&s.OpenedAt, &closedAt, &s.Notes,
	); err != nil {
		return nil, err
	}

	if closedAt.Valid {
		s.ClosedAt = &closedAt.Time
	}

	return &s, nil
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

func (r *Repository) GetSessionByID(ctx context.Context, ownerID, sessionID int64) (*domain.Session, error) {
	query := `
		SELECT id, owner_id, COALESCE(location_id,0), staff_id, status,
		       opening_cash, COALESCE(closing_cash,0), COALESCE(expected_cash,0),
		       COALESCE(cash_difference,0), total_sales, total_orders,
		       total_cash_sales, total_card_sales, total_online_sales, total_refunds,
		       opened_at, closed_at, COALESCE(notes,'')
		FROM pos_sessions
		WHERE owner_id = ? AND id = ?`

	row := r.pool.QueryRowContext(ctx, query, ownerID, sessionID)

	sess, err := scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrSessionNotFound
		}

		logger.Ctx(ctx).Error("Failed to get session by id", logger.Error(err))

		return nil, err
	}

	return sess, nil
}

func (r *Repository) UpdateSessionRefundTotals(ctx context.Context, sessionID int64, refundAmount float64) error {
	query := `UPDATE pos_sessions SET total_refunds = total_refunds + ? WHERE id = ?`

	_, err := r.pool.ExecContext(ctx, query, refundAmount, sessionID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update session refund totals", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) GetOrderSessionID(ctx context.Context, orderID int64) (int64, error) {
	var sessionID sql.NullInt64

	err := r.pool.QueryRowContext(ctx, `SELECT session_id FROM pos_orders WHERE id = ?`, orderID).Scan(&sessionID)
	if err != nil {
		return 0, err
	}

	if sessionID.Valid {
		return sessionID.Int64, nil
	}

	return 0, nil
}

// ─── Split Order ────────────────────────────────────────────────────────────

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

	var newOrderIDs []int64

	for _, splitItemIDs := range splits {
		orderNum, genErr := r.GenerateOrderNumber(ctx, ownerID)
		if genErr != nil {
			txErr = genErr
			return nil, genErr
		}

		res, insertErr := tx.ExecContext(ctx, `
			INSERT INTO pos_orders
			(owner_id, session_id, table_id, location_id, order_number, type, status,
			 customer_id, customer_name, staff_id, subtotal, total, payment_method, notes, kitchen_notes)
			VALUES (?, ?, ?, ?, ?, ?, 'pending', ?, ?, ?, 0, 0, ?, ?, ?)`,
			ownerID,
			nullInt64(order.SessionID),
			nullInt64(order.TableID),
			nullInt64(order.LocationID),
			orderNum,
			string(order.Type),
			nullInt64(order.CustomerID),
			nullString(order.CustomerName),
			nullInt64(order.StaffID),
			string(order.PaymentMethod),
			nullString(order.Notes),
			nullString(order.KitchenNotes),
		)
		if insertErr != nil {
			txErr = insertErr
			logger.Ctx(ctx).Error("Failed to insert split order", logger.Error(insertErr))
			return nil, insertErr
		}

		newID, _ := res.LastInsertId()
		newOrderIDs = append(newOrderIDs, newID)

		for _, itemID := range splitItemIDs {
			_, moveErr := tx.ExecContext(ctx,
				`UPDATE pos_order_items SET order_id = ? WHERE id = ? AND order_id = ?`,
				newID, itemID, orderID,
			)
			if moveErr != nil {
				txErr = moveErr
				return nil, moveErr
			}
		}

		// Recalculate new order totals
		_, recalcErr := tx.ExecContext(ctx, `
			UPDATE pos_orders o SET
				subtotal = (SELECT COALESCE(SUM(total),0) FROM pos_order_items WHERE order_id = o.id),
				total = (SELECT COALESCE(SUM(total),0) FROM pos_order_items WHERE order_id = o.id)
			WHERE id = ?`, newID)
		if recalcErr != nil {
			txErr = recalcErr
			return nil, recalcErr
		}
	}

	// Cancel original order
	_, cancelErr := tx.ExecContext(ctx,
		`UPDATE pos_orders SET status = 'cancelled' WHERE id = ? AND owner_id = ?`,
		orderID, ownerID,
	)
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

// ─── Apply Discount ─────────────────────────────────────────────────────────

func (r *Repository) ApplyDiscount(ctx context.Context, ownerID int64, input domain.DiscountInput) error {
	var total float64
	err := r.pool.QueryRowContext(ctx,
		`SELECT subtotal FROM pos_orders WHERE owner_id = ? AND id = ?`,
		ownerID, input.OrderID,
	).Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrOrderNotFound
		}
		return err
	}

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

	_, err = r.pool.ExecContext(ctx,
		`UPDATE pos_orders SET discount_amount = ?, total = ?, notes = CONCAT(COALESCE(notes,''), ?)
		 WHERE id = ? AND owner_id = ?`,
		discountAmount, newTotal,
		fmt.Sprintf(" [Discount: %s]", input.Reason),
		input.OrderID, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to apply discount", logger.Error(err))
		return err
	}
	return nil
}

// ─── Void Order ─────────────────────────────────────────────────────────────

func (r *Repository) VoidOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_orders SET status = 'voided', notes = CONCAT(COALESCE(notes,''), ?)
		 WHERE id = ? AND owner_id = ?`,
		fmt.Sprintf(" [Voided: %s]", reason),
		orderID, ownerID,
	)
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

// ─── Hold Order ─────────────────────────────────────────────────────────────

func (r *Repository) HoldOrder(ctx context.Context, ownerID, orderID int64, reason string) error {
	result, err := r.pool.ExecContext(ctx,
		`UPDATE pos_orders SET status = 'held', notes = CONCAT(COALESCE(notes,''), ?)
		 WHERE id = ? AND owner_id = ?`,
		fmt.Sprintf(" [Held: %s]", reason),
		orderID, ownerID,
	)
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

// ─── List Held Orders ───────────────────────────────────────────────────────

func (r *Repository) ListHeldOrders(ctx context.Context, ownerID int64) ([]domain.Order, error) {
	return r.ListOrders(ctx, ownerID, 0, "held", "", "")
}

// GenerateOrderNumber produces ORD-YYYYMMDD-XXXX format using today's order count.
func (r *Repository) GenerateOrderNumber(ctx context.Context, ownerID int64) (string, error) {
	today := time.Now().Format("20060102")

	var count int

	err := r.pool.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM pos_orders WHERE owner_id = ? AND DATE(created_at) = CURDATE()`,
		ownerID,
	).Scan(&count)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ORD-%s-%04d", today, count+1), nil
}
