package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool    *sql.DB
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) SumAccountsByType(ctx context.Context, businessID int64, accountType string) (float64, error) {
	result, err := r.queries.SumAccountsByType(ctx, sqlc.SumAccountsByTypeParams{
		BusinessID:  businessID,
		AccountType: sqlc.FinanceAccountsType(accountType),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum accounts by type", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumPostedNetIncome(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumPostedNetIncome(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum posted net income", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumBankBalances(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumBankBalances(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank balances", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumReceivables(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumReceivables(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum receivables", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumPayables(ctx context.Context, businessID int64) (float64, error) {
	result, err := r.queries.SumPayables(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum payables", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumPaidInvoicesInRange(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	result, err := r.queries.SumPaidInvoicesInRange(ctx, sqlc.SumPaidInvoicesInRangeParams{
		BusinessID:   businessID,
		FromPaidDate: sqlNullTimeStr(dateFrom),
		ToPaidDate:   sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum paid invoices in range", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumExpensesInRange(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	dfTime, err := time.Parse("2006-01-02", dateFrom)
	if err != nil && dateFrom != "" {
		logger.Ctx(ctx).Error("Failed to parse dateFrom", logger.Error(err))
		return 0, err
	}
	dtTime, err := time.Parse("2006-01-02", dateTo)
	if err != nil && dateTo != "" {
		logger.Ctx(ctx).Error("Failed to parse dateTo", logger.Error(err))
		return 0, err
	}

	result, err := r.queries.SumExpensesInRange(ctx, sqlc.SumExpensesInRangeParams{
		BusinessID:      businessID,
		FromExpenseDate: dfTime,
		ToExpenseDate:   dtTime,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum expenses in range", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

// NetProfit computes revenue minus expenses for the given date range. A
// positive value means a profit; a negative value means a loss.
func (r *Repository) NetProfit(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	revenue, err := r.SumPaidInvoicesInRange(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return 0, err
	}

	expenses, err := r.SumExpensesInRange(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return 0, err
	}

	return revenue - expenses, nil
}

// WorkingCapital returns current assets (bank balances + receivables) minus
// current liabilities (payables). A positive value indicates the owner can
// meet its short-term obligations.
func (r *Repository) WorkingCapital(ctx context.Context, businessID int64) (float64, error) {
	bankBal, err := r.SumBankBalances(ctx, businessID)
	if err != nil {
		return 0, err
	}

	receivables, err := r.SumReceivables(ctx, businessID)
	if err != nil {
		return 0, err
	}

	payables, err := r.SumPayables(ctx, businessID)
	if err != nil {
		return 0, err
	}

	return (bankBal + receivables) - payables, nil
}

// ---------------------------------------------------------------------------
// Conversion helpers
// ---------------------------------------------------------------------------

func sqlNullTimeStr(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func toFloat64FromInterface(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return parseReportDecimal(string(val))
	case string:
		return parseReportDecimal(val)
	default:
		return parseReportDecimal(fmt.Sprintf("%v", val))
	}
}

func parseReportDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse finance-report decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
