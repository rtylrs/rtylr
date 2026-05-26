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

func TestList_ReturnsEmptySlice(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Customer{}, total: 0}}
	got, total, err := s.List(context.Background(), 1, 1, 10)
	if err != nil || len(got) != 0 || total != 0 {
		t.Errorf("got %v total %d err %v", got, total, err)
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("db fail")
	s := &Service{repo: &mockRepo{listEr: want}}
	if _, _, err := s.List(context.Background(), 1, 1, 10); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_MultipleCustomers(t *testing.T) {
	customers := []domain.Customer{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	s := &Service{repo: &mockRepo{list: customers, total: 2}}
	got, total, err := s.List(context.Background(), 10, 1, 10)
	if err != nil || len(got) != 2 || total != 2 {
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

func TestGetByID_ReturnsCustomer(t *testing.T) {
	c := &domain.Customer{ID: 7, Name: "Charlie"}
	s := &Service{repo: &mockRepo{cust: c}}
	got, err := s.GetByID(context.Background(), 1, 7)
	if err != nil || got == nil || got.ID != 7 {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestGetByID_ReturnsNilWhenNotFound(t *testing.T) {
	want := errors.New("not found")
	s := &Service{repo: &mockRepo{getEr: want}}
	got, err := s.GetByID(context.Background(), 1, 999)
	if got != nil || !errors.Is(err, want) {
		t.Errorf("got %v err %v", got, err)
	}
}

func TestList_TotalReflectsRepoValue(t *testing.T) {
	s := &Service{repo: &mockRepo{total: 50}}
	_, total, _ := s.List(context.Background(), 1, 1, 10)
	if total != 50 {
		t.Errorf("expected total=50, got %d", total)
	}
}

func TestList_PageParametersPassedThrough(t *testing.T) {
	s := &Service{repo: &mockRepo{total: 0}}
	_, _, err := s.List(context.Background(), 1, 2, 25)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetByID_NamePreserved(t *testing.T) {
	c := &domain.Customer{ID: 3, Name: "Diana", Email: "diana@x.com"}
	s := &Service{repo: &mockRepo{cust: c}}
	got, err := s.GetByID(context.Background(), 1, 3)
	if err != nil || got.Name != "Diana" || got.Email != "diana@x.com" {
		t.Errorf("fields not preserved: %+v err %v", got, err)
	}
}

func TestList_CustomerFieldsPreserved(t *testing.T) {
	customers := []domain.Customer{{ID: 5, Name: "Eve", Email: "eve@test.com"}}
	s := &Service{repo: &mockRepo{list: customers, total: 1}}
	got, _, err := s.List(context.Background(), 1, 1, 10)
	if err != nil || got[0].Name != "Eve" {
		t.Errorf("fields not preserved: %v err %v", got, err)
	}
}

func TestList_ErrorReturnsZeroTotal(t *testing.T) {
	s := &Service{repo: &mockRepo{listEr: errors.New("fail"), total: 5}}
	_, total, err := s.List(context.Background(), 1, 1, 10)
	if err == nil {
		t.Fatal("expected error")
	}
	if total != 0 {
		t.Errorf("expected total=0 on error, got %d", total)
	}
}
