package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/reviews/internal/domain"
)

type mockRepo struct {
	list      []domain.Review
	createID  int64
	createErr error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Review, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.ReviewInput) (int64, error) {
	return m.createID, m.createErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Review{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.ReviewInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
