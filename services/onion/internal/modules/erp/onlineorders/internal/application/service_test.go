package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/onlineorders/internal/domain"
)

type storeOrderRepo interface {
	List(ctx context.Context, ownerID int64) ([]domain.StoreOrder, error)
	GetDetail(ctx context.Context, ownerID, orderID int64) (domain.StoreOrderDetail, error)
	Create(ctx context.Context, ownerID int64, input domain.StoreOrderInput) (int64, error)
	Update(ctx context.Context, ownerID, id int64, input domain.StoreOrderInput) error
}

type mockOrderRepo struct {
	list      []domain.StoreOrder
	listErr   error
	detail    domain.StoreOrderDetail
	detailErr error
	createID  int64
	createErr error
	updateErr error
}

func (m *mockOrderRepo) List(_ context.Context, _ int64) ([]domain.StoreOrder, error) {
	return m.list, m.listErr
}

func (m *mockOrderRepo) GetDetail(_ context.Context, _, _ int64) (domain.StoreOrderDetail, error) {
	return m.detail, m.detailErr
}

func (m *mockOrderRepo) Create(_ context.Context, _ int64, _ domain.StoreOrderInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockOrderRepo) Update(_ context.Context, _, _ int64, _ domain.StoreOrderInput) error {
	return m.updateErr
}

type orderTestService struct{ repo storeOrderRepo }

func (s *orderTestService) List(ctx context.Context, ownerID int64) ([]domain.StoreOrder, error) {
	return s.repo.List(ctx, ownerID)
}

func (s *orderTestService) GetDetail(ctx context.Context, ownerID, orderID int64) (domain.StoreOrderDetail, error) {
	return s.repo.GetDetail(ctx, ownerID, orderID)
}

func (s *orderTestService) Create(ctx context.Context, ownerID int64, input domain.StoreOrderInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, input)
}

func (s *orderTestService) Update(ctx context.Context, ownerID, id int64, input domain.StoreOrderInput) error {
	return s.repo.Update(ctx, ownerID, id, input)
}

// --- Tests ---

func TestStoreOrder_List_Delegates(t *testing.T) {
	m := &mockOrderRepo{list: []domain.StoreOrder{{ID: 1}, {ID: 2}}}
	s := &orderTestService{repo: m}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestStoreOrder_List_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &orderTestService{repo: &mockOrderRepo{listErr: want}}
	_, err := s.List(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreOrder_List_Empty(t *testing.T) {
	s := &orderTestService{repo: &mockOrderRepo{list: []domain.StoreOrder{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

func TestStoreOrder_GetDetail_ReturnsDetail(t *testing.T) {
	detail := domain.StoreOrderDetail{
		Order: domain.StoreOrder{ID: 7, CustomerEmail: "a@b.com"},
		Items: []domain.StoreOrderItem{{ID: 1, Name: "Widget"}},
	}
	s := &orderTestService{repo: &mockOrderRepo{detail: detail}}
	got, err := s.GetDetail(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Order.ID != 7 {
		t.Errorf("order ID = %d, want 7", got.Order.ID)
	}
	if len(got.Items) != 1 {
		t.Errorf("items count = %d, want 1", len(got.Items))
	}
}

func TestStoreOrder_GetDetail_PropagatesError(t *testing.T) {
	want := errors.New("not found")
	s := &orderTestService{repo: &mockOrderRepo{detailErr: want}}
	_, err := s.GetDetail(context.Background(), 1, 999)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreOrder_Create_ReturnsID(t *testing.T) {
	s := &orderTestService{repo: &mockOrderRepo{createID: 55}}
	id, err := s.Create(context.Background(), 1, domain.StoreOrderInput{CustomerEmail: "x@y.com"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 55 {
		t.Errorf("id = %d, want 55", id)
	}
}

func TestStoreOrder_Create_PropagatesError(t *testing.T) {
	want := errors.New("insert failed")
	s := &orderTestService{repo: &mockOrderRepo{createErr: want}}
	_, err := s.Create(context.Background(), 1, domain.StoreOrderInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestValidateOrder_RejectsClientTotalWithoutCatalogItems(t *testing.T) {
	if err := validateOrder(domain.StoreOrderInput{Total: 99}); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("validateOrder error = %v, want ErrInvalidAmount", err)
	}
}

func TestValidateOrder_RejectsClientPricedLineItems(t *testing.T) {
	if err := validateOrder(domain.StoreOrderInput{Items: []domain.OrderItemInput{{Quantity: 1, Price: 1, Total: 1}}}); !errors.Is(err, domain.ErrInvalidAmount) {
		t.Fatalf("validateOrder error = %v, want ErrInvalidAmount", err)
	}
}

func TestStoreOrder_Update_HappyPath(t *testing.T) {
	s := &orderTestService{repo: &mockOrderRepo{}}
	if err := s.Update(context.Background(), 1, 5, domain.StoreOrderInput{}); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestStoreOrder_Update_PropagatesError(t *testing.T) {
	want := errors.New("update failed")
	s := &orderTestService{repo: &mockOrderRepo{updateErr: want}}
	if err := s.Update(context.Background(), 1, 5, domain.StoreOrderInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStoreOrder_GetDetail_EmptyItems(t *testing.T) {
	detail := domain.StoreOrderDetail{
		Order: domain.StoreOrder{ID: 3},
		Items: []domain.StoreOrderItem{},
	}
	s := &orderTestService{repo: &mockOrderRepo{detail: detail}}
	got, err := s.GetDetail(context.Background(), 1, 3)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(got.Items))
	}
}
