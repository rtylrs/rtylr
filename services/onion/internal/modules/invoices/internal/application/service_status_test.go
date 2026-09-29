package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

// Every non-terminal current status: a status-only write aimed at a
// money-derived status must be refused from all of them.
var statusWriteOrigins = []string{
	domain.StatusDraft,
	domain.StatusSent,
	domain.StatusViewed,
	domain.StatusPartial,
	domain.StatusOverdue,
}

// "paid" adds total to revenue and takes the invoice out of receivables, so
// writing it without a payment makes the books balance while both sides are
// wrong. "partial" asserts cash was collected that never arrived. Neither may
// be reachable from the status endpoint, from any current status.
func TestUpdateInvoiceStatus_RejectsMoneyDerivedStatus(t *testing.T) {
	for _, target := range []string{domain.StatusPaid, domain.StatusPartial} {
		for _, from := range statusWriteOrigins {
			repo := &mockRepo{invoice: &domain.Invoice{Status: from}}
			s := &Service{repo: repo}

			err := s.UpdateInvoiceStatus(context.Background(), 1, 10, target)
			if !errors.Is(err, domain.ErrInvoiceStatusNotSettable) {
				t.Errorf("%s -> %s error = %v, want ErrInvoiceStatusNotSettable", from, target, err)
			}
			if repo.statusWrites != 0 {
				t.Errorf("%s -> %s wrote %d status rows, want 0", from, target, repo.statusWrites)
			}
		}
	}
}

// The same rejection must hold on a terminal invoice, and must be a money
// rejection rather than a lifecycle conflict.
func TestUpdateInvoiceStatus_RejectsPaidOnPaidInvoice(t *testing.T) {
	repo := &mockRepo{invoice: &domain.Invoice{Status: domain.StatusPaid}}
	s := &Service{repo: repo}

	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, domain.StatusPaid); !errors.Is(err, domain.ErrInvoiceStatusNotSettable) {
		t.Errorf("error = %v, want ErrInvoiceStatusNotSettable", err)
	}
	if repo.statusWrites != 0 {
		t.Errorf("wrote %d status rows, want 0", repo.statusWrites)
	}
}

// An unknown status never reached the ENUM as a typed value: pre-fix it went
// straight into the UPDATE and surfaced as a 500 data-truncation fault.
func TestUpdateInvoiceStatus_RejectsUnknownStatus(t *testing.T) {
	for _, target := range []string{"PAID", " paid", "settled", "refunded", "partially_paid"} {
		repo := &mockRepo{invoice: &domain.Invoice{Status: domain.StatusSent}}
		s := &Service{repo: repo}

		err := s.UpdateInvoiceStatus(context.Background(), 1, 10, target)
		if !errors.Is(err, domain.ErrInvalidInvoiceStatus) {
			t.Errorf("sent -> %q error = %v, want ErrInvalidInvoiceStatus", target, err)
		}
		if repo.statusWrites != 0 {
			t.Errorf("sent -> %q wrote %d status rows, want 0", target, repo.statusWrites)
		}
	}
}

func TestUpdateInvoiceStatus_AllowsOverdueFromSent(t *testing.T) {
	repo := &mockRepo{invoice: &domain.Invoice{Status: domain.StatusSent}}
	s := &Service{repo: repo}

	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, domain.StatusOverdue); err != nil {
		t.Fatalf("sent -> overdue: %v", err)
	}
	if repo.statusWrites != 1 || repo.lastStatusWrite != domain.StatusOverdue {
		t.Fatalf("status writes = %d (%q), want 1 (%q)", repo.statusWrites, repo.lastStatusWrite, domain.StatusOverdue)
	}
}

func TestUpdateInvoiceStatus_AllowsIssueAndCloseFromDraft(t *testing.T) {
	for _, target := range []string{domain.StatusSent, domain.StatusCancelled, domain.StatusVoid} {
		repo := &mockRepo{invoice: &domain.Invoice{Status: domain.StatusDraft}}
		s := &Service{repo: repo}

		if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, target); err != nil {
			t.Errorf("draft -> %s: %v", target, err)
			continue
		}
		if repo.lastStatusWrite != target {
			t.Errorf("draft -> %s wrote %q", target, repo.lastStatusWrite)
		}
	}
}

// A draft that was never sent cannot have been viewed, and it is not a
// receivable that can be overdue.
func TestUpdateInvoiceStatus_RejectsIssuanceSkipsFromDraft(t *testing.T) {
	for _, target := range []string{domain.StatusViewed, domain.StatusOverdue} {
		repo := &mockRepo{invoice: &domain.Invoice{Status: domain.StatusDraft}}
		s := &Service{repo: repo}

		if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, target); !errors.Is(err, domain.ErrInvoiceNotDraft) {
			t.Errorf("draft -> %s error = %v, want ErrInvoiceNotDraft", target, err)
		}
		if repo.statusWrites != 0 {
			t.Errorf("draft -> %s wrote %d status rows, want 0", target, repo.statusWrites)
		}
	}
}

// Terminal invoices stay closed, and an invoice with cash against it is not
// relabelled back to a status that reads as "nothing collected yet".
func TestUpdateInvoiceStatus_RejectsRefusedTransitions(t *testing.T) {
	cases := []struct{ from, to string }{
		{domain.StatusSent, domain.StatusDraft},
		{domain.StatusViewed, domain.StatusDraft},
		{domain.StatusOverdue, domain.StatusDraft},
		{domain.StatusPartial, domain.StatusDraft},
		{domain.StatusPartial, domain.StatusSent},
		{domain.StatusPartial, domain.StatusViewed},
		{domain.StatusPaid, domain.StatusSent},
		{domain.StatusPaid, domain.StatusCancelled},
		{domain.StatusCancelled, domain.StatusSent},
		{domain.StatusCancelled, domain.StatusDraft},
		{domain.StatusVoid, domain.StatusSent},
		{domain.StatusVoid, domain.StatusDraft},
	}
	for _, tc := range cases {
		repo := &mockRepo{invoice: &domain.Invoice{Status: tc.from}}
		s := &Service{repo: repo}

		err := s.UpdateInvoiceStatus(context.Background(), 1, 10, tc.to)
		if !errors.Is(err, domain.ErrInvoiceNotDraft) {
			t.Errorf("%s -> %s error = %v, want ErrInvoiceNotDraft", tc.from, tc.to, err)
		}
		if repo.statusWrites != 0 {
			t.Errorf("%s -> %s wrote %d status rows, want 0", tc.from, tc.to, repo.statusWrites)
		}
	}
}

// A missing invoice is still a 404, not a lifecycle conflict.
func TestUpdateInvoiceStatus_PropagatesMissingInvoice(t *testing.T) {
	want := domain.ErrInvoiceNotFound
	repo := &mockRepo{err: want}
	s := &Service{repo: repo}

	if err := s.UpdateInvoiceStatus(context.Background(), 1, 10, domain.StatusSent); !errors.Is(err, want) {
		t.Errorf("error = %v, want ErrInvoiceNotFound", err)
	}
	if repo.statusWrites != 0 {
		t.Errorf("wrote %d status rows, want 0", repo.statusWrites)
	}
}
