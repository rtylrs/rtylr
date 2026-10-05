package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	reportingCurrency func(context.Context, int64) (string, error)
	pool              *sql.DB
	queries           *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
		pool:    db,
	}
}

func (r *Repository) SumAccountsByType(ctx context.Context, businessID int64, accountType string) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result, err := r.queries.SumAccountsByType(ctx, sqlc.SumAccountsByTypeParams{
		ReportingCurrency: sql.NullString{String: currency, Valid: true},
		BusinessID:        businessID,
		AccountType:       sqlc.FinanceAccountsType(accountType),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum accounts by type", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumPostedNetIncome(ctx context.Context, businessID int64) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result, err := r.queries.SumPostedNetIncome(ctx, sqlc.SumPostedNetIncomeParams{BusinessID: businessID, ReportingCurrency: sql.NullString{String: currency, Valid: true}})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum posted net income", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumBankBalances(ctx context.Context, businessID int64) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result, err := r.queries.SumBankBalances(ctx, sqlc.SumBankBalancesParams{BusinessID: businessID, ReportingCurrency: sql.NullString{String: currency, Valid: true}})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum bank balances", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumReceivables(ctx context.Context, businessID int64) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result, err := r.queries.SumReceivables(ctx, sqlc.SumReceivablesParams{BusinessID: businessID, ReportingCurrency: sql.NullString{String: currency, Valid: true}})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum receivables", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumPayables(ctx context.Context, businessID int64) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result, err := r.queries.SumPayables(ctx, sqlc.SumPayablesParams{BusinessID: businessID, ReportingCurrency: sql.NullString{String: currency, Valid: true}})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum payables", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumPaidInvoicesInRange(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	result, err := r.queries.SumPaidInvoicesInRange(ctx, sqlc.SumPaidInvoicesInRangeParams{
		ReportingCurrency: sql.NullString{String: currency, Valid: true},
		BusinessID:        businessID,
		FromPaidDate:      sqlNullTimeStr(dateFrom),
		ToPaidDate:        sqlNullTimeStr(dateTo),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum paid invoices in range", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumInvoicePaymentsInRange(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	query := `
 SELECT COALESCE(SUM(p.amount), 0)
		FROM finance_payments p
		WHERE p.business_id = ?
		  AND COALESCE(NULLIF(p.currency, ''), 'USD') = ?`
	args := []any{businessID, currency}

	if dateFrom != "" {
		from, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			return 0, fmt.Errorf("parse payment range start %q: %w", dateFrom, err)
		}
		query += " AND p.payment_date >= ?"
		args = append(args, from)
	}
	if dateTo != "" {
		to, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			return 0, fmt.Errorf("parse payment range end %q: %w", dateTo, err)
		}
		query += " AND p.payment_date <= ?"
		args = append(args, to)
	}

	var result interface{}
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&result); err != nil {
		logger.Ctx(ctx).Error("Failed to sum invoice payments in range", logger.Error(err))
		return 0, err
	}
	return toFloat64FromInterface(result), nil
}

func (r *Repository) SumExpensesInRange(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	currency, err := r.reportCurrency(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
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
		ReportingCurrency: sql.NullString{String: currency, Valid: true},
		BusinessID:        businessID,
		FromExpenseDate:   dfTime,
		ToExpenseDate:     dtTime,
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

func (r *Repository) WithReportingCurrency(read func(context.Context, int64) (string, error)) *Repository {
	r.reportingCurrency = read
	return r
}

func (r *Repository) reportCurrency(ctx context.Context, businessID int64) (string, error) {
	if r.reportingCurrency == nil {
		return "", errors.New("finance reporting currency reader is not configured")
	}
	return r.reportingCurrency(ctx, businessID)
}
