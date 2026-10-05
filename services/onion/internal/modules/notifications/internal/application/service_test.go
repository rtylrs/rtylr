package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/notifications/internal/domain"
)

type mockRepo struct {
	list      []domain.Notification
	createErr error
	mrErr     error
	maErr     error
	created   domain.Notification

	markReadBusinessID     int64
	markReadUserID         int64
	markReadNotificationID int64
	markAllBusinessID      int64
	markAllUserID          int64
}

func (m *mockRepo) Create(_ context.Context, notification domain.Notification) error {
	m.created = notification
	return m.createErr
}

func (m *mockRepo) List(_ context.Context, _, _ int64) ([]domain.Notification, error) {
	return m.list, nil
}

func (m *mockRepo) MarkRead(_ context.Context, businessID, userID, notificationID int64) error {
	m.markReadBusinessID = businessID
	m.markReadUserID = userID
	m.markReadNotificationID = notificationID
	return m.mrErr
}

func (m *mockRepo) MarkAllRead(_ context.Context, businessID, userID int64) error {
	m.markAllBusinessID = businessID
	m.markAllUserID = userID
	return m.maErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.Notification{{}}}}
	got, _ := s.List(context.Background(), 1, 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_ForwardsTenantAndRecipient(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	want := domain.Notification{
		OwnerID: 10, BusinessID: 20, UserID: 30,
		Title: "Access request", Body: "Review", Type: "info", Link: "/team",
	}
	if err := s.Create(context.Background(), want); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo.created != want {
		t.Errorf("Create forwarded %+v, want %+v", repo.created, want)
	}
}

func TestMarkRead_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{mrErr: want}}
	if err := s.MarkRead(context.Background(), 1, 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestMarkAllRead_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.MarkAllRead(context.Background(), 1, 1); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

// The tenant resolved by the handler has to reach the statement: without it
// the update matches the caller's rows in every business they belong to.
func TestMarkRead_ForwardsBusinessScope(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	if err := s.MarkRead(context.Background(), 42, 7, 5); err != nil {
		t.Fatalf("MarkRead: %v", err)
	}
	if repo.markReadBusinessID != 42 || repo.markReadUserID != 7 || repo.markReadNotificationID != 5 {
		t.Errorf("MarkRead called with business=%d user=%d notification=%d, want 42/7/5",
			repo.markReadBusinessID, repo.markReadUserID, repo.markReadNotificationID)
	}
}

func TestMarkAllRead_ForwardsBusinessScope(t *testing.T) {
	repo := &mockRepo{}
	s := &Service{repo: repo}
	if err := s.MarkAllRead(context.Background(), 42, 7); err != nil {
		t.Fatalf("MarkAllRead: %v", err)
	}
	if repo.markAllBusinessID != 42 || repo.markAllUserID != 7 {
		t.Errorf("MarkAllRead called with business=%d user=%d, want 42/7",
			repo.markAllBusinessID, repo.markAllUserID)
	}
}
