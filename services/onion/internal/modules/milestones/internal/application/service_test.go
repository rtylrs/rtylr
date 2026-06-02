package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
)

type mockRepo struct {
	list          []domain.Milestone
	createID      int64
	createErr     error
	updErr        error
	getProjectErr error
}

func (m *mockRepo) List(_ context.Context, _ int64, _ *int64) ([]domain.Milestone, error) {
	return m.list, nil
}

func (m *mockRepo) GetProject(_ context.Context, _ int64, _ int64) (int64, error) {
	return 1, m.getProjectErr
}

func (m *mockRepo) Create(_ context.Context, _ domain.MilestoneInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.MilestoneUpdateInput) error {
	return m.updErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Milestone{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1, nil)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.MilestoneInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_RejectsUnownedProject(t *testing.T) {
	s := &Service{repo: &mockRepo{getProjectErr: domain.ErrProjectNotFound}}
	if _, err := s.Create(context.Background(), domain.MilestoneInput{ProjectID: 999}); !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.MilestoneUpdateInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
