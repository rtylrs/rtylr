package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
)

type mockRepo struct {
	listResult []domain.Account
	tree       []domain.Account
	createID   int64
	createErr  error
	updErr     error
	delErr     error
	treeErr    error
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Account, error) {
	return m.listResult, nil
}
func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.AccountInput) (int64, error) {
	return m.createID, m.createErr
}
func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.AccountInput) error {
	return m.updErr
}
func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}
func (m *mockRepo) ListForTree(_ context.Context, _ int64) ([]domain.Account, error) {
	return m.tree, m.treeErr
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Account{{ID: 1}}}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.AccountInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdateAndDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.AccountInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
	if err := s.Delete(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestChartOfAccounts_BuildsTree(t *testing.T) {
	tree := []domain.Account{
		{ID: 1, Name: "Assets", ParentID: 0},
		{ID: 10, Name: "Cash", ParentID: 1},
		{ID: 20, Name: "Bank", ParentID: 1},
		{ID: 100, Name: "Orphan", ParentID: 999}, // parent missing
	}
	s := &Service{repo: &mockRepo{tree: tree}}
	roots, err := s.ChartOfAccounts(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Two roots: Assets (id=1) and Orphan (id=100)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	var assetsNode *domain.AccountTreeNode
	for _, r := range roots {
		if r.ID == 1 {
			assetsNode = r
		}
	}
	if assetsNode == nil {
		t.Fatal("expected Assets root")
	}
	if len(assetsNode.Children) != 2 {
		t.Errorf("expected 2 children of Assets, got %d", len(assetsNode.Children))
	}
}

func TestChartOfAccounts_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{treeErr: want}}
	if _, err := s.ChartOfAccounts(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
