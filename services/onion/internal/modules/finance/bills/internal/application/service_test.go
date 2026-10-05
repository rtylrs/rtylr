package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/finance/bills/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

type mockRepo struct {
	listResult          []domain.Bill
	getResult           domain.Bill
	createID            int64
	listErr             error
	getErr              error
	delErr              error
	createErr           error
	updErr              error
	listPaginatedLimit  int32
	listPaginatedOffset int32
}

func (m *mockRepo) List(_ context.Context, _ int64) ([]domain.Bill, error) {
	return m.listResult, m.listErr
}

func (m *mockRepo) ListPaginated(_ context.Context, _ int64, limit, offset int32) ([]domain.Bill, error) {
	m.listPaginatedLimit = limit
	m.listPaginatedOffset = offset
	return m.listResult, m.listErr
}

func (m *mockRepo) Get(_ context.Context, _, _ int64) (domain.Bill, error) {
	return m.getResult, m.getErr
}

func (m *mockRepo) Delete(_ context.Context, _, _ int64) error {
	return m.delErr
}

func (m *mockRepo) Create(_ context.Context, _ int64, _ domain.BillInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockRepo) Update(_ context.Context, _, _ int64, _ domain.BillInput) error {
	return m.updErr
}

func TestList_Delegates(t *testing.T) {
	want := []domain.Bill{{ID: 1}}
	s := &Service{repo: &mockRepo{listResult: want}}
	got, _ := s.List(context.Background(), 1)
	if len(got) != 1 {
		t.Errorf("got %d", len(got))
	}
}

func TestGet_Delegates(t *testing.T) {
	want := domain.Bill{ID: 7}
	s := &Service{repo: &mockRepo{getResult: want}}
	got, _ := s.GetBill(context.Background(), 1, 7)
	if got.ID != 7 {
		t.Errorf("got %+v", got)
	}
}

func TestCreate_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{createErr: want}}
	if _, err := s.Create(context.Background(), 1, domain.BillInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.BillInput{}); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestDelete_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.DeleteBill(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected: %v", err)
	}
}

func TestList_Empty(t *testing.T) {
	s := &Service{repo: &mockRepo{listResult: []domain.Bill{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestList_PropagatesError(t *testing.T) {
	want := errors.New("list fail")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.List(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestListPage_UsesRequestedPageAndReportsMoreRows(t *testing.T) {
	repo := &mockRepo{listResult: make([]domain.Bill, 4)}
	s := &Service{repo: repo}

	page, err := s.ListPage(context.Background(), 1, 2, 3)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if repo.listPaginatedLimit != 4 {
		t.Fatalf("query limit = %d, want 4", repo.listPaginatedLimit)
	}
	if repo.listPaginatedOffset != 3 {
		t.Fatalf("query offset = %d, want 3", repo.listPaginatedOffset)
	}
	if page.Page != 2 || page.PageSize != 3 {
		t.Fatalf("page metadata = (%d, %d), want (2, 3)", page.Page, page.PageSize)
	}
	if len(page.Bills) != 3 {
		t.Fatalf("returned bills = %d, want 3", len(page.Bills))
	}
	if !page.HasMore {
		t.Fatal("HasMore = false, want true")
	}
}

func TestListPage_UsesDefaultsAndCapsPageSize(t *testing.T) {
	repo := &mockRepo{listResult: make([]domain.Bill, 1)}
	s := &Service{repo: repo}

	page, err := s.ListPage(context.Background(), 1, 0, 10000)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if page.Page != 1 || page.PageSize != int64(maxListLimit) {
		t.Fatalf("page metadata = (%d, %d), want (1, %d)", page.Page, page.PageSize, maxListLimit)
	}
	if repo.listPaginatedLimit != maxListLimit+1 {
		t.Fatalf("query limit = %d, want %d", repo.listPaginatedLimit, maxListLimit+1)
	}
	if repo.listPaginatedOffset != 0 {
		t.Fatalf("query offset = %d, want 0", repo.listPaginatedOffset)
	}
}

func TestGet_PropagatesError(t *testing.T) {
	want := errors.New("get fail")
	s := &Service{repo: &mockRepo{getErr: want}}
	if _, err := s.GetBill(context.Background(), 1, 7); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestDelete_PropagatesError(t *testing.T) {
	want := errors.New("delete fail")
	s := &Service{repo: &mockRepo{delErr: want}}
	if err := s.DeleteBill(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestCreate_ReturnsID(t *testing.T) {
	s := &Service{repo: &mockRepo{createID: 88}}
	id, err := s.Create(context.Background(), 1, domain.BillInput{})
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if id != 88 {
		t.Errorf("expected 88, got %d", id)
	}
}

func TestUpdate_PropagatesError(t *testing.T) {
	want := errors.New("update fail")
	s := &Service{repo: &mockRepo{updErr: want}}
	if err := s.Update(context.Background(), 1, 5, domain.BillInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestList_MultipleBills(t *testing.T) {
	bills := []domain.Bill{
		{ID: 1, Status: "open"},
		{ID: 2, Status: "paid"},
		{ID: 3, Status: "overdue"},
	}
	s := &Service{repo: &mockRepo{listResult: bills}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("expected 3, got %d", len(got))
	}
}

func TestGet_ReturnsFields(t *testing.T) {
	bill := domain.Bill{ID: 7, OwnerID: 1, Amount: money.Amount(50000), Currency: "USD", Status: "open"}
	s := &Service{repo: &mockRepo{getResult: bill}}
	got, err := s.GetBill(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if got.Amount != money.Amount(50000) || got.Currency != "USD" || got.Status != "open" {
		t.Errorf("fields wrong: %+v", got)
	}
}

func (m *mockRepo) ListByPurchaseOrder(_ context.Context, _, _ int64) ([]domain.Bill, error) {
	return nil, nil
}

func (m *mockRepo) CountByPurchaseOrder(_ context.Context, _, _ int64) (int64, error) {
	return 0, nil
}
