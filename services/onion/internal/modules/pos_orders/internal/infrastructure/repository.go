package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/searchutil"
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

// ─── Locations ───────────────────────────────────────────────────────────────

func (r *Repository) ListLocations(ctx context.Context, businessID int64) ([]domain.Location, error) {
	rows, err := r.queries.ListLocationsByOwner(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos locations", logger.Error(err))
		return nil, err
	}

	locations := make([]domain.Location, 0, len(rows))
	for _, row := range rows {
		locations = append(locations, domain.Location{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			Address:    row.Address.String,
			Phone:      row.Phone.String,
			IsActive:   row.IsActive.Bool,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return locations, nil
}

func (r *Repository) CreateLocation(ctx context.Context, ownerID int64, input domain.LocationInput) (int64, error) {
	res, err := r.queries.CreateLocation(ctx, sqlc.CreateLocationParams{
		OwnerID:    ownerID,
		BusinessID: ownerID,
		Name:       input.Name,
		Address:    nullString(input.Address),
		Phone:      nullString(input.Phone),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos location", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// ─── Tables ──────────────────────────────────────────────────────────────────

func (r *Repository) ListTables(ctx context.Context, businessID, locationID int64, status string) ([]domain.Table, error) {
	rows, err := r.queries.ListTablesByOwner(ctx, sqlc.ListTablesByOwnerParams{
		BusinessID: businessID,
		LocationID: sql.NullInt64{Int64: locationID, Valid: true},
		Status:     sqlc.NullPosTablesStatus{PosTablesStatus: sqlc.PosTablesStatus(status), Valid: status != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list pos tables", logger.Error(err))
		return nil, err
	}

	tables := make([]domain.Table, 0, len(rows))
	for _, row := range rows {
		tables = append(tables, domain.Table{
			ID:                    row.ID,
			OwnerID:               row.OwnerID,
			BusinessID:            row.BusinessID,
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

func (r *Repository) CreateTable(ctx context.Context, businessID int64, input domain.TableInput) (int64, error) {
	res, err := r.queries.CreatePosTable(ctx, sqlc.CreatePosTableParams{
		OwnerID:    businessID,
		BusinessID: businessID,
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

func (r *Repository) UpdateTable(ctx context.Context, businessID int64, input domain.TableInput) error {
	err := r.queries.UpdatePosTable(ctx, sqlc.UpdatePosTableParams{
		LocationID: nullInt64(input.LocationID),
		Name:       input.Name,
		Capacity:   sql.NullInt32{Int32: int32(input.Capacity), Valid: true},
		QrCodeUrl:  nullString(input.QRCodeURL),
		BusinessID: businessID,
		ID:         input.ID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update pos table", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) UpdateTableStatus(ctx context.Context, businessID, tableID int64, status domain.TableStatus) error {
	err := r.queries.UpdateTableStatus(ctx, sqlc.UpdateTableStatusParams{
		Status:     sqlc.NullPosTablesStatus{PosTablesStatus: sqlc.PosTablesStatus(status), Valid: true},
		BusinessID: businessID,
		ID:         tableID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update table status", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) DeleteTable(ctx context.Context, businessID, tableID int64) error {
	err := r.queries.DeletePosTable(ctx, sqlc.DeletePosTableParams{
		BusinessID: businessID,
		ID:         tableID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to delete pos table", logger.Error(err))
		return err
	}
	return nil
}

// ─── Sessions ─────────────────────────────────────────────────────────────────

func (r *Repository) GetCurrentOpenSession(ctx context.Context, businessID, staffID int64) (*domain.Session, error) {
	row, err := r.queries.GetCurrentOpenSession(ctx, sqlc.GetCurrentOpenSessionParams{
		BusinessID: businessID,
		StaffID:    staffID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Ctx(ctx).Error("Failed to get current open session", logger.Error(err))
		return nil, err
	}
	s := &domain.Session{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		BusinessID:       row.BusinessID,
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
	return s, nil
}

// dynamic-query: squirrel
func (r *Repository) ListSessions(ctx context.Context, businessID, locationID int64, dateFrom, dateTo string) ([]domain.Session, error) {
	// Optional filters: location and date range.
	builder := sq.
		Select(
			"id", "owner_id", "COALESCE(location_id,0)", "staff_id", "status",
			"opening_cash", "COALESCE(closing_cash,0)", "COALESCE(expected_cash,0)",
			"COALESCE(cash_difference,0)", "total_sales", "total_orders",
			"total_cash_sales", "total_card_sales", "total_online_sales", "total_refunds",
			"opened_at", "closed_at", "COALESCE(notes,'')",
		).
		From("pos_sessions").
		Where(sq.Eq{"business_id": businessID})

	if locationID > 0 {
		builder = builder.Where(sq.Eq{"location_id": locationID})
	}

	if dateFrom != "" {
		builder = builder.Where(sq.GtOrEq{"opened_at": dateFrom})
	}

	if dateTo != "" {
		builder = builder.Where(sq.LtOrEq{"opened_at": dateTo})
	}

	query, args, err := builder.
		OrderBy("opened_at DESC").
		Limit(500).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build list pos sessions query", logger.Error(err))
		return nil, err
	}

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

func (r *Repository) OpenSession(ctx context.Context, businessID, staffID, locationID int64, openingCash float64) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	lockName := fmt.Sprintf("pos_session:%d:%d", businessID, staffID)
	var acquired int
	if err := tx.QueryRowContext(ctx, `SELECT GET_LOCK(?, 5)`, lockName).Scan(&acquired); err != nil {
		return 0, err
	}
	if acquired != 1 {
		return 0, domain.ErrSessionAlreadyOpen
	}
	defer tx.ExecContext(ctx, `SELECT RELEASE_LOCK(?)`, lockName)
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pos_sessions WHERE business_id = ? AND staff_id = ? AND status = 'open')`, businessID, staffID).Scan(&exists); err != nil {
		return 0, err
	}
	if exists == 1 {
		return 0, domain.ErrSessionAlreadyOpen
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO pos_sessions (owner_id, business_id, location_id, staff_id, status, opening_cash)
		VALUES (?, ?, ?, ?, 'open', ?)`, businessID, businessID, locationID, staffID, floatToDecimal(openingCash))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to open pos session", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) CloseSessionForStaff(ctx context.Context, businessID, staffID, sessionID int64, closingCash, expectedCash, cashDiff float64, notes string) error {
	result, err := r.pool.ExecContext(ctx, `
		UPDATE pos_sessions
		SET status = 'closed', closing_cash = ?, expected_cash = ?, cash_difference = ?, closed_at = NOW(), notes = ?
		WHERE business_id = ? AND id = ? AND staff_id = ? AND status = 'open'`,
		floatToDecimal(closingCash), floatToDecimal(expectedCash), floatToDecimal(cashDiff), notes, businessID, sessionID, staffID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrSessionNotFound
	}
	return nil
}

func (r *Repository) CloseSession(ctx context.Context, businessID, sessionID int64, closingCash, expectedCash, cashDiff float64, notes string) error {
	err := r.queries.CloseSession(ctx, sqlc.CloseSessionParams{
		ClosingCash:    sql.NullString{String: floatToDecimal(closingCash), Valid: true},
		ExpectedCash:   sql.NullString{String: floatToDecimal(expectedCash), Valid: true},
		CashDifference: sql.NullString{String: floatToDecimal(cashDiff), Valid: true},
		Notes:          nullString(notes),
		BusinessID:     businessID,
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
		return totalSales, cashSales, cardSales, onlineSales, totalOrders, err
	}
	totalSales = interfaceToFloat(row.TotalSales)
	cashSales = interfaceToFloat(row.CashSales)
	cardSales = interfaceToFloat(row.CardSales)
	onlineSales = interfaceToFloat(row.OnlineSales)
	totalOrders = int(row.TotalOrders)
	return totalSales, cashSales, cardSales, onlineSales, totalOrders, err
}

// GetSessionCashAdjustments returns the net drawer movement and refunds for a
// session. Drops, pickups, and cash-outs leave the drawer; cash-ins and
// adjustments add to it. Refunds reduce the cash that should remain.
func (r *Repository) GetSessionCashAdjustments(ctx context.Context, businessID, sessionID int64) (netMovement, refunds float64, err error) {
	err = r.pool.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(CASE
			WHEN type IN ('drop', 'pickup', 'cash_out') THEN -amount
			WHEN type IN ('cash_in', 'adjustment') THEN amount
			ELSE 0
		END), 0)
		FROM pos_cash_movements
		WHERE business_id = ? AND session_id = ?`, businessID, sessionID).Scan(&netMovement)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session cash movements", logger.Error(err))
		return 0, 0, err
	}

	err = r.pool.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(r.amount), 0)
		FROM pos_refunds r
		JOIN pos_orders o ON o.id = r.order_id AND o.business_id = r.business_id
		WHERE r.business_id = ? AND o.session_id = ?`, businessID, sessionID).Scan(&refunds)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get session refunds", logger.Error(err))
		return 0, 0, err
	}

	return netMovement, refunds, nil
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

func (r *Repository) GetSessionByID(ctx context.Context, businessID, sessionID int64) (*domain.Session, error) {
	row, err := r.queries.GetSessionByID(ctx, sqlc.GetSessionByIDParams{
		BusinessID: businessID,
		ID:         sessionID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		logger.Ctx(ctx).Error("Failed to get session by id", logger.Error(err))
		return nil, err
	}
	s := &domain.Session{
		ID:               row.ID,
		OwnerID:          row.OwnerID,
		BusinessID:       row.BusinessID,
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
	return s, nil
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
		Limit(uint64(limit + 1)).
		Offset(uint64(offset)).
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
			Quantity:       decimalToFloat(row.Quantity),
			UnitPrice:      decimalToFloat(row.UnitPrice),
			DiscountPct:    decimalToFloat(row.DiscountPct),
			DiscountAmount: decimalToFloat(row.DiscountAmount),
			TaxRate:        decimalToFloat(row.TaxRate),
			Total:          decimalToFloat(row.Total),
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
		TipAmount:      decimalToFloat(row.TipAmount.String),
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

func (r *Repository) CreateOrder(ctx context.Context, businessID, staffID int64, orderNumber string, input domain.OrderInput) (int64, error) {
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
		SessionID:     nullInt64(input.SessionID),
		TableID:       nullInt64(input.TableID),
		LocationID:    nullInt64(input.LocationID),
		OrderNumber:   orderNumber,
		Type:          sqlc.NullPosOrdersType{PosOrdersType: sqlc.PosOrdersType(input.Type), Valid: true},
		CustomerID:    nullInt64(input.CustomerID),
		CustomerName:  nullString(input.CustomerName),
		StaffID:       nullInt64(staffID),
		Subtotal:      sql.NullString{String: floatToDecimal(subtotal), Valid: true},
		TaxAmount:     sql.NullString{String: floatToDecimal(taxAmount), Valid: true},
		Total:         sql.NullString{String: floatToDecimal(total), Valid: true},
		PaymentMethod: sqlc.NullPosOrdersPaymentMethod{PosOrdersPaymentMethod: sqlc.PosOrdersPaymentMethod(input.PaymentMethod), Valid: input.PaymentMethod != ""},
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

	for i, item := range input.Items {
		modifierJSON, marshalErr := json.Marshal(item.ModifierIDs)
		if marshalErr != nil {
			return 0, marshalErr
		}
		_, err = qtx.CreateOrderItem(ctx, sqlc.CreateOrderItemParams{
			OrderID:        orderID,
			MenuItemID:     nullInt64(item.MenuItemID),
			Name:           item.Name,
			Sku:            nullString(item.SKU),
			Quantity:       floatToDecimal3(item.Quantity),
			UnitPrice:      floatToDecimal(item.UnitPrice),
			DiscountPct:    sql.NullString{String: floatToDecimal(item.DiscountPct), Valid: true},
			DiscountAmount: sql.NullString{String: floatToDecimal(discAmts[i]), Valid: true},
			TaxRate:        sql.NullString{String: floatToDecimal(item.TaxRate), Valid: true},
			Total:          floatToDecimal(itemTotals[i]),
			Notes:          nullString(item.Notes),
			Modifiers:      modifierJSON,
			ComboID:        nullInt64(item.ComboID),
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

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit create order transaction", logger.Error(err))
		return 0, err
	}

	return orderID, nil
}

func (r *Repository) resolveOrderItems(ctx context.Context, qtx *sqlc.Queries, businessID, staffID int64, input domain.OrderInput) ([]domain.OrderItemInput, error) {
	items := make([]domain.OrderItemInput, len(input.Items))
	copy(items, input.Items)

	needsDiscountCapability := false
	for _, item := range items {
		if item.DiscountPct > 0 {
			needsDiscountCapability = true
			break
		}
	}

	var maxDiscountPct float64
	var canApplyDiscount bool
	if needsDiscountCapability {
		capability, err := qtx.GetPOSUserCapabilities(ctx, sqlc.GetPOSUserCapabilitiesParams{
			BusinessID:   businessID,
			UserID:       staffID,
			LocationID:   input.LocationID,
			Column4:      input.LocationID,
			LocationID_2: input.LocationID,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, domain.ErrDiscountNotAllowed
			}
			return nil, err
		}
		canApplyDiscount = capability.CanApplyDiscount.Valid && capability.CanApplyDiscount.Bool
		maxDiscountPct = decimalToFloat(capability.MaxDiscountPct.String)
		if maxDiscountPct > 100 {
			maxDiscountPct = 100
		}
	}

	for i := range items {
		item := &items[i]
		if item.DiscountPct < 0 || item.DiscountPct > 100 {
			return nil, domain.ErrDiscountExceedsLimit
		}
		if item.DiscountPct > 0 {
			if !canApplyDiscount {
				return nil, domain.ErrDiscountNotAllowed
			}
			if item.DiscountPct > maxDiscountPct {
				return nil, domain.ErrDiscountExceedsLimit
			}
		}

		modifierTotal, err := r.resolveModifierTotal(ctx, qtx, businessID, item.ModifierIDs)
		if err != nil {
			return nil, err
		}

		if item.ComboID > 0 {
			if item.MenuItemID > 0 {
				return nil, domain.ErrPricedItemNotFound
			}
			combo, comboErr := qtx.GetPOSComboForOrder(ctx, sqlc.GetPOSComboForOrderParams{
				BusinessID: businessID,
				ID:         item.ComboID,
			})
			if comboErr != nil {
				if errors.Is(comboErr, sql.ErrNoRows) {
					return nil, domain.ErrPricedItemNotFound
				}
				return nil, comboErr
			}
			item.Name = combo.Name
			item.SKU = ""
			item.UnitPrice = decimalToFloat(combo.Price) + modifierTotal
			item.TaxRate = 0
		} else {
			if item.MenuItemID <= 0 {
				return nil, domain.ErrPricedItemNotFound
			}
			product, productErr := qtx.GetPublicMenuItem(ctx, sqlc.GetPublicMenuItemParams{
				BusinessID: businessID,
				ID:         item.MenuItemID,
			})
			if productErr != nil {
				if errors.Is(productErr, sql.ErrNoRows) {
					return nil, domain.ErrPricedItemNotFound
				}
				return nil, productErr
			}
			item.Name = product.Name
			item.SKU = product.Sku.String
			item.UnitPrice = decimalToFloat(product.SalePrice) + modifierTotal
			item.TaxRate = decimalToFloat(product.TaxRate)
		}
		if item.UnitPrice < 0 {
			return nil, domain.ErrPricedItemNotFound
		}
	}

	return items, nil
}

func (r *Repository) resolveModifierTotal(ctx context.Context, qtx *sqlc.Queries, businessID int64, modifierIDs []int64) (float64, error) {
	if len(modifierIDs) == 0 {
		return 0, nil
	}
	seen := make(map[int64]struct{}, len(modifierIDs))
	for _, id := range modifierIDs {
		if id <= 0 {
			return 0, domain.ErrPricedItemNotFound
		}
		if _, exists := seen[id]; exists {
			return 0, domain.ErrPricedItemNotFound
		}
		seen[id] = struct{}{}
	}

	row, err := qtx.GetPOSModifierTotal(ctx, sqlc.GetPOSModifierTotalParams{
		BusinessID:  businessID,
		ModifierIds: modifierIDs,
	})
	if err != nil {
		return 0, err
	}
	if row.ModifierCount != int64(len(modifierIDs)) {
		return 0, domain.ErrPricedItemNotFound
	}
	return interfaceToFloat(row.PriceDelta), nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, businessID, orderID int64, status domain.OrderStatus) error {
	err := r.queries.UpdateOrderStatus(ctx, sqlc.UpdateOrderStatusParams{
		Status:     sqlc.NullPosOrdersStatus{PosOrdersStatus: sqlc.PosOrdersStatus(status), Valid: true},
		BusinessID: businessID,
		ID:         orderID,
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

	items := make([]domain.OrderItem, 0, len(rows))
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

	total := decimalToFloat(paymentRow.Total.String)
	change, err := domain.ComputeChange(total, input.AmountPaid)
	if err != nil {
		return err
	}
	changeDue := change.Change

	now := time.Now().UTC()

	markResult, err := qtx.MarkOrderPaid(ctx, sqlc.MarkOrderPaidParams{
		PaymentMethod: sqlc.NullPosOrdersPaymentMethod{PosOrdersPaymentMethod: sqlc.PosOrdersPaymentMethod(input.PaymentMethod), Valid: true},
		AmountPaid:    sql.NullString{String: floatToDecimal(input.AmountPaid), Valid: true},
		ChangeDue:     sql.NullString{String: floatToDecimal(changeDue), Valid: true},
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
		err = qtx.FreeTable(ctx, paymentRow.TableID.Int64)
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
	if status != "paid" && status != "refunded" {
		err = domain.ErrOrderNotPaid
		return 0, err
	}
	prevRefunds, err := qtx.SumRefundsForOrder(ctx, sqlc.SumRefundsForOrderParams{
		BusinessID: businessID,
		OrderID:    orderID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum prior refunds", logger.Error(err))
		return 0, err
	}
	if interfaceToFloat(prevRefunds)+amount > decimalToFloat(order.Total.String) {
		err = domain.ErrRefundExceedsTotal
		return 0, err
	}

	res, err := qtx.CreateRefund(ctx, sqlc.CreateRefundParams{
		OwnerID:    businessID,
		BusinessID: businessID,
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
		BusinessID: businessID,
		ID:         orderID,
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

	if _, err := qtx.VoidOrderItem(ctx, sqlc.VoidOrderItemParams{
		ID:      itemID,
		OrderID: orderID,
	}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback void-order-item update", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to cancel order item", logger.Error(err))
		return err
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

	if err := qtx.OccupyTableByID(ctx, sqlc.OccupyTableByIDParams{
		ID:         newTableID,
		BusinessID: businessID,
	}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Ctx(ctx).Error("failed to rollback occupy table", logger.Error(rollbackErr))
		}
		logger.Ctx(ctx).Error("Failed to mark new table occupied during transfer", logger.Error(err))
		return err
	}

	if err := tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit transfer-table transaction", logger.Error(err))
		return err
	}

	return nil
}

// ─── Dashboard ───────────────────────────────────────────────────────────────

func (r *Repository) GetDashboardStats(ctx context.Context, businessID int64) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	salesRow, err := r.queries.GetDashboardSalesToday(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get dashboard stats", logger.Error(err))
		return stats, err
	}

	stats.TotalSales = interfaceToFloat(salesRow.TotalSales)
	stats.OrderCount = int(salesRow.OrderCount)
	stats.AvgOrder = interfaceToFloat(salesRow.AvgOrder)

	turnover, err := r.queries.GetTableTurnover(ctx, businessID)
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

func (r *Repository) SplitOrder(ctx context.Context, businessID, orderID int64, splits [][]int64) ([]int64, error) {
	order, err := r.GetOrder(ctx, businessID, orderID)
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
	newOrderIDs := make([]int64, 0, len(splits))

	for _, splitItemIDs := range splits {
		orderNum, genErr := r.GenerateOrderNumber(ctx, businessID)
		if genErr != nil {
			txErr = genErr
			return nil, genErr
		}

		res, insertErr := qtx.CreateOrder(ctx, sqlc.CreateOrderParams{
			OwnerID:       businessID,
			BusinessID:    businessID,
			SessionID:     nullInt64(order.SessionID),
			TableID:       nullInt64(order.TableID),
			LocationID:    nullInt64(order.LocationID),
			OrderNumber:   orderNum,
			Type:          sqlc.NullPosOrdersType{PosOrdersType: sqlc.PosOrdersType(order.Type), Valid: true},
			CustomerID:    nullInt64(order.CustomerID),
			CustomerName:  nullString(order.CustomerName),
			StaffID:       nullInt64(order.StaffID),
			Subtotal:      sql.NullString{String: "0", Valid: true},
			TaxAmount:     sql.NullString{String: "0", Valid: true},
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

		itemsTotal, totalErr := qtx.GetOrderItemsTotal(ctx, newID)
		if totalErr != nil {
			txErr = totalErr
			return nil, totalErr
		}

		itemsTotalDecimal := nullableDecimal(itemsTotal)
		recalcErr := qtx.UpdateOrderTotals(ctx, sqlc.UpdateOrderTotalsParams{
			Subtotal: itemsTotalDecimal,
			Total:    itemsTotalDecimal,
			ID:       newID,
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
	maxDiscountPct := decimalToFloat(capability.MaxDiscountPct.String)
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
	totalDiscount := decimalToFloat(financials.ExistingDiscount) + discountAmount

	newTotal := subtotal + taxAmount - totalDiscount
	if newTotal < 0 {
		newTotal = 0
	}

	err = qtx.ApplyDiscount(ctx, sqlc.ApplyDiscountParams{
		DiscountAmount: sql.NullString{String: floatToDecimal(totalDiscount), Valid: true},
		Total:          sql.NullString{String: floatToDecimal(newTotal), Valid: true},
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

func (r *Repository) VoidOrder(ctx context.Context, businessID, orderID int64, reason string) error {
	result, err := r.queries.VoidOrder(ctx, sqlc.VoidOrderParams{
		CONCAT:     fmt.Sprintf(" [Voided: %s]", reason),
		ID:         orderID,
		BusinessID: businessID,
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

func (r *Repository) GenerateOrderNumber(ctx context.Context, businessID int64) (string, error) {
	today := time.Now().Format("20060102")

	count, err := r.queries.CountTodaysOrders(ctx, businessID)
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
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse decimal from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

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
	res, err := r.queries.MarkOrderPaidByPaddleTransaction(ctx, sql.NullString{String: txID, Valid: true})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to mark order paid by paddle tx", logger.Error(err))
		return 0, err
	}
	return res.RowsAffected()
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
		UnitPrice:  decimalToFloat(row.SalePrice),
		TaxRate:    decimalToFloat(row.TaxRate),
	}
	return item, nil
}

func (r *Repository) CreatePublicOrder(ctx context.Context, businessID int64, orderNumber, customerName, customerPhone, notes string, items []domain.OrderItemInput) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin public order transaction", logger.Error(err))
		return 0, err
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
	res, err := qtx.CreatePublicOrder(ctx, sqlc.CreatePublicOrderParams{
		OwnerID:       businessID,
		BusinessID:    businessID,
		OrderNumber:   orderNumber,
		CustomerID:    sql.NullInt64{},
		CustomerName:  nullString(customerName),
		CustomerPhone: nullString(customerPhone),
		Subtotal:      sql.NullString{String: floatToDecimal(totals.Subtotal), Valid: true},
		TaxAmount:     sql.NullString{String: floatToDecimal(totals.TaxAmount), Valid: true},
		Total:         sql.NullString{String: floatToDecimal(totals.Total), Valid: true},
		Notes:         nullString(notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert public order", logger.Error(err))
		return 0, err
	}

	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		itemTotal := item.UnitPrice * item.Quantity
		if err := qtx.CreatePublicOrderItem(ctx, sqlc.CreatePublicOrderItemParams{
			OrderID:    orderID,
			MenuItemID: sql.NullInt64{Int64: item.MenuItemID, Valid: item.MenuItemID != 0},
			Name:       item.Name,
			Quantity:   floatToDecimal3(item.Quantity),
			UnitPrice:  floatToDecimal(item.UnitPrice),
			Total:      floatToDecimal(itemTotal),
			Notes:      nullString(item.Notes),
		}); err != nil {
			logger.Ctx(ctx).Error("Failed to insert public order item", logger.Error(err))
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return orderID, nil
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
