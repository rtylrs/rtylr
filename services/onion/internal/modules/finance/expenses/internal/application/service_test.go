package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/expenses/internal/domain"
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

func (m *mockRepo) ListPaginated(_ context.Context, _ int64, _, _ int32) ([]domain.Expense, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) GetByID(_ context.Context, _, _ int64) (domain.Expense, error) {
	if len(m.listResult) > 0 {
		return m.listResult[0], m.listErr
	}
	return domain.Expense{}, m.listErr
}

func (m *mockRepo) Create(_ context.Context, _ domain.ExpenseInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.ExpenseUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) UpdateStatus(_ context.Context, _, _ int64, _ string) error {
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

func TestList_EmptyReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Expense{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0, got %d", len(got))
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("list fail")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.List(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 55}}
	id, err := s.Create(context.Background(), domain.ExpenseInput{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if id != 55 {
		t.Errorf("expected 55, got %d", id)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.Update(context.Background(), domain.ExpenseUpdateInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := &Service{repo: &mockRepo{delErr: want}}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateStatus_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.UpdateStatus(context.Background(), 1, 5, "approved"); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestUpdateStatus_PropagatesError(t *testing.T) {
	want := errors.New("status fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.UpdateStatus(context.Background(), 1, 5, "approved"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetByID_ReturnsFirstResult(t *testing.T) {
	exp := domain.Expense{ID: 7, Amount: 100}
	s := &Service{repo: &mockRepo{listResult: []domain.Expense{exp}}}
	got, err := s.GetByID(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.ID != 7 {
		t.Errorf("expected id=7, got %d", got.ID)
	}
}
