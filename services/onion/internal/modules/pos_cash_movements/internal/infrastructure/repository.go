package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_cash_movements/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) ListBySession(ctx context.Context, businessID, sessionID int64) ([]domain.CashMovement, error) {
	rows, err := r.queries.ListCashMovementsBySession(ctx, sqlc.ListCashMovementsBySessionParams{
		BusinessID: businessID,
		SessionID:  sessionID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list cash movements", logger.Error(err))
		return nil, err
	}

	movements := make([]domain.CashMovement, 0, len(rows))
	for _, row := range rows {
		m := domain.CashMovement{
			ID:         row.ID,
			BusinessID: row.BusinessID,
			BranchID:   row.BranchID,
			SessionID:  row.SessionID,
			EmployeeID: row.EmployeeID,
			Type:       string(row.Type),
			Amount:     sqlutil.DecimalToFloat(row.Amount),
			CreatedAt:  row.CreatedAt.Time,
		}
		if row.TerminalID.Valid {
			m.TerminalID = &row.TerminalID.Int64
		}
		if row.Reason.Valid {
			m.Reason = &row.Reason.String
		}
		if row.ApprovedBy.Valid {
			m.ApprovedBy = &row.ApprovedBy.Int64
		}
		movements = append(movements, m)
	}
	return movements, nil
}

func (r *Repository) Create(ctx context.Context, input domain.CreateCashMovementInput) (int64, error) {
	var sessionOK int
	if err := r.pool.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM pos_sessions
		WHERE id = ? AND business_id = ? AND staff_id = ? AND status = 'open')`, input.SessionID, input.BusinessID, input.EmployeeID).Scan(&sessionOK); err != nil {
		return 0, err
	}
	if sessionOK != 1 {
		return 0, fmt.Errorf("cash movement requires the employee's open session")
	}
	var termID sql.NullInt64
	if input.TerminalID != nil {
		termID = sql.NullInt64{Int64: *input.TerminalID, Valid: true}
	}
	var reason sql.NullString
	if input.Reason != nil {
		reason = sql.NullString{String: *input.Reason, Valid: true}
	}
	var approvedBy sql.NullInt64
	if input.ApprovedBy != nil {
		approvedBy = sql.NullInt64{Int64: *input.ApprovedBy, Valid: true}
	}

	result, err := r.queries.CreateCashMovement(ctx, sqlc.CreateCashMovementParams{
		BusinessID: input.BusinessID,
		BranchID:   input.BranchID,
		TerminalID: termID,
		SessionID:  input.SessionID,
		EmployeeID: input.EmployeeID,
		Type:       sqlc.PosCashMovementsType(input.Type),
		Amount:     sqlutil.FloatToDecimal(input.Amount),
		Reason:     reason,
		ApprovedBy: approvedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create cash movement", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}
