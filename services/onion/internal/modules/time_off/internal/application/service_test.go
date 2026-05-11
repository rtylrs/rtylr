package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
)

type mockRepo struct {
	list      []domain.TimeOffRequest
	createID  int64
	listErr   error
	createErr error
	apvErr    error
}

func (m *mockRepo) ListTimeOffRequests(_ context.Context, _ int64) ([]domain.TimeOffRequest, error) {
	return m.list, m.listErr
}

func (m *mockRepo) CreateTimeOffRequest(_ context.Context, _ domain.TimeOffInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) ApproveTimeOffRequest(_ context.Context, _ int64, _ domain.TimeOffApprovalInput) error {
	return m.apvErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.TimeOffRequest{{}}}}
	got, _ := s.ListTimeOffRequests(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestSubmit_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.SubmitTimeOffRequest(context.Background(), domain.TimeOffInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestApprove_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{apvErr: want}}
	if err := s.ApproveTimeOffRequest(context.Background(), 1, domain.TimeOffApprovalInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
