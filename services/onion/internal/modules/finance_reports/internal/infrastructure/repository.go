package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) SumAccountsByType(ctx context.Context, ownerID int64, accountType string) (float64, error) {
	var total sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM finance_accounts WHERE owner_id = ? AND type = ? AND is_active = TRUE`,
		ownerID, accountType,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum accounts by type", logger.Error(err))
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *Repository) SumBankBalances(ctx context.Context, ownerID int64) (float64, error) {
	var total sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM finance_bank_accounts WHERE owner_id = ?`,
		ownerID,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank balances", logger.Error(err))
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *Repository) SumReceivables(ctx context.Context, ownerID int64) (float64, error) {
	var total sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount_due), 0) FROM finance_invoices WHERE owner_id = ? AND status NOT IN ('paid','cancelled','void')`,
		ownerID,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum receivables", logger.Error(err))
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *Repository) SumPayables(ctx context.Context, ownerID int64) (float64, error) {
	var total sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM finance_bills WHERE owner_id = ? AND status = 'pending'`,
		ownerID,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum payables", logger.Error(err))
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *Repository) SumPaidInvoicesInRange(ctx context.Context, ownerID int64, dateFrom, dateTo string) (float64, error) {
	var total sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total), 0) FROM finance_invoices
		 WHERE owner_id = ? AND status = 'paid' AND paid_date BETWEEN ? AND ?`,
		ownerID, dateFrom, dateTo,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum paid invoices in range", logger.Error(err))
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}

func (r *Repository) SumExpensesInRange(ctx context.Context, ownerID int64, dateFrom, dateTo string) (float64, error) {
	var total sql.NullFloat64
	err := r.pool.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM finance_expenses
		 WHERE owner_id = ? AND expense_date BETWEEN ? AND ?`,
		ownerID, dateFrom, dateTo,
	).Scan(&total)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses in range", logger.Error(err))
		return 0, err
	}
	if total.Valid {
		return total.Float64, nil
	}
	return 0, nil
}
