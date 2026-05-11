package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/expenses/internal/domain"
)

type mockRepo struct {
	listResult []domain.Expense
	catResult  []domain.ExpenseCategory
	createID   int64
	listErr    error
	createErr  error
	updErr     error
	delErr     error
	catErr     error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Expense, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) Create(_ context.Context, _ domain.ExpenseInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.ExpenseUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func (m *mockRepo) ListCategories(_ context.Context, _ int64) ([]domain.ExpenseCategory, error) {
	return m.catResult, m.catErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Expense{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.ExpenseInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.ExpenseUpdateInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestListCategories_Delegates(t *testing.T) {
	want := []domain.ExpenseCategory{{ID: 1, Name: "Travel"}}
	s := &Service{repo: &mockRepo{catResult: want}}
	got, _ := s.ListCategories(context.Background(), 1)
	if len(got) != 1 || got[0].Name != "Travel" {
		t.Errorf("got %v", got)
	}
}
