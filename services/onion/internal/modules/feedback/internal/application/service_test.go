package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/feedback/internal/domain"
)

type mockRepo struct {
	listResult []domain.Feedback
	listTotal  int
	listErr    error
	createID   int64
	createErr  error
	listGot    struct {
		ownerID  int64
		filters  domain.ListFilters
		page     int
		pageSize int
	}
}

func (m *mockRepo) List(_ context.Context, ownerID int64, filters domain.ListFilters, page, pageSize int) ([]domain.Feedback, int, error) {
	m.listGot.ownerID = ownerID
	m.listGot.filters = filters
	m.listGot.page = page
	m.listGot.pageSize = pageSize
	return m.listResult, m.listTotal, m.listErr
}

func (m *mockRepo) CustomerExistsInBusiness(_ context.Context, _, _ int64) bool { return true }

func (m *mockRepo) Create(_ context.Context, _ domain.FeedbackInput) (int64, error) {
	return m.createID, m.createErr
}

func TestList_PassesPaginationAndFilters(t *testing.T) {
	m := &mockRepo{listResult: []domain.Feedback{{ID: 1}}, listTotal: 100}
	s := &Service{repo: m}

	got, total, err := s.List(context.Background(), 7, domain.ListFilters{}, 2, 25)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 1 || total != 100 {
		t.Errorf("got %d items, total %d", len(got), total)
	}
	if m.listGot.ownerID != 7 || m.listGot.page != 2 || m.listGot.pageSize != 25 {
		t.Errorf("repo got %+v", m.listGot)
	}
}

func TestCreate_Delegates(t *testing.T) {
	m := &mockRepo{createID: 9}
	s := &Service{repo: m}
	id, err := s.Create(context.Background(), domain.FeedbackInput{})
	if err != nil || id != 9 {
		t.Errorf("id=%d err=%v", id, err)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), domain.FeedbackInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
