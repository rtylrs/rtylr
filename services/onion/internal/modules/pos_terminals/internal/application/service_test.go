package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_terminals/internal/domain"
)

type mockRepo struct {
	list      []domain.Terminal
	createID  int64
	createErr error
	updErr    error
	hbErr     error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Terminal, error) {
	return m.list, nil
}

func (m *mockRepo) Create(_ context.Context, _ domain.TerminalInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _ domain.TerminalUpdateInput) error {
	return m.updErr
}

func (m *mockRepo) Heartbeat(_ context.Context, _, _ int64) error {
	return m.hbErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Terminal{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_ValidationRejectsEmpty(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.Create(context.Background(), domain.TerminalInput{}); err == nil {
		t.Error("expected validation error for empty input")
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	validInput := domain.TerminalInput{BusinessID: 1, BranchID: 1, Name: "POS-1"}
	if _, err := s.Create(context.Background(), validInput); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndHeartbeat_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	validUpdate := domain.TerminalUpdateInput{ID: 1, BusinessID: 1, Name: "POS-1"}
	if err := s.Update(context.Background(), validUpdate); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Heartbeat(context.Background(), 5, 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestList_MultipleTerminals(t *testing.T) {
	terminals := []domain.Terminal{{ID: 1}, {ID: 2}}
	s := &Service{repo: &mockRepo{list: terminals}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 2 {
		t.Errorf("expected 2 terminals, got %d", len(got))
	}
}

func TestList_EmptyTerminals(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Terminal{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 88}}
	id, err := s.Create(context.Background(), domain.TerminalInput{BusinessID: 1, BranchID: 1, Name: "POS-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 88 {
		t.Errorf("expected id 88, got %d", id)
	}
}

func TestCreate_ZeroIDOnError(t *testing.T) {
	s := &Service{repo: &mockRepo{createErr: errors.New("fail"), createID: 50}}
	id, _ := s.Create(context.Background(), domain.TerminalInput{BusinessID: 1, BranchID: 1, Name: "POS-1"})
	if id != 0 {
		t.Errorf("expected 0 on error, got %d", id)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	validInput := domain.TerminalUpdateInput{ID: 1, BusinessID: 1, Name: "POS-1"}
	if err := s.Update(context.Background(), validInput); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestHeartbeat_PropagatesError(t *testing.T) {
	want := errors.New("heartbeat fail")
	s := &Service{repo: &mockRepo{hbErr: want}}
	if err := s.Heartbeat(context.Background(), 1, 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_WithDeviceID(t *testing.T) {
	devID := "device-abc"
	s := &Service{repo: &mockRepo{createID: 5}}
	id, err := s.Create(context.Background(), domain.TerminalInput{
		BusinessID: 1,
		BranchID:   1,
		Name:       "Terminal-2",
		DeviceID:   &devID,
	})
	if err != nil || id != 5 {
		t.Errorf("unexpected: id=%d err=%v", id, err)
	}
}

func TestUpdate_IsActiveFlag(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), domain.TerminalUpdateInput{ID: 1, BusinessID: 1, Name: "POS-1", IsActive: false}); err != nil {
		t.Errorf("deactivating terminal should succeed: %v", err)
	}
}

func TestHeartbeat_ZeroTerminalIDAllowed(t *testing.T) {
	// Repo decides if zero ID is valid; service passes it through.
	s := &Service{repo: &mockRepo{}}
	if err := s.Heartbeat(context.Background(), 0, 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}
