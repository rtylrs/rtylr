package domain

import (
	"testing"
)

func TestIsPosted_TrueWhenStatusPosted(t *testing.T) {
	j := &JournalEntry{Status: "posted"}
	if !j.IsPosted() {
		t.Error("expected IsPosted=true for status=posted")
	}
}

func TestIsPosted_FalseWhenStatusDraft(t *testing.T) {
	j := &JournalEntry{Status: "draft"}
	if j.IsPosted() {
		t.Error("expected IsPosted=false for status=draft")
	}
}

func TestIsDraft_TrueWhenStatusDraft(t *testing.T) {
	j := &JournalEntry{Status: "draft"}
	if !j.IsDraft() {
		t.Error("expected IsDraft=true for status=draft")
	}
}

func TestIsDraft_TrueWhenStatusEmpty(t *testing.T) {
	j := &JournalEntry{}
	if !j.IsDraft() {
		t.Error("expected IsDraft=true for empty status")
	}
}

func TestIsDraft_FalseWhenPosted(t *testing.T) {
	j := &JournalEntry{Status: "posted"}
	if j.IsDraft() {
		t.Error("expected IsDraft=false for status=posted")
	}
}

func TestTotalDebits_SumsAllLines(t *testing.T) {
	j := &JournalEntry{
		Lines: []JournalLine{
			{Debit: 100},
			{Debit: 200},
			{Debit: 50},
		},
	}
	want := 350.0
	if got := j.TotalDebits(); got != want {
		t.Errorf("expected TotalDebits=%f, got %f", want, got)
	}
}

func TestTotalCredits_SumsAllLines(t *testing.T) {
	j := &JournalEntry{
		Lines: []JournalLine{
			{Credit: 150},
			{Credit: 200},
		},
	}
	want := 350.0
	if got := j.TotalCredits(); got != want {
		t.Errorf("expected TotalCredits=%f, got %f", want, got)
	}
}

func TestTotalDebits_ZeroForNoLines(t *testing.T) {
	j := &JournalEntry{}
	if got := j.TotalDebits(); got != 0 {
		t.Errorf("expected 0 for no lines, got %f", got)
	}
}

func TestTotalCredits_ZeroForNoLines(t *testing.T) {
	j := &JournalEntry{}
	if got := j.TotalCredits(); got != 0 {
		t.Errorf("expected 0 for no lines, got %f", got)
	}
}

func TestTotalDebitsEqualsCredits_ForBalancedEntry(t *testing.T) {
	j := &JournalEntry{
		Lines: []JournalLine{
			{AccountID: 1, Debit: 500, Credit: 0},
			{AccountID: 2, Debit: 0, Credit: 500},
		},
	}
	if j.TotalDebits() != j.TotalCredits() {
		t.Errorf("expected balanced entry: debits=%f credits=%f", j.TotalDebits(), j.TotalCredits())
	}
}
