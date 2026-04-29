package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/shifts/internal/domain"
)

type mockRepo struct {
	list      []domain.Shift
	createID  int64
	createErr error
	updErr    error
	delErr    error
}

func (m *mockRepo) ListShifts(_ context.Context, _, _ int64, _ string) ([]domain.Shift, error) {
	return m.list, nil
}
func (m *mockRepo) CreateShift(_ context.Context, _ domain.ShiftInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) UpdateShift(_ context.Context, _ int64, _ domain.ShiftUpdateInput) error {
	return m.updErr
}
func (m *mockRepo) DeleteShift(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestListShifts_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Shift{{}}}}
	got, _ := s.ListShifts(context.Background(), 1, 2, "2026-04-28")
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreateShift_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.CreateShift(context.Background(), domain.ShiftInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateShift(context.Background(), 1, domain.ShiftUpdateInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.DeleteShift(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
