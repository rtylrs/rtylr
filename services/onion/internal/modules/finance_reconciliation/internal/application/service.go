package application

import (
	"context"
	"database/sql"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/domain"
	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	financeReconScope = "finance_reconciliation"
	financeReconTTL   = 2 * time.Minute
)

// maxSuggestions caps the suggestion list so the response stays small even
// when both sides have hundreds of rows.
const maxSuggestions = 50

// reconciliationPreviewLimit keeps the overview and its suggestion matrix
// bounded. Truncation is explicit in the response so callers never mistake a
// preview for a complete unmatched-record list.
const reconciliationPreviewLimit = 200

// suggestionScoreCutoff discards pairings whose combined amount + date
// distance is too large to be worth showing. The score below is "amount
// difference in minor units + (days_apart * dayWeightMinorUnits)", so this
// cutoff is generous for a one-dollar, one-week mismatch and tight enough to
// reject random noise.
const (
	suggestionScoreCutoff = 500000.0 // $5000 equivalent
	dayWeightMinorUnits   = 100.0    // one day ≈ $1 of penalty
)

type repository interface {
	ListUnmatchedBankFeeds(ctx context.Context, businessID int64) ([]domain.UnmatchedBankTransaction, error)
	ListUnmatchedInvoices(ctx context.Context, businessID int64) ([]domain.UnmatchedInvoice, error)
	ListUnmatchedExpenses(ctx context.Context, businessID int64) ([]domain.UnmatchedInvoice, error)
	BankFeedExists(ctx context.Context, businessID, id int64) (bool, error)
	TargetExists(ctx context.Context, businessID int64, input domain.MatchInput) (bool, error)
	CreateMatch(ctx context.Context, businessID int64, input domain.MatchInput) (int64, error)
}

type Service struct {
	repo repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

func (s *Service) Overview(ctx context.Context, businessID int64) (domain.ReconciliationOverview, error) {
	return cacher.Of[domain.ReconciliationOverview]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(financeReconScope, businessID, "overview"),
		func(ctx context.Context) (domain.ReconciliationOverview, error) {
			return s.computeOverview(ctx, businessID)
		},
		cacher.WithExpiry(ctx, financeReconTTL),
	)
}

func (s *Service) computeOverview(ctx context.Context, businessID int64) (domain.ReconciliationOverview, error) {
	bankTxns, err := s.repo.ListUnmatchedBankFeeds(ctx, businessID)
	if err != nil {
		return domain.ReconciliationOverview{}, err
	}

	invoices, err := s.repo.ListUnmatchedInvoices(ctx, businessID)
	if err != nil {
		return domain.ReconciliationOverview{}, err
	}

	expenses, err := s.repo.ListUnmatchedExpenses(ctx, businessID)
	if err != nil {
		return domain.ReconciliationOverview{}, err
	}

	bankTxns, bankTxnsTruncated := truncatePreview(bankTxns)
	invoices, invoicesTruncated := truncatePreview(invoices)
	expenses, expensesTruncated := truncatePreview(expenses)

	suggestions := computeSuggestions(bankTxns, invoices, expenses)

	return domain.ReconciliationOverview{
		UnmatchedBankTxns: bankTxns,
		UnmatchedInvoices: invoices,
		Suggestions:       suggestions,
		BankTxnsTruncated: bankTxnsTruncated,
		InvoicesTruncated: invoicesTruncated,
		ExpensesTruncated: expensesTruncated,
		PreviewLimit:      reconciliationPreviewLimit,
	}, nil
}

func truncatePreview[T any](items []T) ([]T, bool) {
	if len(items) <= reconciliationPreviewLimit {
		return items, false
	}

	return items[:reconciliationPreviewLimit], true
}

func (s *Service) Match(ctx context.Context, businessID int64, input domain.MatchInput) (int64, error) {
	if input.TargetType != domain.TargetTypeInvoice && input.TargetType != domain.TargetTypeExpense {
		return 0, domain.ErrInvalidTarget
	}
	if input.BankFeedID <= 0 || input.TargetID <= 0 {
		return 0, domain.ErrInvalidTarget
	}

	exists, err := s.repo.BankFeedExists(ctx, businessID, input.BankFeedID)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, domain.ErrBankFeedNotFound
	}
	targetExists, err := s.repo.TargetExists(ctx, businessID, input)
	if err != nil {
		return 0, err
	}
	if !targetExists {
		return 0, domain.ErrTargetNotFound
	}

	id, err := s.repo.CreateMatch(ctx, businessID, input)
	if err == nil {
		cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(financeReconScope, businessID))
	}
	return id, err
}

// computeSuggestions scores every same-currency (bank_txn × target) pair by
// amount and date proximity and returns the top matches. Amounts are compared
// in absolute value so a negative bank amount (money out) still pairs with a
// positive expense amount.
func computeSuggestions(bankTxns []domain.UnmatchedBankTransaction, invoices, expenses []domain.UnmatchedInvoice) []domain.MatchSuggestion {
	if len(bankTxns) == 0 {
		return nil
	}
	type target struct {
		id       int64
		kind     string
		amount   int64
		currency string
		date     int64 // unix day
	}
	targets := make([]target, 0, len(invoices)+len(expenses))
	for _, inv := range invoices {
		targets = append(targets, target{
			id:       inv.ID,
			kind:     domain.TargetTypeInvoice,
			amount:   inv.AmountMinorUnits,
			currency: inv.Currency,
			date:     inv.DueDate.Unix() / 86400,
		})
	}
	for _, exp := range expenses {
		targets = append(targets, target{
			id:       exp.ID,
			kind:     domain.TargetTypeExpense,
			amount:   exp.AmountMinorUnits,
			currency: exp.Currency,
			date:     exp.IssueDate.Unix() / 86400,
		})
	}
	if len(targets) == 0 {
		return nil
	}

	suggestions := make([]domain.MatchSuggestion, 0, len(bankTxns))
	for _, bt := range bankTxns {
		bankCurrency := normalizeCurrency(bt.Currency)
		bankAmt := absInt64(bt.AmountMinorUnits)
		bankDate := bt.OccurredAt.Unix() / 86400
		for _, t := range targets {
			if bankCurrency == "" || bankCurrency != normalizeCurrency(t.currency) {
				continue
			}
			amountDiff := math.Abs(float64(bankAmt - absInt64(t.amount)))
			daysApart := absInt64(bankDate - t.date)
			score := amountDiff + float64(daysApart)*dayWeightMinorUnits
			if score > suggestionScoreCutoff {
				continue
			}
			suggestions = append(suggestions, domain.MatchSuggestion{
				BankFeedID:       bt.ID,
				TargetType:       t.kind,
				TargetID:         t.id,
				AmountMinorUnits: t.amount,
				DaysApart:        int(daysApart),
				Score:            score,
			})
		}
	}

	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score < suggestions[j].Score
	})
	if len(suggestions) > maxSuggestions {
		suggestions = suggestions[:maxSuggestions]
	}
	return suggestions
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func absInt64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
