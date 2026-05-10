package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/billing/internal/domain"
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
