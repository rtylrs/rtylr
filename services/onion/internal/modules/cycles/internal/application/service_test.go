package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
)

type mockRepo struct {
	list      []domain.Cycle
	createID  int64
	createErr error
	updErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64, _ *int64) ([]domain.Cycle, error) {
	return m.list, nil
}
func (m *mockRepo) Create(_ context.Context, _ domain.CycleInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) Update(_ context.Context, _ domain.CycleUpdateInput) error {
	return m.updErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Cycle{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1, nil)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.CycleInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.CycleUpdateInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
