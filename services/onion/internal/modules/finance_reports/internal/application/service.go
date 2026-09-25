package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reports/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_reports/internal/infrastructure"
)

type repository interface {
	SumAccountsByType(ctx context.Context, businessID int64, accountType string) (float64, error)
	SumPostedNetIncome(ctx context.Context, businessID int64) (float64, error)
	SumBankBalances(ctx context.Context, businessID int64) (float64, error)
	SumReceivables(ctx context.Context, businessID int64) (float64, error)
	SumPayables(ctx context.Context, businessID int64) (float64, error)
	SumPaidInvoicesInRange(ctx context.Context, businessID int64, from, to string) (float64, error)
	SumExpensesInRange(ctx context.Context, businessID int64, from, to string) (float64, error)
}

type invoicePaymentRepository interface {
	SumInvoicePaymentsInRange(ctx context.Context, businessID int64, from, to string) (float64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) GetBalanceSheet(ctx context.Context, businessID int64) (domain.BalanceSheet, error) {
	return s.computeBalanceSheet(ctx, businessID)
}

func (s *Service) computeBalanceSheet(ctx context.Context, businessID int64) (domain.BalanceSheet, error) {
	assets, err := s.repo.SumAccountsByType(ctx, businessID, "asset")
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	liabilities, err := s.repo.SumAccountsByType(ctx, businessID, "liability")
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	equity, err := s.repo.SumAccountsByType(ctx, businessID, "equity")
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	netIncome, err := s.repo.SumPostedNetIncome(ctx, businessID)
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	cash, err := s.repo.SumBankBalances(ctx, businessID)
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	receivables, err := s.repo.SumReceivables(ctx, businessID)
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	payables, err := s.repo.SumPayables(ctx, businessID)
	if err != nil {
		return domain.BalanceSheet{}, err
	}

	return domain.BalanceSheet{
		// Account totals combine the configured opening balance with posted
		// journal deltas. The bank/invoice/bill tables are detail views and
		// must not be added to the same totals a second time.
		TotalAssets:        assets,
		TotalLiabilities:   liabilities,
		TotalEquity:        equity + netIncome,
		CashBalance:        cash,
		AccountsReceivable: receivables,
		AccountsPayable:    payables,
	}, nil
}

func (s *Service) GetCashFlow(ctx context.Context, businessID int64, period string) (domain.CashFlow, error) {
	return s.computeCashFlow(ctx, businessID, period)
}

func (s *Service) computeCashFlow(ctx context.Context, businessID int64, period string) (domain.CashFlow, error) {
	dateFrom, dateTo := resolvePeriodRange(period)

	inflow, err := s.sumInvoicePaymentsInRange(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return domain.CashFlow{}, err
	}

	outflow, err := s.repo.SumExpensesInRange(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return domain.CashFlow{}, err
	}

	netOperating := inflow - outflow

	return domain.CashFlow{
		Period:           period,
		OperatingInflow:  inflow,
		OperatingOutflow: outflow,
		NetOperating:     netOperating,
		NetCashFlow:      netOperating,
	}, nil
}

func (s *Service) sumInvoicePaymentsInRange(ctx context.Context, businessID int64, dateFrom, dateTo string) (float64, error) {
	if repo, ok := s.repo.(invoicePaymentRepository); ok {
		return repo.SumInvoicePaymentsInRange(ctx, businessID, dateFrom, dateTo)
	}
	return s.repo.SumPaidInvoicesInRange(ctx, businessID, dateFrom, dateTo)
}

func resolvePeriodRange(period string) (string, string) {
	now := time.Now().UTC()

	switch period {
	case "quarter":
		quarter := (int(now.Month()) - 1) / 3
		fromMonth := time.Month(quarter*3 + 1)
		from := time.Date(now.Year(), fromMonth, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 3, -1)
		return from.Format("2006-01-02"), to.Format("2006-01-02")
	case "year":
		from := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(now.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
		return from.Format("2006-01-02"), to.Format("2006-01-02")
	default:
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, -1)
		return from.Format("2006-01-02"), to.Format("2006-01-02")
	}
}
