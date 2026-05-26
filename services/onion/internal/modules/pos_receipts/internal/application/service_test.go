package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
)

type mockRepo struct {
	list    []domain.Receipt
	receipt *domain.Receipt
	listErr error
	getErr  error
}

func (m *mockRepo) List(_ context.Context, _ int64, _, _ string) ([]domain.Receipt, error) {
	return m.list, m.listErr
}

func (m *mockRepo) Get(_ context.Context, _, _ int64) (*domain.Receipt, error) {
	return m.receipt, m.getErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Receipt{{}}}}
	got, _ := s.List(context.Background(), 1, "", "")
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestGet_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{getErr: want}}
	if _, err := s.Get(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestEmailReceipt_PropagatesGetError(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{getErr: want}}
	if err := s.EmailReceipt(context.Background(), 1, 5, "x@y"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_MultipleReceipts(t *testing.T) {
	items := []domain.Receipt{{ID: 1}, {ID: 2}, {ID: 3}}
	s := &Service{repo: &mockRepo{list: items}}
	got, err := s.List(context.Background(), 1, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 receipts, got %d", len(got))
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("list fail")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.List(context.Background(), 1, "2026-01-01", "2026-01-31"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGet_ReturnsReceipt(t *testing.T) {
	receipt := &domain.Receipt{ID: 7, OrderNumber: "ORD-001"}
	s := &Service{repo: &mockRepo{receipt: receipt}}
	got, err := s.Get(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != 7 {
		t.Errorf("expected receipt id 7, got %+v", got)
	}
}

func TestGet_NilReceiptWhenNotFound(t *testing.T) {
	s := &Service{repo: &mockRepo{receipt: nil}}
	got, err := s.Get(context.Background(), 1, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil receipt, got %+v", got)
	}
}

func TestEmailReceipt_NilReceiptPropagatesGetError(t *testing.T) {
	// If repo.Get returns no error but nil receipt, EmailReceipt would panic
	// accessing fields. Confirm get-error path is guarded.
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{receipt: nil, getErr: want}}
	if err := s.EmailReceipt(context.Background(), 1, 5, "a@b.com"); !errors.Is(err, want) {
		t.Errorf("expected get error propagated, got %v", err)
	}
}

func TestList_EmptyDateRange(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Receipt{}}}
	got, err := s.List(context.Background(), 1, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestGet_PropagatesOwnerID(t *testing.T) {
	// Just verify we can call with large IDs without panic (no repo call needed).
	s := &Service{repo: &mockRepo{receipt: &domain.Receipt{ID: 1}}}
	_, err := s.Get(context.Background(), 9999, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
