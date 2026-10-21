package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
)

func (m *mockRepo) ListInvoicesByCustomer(_ context.Context, _, _ int64, _, _ int) ([]domain.Invoice, error) {
	if m.invoice != nil {
		return []domain.Invoice{*m.invoice}, nil
	}
	return nil, m.err
}

func (m *mockRepo) CountInvoicesByCustomer(_ context.Context, _, _ int64) (int64, error) {
	if m.invoice != nil {
		return 1, nil
	}
	return 0, m.err
}

func TestApplyHeaderDiscount_RejectsOverDiscount(t *testing.T) {
	for _, tc := range []struct {
		name  string
		dtype string
		value float64
	}{
		{"fixed exceeds subtotal", "fixed", 200},
		{"percentage over 100", "percentage", 150},
		{"negative fixed", "fixed", -10},
		{"negative percentage", "percentage", -5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := applyHeaderDiscount(150, tc.dtype, tc.value); !errors.Is(err, domain.ErrDiscountExceedsTotal) {
				t.Fatalf("error = %v, want ErrDiscountExceedsTotal", err)
			}
		})
	}
}

func TestApplyHeaderDiscount_AcceptsBoundary(t *testing.T) {
	amount, err := applyHeaderDiscount(150, "fixed", 150)
	if err != nil || !almostEqualAmount(amount, 150) {
		t.Fatalf("amount = %f, err = %v, want 150, nil", amount, err)
	}
	amount, err = applyHeaderDiscount(150, "percentage", 100)
	if err != nil || !almostEqualAmount(amount, 150) {
		t.Fatalf("amount = %f, err = %v, want 150, nil", amount, err)
	}
}

func TestCreateInvoice_RejectsOverDiscount(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate:     "2024-06-15",
		DueDate:       "2024-06-20",
		DiscountValue: 200,
		Items: []domain.InvoiceItemInput{
			{Quantity: 1, UnitPrice: 150},
		},
	})
	if !errors.Is(err, domain.ErrDiscountExceedsTotal) {
		t.Fatalf("error = %v, want ErrDiscountExceedsTotal", err)
	}
	if repo.created != nil {
		t.Error("over-discounted invoice reached the repository")
	}
}

// ---- CRM customer link ----

type fakeCustomerSnapshots struct {
	name    string
	email   string
	address string
	err     error
	calls   int
}

func (f *fakeCustomerSnapshots) CustomerSnapshot(_ context.Context, _, _ int64) (string, string, string, error) {
	f.calls++
	return f.name, f.email, f.address, f.err
}

func TestCreateInvoice_PersistsVerifiedCustomerLink(t *testing.T) {
	repo := &mockRepo{generatedID: 77}
	provider := &fakeCustomerSnapshots{name: "Acme Ltd", email: "billing@acme.test", address: "1 Test Way"}
	s := &Service{repo: repo}
	s.SetCustomerSnapshotProvider(provider)

	inv, err := s.CreateInvoice(context.Background(), 1001, domain.InvoiceInput{
		CustomerID: 42,
		IssueDate:  "2024-06-15",
		DueDate:    "2024-06-30",
		Items: []domain.InvoiceItemInput{
			{Description: "Widget", Quantity: 1, UnitPrice: 10},
		},
	})
	if err != nil {
		t.Fatalf("CreateInvoice() error = %v", err)
	}
	if provider.calls != 1 {
		t.Fatalf("snapshot provider calls = %d, want 1", provider.calls)
	}
	if repo.created == nil {
		t.Fatal("invoice never reached the repository")
	}
	if repo.created.CustomerID == nil || *repo.created.CustomerID != 42 {
		t.Fatalf("persisted customer_id = %v, want 42", repo.created.CustomerID)
	}
	if inv.ClientName != "Acme Ltd" || inv.ClientEmail != "billing@acme.test" || inv.ClientAddress != "1 Test Way" {
		t.Fatalf("snapshot not applied: name=%q email=%q address=%q", inv.ClientName, inv.ClientEmail, inv.ClientAddress)
	}
}

func TestCreateInvoice_RejectsUnverifiableCustomerLink(t *testing.T) {
	baseInput := domain.InvoiceInput{
		CustomerID: 42,
		IssueDate:  "2024-06-15",
		DueDate:    "2024-06-30",
		Items: []domain.InvoiceItemInput{
			{Description: "Widget", Quantity: 1, UnitPrice: 10},
		},
	}

	// No provider wired: the link cannot be verified, so the invoice must not
	// be written.
	unwired := &mockRepo{}
	s := &Service{repo: unwired}
	if _, err := s.CreateInvoice(context.Background(), 1001, baseInput); !errors.Is(err, ErrCustomerVerificationUnavailable) {
		t.Fatalf("unwired provider error = %v, want ErrCustomerVerificationUnavailable", err)
	}
	if unwired.created != nil {
		t.Error("invoice persisted without a customer snapshot provider")
	}

	// Unknown/foreign customer: the provider error surfaces as
	// ErrCustomerNotFound and the invoice must not be written.
	failing := &mockRepo{}
	s = &Service{repo: failing}
	s.SetCustomerSnapshotProvider(&fakeCustomerSnapshots{err: errors.New("no rows")})
	if _, err := s.CreateInvoice(context.Background(), 1001, baseInput); !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("failed lookup error = %v, want ErrCustomerNotFound", err)
	}
	if failing.created != nil {
		t.Error("invoice persisted with an unverified customer link")
	}
}
