package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
)

func TestCreateExpenseCategory_PropagatesRepoError(t *testing.T) {
	want := errors.New("cat error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateExpenseCategory(context.Background(), 1, domain.ExpenseCategoryInput{Name: "Office"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateTaxConfig ----

func TestCreateTaxConfig_DefaultsTypeToVAT(t *testing.T) {
	s := &Service{repo: &mockRepo{generatedID: 5}}
	tc, err := s.CreateTaxConfig(context.Background(), 1, domain.TaxConfigInput{Name: "GST", Rate: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tc.Type != "vat" {
		t.Errorf("expected type=vat, got %s", tc.Type)
	}
}

func TestCreateTaxConfig_PropagatesRepoError(t *testing.T) {
	want := errors.New("tax error")
	s := &Service{repo: &mockRepo{err: want}}
	_, err := s.CreateTaxConfig(context.Background(), 1, domain.TaxConfigInput{Name: "VAT", Rate: 20})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---- CreateInvoice date validation ----

func TestCreateInvoice_RejectsDueDateBeforeIssueDate(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-10",
	})
	if !errors.Is(err, domain.ErrInvalidInvoiceDate) {
		t.Errorf("expected ErrInvalidInvoiceDate, got %v", err)
	}
}

func TestCreateInvoice_RejectsInvalidIssueDate(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "not-a-date",
		DueDate:   "2024-06-30",
	})
	if err == nil {
		t.Error("expected parse error, got nil")
	}
}

// ---- Invoice atomic writes ----

func TestCreateInvoice_WritesHeaderAndItemsAtomically(t *testing.T) {
	repo := &mockRepo{generatedID: 42}
	s := &Service{repo: repo}

	inv, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		ClientName: "Acme",
		IssueDate:  "2024-06-15",
		DueDate:    "2024-06-30",
		Items: []domain.InvoiceItemInput{
			{Description: "Widget", Quantity: 2, UnitPrice: 10},
			{Description: "Gadget", Quantity: 1, UnitPrice: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.createWithItemsCalls != 1 {
		t.Fatalf("CreateInvoiceWithItems calls = %d, want exactly 1", repo.createWithItemsCalls)
	}
	if len(repo.lastCreatedItems) != 2 {
		t.Fatalf("expected items persisted together with header, got %+v", repo.lastCreatedItems)
	}
	if repo.lastCreatedItems[0].Description != "Widget" || repo.lastCreatedItems[0].SortOrder != 0 ||
		repo.lastCreatedItems[1].Description != "Gadget" || repo.lastCreatedItems[1].SortOrder != 1 {
		t.Errorf("unexpected items passed to repo: %+v", repo.lastCreatedItems)
	}
	if inv.ID != 42 {
		t.Errorf("expected ID=42, got %d", inv.ID)
	}
	if len(inv.Items) != 2 {
		t.Errorf("expected returned invoice to carry 2 items, got %d", len(inv.Items))
	}
	if inv.Subtotal != 25 || inv.Total != 25 {
		t.Errorf("expected subtotal=total=25, got subtotal=%f total=%f", inv.Subtotal, inv.Total)
	}
}

func TestCreateInvoice_RepoFailurePropagatesWithoutPartialWrites(t *testing.T) {
	want := errors.New("atomic create failed")
	repo := &mockRepo{writeErr: want}
	s := &Service{repo: repo}

	_, err := s.CreateInvoice(context.Background(), 1, domain.InvoiceInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-30",
		Items: []domain.InvoiceItemInput{
			{Description: "Widget", Quantity: 1, UnitPrice: 10},
		},
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	if repo.createWithItemsCalls != 1 {
		t.Errorf("CreateInvoiceWithItems calls = %d, want exactly 1 (no split writes)", repo.createWithItemsCalls)
	}
	if repo.updateWithItemsCalls != 0 {
		t.Errorf("UpdateInvoiceWithItems calls = %d, want 0", repo.updateWithItemsCalls)
	}
}

func TestUpdateInvoice_WritesHeaderAndItemsAtomically(t *testing.T) {
	repo := &mockRepo{invoice: &domain.Invoice{ID: 10, Status: domain.StatusDraft}}
	s := &Service{repo: repo}

	updated, err := s.UpdateInvoice(context.Background(), 1, 10, domain.InvoiceUpdateInput{
		ClientName: "Acme",
		IssueDate:  "2024-06-15",
		DueDate:    "2024-06-30",
		Items: []domain.InvoiceItemInput{
			{Description: "Gadget", Quantity: 3, UnitPrice: 5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.updateWithItemsCalls != 1 {
		t.Fatalf("UpdateInvoiceWithItems calls = %d, want exactly 1", repo.updateWithItemsCalls)
	}
	if len(repo.lastUpdatedItems) != 1 || repo.lastUpdatedItems[0].Description != "Gadget" {
		t.Fatalf("expected replacement items written with header in one call, got %+v", repo.lastUpdatedItems)
	}
	if updated.Items == nil || len(updated.Items) != 1 {
		t.Errorf("expected returned invoice to carry 1 item, got %+v", updated.Items)
	}
	if updated.Total != 15 || updated.AmountDue != 15 {
		t.Errorf("expected total=amountDue=15, got total=%f amountDue=%f", updated.Total, updated.AmountDue)
	}
}

func TestUpdateInvoice_RepoFailurePropagatesWithoutPartialWrites(t *testing.T) {
	want := errors.New("atomic update failed")
	repo := &mockRepo{
		invoice:  &domain.Invoice{ID: 10, Status: domain.StatusDraft},
		writeErr: want,
	}
	s := &Service{repo: repo}

	_, err := s.UpdateInvoice(context.Background(), 1, 10, domain.InvoiceUpdateInput{
		IssueDate: "2024-06-15",
		DueDate:   "2024-06-30",
		Items: []domain.InvoiceItemInput{
			{Description: "Widget", Quantity: 1, UnitPrice: 10},
		},
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected sentinel, got %v", err)
	}
	if repo.updateWithItemsCalls != 1 {
		t.Errorf("UpdateInvoiceWithItems calls = %d, want exactly 1 (no split delete+recreate)", repo.updateWithItemsCalls)
	}
	if repo.createWithItemsCalls != 0 {
		t.Errorf("CreateInvoiceWithItems calls = %d, want 0", repo.createWithItemsCalls)
	}
}

// ---- CreateInvoice totals ----

func almostEqualAmount(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestCreateInvoice_Totals_Table(t *testing.T) {
	tests := []struct {
		name string
		in   domain.InvoiceInput
		want domain.Invoice
	}{
		{
			name: "header tax not stacked on item tax",
			// $100 line at 10% item tax plus 10% header tax: one $10 item tax
			// and one $10 header tax on the same base, not $21.
			in: domain.InvoiceInput{
				IssueDate: "2024-06-15",
				DueDate:   "2024-06-20",
				TaxRate:   10,
				Items: []domain.InvoiceItemInput{
					{Quantity: 1, UnitPrice: 100, TaxRate: 10},
				},
			},
			want: domain.Invoice{Subtotal: 100, DiscountAmount: 0, TaxAmount: 20, Total: 120},
		},
		{
			// A $10 discount leaves $90 payable, so the $10 item tax and the
			// $10 header tax are each 10% of $90. Taxing the item on the
			// undiscounted $100 would declare $19 of tax on $90 of revenue.
			name: "header discount also reduces the item tax base",
			in: domain.InvoiceInput{
				IssueDate:     "2024-06-15",
				DueDate:       "2024-06-20",
				DiscountType:  "percentage",
				DiscountValue: 10,
				TaxRate:       10,
				Items: []domain.InvoiceItemInput{
					{Quantity: 1, UnitPrice: 100, TaxRate: 10},
				},
			},
			want: domain.Invoice{Subtotal: 100, DiscountAmount: 10, TaxAmount: 18, Total: 108},
		},
		{
			name: "no item tax keeps legacy behavior",
			in: domain.InvoiceInput{
				IssueDate: "2024-06-15",
				DueDate:   "2024-06-20",
				TaxRate:   5,
				Items: []domain.InvoiceItemInput{
					{Quantity: 2, UnitPrice: 100},
				},
			},
			want: domain.Invoice{Subtotal: 200, DiscountAmount: 0, TaxAmount: 10, Total: 210},
		},
		{
			// $20 off a $150 basket leaves $130 payable, split $86.67/$43.33
			// across the lines. Only the second line is taxed, on its own
			// $43.33 share: 4.33, not 10% of the undiscounted $50.
			name: "fixed discount with mixed item tax rates",
			in: domain.InvoiceInput{
				IssueDate:     "2024-06-15",
				DueDate:       "2024-06-20",
				DiscountValue: 20,
				Items: []domain.InvoiceItemInput{
					{Quantity: 1, UnitPrice: 100},
					{Quantity: 1, UnitPrice: 50, TaxRate: 10},
				},
			},
			want: domain.Invoice{Subtotal: 150, DiscountAmount: 20, TaxAmount: 4.33, Total: 134.33},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{}
			s := &Service{repo: repo}
			inv, err := s.CreateInvoice(context.Background(), 1, tt.in)
			if err != nil {
				t.Fatalf("CreateInvoice() error = %v", err)
			}
			if repo.created == nil {
				t.Fatal("expected invoice to reach the repository")
			}
			if !almostEqualAmount(inv.Subtotal, tt.want.Subtotal) {
				t.Errorf("Subtotal = %f, want %f", inv.Subtotal, tt.want.Subtotal)
			}
			if !almostEqualAmount(inv.DiscountAmount, tt.want.DiscountAmount) {
				t.Errorf("DiscountAmount = %f, want %f", inv.DiscountAmount, tt.want.DiscountAmount)
			}
			if !almostEqualAmount(inv.TaxAmount, tt.want.TaxAmount) {
				t.Errorf("TaxAmount = %f, want %f", inv.TaxAmount, tt.want.TaxAmount)
			}
			if !almostEqualAmount(inv.Total, tt.want.Total) {
				t.Errorf("Total = %f, want %f", inv.Total, tt.want.Total)
			}
			if !almostEqualAmount(inv.AmountDue, tt.want.Total) {
				t.Errorf("AmountDue = %f, want %f", inv.AmountDue, tt.want.Total)
			}
			if inv.Total != inv.Subtotal-inv.DiscountAmount+inv.TaxAmount && !almostEqualAmount(inv.Total, inv.Subtotal-inv.DiscountAmount+inv.TaxAmount) {
				t.Errorf("invariant broken: Subtotal - DiscountAmount + TaxAmount = %f, Total = %f", inv.Subtotal-inv.DiscountAmount+inv.TaxAmount, inv.Total)
			}
		})
	}
}
