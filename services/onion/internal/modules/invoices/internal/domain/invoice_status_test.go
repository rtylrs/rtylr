package domain

import (
	"errors"
	"testing"
)

// expectedInvoiceStatusTransitions is the from -> to matrix a status-only write
// must implement. It is written out longhand on purpose: the table in
// invoice_status.go is the money rule, and a test that recomputed it from the
// same map would only restate it.
var expectedInvoiceStatusTransitions = map[string][]string{
	StatusDraft:     {StatusDraft, StatusSent, StatusCancelled, StatusVoid},
	StatusSent:      {StatusSent, StatusViewed, StatusOverdue, StatusCancelled, StatusVoid},
	StatusViewed:    {StatusSent, StatusViewed, StatusOverdue, StatusCancelled, StatusVoid},
	StatusPartial:   {StatusOverdue, StatusCancelled, StatusVoid},
	StatusOverdue:   {StatusSent, StatusViewed, StatusOverdue, StatusCancelled, StatusVoid},
	StatusPaid:      {},
	StatusCancelled: {},
	StatusVoid:      {},
}

func TestInvoiceStatusTransitions_MatchTheDocumentedMatrix(t *testing.T) {
	if len(invoiceStatusTransitions) != len(expectedInvoiceStatusTransitions) {
		t.Fatalf("transition table has %d rows, want %d", len(invoiceStatusTransitions), len(expectedInvoiceStatusTransitions))
	}
	for from, want := range expectedInvoiceStatusTransitions {
		allowed, ok := invoiceStatusTransitions[from]
		if !ok {
			t.Errorf("status %q has no transition row", from)
			continue
		}
		if len(allowed) != len(want) {
			t.Errorf("from %q allowed %d targets, want %d (%v)", from, len(allowed), len(want), want)
		}
		for _, to := range want {
			if _, ok := allowed[to]; !ok {
				t.Errorf("from %q, transition to %q should be allowed", from, to)
			}
		}
	}
}

func TestInvoiceStatusTransitions_CoverEveryValidStatus(t *testing.T) {
	for status := range ValidInvoiceStatuses {
		if _, ok := invoiceStatusTransitions[status]; !ok {
			t.Errorf("status %q is valid but has no transition row", status)
		}
	}
}

// A status-only write must never be able to reach a status that asserts money
// was collected, from any current status.
func TestInvoiceStatusTransitions_NeverReachMoneyDerivedStatus(t *testing.T) {
	for from, allowed := range invoiceStatusTransitions {
		for to := range allowed {
			if !IsManuallyAssignableStatus(to) {
				t.Errorf("from %q, target %q is money-derived and must not be writable", from, to)
			}
		}
	}
}

func TestInvoiceStatusTransitions_TerminalStatusesAreDeadEnds(t *testing.T) {
	for status := range TerminalStatuses {
		if allowed := invoiceStatusTransitions[status]; len(allowed) != 0 {
			t.Errorf("terminal status %q still allows %d transitions", status, len(allowed))
		}
	}
}

func TestValidateSettableInvoiceStatus(t *testing.T) {
	for _, status := range []string{StatusDraft, StatusSent, StatusViewed, StatusOverdue, StatusCancelled, StatusVoid} {
		if err := ValidateSettableInvoiceStatus(status); err != nil {
			t.Errorf("ValidateSettableInvoiceStatus(%q) = %v, want nil", status, err)
		}
	}
	for _, status := range []string{StatusPaid, StatusPartial} {
		if err := ValidateSettableInvoiceStatus(status); !errors.Is(err, ErrInvoiceStatusNotSettable) {
			t.Errorf("ValidateSettableInvoiceStatus(%q) = %v, want ErrInvoiceStatusNotSettable", status, err)
		}
	}
	for _, status := range []string{"", "PAID", " paid", "settled", "refunded"} {
		if err := ValidateSettableInvoiceStatus(status); !errors.Is(err, ErrInvalidInvoiceStatus) {
			t.Errorf("ValidateSettableInvoiceStatus(%q) = %v, want ErrInvalidInvoiceStatus", status, err)
		}
	}
}

func TestCanTransitionInvoiceStatus(t *testing.T) {
	for from, targets := range expectedInvoiceStatusTransitions {
		for to := range ValidInvoiceStatuses {
			err := CanTransitionInvoiceStatus(from, to)
			allowed := false
			for _, candidate := range targets {
				if candidate == to {
					allowed = true
				}
			}
			switch {
			case allowed && err != nil:
				t.Errorf("%s -> %s should be allowed, got %v", from, to, err)
			case !allowed && err == nil:
				t.Errorf("%s -> %s should be refused", from, to)
			case !allowed && !errors.Is(err, ErrInvoiceNotDraft) && !errors.Is(err, ErrInvoiceStatusNotSettable):
				t.Errorf("%s -> %s error = %v, want a lifecycle or money rejection", from, to, err)
			}
		}
	}
}

// An invoice whose status is not a known value cannot be re-labelled: the
// lifecycle it is actually in is unknown.
func TestCanTransitionInvoiceStatus_RefusesUnknownCurrentStatus(t *testing.T) {
	if err := CanTransitionInvoiceStatus("written-off", StatusSent); !errors.Is(err, ErrInvoiceNotDraft) {
		t.Errorf("error = %v, want ErrInvoiceNotDraft", err)
	}
}
