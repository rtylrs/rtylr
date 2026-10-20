package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/finance/journal/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/financeaccounts"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	journalScope = "journal"
	journalTTL   = 5 * time.Minute
)

type repository interface {
	List(ctx context.Context, businessID int64, limit, offset int32) ([]domain.JournalEntry, error)
	Get(ctx context.Context, businessID, id int64) (domain.JournalEntry, error)
	Create(ctx context.Context, ownerID, businessID int64, input domain.JournalEntryInput) (int64, error)
	Post(ctx context.Context, businessID, id int64) error
	PostPayrollJournal(ctx context.Context, input domain.PayrollJournalInput) error
	SetPayrollAccountProvider(financeaccounts.PayrollAccountProvider)
	ListUnpostedSources(ctx context.Context, businessID int64) ([]domain.OperationalSource, error)
	PostOperationalEntry(ctx context.Context, src domain.OperationalSource, debitAccountID, creditAccountID int64) (bool, error)
}

func (s *Service) SetPayrollAccountProvider(provider financeaccounts.PayrollAccountProvider) {
	s.repo.SetPayrollAccountProvider(provider)
}

type Service struct {
	repo           repository
	ledgerAccounts financeaccounts.LedgerAccountProvider
}

func (s *Service) SetLedgerAccountProvider(provider financeaccounts.LedgerAccountProvider) {
	s.ledgerAccounts = provider
}

// SyncOperationalPostings posts every approved expense and recorded invoice
// payment that is not in the journal yet, as a balanced two-line entry:
//
//	expense:          Dr Operating expenses   Cr Cash or Bank
//	invoice payment:  Dr Cash or Bank         Cr Sales
//
// Cash is used when the payment method is cash, Bank otherwise. Before this,
// nothing but payroll and manual entries reached the journal, so the balance
// sheet, journal and account balances never saw an expense or a payment.
// Each source posts at most once, so callers run it before every ledger read.
func (s *Service) SyncOperationalPostings(ctx context.Context, businessID int64) error {
	if s.ledgerAccounts == nil {
		return nil
	}
	sources, err := s.repo.ListUnpostedSources(ctx, businessID)
	if err != nil || len(sources) == 0 {
		return err
	}
	accountsByCurrency := map[string]financeaccounts.LedgerAccounts{}
	posted := false
	for _, src := range sources {
		accounts, ok := accountsByCurrency[src.Currency]
		if !ok {
			// Accounts follow the module convention of owner_id = business id.
			accounts, err = s.ledgerAccounts.EnsureLedgerAccounts(ctx, businessID, businessID, src.Currency)
			if err != nil {
				return err
			}
			accountsByCurrency[src.Currency] = accounts
		}
		moneyAccount := accounts.Bank
		if src.PaidInCash {
			moneyAccount = accounts.Cash
		}
		debit, credit := accounts.OperatingExpenses, moneyAccount
		if src.Kind == domain.SourceInvoicePayment {
			debit, credit = moneyAccount, accounts.Sales
		}
		ok, err := s.repo.PostOperationalEntry(ctx, src, debit, credit)
		if err != nil {
			return err
		}
		posted = posted || ok
	}
	if posted {
		invalidateJournalScope(ctx, businessID)
	}
	return nil
}

func NewService(_ context.Context, repo repository) *Service {
	return &Service{repo: repo}
}

func invalidateJournalScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(journalScope, businessID))
}

func (s *Service) List(ctx context.Context, businessID, page, pageSize int64) ([]domain.JournalEntry, bool, error) {
	// A failed sync still lists what is posted; the next read retries it.
	if err := s.SyncOperationalPostings(ctx, businessID); err != nil {
		logger.Ctx(ctx).Warn("Ledger sync before listing the journal failed", logger.Error(err))
	}
	limit, offset := paginate(page, pageSize)
	key := cacheKeys.Scoped(journalScope, businessID, fmt.Sprintf("list|p=%d|ps=%d", page, pageSize))
	rows, err := cacher.Of[[]domain.JournalEntry]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) ([]domain.JournalEntry, error) {
			return s.repo.List(ctx, businessID, int32(limit+1), int32(offset)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		},
		cacher.WithExpiry(ctx, journalTTL),
	)
	if err != nil {
		return nil, false, err
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, nil
}

func (s *Service) Get(ctx context.Context, businessID, id int64) (domain.JournalEntry, error) {
	return s.repo.Get(ctx, businessID, id)
}

func paginate(page, pageSize int64) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	return int(pageSize), int((page - 1) * pageSize)
}

func (s *Service) Create(ctx context.Context, ownerID, businessID int64, input domain.JournalEntryInput) (int64, error) {
	if err := domain.ValidateJournalEntryInput(input); err != nil {
		return 0, fmt.Errorf("%w: %w", domain.ErrInvalidJournalEntry, err)
	}

	id, err := s.repo.Create(ctx, ownerID, businessID, input)
	if err != nil {
		return 0, err
	}
	invalidateJournalScope(ctx, businessID)
	return id, nil
}

func (s *Service) Post(ctx context.Context, businessID, id int64) error {
	if err := s.repo.Post(ctx, businessID, id); err != nil {
		return err
	}
	invalidateJournalScope(ctx, businessID)
	return nil
}

func (s *Service) PostPayrollJournal(ctx context.Context, input domain.PayrollJournalInput) error {
	if err := validatePayrollJournalInput(input); err != nil {
		return err
	}
	if err := s.repo.PostPayrollJournal(ctx, input); err != nil {
		return err
	}
	invalidateJournalScope(ctx, input.BusinessID)
	return nil
}

func validatePayrollJournalInput(input domain.PayrollJournalInput) error {
	if input.OwnerID <= 0 || input.BusinessID <= 0 || input.RunID <= 0 {
		return fmt.Errorf("payroll journal owner, business, and run IDs must be positive")
	}
	if input.TotalGross <= 0 {
		return fmt.Errorf("payroll journal gross total must be positive")
	}
	if input.TotalNet <= 0 || input.TotalNet > input.TotalGross {
		return fmt.Errorf("payroll journal net total must be positive and no greater than gross total")
	}
	if _, err := time.Parse("2006-01-02", input.PeriodStart); err != nil {
		return fmt.Errorf("payroll journal period start must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", input.PeriodEnd); err != nil {
		return fmt.Errorf("payroll journal period end must be in YYYY-MM-DD format")
	}
	return nil
}
