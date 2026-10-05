package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance/finance_reconciliation/internal/domain"
)

func TestTruncatePreviewSignalsOverflow(t *testing.T) {
	items := make([]int, reconciliationPreviewLimit+1)
	preview, truncated := truncatePreview(items)
	if !truncated {
		t.Fatal("truncated = false, want true")
	}
	if len(preview) != reconciliationPreviewLimit {
		t.Fatalf("preview length = %d, want %d", len(preview), reconciliationPreviewLimit)
	}
}

type mockRepo struct {
	bank         []domain.UnmatchedBankTransaction
	invoices     []domain.UnmatchedInvoice
	expenses     []domain.UnmatchedInvoice
	bankErr      error
	invErr       error
	expErr       error
	exists       bool
	existsErr    error
	targetExists bool
	matchID      int64
	matchErr     error
}

func (m *mockRepo) ListUnmatchedBankFeeds(_ context.Context, _ int64) ([]domain.UnmatchedBankTransaction, error) {
	return m.bank, m.bankErr
}

func (m *mockRepo) ListUnmatchedInvoices(_ context.Context, _ int64) ([]domain.UnmatchedInvoice, error) {
	return m.invoices, m.invErr
}

func (m *mockRepo) ListUnmatchedExpenses(_ context.Context, _ int64) ([]domain.UnmatchedInvoice, error) {
	return m.expenses, m.expErr
}

func (*mockRepo) PopulateOutstandingInvoiceAmounts(_ context.Context, _ int64, _ []domain.UnmatchedInvoice) error {
	return nil
}

func (m *mockRepo) BankFeedExists(_ context.Context, _, _ int64) (bool, error) {
	return m.exists, m.existsErr
}

func (m *mockRepo) TargetExists(_ context.Context, _ int64, _ domain.MatchInput) (bool, error) {
	return m.targetExists, nil
}

func (m *mockRepo) CreateMatch(_ context.Context, _ int64, _ domain.MatchInput) (int64, error) {
	return m.matchID, m.matchErr
}

func TestOverview_PropagatesEachListError(t *testing.T) {
	want := errors.New("boom")

	s := &Service{repo: &mockRepo{bankErr: want}}
	if _, err := s.Overview(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("bank err: %v", err)
	}

	s = &Service{repo: &mockRepo{invErr: want}}
	if _, err := s.Overview(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("inv err: %v", err)
	}

	s = &Service{repo: &mockRepo{expErr: want}}
	if _, err := s.Overview(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("exp err: %v", err)
	}
}

func TestOverview_ReturnsBundle(t *testing.T) {
	s := &Service{repo: &mockRepo{
		bank:     []domain.UnmatchedBankTransaction{{ID: 1}},
		invoices: []domain.UnmatchedInvoice{{ID: 2}},
	}}
	got, err := s.Overview(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got.UnmatchedBankTxns) != 1 || len(got.UnmatchedInvoices) != 1 {
		t.Errorf("bundle: %+v", got)
	}
}

func TestMatch_RejectsBadInputs(t *testing.T) {
	s := &Service{repo: &mockRepo{exists: true}}

	if _, err := s.Match(context.Background(), 1, domain.MatchInput{TargetType: "fake", BankFeedID: 1, TargetID: 1}); !errors.Is(err, domain.ErrInvalidTarget) {
		t.Errorf("bad target type: %v", err)
	}
	if _, err := s.Match(context.Background(), 1, domain.MatchInput{TargetType: domain.TargetTypeInvoice, BankFeedID: 0, TargetID: 1}); !errors.Is(err, domain.ErrInvalidTarget) {
		t.Errorf("zero bank id: %v", err)
	}
	if _, err := s.Match(context.Background(), 1, domain.MatchInput{TargetType: domain.TargetTypeInvoice, BankFeedID: 1, TargetID: 0}); !errors.Is(err, domain.ErrInvalidTarget) {
		t.Errorf("zero target id: %v", err)
	}
}

func TestMatch_RejectsMissingBankFeed(t *testing.T) {
	s := &Service{repo: &mockRepo{exists: false}}
	_, err := s.Match(context.Background(), 1, domain.MatchInput{TargetType: domain.TargetTypeInvoice, BankFeedID: 5, TargetID: 7})
	if !errors.Is(err, domain.ErrBankFeedNotFound) {
		t.Errorf("expected ErrBankFeedNotFound, got %v", err)
	}
}

func TestMatch_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{exists: true, targetExists: true, matchID: 99}}
	id, err := s.Match(context.Background(), 1, domain.MatchInput{TargetType: domain.TargetTypeInvoice, BankFeedID: 5, TargetID: 7})
	if err != nil || id != 99 {
		t.Errorf("id=%d err=%v", id, err)
	}
}

func TestMatch_RejectsMissingTarget(t *testing.T) {
	s := &Service{repo: &mockRepo{exists: true}}
	_, err := s.Match(context.Background(), 1, domain.MatchInput{TargetType: domain.TargetTypeInvoice, BankFeedID: 5, TargetID: 7})
	if !errors.Is(err, domain.ErrTargetNotFound) {
		t.Errorf("expected ErrTargetNotFound, got %v", err)
	}
}

func TestAbsInt64(t *testing.T) {
	cases := map[int64]int64{
		0:        0,
		5:        5,
		-5:       5,
		-1 << 30: 1 << 30,
	}
	for in, want := range cases {
		if got := absInt64(in); got != want {
			t.Errorf("absInt64(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestComputeSuggestions_EmptyInputs(t *testing.T) {
	if got := computeSuggestions(nil, nil, nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
	if got := computeSuggestions([]domain.UnmatchedBankTransaction{{ID: 1}}, nil, nil); got != nil {
		t.Errorf("no targets -> nil, got %v", got)
	}
}

func TestComputeSuggestions_PerfectMatchSortsFirst(t *testing.T) {
	now := time.Now().Truncate(24 * time.Hour)
	bank := []domain.UnmatchedBankTransaction{
		{ID: 100, AmountMinorUnits: -10000, Currency: "usd", OccurredAt: now},
	}
	invoices := []domain.UnmatchedInvoice{
		{ID: 1, AmountMinorUnits: 10000, Currency: "USD", DueDate: now}, // exact same day, exact amount → score 0
		{ID: 2, AmountMinorUnits: 10000, Currency: "USD", DueDate: now.AddDate(0, 0, 5)},
	}
	suggestions := computeSuggestions(bank, invoices, nil)
	if len(suggestions) == 0 {
		t.Fatal("expected suggestions")
	}
	if suggestions[0].TargetID != 1 {
		t.Errorf("expected exact match (target 1) to sort first, got %d", suggestions[0].TargetID)
	}
	if suggestions[0].Score >= suggestions[len(suggestions)-1].Score+1 {
		t.Errorf("scores should be ascending")
	}
}

func TestComputeSuggestions_DropsHighScoreCandidates(t *testing.T) {
	now := time.Now()
	bank := []domain.UnmatchedBankTransaction{
		{ID: 100, AmountMinorUnits: -1_000_000, Currency: "USD", OccurredAt: now},
	}
	invoices := []domain.UnmatchedInvoice{
		{ID: 1, AmountMinorUnits: 10, Currency: "USD", DueDate: now.AddDate(0, 0, 365)}, // huge amount + date diff
	}
	got := computeSuggestions(bank, invoices, nil)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for high-score candidate, got %v", got)
	}
}

func TestComputeSuggestions_ExpensesAreTagged(t *testing.T) {
	now := time.Now()
	bank := []domain.UnmatchedBankTransaction{
		{ID: 7, AmountMinorUnits: 5000, Currency: "USD", OccurredAt: now},
	}
	expenses := []domain.UnmatchedInvoice{
		{ID: 42, AmountMinorUnits: 5000, Currency: "USD", IssueDate: now},
	}
	got := computeSuggestions(bank, nil, expenses)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0].TargetType != domain.TargetTypeExpense {
		t.Errorf("expected target type %q, got %q", domain.TargetTypeExpense, got[0].TargetType)
	}
}

func TestComputeSuggestions_DropsCrossCurrencyCandidates(t *testing.T) {
	now := time.Now()
	bank := []domain.UnmatchedBankTransaction{
		{ID: 1, AmountMinorUnits: 10000, Currency: "USD", OccurredAt: now},
	}
	invoices := []domain.UnmatchedInvoice{
		{ID: 2, AmountMinorUnits: 10000, Currency: "EUR", DueDate: now},
	}
	if got := computeSuggestions(bank, invoices, nil); len(got) != 0 {
		t.Fatalf("expected no cross-currency suggestions, got %v", got)
	}
}

func TestComputeSuggestions_CapsAtMaxSuggestions(t *testing.T) {
	now := time.Now()
	bank := []domain.UnmatchedBankTransaction{}
	invoices := []domain.UnmatchedInvoice{}
	for i := 0; i < 100; i++ {
		bank = append(bank, domain.UnmatchedBankTransaction{
			ID: int64(i + 1), AmountMinorUnits: 1000, Currency: "USD", OccurredAt: now,
		})
	}
	for i := 0; i < 100; i++ {
		invoices = append(invoices, domain.UnmatchedInvoice{
			ID: int64(i + 1000), AmountMinorUnits: 1000, Currency: "USD", DueDate: now,
		})
	}
	got := computeSuggestions(bank, invoices, nil)
	if len(got) > maxSuggestions {
		t.Errorf("expected at most %d, got %d", maxSuggestions, len(got))
	}
}
