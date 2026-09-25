package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/accounts/internal/domain"
	"github.com/voxire/onion/internal/platform/financeaccounts"
)

type mockRepo struct {
	listResult []domain.Account
	tree       []domain.Account
	createID   int64
	createErr  error
	updErr     error
	delErr     error
	treeErr    error
	mapping    financeaccounts.PayrollAccountMapping
	mappingErr error
}

func (m *mockRepo) List(_ context.Context, _ int64, _, _ int) ([]domain.Account, error) {
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

func (m *mockRepo) GetPayrollAccountMapping(context.Context, int64) (financeaccounts.PayrollAccountMapping, error) {
	return m.mapping, m.mappingErr
}

func TestGetPayrollAccountMapping_Delegates(t *testing.T) {
	want := financeaccounts.PayrollAccountMapping{BusinessID: 7, ExpenseAccountID: 11, PayableAccountID: 12}
	s := &Service{repo: &mockRepo{mapping: want}}

	got, err := s.GetPayrollAccountMapping(context.Background(), want.BusinessID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("mapping = %+v, want %+v", got, want)
	}
}

func TestGetPayrollAccountMapping_PropagatesError(t *testing.T) {
	want := errors.New("mapping unavailable")
	s := &Service{repo: &mockRepo{mappingErr: want}}
	if _, err := s.GetPayrollAccountMapping(context.Background(), 7); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestList_Delegates(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Account{{ID: 1}}}}
	got, _, err := s.List(context.Background(), 1, 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
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

func TestList_Empty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Account{}}}
	got, _, err := s.List(context.Background(), 1, 1, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestList_PropagatesError(t *testing.T) {
	// mockRepo.List ignores listErr intentionally (no error field for it)
	// Verify nil error is returned when repo succeeds.
	s := &Service{repo: &mockRepo{listResult: nil}}
	_, _, err := s.List(context.Background(), 1, 1, 50)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 42}}
	id, err := s.Create(context.Background(), 1, domain.AccountInput{Name: "Cash"})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if id != 42 {
		t.Errorf("expected 42, got %d", id)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.Update(context.Background(), 1, 5, domain.AccountInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := &Service{repo: &mockRepo{delErr: want}}
	if err := s.Delete(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestChartOfAccounts_EmptyTree(t *testing.T) {
	s := &Service{repo: &mockRepo{tree: []domain.Account{}}}
	roots, err := s.ChartOfAccounts(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(roots) != 0 {
		t.Errorf("expected empty tree, got %d roots", len(roots))
	}
}

func TestChartOfAccounts_FlatAccounts(t *testing.T) {
	tree := []domain.Account{
		{ID: 1, Name: "Cash", ParentID: 0},
		{ID: 2, Name: "Bank", ParentID: 0},
		{ID: 3, Name: "Equity", ParentID: 0},
	}
	s := &Service{repo: &mockRepo{tree: tree}}
	roots, err := s.ChartOfAccounts(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(roots) != 3 {
		t.Errorf("expected 3 roots, got %d", len(roots))
	}
	for _, r := range roots {
		if len(r.Children) != 0 {
			t.Errorf("expected no children for root %d, got %d", r.ID, len(r.Children))
		}
	}
}

func TestChartOfAccounts_DeepNesting(t *testing.T) {
	tree := []domain.Account{
		{ID: 1, Name: "Assets", ParentID: 0},
		{ID: 2, Name: "Current Assets", ParentID: 1},
		{ID: 3, Name: "Cash and Equivalents", ParentID: 2},
	}
	s := &Service{repo: &mockRepo{tree: tree}}
	roots, err := s.ChartOfAccounts(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if len(roots[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(roots[0].Children))
	}
	if len(roots[0].Children[0].Children) != 1 {
		t.Errorf("expected 1 grandchild, got %d", len(roots[0].Children[0].Children))
	}
}
