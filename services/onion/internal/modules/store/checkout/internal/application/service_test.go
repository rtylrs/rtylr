package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/store/checkout/internal/domain"
)

type mockRepo struct {
	createID  int64
	createErr error
	compErr   error
}

func (m *mockRepo) Create(_ context.Context, _, _ int64, _ domain.CheckoutCreateInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Complete(_ context.Context, _, _, _ int64) error {
	return m.compErr
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, 1, domain.CheckoutCreateInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestComplete_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{compErr: want}}
	if err := s.Complete(context.Background(), 1, 5, 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
