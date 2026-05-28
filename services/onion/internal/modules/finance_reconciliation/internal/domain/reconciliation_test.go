package domain

import (
	"testing"
)

func TestIsExactAmountMatch_TrueWhenZeroAmountDiff(t *testing.T) {
	s := &MatchSuggestion{AmountMinorUnits: 0}
	if !s.IsExactAmountMatch() {
		t.Error("expected IsExactAmountMatch=true when AmountMinorUnits=0")
	}
}

func TestIsExactAmountMatch_FalseWhenNonZero(t *testing.T) {
	s := &MatchSuggestion{AmountMinorUnits: 150}
	if s.IsExactAmountMatch() {
		t.Error("expected IsExactAmountMatch=false when AmountMinorUnits=150")
	}
}

func TestIsHighConfidence_TrueWhenScoreBelow500(t *testing.T) {
	s := &MatchSuggestion{Score: 499}
	if !s.IsHighConfidence() {
		t.Error("expected IsHighConfidence=true for score=499")
	}
}

func TestIsHighConfidence_FalseWhenScoreAt500(t *testing.T) {
	s := &MatchSuggestion{Score: 500}
	if s.IsHighConfidence() {
		t.Error("expected IsHighConfidence=false for score=500")
	}
}

func TestIsHighConfidence_FalseWhenScoreAbove500(t *testing.T) {
	s := &MatchSuggestion{Score: 1000}
	if s.IsHighConfidence() {
		t.Error("expected IsHighConfidence=false for score=1000")
	}
}

func TestIsValidTargetType_TrueForInvoice(t *testing.T) {
	if !IsValidTargetType(TargetTypeInvoice) {
		t.Error("expected IsValidTargetType=true for invoice")
	}
}

func TestIsValidTargetType_TrueForExpense(t *testing.T) {
	if !IsValidTargetType(TargetTypeExpense) {
		t.Error("expected IsValidTargetType=true for expense")
	}
}

func TestIsValidTargetType_FalseForUnknown(t *testing.T) {
	if IsValidTargetType("payment") {
		t.Error("expected IsValidTargetType=false for unknown type")
	}
}

func TestHasSuggestions_TrueWhenSuggestionsPresent(t *testing.T) {
	o := &ReconciliationOverview{Suggestions: []MatchSuggestion{{BankFeedID: 1}}}
	if !o.HasSuggestions() {
		t.Error("expected HasSuggestions=true when suggestions are present")
	}
}

func TestHasSuggestions_FalseWhenEmpty(t *testing.T) {
	o := &ReconciliationOverview{}
	if o.HasSuggestions() {
		t.Error("expected HasSuggestions=false when no suggestions")
	}
}

func TestHasUnmatchedItems_TrueWhenBankTxnsPresent(t *testing.T) {
	o := &ReconciliationOverview{UnmatchedBankTxns: []UnmatchedBankTransaction{{ID: 1}}}
	if !o.HasUnmatchedItems() {
		t.Error("expected HasUnmatchedItems=true when bank txns present")
	}
}

func TestHasUnmatchedItems_TrueWhenInvoicesPresent(t *testing.T) {
	o := &ReconciliationOverview{UnmatchedInvoices: []UnmatchedInvoice{{ID: 1}}}
	if !o.HasUnmatchedItems() {
		t.Error("expected HasUnmatchedItems=true when unmatched invoices present")
	}
}

func TestHasUnmatchedItems_FalseWhenBothEmpty(t *testing.T) {
	o := &ReconciliationOverview{}
	if o.HasUnmatchedItems() {
		t.Error("expected HasUnmatchedItems=false when both slices empty")
	}
}
