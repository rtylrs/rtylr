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
