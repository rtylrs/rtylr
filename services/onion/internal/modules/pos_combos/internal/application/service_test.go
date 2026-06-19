package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
)

type mockRepo struct {
	list      []domain.Combo
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Combo, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _, _ int64, _ domain.ComboInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.ComboInput) error {
	return m.updateErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.deleteErr
}

func TestCreate_RejectsBlankName(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 1}}
	_, err := s.Create(context.Background(), 1, 1, domain.ComboInput{Price: 10})
	if err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestCreate_RejectsNegativePrice(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 1}}
	_, err := s.Create(context.Background(), 1, 1, domain.ComboInput{Name: "Meal", Price: -5})
	if err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{createErr: want}}
	_, err := s.Create(context.Background(), 1, 1, domain.ComboInput{Name: "Valid", Price: 10})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Combo{{ID: 1, Name: "Meal"}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("expected 1 combo, got %d", len(got))
	}
}

func TestDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete error")
	s := &Service{repo: &mockRepo{deleteErr: want}}
	if err := s.Delete(context.Background(), 1, 2); !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
