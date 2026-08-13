package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/pos_audit/internal/domain"
	"github.com/voxire/onion/internal/platform/cursor"
)

type mockRepo struct {
	list      []domain.AuditLog
	createID  int64
	listErr   error
	createErr error
}

type pagedMockRepo struct {
	mockRepo
	filter domain.AuditLogFilter
}

func (m *pagedMockRepo) List(_ context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, error) {
	m.filter = filter
	return m.list, m.listErr
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

func TestListPageCapsLimitAndProvidesContinuation(t *testing.T) {
	logs := make([]domain.AuditLog, 51)
	for i := range logs {
		logs[i] = domain.AuditLog{ID: int64(100 - i), CreatedAt: time.Date(2026, 8, 13, 12, 0, i, 0, time.UTC)}
	}
	repo := &pagedMockRepo{mockRepo: mockRepo{list: logs}}
	s := &Service{repo: repo}

	page, err := s.ListPage(context.Background(), domain.AuditLogFilter{BusinessID: 1, Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Logs) != 50 {
		t.Fatalf("logs = %d, want 50", len(page.Logs))
	}
	if page.NextCursor == "" {
		t.Fatal("next cursor is empty")
	}
	boundary, err := cursor.DecodeCreatedAtID(page.NextCursor)
	if err != nil {
		t.Fatal(err)
	}
	if boundary.ID != logs[49].ID {
		t.Fatalf("cursor id = %d, want final returned id %d", boundary.ID, logs[49].ID)
	}
	if repo.filter.Limit != 51 {
		t.Fatalf("query limit = %d, want 51", repo.filter.Limit)
	}
}
