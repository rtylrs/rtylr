package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
)

type mockRepo struct {
	list []domain.Customer
	err  error
	got  struct {
		owner int64
		query string
	}
}

func (m *mockRepo) Search(_ context.Context, owner int64, query string) ([]domain.Customer, error) {
	m.got.owner = owner
	m.got.query = query
	return m.list, m.err
}

func TestSearch_Delegates(t *testing.T) {
	m := &mockRepo{list: []domain.Customer{{}}}
	s := &Service{repo: m}
	if _, err := s.Search(context.Background(), 7, "alice"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if m.got.owner != 7 || m.got.query != "alice" {
		t.Errorf("got %+v", m.got)
	}
}

func TestSearch_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.Search(context.Background(), 1, ""); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
