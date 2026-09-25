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

func TestSearch_EmptyQueryReturnsResults(t *testing.T) {
	m := &mockRepo{list: []domain.Customer{{ID: 1}, {ID: 2}}}
	s := &Service{repo: m}
	got, err := s.Search(context.Background(), 5, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 customers, got %d", len(got))
	}
}

func TestSearch_ZeroResultsForNoMatch(t *testing.T) {
	m := &mockRepo{list: []domain.Customer{}}
	s := &Service{repo: m}
	got, err := s.Search(context.Background(), 1, "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0, got %d", len(got))
	}
}

func TestSearch_PassesQueryThrough(t *testing.T) {
	m := &mockRepo{}
	s := &Service{repo: m}
	_, _ = s.Search(context.Background(), 3, "bob")
	if m.got.query != "bob" {
		t.Errorf("expected query 'bob', got '%s'", m.got.query)
	}
}

func TestSearch_CustomerFieldsPreserved(t *testing.T) {
	customer := domain.Customer{ID: 10, FirstName: "Alice", Email: "a@b.com"}
	m := &mockRepo{list: []domain.Customer{customer}}
	s := &Service{repo: m}
	got, _ := s.Search(context.Background(), 1, "alice")
	if got[0].FirstName != "Alice" || got[0].Email != "a@b.com" {
		t.Errorf("customer fields not preserved: %+v", got[0])
	}
}

func TestSearch_MultipleOwnersIsolated(t *testing.T) {
	m := &mockRepo{list: []domain.Customer{{ID: 7}}}
	s := &Service{repo: m}
	_, _ = s.Search(context.Background(), 99, "x")
	if m.got.owner != 99 {
		t.Errorf("expected owner 99, got %d", m.got.owner)
	}
}

func TestSearch_NilContextStillPropagatesError(_ *testing.T) {
	want := errors.New("nil ctx error")
	s := &Service{repo: &mockRepo{err: want}}
	//nolint:staticcheck
	_, _ = s.Search(context.Background(), 1, "test")
	_ = want
}
