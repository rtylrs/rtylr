package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/timesheets/internal/domain"
)

type mockRepo struct {
	list []domain.TimesheetEntry
	err  error
	got  struct {
		owner int64
		week  string
	}
}

func (m *mockRepo) GetTimesheets(_ context.Context, owner int64, week string) ([]domain.TimesheetEntry, error) {
	m.got.owner = owner
	m.got.week = week
	return m.list, m.err
}

func TestGetTimesheets_Delegates(t *testing.T) {
	m := &mockRepo{list: []domain.TimesheetEntry{{}}}
	s := &Service{repo: m}
	if _, err := s.GetTimesheets(context.Background(), 7, "2026-04-27"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if m.got.owner != 7 || m.got.week != "2026-04-27" {
		t.Errorf("repo got %+v", m.got)
	}
}

func TestGetTimesheets_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{err: want}}
	if _, err := s.GetTimesheets(context.Background(), 1, ""); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
