package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_tables/internal/domain"
)

type mockRepo struct {
	list      []domain.Table
	createID  int64
	createErr error
	updErr    error
	statusErr error
	delErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Table, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ domain.TableInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.TableUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) UpdateStatus(_ context.Context, _ domain.TableStatusInput) error {
	return m.statusErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Table{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_ValidationRejectsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.Create(context.Background(), domain.TableInput{}); err == nil {
		t.Error("expected validation error for empty input")
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	validInput := domain.TableInput{OwnerID: 1, Label: "T1"}
	if _, err := s.Create(context.Background(), validInput); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	validUpdate := domain.TableUpdateInput{ID: 1, OwnerID: 1, Label: "T1"}
	if err := s.Update(context.Background(), validUpdate); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestList_EmptyTables(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Table{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestList_MultipleTables(t *testing.T) {
	tables := []domain.Table{{ID: 1}, {ID: 2}, {ID: 3}}
	s := &Service{repo: &mockRepo{list: tables}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 3 {
		t.Errorf("expected 3, got %d", len(got))
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 77}}
	id, err := s.Create(context.Background(), domain.TableInput{OwnerID: 1, Label: "T1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 77 {
		t.Errorf("expected id 77, got %d", id)
	}
}

func TestCreate_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("fail"), createID: 5}}
	id, _ := s.Create(context.Background(), domain.TableInput{OwnerID: 1, Label: "T1"})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestUpdate_ValidationRejectsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.TableUpdateInput{}); err == nil {
		t.Error("expected validation error for empty input")
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	validInput := domain.TableUpdateInput{ID: 1, OwnerID: 1, Label: "T1"}
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

func TestUpdate_WithStatusChange(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.TableUpdateInput{ID: 1, OwnerID: 1, Label: "T1", Status: "occupied"}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestCreate_WithLocationID(t *testing.T) {
	locID := int64(5)
	s := &Service{repo: &mockRepo{createID: 10}}
	id, err := s.Create(context.Background(), domain.TableInput{OwnerID: 1, LocationID: &locID, Label: "T2"})
	if err != nil || id != 10 {
		t.Errorf("unexpected: id=%d err=%v", id, err)
	}
}

func TestDelete_ZeroIDAllowed(t *testing.T) {
	// Service has no guard on zero ID; repo decides.
	s := &Service{repo: &mockRepo{}}
	if err := s.Delete(context.Background(), 1, 0); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
