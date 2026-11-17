package domain

import (
	"errors"
	"testing"
)

func TestCanChangeInvoiceProject(t *testing.T) {
	tests := []struct {
		status string
		want   error
	}{
		{status: StatusDraft},
		{status: StatusSent},
		{status: StatusViewed},
		{status: StatusPartial},
		{status: StatusOverdue},
		// Paid is terminal for money but is what a project's collected total
		// counts, so its tag stays editable.
		{status: StatusPaid},
		{status: StatusCancelled, want: ErrInvoiceProjectLocked},
		{status: StatusVoid, want: ErrInvoiceProjectLocked},
		{status: "", want: ErrInvalidInvoiceStatus},
		{status: "archived", want: ErrInvalidInvoiceStatus},
	}
	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			err := CanChangeInvoiceProject(tc.status)
			if tc.want == nil && err != nil {
				t.Fatalf("CanChangeInvoiceProject(%q) = %v, want nil", tc.status, err)
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("CanChangeInvoiceProject(%q) = %v, want %v", tc.status, err, tc.want)
			}
		})
	}
}

func TestCanChangeInvoiceProject_CoversEveryStatus(t *testing.T) {
	for status := range ValidInvoiceStatuses {
		err := CanChangeInvoiceProject(status)
		_, locked := projectTagLockedStatuses[status]
		if locked != errors.Is(err, ErrInvoiceProjectLocked) {
			t.Fatalf("status %q: locked = %v but CanChangeInvoiceProject = %v", status, locked, err)
		}
		if !locked && err != nil {
			t.Fatalf("status %q: unexpected error %v", status, err)
		}
	}
}
