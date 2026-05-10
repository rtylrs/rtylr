package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_floors/internal/domain"
)

type mockRepo struct {
	list      []domain.Floor
	createID  int64
	createErr error
	updErr    error
	delErr    error
}

func (m *mockRepo) List(_ context.Context, _, _ int64) ([]domain.Floor, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ domain.FloorInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.FloorUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Floor{{}}}}
	got, _ := s.List(context.Background(), 1, 2)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.FloorInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.FloorUpdateInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
