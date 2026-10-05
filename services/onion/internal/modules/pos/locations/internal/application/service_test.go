package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/locations/internal/domain"
)

type mockRepo struct {
	list      []domain.Location
	createID  int64
	createErr error
	updErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Location, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ domain.LocationInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.LocationUpdateInput) error {
	return m.updErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Location{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_ValidationRejectsBlankInput(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.Create(context.Background(), domain.LocationInput{}); err == nil {
		t.Error("expected validation error for empty input")
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	validInput := domain.LocationInput{OwnerID: 1, Name: "HQ"}
	if _, err := s.Create(context.Background(), validInput); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdate_ValidationRejectsBlankInput(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.LocationUpdateInput{}); err == nil {
		t.Error("expected validation error for empty input")
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{updErr: want}}
	validInput := domain.LocationUpdateInput{ID: 1, OwnerID: 1, Name: "HQ"}
	if err := s.Update(context.Background(), validInput); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
