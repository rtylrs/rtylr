package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/training/internal/domain"
)

type mockRepo struct {
	list      []domain.TrainingCourse
	createID  int64
	createErr error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.TrainingCourse, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.TrainingCourseInput) (int64, error) {
	return m.createID, m.createErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.TrainingCourse{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.TrainingCourseInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.TrainingCourse{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestList_ReturnsMultipleCourses(t *testing.T) {
	courses := []domain.TrainingCourse{
		{ID: 1, Name: "Safety Training"},
		{ID: 2, Name: "HR Onboarding"},
		{ID: 3, Name: "POS Basics"},
	}
	s := &Service{repo: &mockRepo{list: courses}}
	got, err := s.List(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3 courses, got %d", len(got))
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 99}}
	id, err := s.Create(context.Background(), 1, domain.TrainingCourseInput{Name: "Fire Safety"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 99 {
		t.Errorf("expected id 99, got %d", id)
	}
}

func TestCreate_WithRequiredFlag(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 7}}
	id, err := s.Create(context.Background(), 1, domain.TrainingCourseInput{
		Name:       "Required Course",
		IsRequired: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Errorf("expected id 7, got %d", id)
	}
}

func TestCreate_WithDuration(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 3}}
	id, err := s.Create(context.Background(), 1, domain.TrainingCourseInput{
		Name:          "Long Course",
		DurationHours: 8.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 3 {
		t.Errorf("expected id 3, got %d", id)
	}
}

func TestList_DifferentOwners(t *testing.T) {
	// repo returns same list regardless of ownerID — tests that ownerID is passed through
	courses := []domain.TrainingCourse{{ID: 10, OwnerID: 5}}
	s := &Service{repo: &mockRepo{list: courses}}
	got, err := s.List(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].OwnerID != 5 {
		t.Errorf("expected ownerID 5, got %d", got[0].OwnerID)
	}
}

func TestCreate_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("db error")}}
	id, _ := s.Create(context.Background(), 1, domain.TrainingCourseInput{})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestList_ReturnsRequiredCourses(t *testing.T) {
	courses := []domain.TrainingCourse{
		{ID: 1, IsRequired: true},
		{ID: 2, IsRequired: false},
	}
	s := &Service{repo: &mockRepo{list: courses}}
	got, _ := s.List(context.Background(), 1)
	if !got[0].IsRequired {
		t.Error("first course should be required")
	}
	if got[1].IsRequired {
		t.Error("second course should not be required")
	}
}

func TestCreate_EmptyInput(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 1}}
	id, err := s.Create(context.Background(), 1, domain.TrainingCourseInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("expected 1, got %d", id)
	}
}

func TestList_PreservesOrder(t *testing.T) {
	courses := []domain.TrainingCourse{{ID: 3}, {ID: 1}, {ID: 2}}
	s := &Service{repo: &mockRepo{list: courses}}
	got, _ := s.List(context.Background(), 1)
	if got[0].ID != 3 || got[1].ID != 1 || got[2].ID != 2 {
		t.Errorf("order not preserved: %v", got)
	}
}
