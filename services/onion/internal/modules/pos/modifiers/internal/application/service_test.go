package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/modifiers/internal/domain"
)

type mockModifierRepo struct {
	list      []domain.Modifier
	createID  int64
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockModifierRepo) List(_ context.Context, _ int64) ([]domain.Modifier, error) {
	return m.list, nil
}

func (m *mockModifierRepo) Create(_ context.Context, _, _ int64, _ domain.ModifierInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockModifierRepo) Update(_ context.Context, _, _ int64, _ domain.ModifierInput) error {
	return m.updateErr
}

func (m *mockModifierRepo) Delete(_ context.Context, _, _ int64) error {
	return m.deleteErr
}

func TestModifierCreate_RejectsBlankName(t *testing.T) {
	s := &Service{repo: &mockModifierRepo{createID: 1}}
	_, err := s.Create(context.Background(), 1, 2, domain.ModifierInput{})
	if err == nil {
		t.Fatal("expected error for blank name")
	}
}

func TestModifierCreate_AcceptsValid(t *testing.T) {
	s := &Service{repo: &mockModifierRepo{createID: 5}}
	id, err := s.Create(context.Background(), 1, 2, domain.ModifierInput{Name: "Extra Cheese", PriceDelta: 1.50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 5 {
		t.Errorf("expected id=5, got %d", id)
	}
}

func TestModifierCreate_PropagatesRepoError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockModifierRepo{createErr: want}}
	_, err := s.Create(context.Background(), 1, 2, domain.ModifierInput{Name: "Valid"})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestModifierList_Delegates(t *testing.T) {
	s := &Service{repo: &mockModifierRepo{list: []domain.Modifier{{ID: 1}}}}
	got, _ := s.List(context.Background(), 2)
	if len(got) != 1 {
		t.Errorf("expected 1 modifier, got %d", len(got))
	}
}

func TestModifierDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete error")
	s := &Service{repo: &mockModifierRepo{deleteErr: want}}
	if err := s.Delete(context.Background(), 2, 3); !errors.Is(err, want) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
