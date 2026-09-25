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

func TestList_ReturnsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.TimeOffRequest{}}}
	got, err := s.ListTimeOffRequests(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestList_ReturnsMultiple(t *testing.T) {
	reqs := []domain.TimeOffRequest{
		{ID: 1, StaffID: 10, Type: "vacation"},
		{ID: 2, StaffID: 11, Type: "sick"},
	}
	s := &Service{repo: &mockRepo{list: reqs}}
	got, err := s.ListTimeOffRequests(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2, got %d", len(got))
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &Service{repo: &mockRepo{listErr: want}}
	_, err := s.ListTimeOffRequests(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestSubmit_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 42}}
	id, err := s.SubmitTimeOffRequest(context.Background(), domain.TimeOffInput{OwnerID: 1, StaffID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestSubmit_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("fail")}}
	id, _ := s.SubmitTimeOffRequest(context.Background(), domain.TimeOffInput{})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestApprove_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.ApproveTimeOffRequest(context.Background(), 1, domain.TimeOffApprovalInput{ID: 5, Status: "approved"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubmit_VacationType(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 10}}
	id, err := s.SubmitTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID: 1, StaffID: 2, Type: "vacation", DaysCount: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 10 {
		t.Errorf("expected 10, got %d", id)
	}
}

func TestApprove_WithNotes(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	err := s.ApproveTimeOffRequest(context.Background(), 1, domain.TimeOffApprovalInput{
		ID:         3,
		Status:     "approved",
		ApprovedBy: 99,
		Notes:      "approved by manager",
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestList_PreservesRequestData(t *testing.T) {
	reqs := []domain.TimeOffRequest{
		{ID: 7, StaffID: 20, Type: "sick", DaysCount: 2.0, Status: "pending"},
	}
	s := &Service{repo: &mockRepo{list: reqs}}
	got, _ := s.ListTimeOffRequests(context.Background(), 1)
	if got[0].Type != "sick" || got[0].DaysCount != 2.0 {
		t.Errorf("data mismatch: %+v", got[0])
	}
}

func TestSubmit_SickType(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 15}}
	id, err := s.SubmitTimeOffRequest(context.Background(), domain.TimeOffInput{
		OwnerID: 1, StaffID: 3, Type: "sick", DaysCount: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 15 {
		t.Errorf("expected 15, got %d", id)
	}
}
