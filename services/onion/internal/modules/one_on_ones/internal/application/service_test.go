package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/one_on_ones/internal/domain"
)

type mockRepo struct {
	list      []domain.OneOnOne
	createID  int64
	createErr error
	listErr   error
	gotOwner  int64
	gotInput  domain.OneOnOneInput
}

func (m *mockRepo) List(_ context.Context, ownerID int64) ([]domain.OneOnOne, error) {
	m.gotOwner = ownerID
	return m.list, m.listErr
}

func (m *mockRepo) Create(_ context.Context, _ int64, input domain.OneOnOneInput) (int64, error) {
	m.gotInput = input
	return m.createID, m.createErr
}

// ─── List ────────────────────────────────────────────────────────────────────

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.OneOnOne{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestList_EmptyReturnsEmptySlice(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.OneOnOne{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestList_PropagatesOwnerID(t *testing.T) {
	m := &mockRepo{list: []domain.OneOnOne{}}
	s := &Service{repo: m}
	_, _ = s.List(context.Background(), 99)
	if m.gotOwner != 99 {
		t.Errorf("expected ownerID 99, got %d", m.gotOwner)
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.List(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_MultipleItems(t *testing.T) {
	items := []domain.OneOnOne{{ID: 1}, {ID: 2}, {ID: 3}}
	s := &Service{repo: &mockRepo{list: items}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 3 {
		t.Errorf("expected 3 items, got %d", len(got))
	}
}

// ─── Create ──────────────────────────────────────────────────────────────────

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.OneOnOneInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 42}}
	id, err := s.Create(context.Background(), 1, domain.OneOnOneInput{ManagerID: 2, StaffID: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected id 42, got %d", id)
	}
}

func TestCreate_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("fail"), createID: 99}}
	id, err := s.Create(context.Background(), 1, domain.OneOnOneInput{})
	if err == nil {
		t.Fatal("expected error")
	}
	if id != 0 {
		t.Errorf("expected id 0 on error, got %d", id)
	}
}

func TestCreate_InputPassedThrough(t *testing.T) {
	m := &mockRepo{createID: 1}
	s := &Service{repo: m}
	input := domain.OneOnOneInput{ManagerID: 10, StaffID: 20, Status: "scheduled"}
	_, _ = s.Create(context.Background(), 1, input)
	if m.gotInput.ManagerID != 10 || m.gotInput.StaffID != 20 {
		t.Errorf("input not passed through: %+v", m.gotInput)
	}
}

func TestCreate_ZeroOwnerIDAllowed(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 1}}
	id, err := s.Create(context.Background(), 0, domain.OneOnOneInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}
}
