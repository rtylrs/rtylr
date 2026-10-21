package domain

import (
	"testing"
	"time"
)

func TestValidateTerminalInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateTerminalInput(TerminalInput{BranchID: 1, Name: "POS 1"}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateTerminalInput_RejectsMissingBranchID(t *testing.T) {
	if err := ValidateTerminalInput(TerminalInput{BusinessID: 1, Name: "POS 1"}); err == nil {
		t.Fatal("expected error for missing branch id")
	}
}

func TestValidateTerminalInput_RejectsBlankName(t *testing.T) {
	if err := ValidateTerminalInput(TerminalInput{BusinessID: 1, BranchID: 2}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateTerminalInput_AcceptsValid(t *testing.T) {
	if err := ValidateTerminalInput(TerminalInput{BusinessID: 1, BranchID: 2, Name: "Terminal A"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTerminalUpdateInput_RejectsMissingID(t *testing.T) {
	if err := ValidateTerminalUpdateInput(TerminalUpdateInput{BusinessID: 1, Name: "POS 1"}); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestValidateTerminalUpdateInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateTerminalUpdateInput(TerminalUpdateInput{ID: 1, Name: "POS 1"}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateTerminalUpdateInput_RejectsBlankName(t *testing.T) {
	if err := ValidateTerminalUpdateInput(TerminalUpdateInput{ID: 1, BusinessID: 2}); err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestValidateTerminalUpdateInput_AcceptsValid(t *testing.T) {
	if err := ValidateTerminalUpdateInput(TerminalUpdateInput{ID: 1, BusinessID: 2, Name: "Updated Terminal"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTerminal_IsOnline_NilLastSeen(t *testing.T) {
	term := Terminal{LastSeenAt: nil}
	if term.IsOnline() {
		t.Error("expected IsOnline=false for nil LastSeenAt")
	}
}

func TestTerminal_IsOnline_WithLastSeen(t *testing.T) {
	now := time.Now()
	term := Terminal{LastSeenAt: &now}
	if !term.IsOnline() {
		t.Error("expected IsOnline=true when LastSeenAt is set")
	}
}
