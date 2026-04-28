package application

import (
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/finance_reconciliation/internal/domain"
)

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
		{ID: 100, AmountMinorUnits: -10000, OccurredAt: now},
	}
	invoices := []domain.UnmatchedInvoice{
		{ID: 1, AmountMinorUnits: 10000, DueDate: now}, // exact same day, exact amount → score 0
		{ID: 2, AmountMinorUnits: 10000, DueDate: now.AddDate(0, 0, 5)},
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
		{ID: 100, AmountMinorUnits: -1_000_000, OccurredAt: now},
	}
	invoices := []domain.UnmatchedInvoice{
		{ID: 1, AmountMinorUnits: 10, DueDate: now.AddDate(0, 0, 365)}, // huge amount + date diff
	}
	got := computeSuggestions(bank, invoices, nil)
	if len(got) != 0 {
		t.Errorf("expected no suggestions for high-score candidate, got %v", got)
	}
}

func TestComputeSuggestions_ExpensesAreTagged(t *testing.T) {
	now := time.Now()
	bank := []domain.UnmatchedBankTransaction{
		{ID: 7, AmountMinorUnits: 5000, OccurredAt: now},
	}
	expenses := []domain.UnmatchedInvoice{
		{ID: 42, AmountMinorUnits: 5000, IssueDate: now},
	}
	got := computeSuggestions(bank, nil, expenses)
	if len(got) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(got))
	}
	if got[0].TargetType != domain.TargetTypeExpense {
		t.Errorf("expected target type %q, got %q", domain.TargetTypeExpense, got[0].TargetType)
	}
}

func TestComputeSuggestions_CapsAtMaxSuggestions(t *testing.T) {
	now := time.Now()
	bank := []domain.UnmatchedBankTransaction{}
	invoices := []domain.UnmatchedInvoice{}
	for i := 0; i < 100; i++ {
		bank = append(bank, domain.UnmatchedBankTransaction{
			ID: int64(i + 1), AmountMinorUnits: 1000, OccurredAt: now,
		})
	}
	for i := 0; i < 100; i++ {
		invoices = append(invoices, domain.UnmatchedInvoice{
			ID: int64(i + 1000), AmountMinorUnits: 1000, DueDate: now,
		})
	}
	got := computeSuggestions(bank, invoices, nil)
	if len(got) > maxSuggestions {
		t.Errorf("expected at most %d, got %d", maxSuggestions, len(got))
	}
}
