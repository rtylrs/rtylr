package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
)

type mockRepo struct {
	list      []domain.MenuItem
	createID  int64
	createErr error
	updateErr error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.MenuItem, error) {
	return m.list, nil
}

func (m *mockRepo) IsOnlineOrderingEnabled(_ context.Context, _ int64) (bool, error) {
	return true, nil
}

func (m *mockRepo) QuickCreate(_ context.Context, _ int64, _ domain.MenuItemInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.MenuItemUpdateInput) error {
	return m.updateErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{list: []domain.MenuItem{{}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestQuickCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.QuickCreate(context.Background(), 1, domain.MenuItemInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
