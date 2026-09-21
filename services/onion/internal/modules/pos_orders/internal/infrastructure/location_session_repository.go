package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
)

// Locations
// ─────────────────────────────────────────────────────────────────────────────

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
		Address:    sqlutil.NullString(input.Address),
		Phone:      sqlutil.NullString(input.Phone),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos location", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// Tables
// ─────────────────────────────────────────────────────────────────────────────

func (r *Repository) ListTables(ctx context.Context, businessID, locationID int64, status string) ([]domain.Table, error) {
	rows, err := r.queries.ListTablesByOwner(ctx, sqlc.ListTablesByOwnerParams{
		BusinessID: businessID,
		LocationID: sql.NullInt64{Int64: locationID, Valid: true},
		Status:     sqlc.NullPosTablesStatus{PosTablesStatus: sqlc.PosTablesStatus(status), Valid: true},
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
		LocationID: sqlutil.NullInt64(input.LocationID),
		Name:       input.Name,
		Capacity:   sql.NullInt32{Int32: int32(input.Capacity), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		QrCodeUrl:  sqlutil.NullString(input.QRCodeURL),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create pos table", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateTable(ctx context.Context, businessID int64, input domain.TableInput) error {
	err := r.queries.UpdatePosTable(ctx, sqlc.UpdatePosTableParams{
		LocationID: sqlutil.NullInt64(input.LocationID),
		Name:       input.Name,
		Capacity:   sql.NullInt32{Int32: int32(input.Capacity), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		QrCodeUrl:  sqlutil.NullString(input.QRCodeURL),
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

// Sessions
// ─────────────────────────────────────────────────────────────────────────────

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
		OpeningCash:      sqlutil.DecimalToFloat(row.OpeningCash.String),
		ClosingCash:      sqlutil.DecimalToFloat(row.ClosingCash.String),
		ExpectedCash:     sqlutil.DecimalToFloat(row.ExpectedCash.String),
		CashDifference:   sqlutil.DecimalToFloat(row.CashDifference.String),
		TotalSales:       sqlutil.DecimalToFloat(row.TotalSales.String),
		TotalOrders:      int(row.TotalOrders.Int32),
		TotalCashSales:   sqlutil.DecimalToFloat(row.TotalCashSales.String),
		TotalCardSales:   sqlutil.DecimalToFloat(row.TotalCardSales.String),
		TotalOnlineSales: sqlutil.DecimalToFloat(row.TotalOnlineSales.String),
		TotalRefunds:     sqlutil.DecimalToFloat(row.TotalRefunds.String),
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
		// date_to is a calendar day and dateFrom already includes the whole
		// start day; comparing the DATETIME column with a date-only string
		// would cut the end day off at 00:00:00. Use the exclusive next-day
		// boundary, matching applyPOSOrderListFilters.
		builder = builder.Where("opened_at < DATE_ADD(?, INTERVAL 1 DAY)", dateTo)
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
	// MySQL named locks are connection-scoped: COMMIT/ROLLBACK do not release
	// them, and database/sql refuses to run statements on a finished Tx, so a
	// deferred RELEASE_LOCK after Commit never reaches the server and leaks
	// the lock onto the pooled connection. Keep a deferred best-effort release
	// for the early-return paths and release explicitly (with the result
	// checked) while the transaction is still open on the success path below.
	defer func() {
		_ = releaseNamedLock(ctx, tx, lockName)
	}()
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pos_sessions WHERE business_id = ? AND staff_id = ? AND status = 'open')`, businessID, staffID).Scan(&exists); err != nil {
		return 0, err
	}
	if exists == 1 {
		return 0, domain.ErrSessionAlreadyOpen
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO pos_sessions (owner_id, business_id, location_id, staff_id, status, opening_cash)
		VALUES (?, ?, ?, ?, 'open', ?)`, businessID, businessID, locationID, staffID, sqlutil.FloatToDecimal(openingCash))
	if err != nil {
		logger.Ctx(ctx).Error("Failed to open pos session", logger.Error(err))
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := releaseNamedLock(ctx, tx, lockName); err != nil {
		logger.Ctx(ctx).Error("Failed to release session lock", logger.Error(err))
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// releaseNamedLock releases a MySQL named lock held by tx. It must be called
// while tx is still open: MySQL releases named locks at connection close, not
// at COMMIT/ROLLBACK, and a finished *sql.Tx returns ErrTxDone without sending
// the statement, which would leak the lock on the pooled connection.
func releaseNamedLock(ctx context.Context, tx *sql.Tx, name string) error {
	var released sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT RELEASE_LOCK(?)`, name).Scan(&released); err != nil {
		return err
	}
	if !released.Valid || released.Int64 != 1 {
		return fmt.Errorf("mysql lock %q was not held by this connection", name)
	}
	return nil
}

// CloseSessionForStaff closes the session and writes its aggregate totals in
// one UPDATE. The totals are computed before the call from the session's
// orders (GetSessionTotals), so folding them in keeps the close atomic: a
// failure leaves the session open and retryable instead of reporting an error
// for a session that is already closed but carries stale totals.
func (r *Repository) CloseSessionForStaff(ctx context.Context, businessID, staffID, sessionID int64, closingCash, expectedCash, cashDiff float64, notes string, totalSales, cashSales, cardSales, onlineSales float64, totalOrders int) error {
	result, err := r.pool.ExecContext(ctx, `
		UPDATE pos_sessions
		SET status = 'closed', closing_cash = ?, expected_cash = ?, cash_difference = ?, closed_at = NOW(), notes = ?,
		    total_sales = ?, total_cash_sales = ?, total_card_sales = ?, total_online_sales = ?, total_orders = ?
		WHERE business_id = ? AND id = ? AND staff_id = ? AND status = 'open'`,
		sqlutil.FloatToDecimal(closingCash), sqlutil.FloatToDecimal(expectedCash), sqlutil.FloatToDecimal(cashDiff), notes,
		sqlutil.FloatToDecimal(totalSales), sqlutil.FloatToDecimal(cashSales), sqlutil.FloatToDecimal(cardSales),
		sqlutil.FloatToDecimal(onlineSales), totalOrders, businessID, sessionID, staffID)
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
		ClosingCash:    sql.NullString{String: sqlutil.FloatToDecimal(closingCash), Valid: true},
		ExpectedCash:   sql.NullString{String: sqlutil.FloatToDecimal(expectedCash), Valid: true},
		CashDifference: sql.NullString{String: sqlutil.FloatToDecimal(cashDiff), Valid: true},
		Notes:          sqlutil.NullString(notes),
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
		TotalSales:       sql.NullString{String: sqlutil.FloatToDecimal(totalSales), Valid: true},
		TotalCashSales:   sql.NullString{String: sqlutil.FloatToDecimal(cashSales), Valid: true},
		TotalCardSales:   sql.NullString{String: sqlutil.FloatToDecimal(cardSales), Valid: true},
		TotalOnlineSales: sql.NullString{String: sqlutil.FloatToDecimal(onlineSales), Valid: true},
		TotalOrders:      sql.NullInt32{Int32: int32(totalOrders), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
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
		OpeningCash:      sqlutil.DecimalToFloat(row.OpeningCash.String),
		ClosingCash:      sqlutil.DecimalToFloat(row.ClosingCash.String),
		ExpectedCash:     sqlutil.DecimalToFloat(row.ExpectedCash.String),
		CashDifference:   sqlutil.DecimalToFloat(row.CashDifference.String),
		TotalSales:       sqlutil.DecimalToFloat(row.TotalSales.String),
		TotalOrders:      int(row.TotalOrders.Int32),
		TotalCashSales:   sqlutil.DecimalToFloat(row.TotalCashSales.String),
		TotalCardSales:   sqlutil.DecimalToFloat(row.TotalCardSales.String),
		TotalOnlineSales: sqlutil.DecimalToFloat(row.TotalOnlineSales.String),
		TotalRefunds:     sqlutil.DecimalToFloat(row.TotalRefunds.String),
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
		TotalRefunds: sql.NullString{String: sqlutil.FloatToDecimal(refundAmount), Valid: true},
		ID:           sessionID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update session refund totals", logger.Error(err))
		return err
	}
	return nil
}
