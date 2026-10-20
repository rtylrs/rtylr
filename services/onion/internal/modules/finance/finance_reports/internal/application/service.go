package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_reports/internal/domain"
	"github.com/voxire/onion/internal/platform/financejournal"
	"github.com/voxire/onion/internal/platform/logger"
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

// balanceSheetLineRepository reads the ledger account by account. Without it
// (narrow test doubles) the balance sheet carries totals only.
type balanceSheetLineRepository interface {
	ListBalanceSheetAccounts(ctx context.Context, businessID int64, asOf time.Time) ([]domain.BalanceSheetLine, error)
	SumPostedNetIncomeAsOf(ctx context.Context, businessID int64, asOf time.Time) (float64, error)
}

type invoicePaymentRepository interface {
	SumInvoicePaymentsInRange(ctx context.Context, businessID int64, from, to string) (float64, error)
}

type Service struct {
	repo   repository
	ledger financejournal.LedgerSyncer
}

// SetLedgerSyncer makes the balance sheet post outstanding expenses and
// invoice payments to the journal before reading it.
func (s *Service) SetLedgerSyncer(ledger financejournal.LedgerSyncer) {
	s.ledger = ledger
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetBalanceSheet(ctx context.Context, businessID int64) (domain.BalanceSheet, error) {
	return s.GetBalanceSheetAsOf(ctx, businessID, "")
}

// GetBalanceSheetAsOf reports the ledger on asOf (YYYY-MM-DD, today when
// empty). Approved expenses and invoice payments are posted first, so the
// statement includes everything the operational screens show.
func (s *Service) GetBalanceSheetAsOf(ctx context.Context, businessID int64, asOf string) (domain.BalanceSheet, error) {
	date := time.Now().UTC()
	if asOf != "" {
		parsed, err := time.Parse("2006-01-02", asOf)
		if err != nil {
			return domain.BalanceSheet{}, ErrInvalidAsOf
		}
		date = parsed
	}
	if s.ledger != nil {
		if err := s.ledger.SyncOperationalPostings(ctx, businessID); err != nil {
			logger.Ctx(ctx).Warn("Ledger sync before the balance sheet failed", logger.Error(err))
		}
	}
	bs, err := s.computeBalanceSheet(ctx, businessID)
	if err != nil {
		return domain.BalanceSheet{}, err
	}
	bs.AsOf = date.Format("2006-01-02")
	lines, ok := s.repo.(balanceSheetLineRepository)
	if !ok {
		return bs, nil
	}
	accounts, err := lines.ListBalanceSheetAccounts(ctx, businessID, date)
	if err != nil {
		return domain.BalanceSheet{}, err
	}
	netIncome, err := lines.SumPostedNetIncomeAsOf(ctx, businessID, date)
	if err != nil {
		return domain.BalanceSheet{}, err
	}
	// The totals are the sums of the lines shown, so the statement adds up on
	// the page; they used to be ledger totals beside rows from other tables.
	bs.Lines = bs.Lines[:0]
	bs.TotalAssets, bs.TotalLiabilities, bs.TotalEquity = 0, 0, netIncome
	for _, line := range accounts {
		if line.Amount == 0 {
			continue
		}
		bs.Lines = append(bs.Lines, line)
		switch line.Type {
		case "asset":
			bs.TotalAssets += line.Amount
		case "liability":
			bs.TotalLiabilities += line.Amount
		case "equity":
			bs.TotalEquity += line.Amount
		}
	}
	bs.NetIncome = netIncome
	return bs, nil
}

// ErrInvalidAsOf reports an as-of date that is not YYYY-MM-DD.
var ErrInvalidAsOf = errors.New("as_of must be a date in YYYY-MM-DD format")

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
