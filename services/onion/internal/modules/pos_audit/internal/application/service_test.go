package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
)

type mockRepo struct {
	list      []domain.AuditLog
	createID  int64
	listErr   error
	createErr error
}

func (m *mockRepo) List(_ context.Context, _ domain.AuditLogFilter) ([]domain.AuditLog, error) {
	return m.list, m.listErr
}

func (m *mockRepo) Create(_ context.Context, _ domain.AuditLogInput) (int64, error) {
	return m.createID, m.createErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.AuditLog{{}}}}
	got, _ := s.List(context.Background(), domain.AuditLogFilter{})
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.AuditLogInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
