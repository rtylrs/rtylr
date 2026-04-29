package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/bills/internal/domain"
)

type mockRepo struct {
	listResult []domain.Bill
	getResult  domain.Bill
	createID   int64
	listErr    error
	getErr     error
	delErr     error
	createErr  error
	updErr     error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Bill, error) {
	return m.listResult, m.listErr
}
func (m *mockRepo) Get(_ context.Context, _, _ int64) (domain.Bill, error) {
	return m.getResult, m.getErr
}
func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}
func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.BillInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.BillInput) error {
	return m.updErr
}

func TestList_Delegates(t *testing.T) {
	want := []domain.Bill{{ID: 1}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestGet_Delegates(t *testing.T) {
	want := domain.Bill{ID: 7}
	s := &Service{repo: &mockRepo{getResult: want}}
	got, _ := s.GetBill(context.Background(), 1, 7)
	if got.ID != 7 {
		t.Errorf("got %+v", got)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.BillInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.BillInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteBill(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
