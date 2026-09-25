package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/billing/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

type mockRepo struct {
	summary  domain.BillingSummary
	invoices []domain.Invoice
	sErr     error
	iErr     error
}

func (m *mockRepo) GetSummary(_ context.Context, _ int64) (domain.BillingSummary, error) {
	return m.summary, m.sErr
}

func (m *mockRepo) ListInvoices(_ context.Context, _ int64) ([]domain.Invoice, error) {
	return m.invoices, m.iErr
}

func TestGetSummary_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.GetSummary(context.Background(), 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestListInvoices_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{iErr: want}}
	if _, err := s.ListInvoices(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetSummary_ReturnsSummaryFields(t *testing.T) {
	summary := domain.BillingSummary{
		Plan:         "pro",
		Status:       "active",
		CurrentUsage: 42.5,
		UsageLimit:   100.0,
	}
	s := &Service{repo: &mockRepo{summary: summary}}
	got, err := s.GetSummary(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.Plan != "pro" || got.Status != "active" {
		t.Errorf("summary fields wrong: %+v", got)
	}
	if got.CurrentUsage != 42.5 || got.UsageLimit != 100.0 {
		t.Errorf("usage fields wrong: %+v", got)
	}
}

func TestGetSummary_PropagatesError(t *testing.T) {
	want := errors.New("summary fail")
	s := &Service{repo: &mockRepo{sErr: want}}
	if _, err := s.GetSummary(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListInvoices_Empty(t *testing.T) {
	s := &Service{repo: &mockRepo{invoices: []domain.Invoice{}}}
	got, err := s.ListInvoices(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestListInvoices_Delegates(t *testing.T) {
	invoices := []domain.Invoice{
		{ID: 1, Amount: money.Amount(9999), Currency: "USD", Status: "paid"},
		{ID: 2, Amount: money.Amount(14900), Currency: "USD", Status: "pending"},
	}
	s := &Service{repo: &mockRepo{invoices: invoices}}
	got, err := s.ListInvoices(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Errorf("wrong order or IDs: %v", got)
	}
}

func TestGetSummary_WithNextBillingDate(t *testing.T) {
	now := time.Now()
	s := &Service{repo: &mockRepo{summary: domain.BillingSummary{
		Plan:            "enterprise",
		NextBillingDate: &now,
	}}}
	got, err := s.GetSummary(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.NextBillingDate == nil {
		t.Error("expected NextBillingDate to be set")
	}
}

func TestListInvoices_PreservesOrder(t *testing.T) {
	invoices := []domain.Invoice{
		{ID: 3, Status: "paid"},
		{ID: 1, Status: "pending"},
		{ID: 2, Status: "overdue"},
	}
	s := &Service{repo: &mockRepo{invoices: invoices}}
	got, _ := s.ListInvoices(context.Background(), 1)
	if got[0].ID != 3 || got[1].ID != 1 || got[2].ID != 2 {
		t.Errorf("order changed: %v", got)
	}
}

func TestGetSummary_ZeroValue(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	got, err := s.GetSummary(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.Plan != "" || got.CurrentUsage != 0 {
		t.Errorf("expected zero value summary: %+v", got)
	}
}
