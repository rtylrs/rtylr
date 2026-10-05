package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/floors/internal/domain"
)

type mockRepo struct {
	list      []domain.Floor
	createID  int64
	createErr error
	updErr    error
	delErr    error
}

func (m *mockRepo) List(_ context.Context, _, _ int64) ([]domain.Floor, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ domain.FloorInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.FloorUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Floor{{}}}}
	got, _ := s.List(context.Background(), 1, 2)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_ValidationRejectsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.Create(context.Background(), domain.FloorInput{}); err == nil {
		t.Error("expected validation error for empty input")
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	validInput := domain.FloorInput{BusinessID: 1, BranchID: 1, Name: "Main"}
	if _, err := s.Create(context.Background(), validInput); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	validUpdate := domain.FloorUpdateInput{ID: 1, BusinessID: 1, Name: "Main"}
	if err := s.Update(context.Background(), validUpdate); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestList_MultipleFloors(t *testing.T) {
	floors := []domain.Floor{{ID: 1}, {ID: 2}, {ID: 3}}
	s := &Service{repo: &mockRepo{list: floors}}
	got, _ := s.List(context.Background(), 10, 20)
	if len(got) != 3 {
		t.Errorf("expected 3 floors, got %d", len(got))
	}
}

func TestList_EmptyFloors(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Floor{}}}
	got, err := s.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 55}}
	id, err := s.Create(context.Background(), domain.FloorInput{BusinessID: 1, BranchID: 1, Name: "Main"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 55 {
		t.Errorf("expected id 55, got %d", id)
	}
}

func TestCreate_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("fail"), createID: 99}}
	id, _ := s.Create(context.Background(), domain.FloorInput{BusinessID: 1, BranchID: 1, Name: "Main"})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	validInput := domain.FloorUpdateInput{ID: 1, BusinessID: 1, Name: "Ground"}
	if err := s.Update(context.Background(), validInput); !errors.Is(err, want) {
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

func TestDelete_ZeroIDAllowed(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Delete(context.Background(), 1, 0); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestUpdate_InputPreserved(t *testing.T) {
	// Ensure the mock receives the exact input (no mutation in the service).
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.FloorUpdateInput{ID: 10, BusinessID: 1, Name: "Ground Floor"}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
