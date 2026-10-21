package domain

import "testing"

func balanced(date string, lines []JournalLineInput) JournalEntryInput {
	return JournalEntryInput{Date: date, Lines: lines}
}

func TestValidateJournalEntryInput_HappyPath(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 1, Debit: 100},
		{AccountID: 2, Credit: 100},
	})
	if err := ValidateJournalEntryInput(input); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateJournalEntryInput_MissingDate(t *testing.T) {
	input := JournalEntryInput{Lines: []JournalLineInput{
		{AccountID: 1, Debit: 100},
		{AccountID: 2, Credit: 100},
	}}
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for missing date")
	}
}

func TestValidateJournalEntryInput_BadDateFormat(t *testing.T) {
	input := balanced("01-01-2026", []JournalLineInput{
		{AccountID: 1, Debit: 100},
		{AccountID: 2, Credit: 100},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for bad date format")
	}
}

func TestValidateJournalEntryInput_TooFewLines(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 1, Debit: 100},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for fewer than two lines")
	}
}

func TestValidateJournalEntryInput_ZeroAccountID(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 0, Debit: 100},
		{AccountID: 2, Credit: 100},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for zero account_id")
	}
}

func TestValidateJournalEntryInput_NegativeDebit(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 1, Debit: -50},
		{AccountID: 2, Credit: -50},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for negative debit")
	}
}

func TestValidateJournalEntryInput_BothDebitAndCreditZero(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 1, Debit: 0, Credit: 0},
		{AccountID: 2, Debit: 0, Credit: 0},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for all-zero line")
	}
}

func TestValidateJournalEntryInput_Unbalanced(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 1, Debit: 200},
		{AccountID: 2, Credit: 100},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for unbalanced entry")
	}
}

func TestValidateJournalEntryInput_MultiLineBalanced(t *testing.T) {
	input := balanced("2026-06-15", []JournalLineInput{
		{AccountID: 1, Debit: 300},
		{AccountID: 2, Credit: 200},
		{AccountID: 3, Credit: 100},
	})
	if err := ValidateJournalEntryInput(input); err != nil {
		t.Errorf("unexpected error for multi-line balanced entry: %v", err)
	}
}

func TestValidateJournalEntryInput_EmptyLines(t *testing.T) {
	input := JournalEntryInput{Date: "2026-01-01", Lines: []JournalLineInput{}}
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for empty lines")
	}
}

func TestValidateJournalEntryInput_BothDebitAndCreditPositive(t *testing.T) {
	input := balanced("2026-01-01", []JournalLineInput{
		{AccountID: 1, Debit: 100, Credit: 50},
		{AccountID: 2, Credit: 50},
	})
	if err := ValidateJournalEntryInput(input); err == nil {
		t.Fatal("expected error for line having both positive debit and credit")
	}
}
