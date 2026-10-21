package domain

import "testing"

func TestValidateOpenSessionInput_RejectsMissingOwnerID(t *testing.T) {
	if err := ValidateOpenSessionInput(OpenSessionInput{OpeningFloat: 100}); err == nil {
		t.Fatal("expected error for missing owner id")
	}
}

func TestValidateOpenSessionInput_RejectsNegativeOpeningFloat(t *testing.T) {
	if err := ValidateOpenSessionInput(OpenSessionInput{OwnerID: 1, OpeningFloat: -10}); err == nil {
		t.Fatal("expected error for negative opening float")
	}
}

func TestValidateOpenSessionInput_AcceptsZeroOpeningFloat(t *testing.T) {
	if err := ValidateOpenSessionInput(OpenSessionInput{OwnerID: 1, OpeningFloat: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOpenSessionInput_AcceptsValid(t *testing.T) {
	if err := ValidateOpenSessionInput(OpenSessionInput{OwnerID: 1, OpeningFloat: 200}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCloseSessionInput_RejectsMissingSessionID(t *testing.T) {
	if err := ValidateCloseSessionInput(CloseSessionInput{OwnerID: 1, ClosingFloat: 100}); err == nil {
		t.Fatal("expected error for missing session id")
	}
}

func TestValidateCloseSessionInput_RejectsMissingOwnerID(t *testing.T) {
	if err := ValidateCloseSessionInput(CloseSessionInput{ID: 1, ClosingFloat: 100}); err == nil {
		t.Fatal("expected error for missing owner id")
	}
}

func TestValidateCloseSessionInput_RejectsNegativeClosingFloat(t *testing.T) {
	if err := ValidateCloseSessionInput(CloseSessionInput{ID: 1, OwnerID: 2, ClosingFloat: -5}); err == nil {
		t.Fatal("expected error for negative closing float")
	}
}

func TestValidateCloseSessionInput_AcceptsValid(t *testing.T) {
	if err := ValidateCloseSessionInput(CloseSessionInput{ID: 1, OwnerID: 2, ClosingFloat: 150}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSession_IsOpen(t *testing.T) {
	s := Session{Status: "open"}
	if !s.IsOpen() {
		t.Error("expected IsOpen=true for status 'open'")
	}
}

func TestSession_IsOpen_WhenClosed(t *testing.T) {
	s := Session{Status: "closed"}
	if s.IsOpen() {
		t.Error("expected IsOpen=false for status 'closed'")
	}
}

func TestSession_IsClosed(t *testing.T) {
	s := Session{Status: "closed"}
	if !s.IsClosed() {
		t.Error("expected IsClosed=true for status 'closed'")
	}
}

func TestSession_IsClosed_WhenOpen(t *testing.T) {
	s := Session{Status: "open"}
	if s.IsClosed() {
		t.Error("expected IsClosed=false for status 'open'")
	}
}
