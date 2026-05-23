package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/support_customers/internal/domain"
)

type mockRepo struct {
	list   []domain.Customer
	total  int
	cust   *domain.Customer
	listEr error
	getEr  error
}

func (m *mockRepo) List(_ context.Context, _ int64, _, _ int) ([]domain.Customer, int, error) {
	return m.list, m.total, m.listEr
}

func (m *mockRepo) GetByID(_ context.Context, _, _ int64) (*domain.Customer, error) {
	return m.cust, m.getEr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Customer{{}}, total: 1}}
	got, total, err := s.List(context.Background(), 1, 1, 10)
	if err != nil || len(got) != 1 || total != 1 {
		t.Errorf("got %v total %d err %v", got, total, err)
	}
}

func TestGetByID_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{getEr: want}}
	if _, err := s.GetByID(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
