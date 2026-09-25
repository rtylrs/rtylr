package domain

import (
	"errors"
	"testing"
)

func TestValidateMatchInput_HappyPath(t *testing.T) {
	input := MatchInput{BankFeedID: 1, TargetID: 2, TargetType: TargetTypeInvoice}
	if err := ValidateMatchInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateMatchInput_ExpenseTargetType(t *testing.T) {
	input := MatchInput{BankFeedID: 1, TargetID: 2, TargetType: TargetTypeExpense}
	if err := ValidateMatchInput(input); err != nil {
		t.Errorf("unexpected error for expense target: %v", err)
	}
}

func TestValidateMatchInput_ZeroBankFeedID(t *testing.T) {
	input := MatchInput{BankFeedID: 0, TargetID: 1, TargetType: TargetTypeInvoice}
	err := ValidateMatchInput(input)
	if !errors.Is(err, ErrInvalidTarget) {
		t.Errorf("expected ErrInvalidTarget, got %v", err)
	}
}

func TestValidateMatchInput_ZeroTargetID(t *testing.T) {
	input := MatchInput{BankFeedID: 1, TargetID: 0, TargetType: TargetTypeInvoice}
	err := ValidateMatchInput(input)
	if !errors.Is(err, ErrInvalidTarget) {
		t.Errorf("expected ErrInvalidTarget, got %v", err)
	}
}

func TestValidateMatchInput_UnknownTargetType(t *testing.T) {
	input := MatchInput{BankFeedID: 1, TargetID: 2, TargetType: "unknown"}
	err := ValidateMatchInput(input)
	if !errors.Is(err, ErrInvalidTarget) {
		t.Errorf("expected ErrInvalidTarget, got %v", err)
	}
}

func TestScoreDescription_Exact(t *testing.T) {
	if got := ScoreDescription(0); got != "exact" {
		t.Errorf("score=0 -> %q, want 'exact'", got)
	}
}

func TestScoreDescription_High(t *testing.T) {
	if got := ScoreDescription(100); got != "high" {
		t.Errorf("score=100 -> %q, want 'high'", got)
	}
}

func TestScoreDescription_Medium(t *testing.T) {
	if got := ScoreDescription(1000); got != "medium" {
		t.Errorf("score=1000 -> %q, want 'medium'", got)
	}
}

func TestScoreDescription_Low(t *testing.T) {
	if got := ScoreDescription(100000); got != "low" {
		t.Errorf("score=100000 -> %q, want 'low'", got)
	}
}
