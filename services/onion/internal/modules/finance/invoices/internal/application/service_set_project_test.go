package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

// sentInvoice is a billed invoice with money on it, so a test can tell
// whether anything besides the project tag moved.
func sentInvoice(status string, projectID *int64) *domain.Invoice {
	return &domain.Invoice{
		ID:         7,
		Status:     status,
		ProjectID:  projectID,
		Subtotal:   200,
		TaxAmount:  20,
		Total:      220,
		AmountPaid: 50,
		AmountDue:  170,
	}
}

func TestSetInvoiceProject_TagsSentAndPaidInvoices(t *testing.T) {
	for _, status := range []string{domain.StatusSent, domain.StatusViewed, domain.StatusPartial, domain.StatusOverdue, domain.StatusPaid, domain.StatusDraft} {
		t.Run(status, func(t *testing.T) {
			repo := &mockRepo{invoice: sentInvoice(status, nil)}
			s := &Service{repo: repo, projectReferences: &fakeProjects{byBusiness: map[int64][]int64{1: {55}}}}

			inv, err := s.SetInvoiceProject(context.Background(), 1, 7, 55)
			if err != nil {
				t.Fatalf("SetInvoiceProject: %v", err)
			}
			if inv.ProjectID == nil || *inv.ProjectID != 55 {
				t.Fatalf("project = %v, want 55", inv.ProjectID)
			}
			if inv.Status != status {
				t.Fatalf("status = %q, want %q unchanged", inv.Status, status)
			}
			if inv.Total != 220 || inv.AmountPaid != 50 || inv.AmountDue != 170 || inv.Subtotal != 200 || inv.TaxAmount != 20 {
				t.Fatalf("totals moved: %+v", inv)
			}
			if repo.updateWithItemsCalls != 0 || repo.statusWrites != 0 {
				t.Fatalf("a document or status write ran: update=%d status=%d", repo.updateWithItemsCalls, repo.statusWrites)
			}
		})
	}
}

func TestSetInvoiceProject_ZeroClearsTheTag(t *testing.T) {
	tagged := int64(55)
	repo := &mockRepo{invoice: sentInvoice(domain.StatusSent, &tagged)}
	projects := &fakeProjects{}
	s := &Service{repo: repo, projectReferences: projects}

	inv, err := s.SetInvoiceProject(context.Background(), 1, 7, 0)
	if err != nil {
		t.Fatalf("SetInvoiceProject(0): %v", err)
	}
	if inv.ProjectID != nil {
		t.Fatalf("project = %d after clear, want none", *inv.ProjectID)
	}
	if projects.calls != 0 {
		t.Fatalf("a clear was verified %d times", projects.calls)
	}
}

func TestSetInvoiceProject_RefusesCancelledAndVoid(t *testing.T) {
	for _, status := range []string{domain.StatusCancelled, domain.StatusVoid} {
		t.Run(status, func(t *testing.T) {
			tagged := int64(55)
			repo := &mockRepo{invoice: sentInvoice(status, &tagged)}
			s := &Service{repo: repo, projectReferences: &fakeProjects{byBusiness: map[int64][]int64{1: {55, 77}}}}

			if _, err := s.SetInvoiceProject(context.Background(), 1, 7, 77); !errors.Is(err, domain.ErrInvoiceProjectLocked) {
				t.Fatalf("error = %v, want ErrInvoiceProjectLocked", err)
			}
			if repo.projectWrites != 0 || *repo.invoice.ProjectID != 55 {
				t.Fatalf("tag changed on a %s invoice: writes=%d project=%d", status, repo.projectWrites, *repo.invoice.ProjectID)
			}
		})
	}
}

func TestSetInvoiceProject_RejectsProjectFromAnotherBusiness(t *testing.T) {
	repo := &mockRepo{invoice: sentInvoice(domain.StatusSent, nil)}
	s := &Service{repo: repo, projectReferences: &fakeProjects{byBusiness: map[int64][]int64{2: {88}}}}

	if _, err := s.SetInvoiceProject(context.Background(), 1, 7, 88); !errors.Is(err, ErrProjectNotFound) {
		t.Fatalf("error = %v, want ErrProjectNotFound", err)
	}
	if repo.projectWrites != 0 {
		t.Fatalf("a foreign project was written")
	}
}

func TestSetInvoiceProject_FailsClosedWhenUnwired(t *testing.T) {
	repo := &mockRepo{invoice: sentInvoice(domain.StatusSent, nil)}
	s := &Service{repo: repo}

	if _, err := s.SetInvoiceProject(context.Background(), 1, 7, 55); !errors.Is(err, ErrProjectVerificationUnavailable) {
		t.Fatalf("error = %v, want ErrProjectVerificationUnavailable", err)
	}
	if repo.projectWrites != 0 {
		t.Fatalf("an unverified project was written")
	}
}

func TestSetInvoiceProject_MissingInvoiceIsNotFound(t *testing.T) {
	s := &Service{repo: &mockRepo{}, projectReferences: &fakeProjects{byBusiness: map[int64][]int64{1: {55}}}}

	if _, err := s.SetInvoiceProject(context.Background(), 1, 7, 55); !errors.Is(err, domain.ErrInvoiceNotFound) {
		t.Fatalf("error = %v, want ErrInvoiceNotFound", err)
	}
}
