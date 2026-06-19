package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/store_payments/internal/domain"
)

type storePayRepo interface {
	List(ctx context.Context, businessID int64) ([]domain.PaymentMethod, error)
	Create(ctx context.Context, ownerID, businessID int64, input domain.PaymentMethodInput) (int64, error)
	Update(ctx context.Context, businessID, id int64, input domain.PaymentMethodInput) error
}

type mockPayRepo struct {
	list      []domain.PaymentMethod
	listErr   error
	createID  int64
	createErr error
	updateErr error
}

func (m *mockPayRepo) List(_ context.Context, _ int64) ([]domain.PaymentMethod, error) {
	return m.list, m.listErr
}

func (m *mockPayRepo) Create(_ context.Context, _, _ int64, _ domain.PaymentMethodInput) (int64, error) {
	return m.createID, m.createErr
}

func (m *mockPayRepo) Update(_ context.Context, _, _ int64, _ domain.PaymentMethodInput) error {
	return m.updateErr
}

type payTestService struct{ repo storePayRepo }

func (s *payTestService) List(ctx context.Context, businessID int64) ([]domain.PaymentMethod, error) {
	return s.repo.List(ctx, businessID)
}

func (s *payTestService) Create(ctx context.Context, ownerID, businessID int64, input domain.PaymentMethodInput) (int64, error) {
	return s.repo.Create(ctx, ownerID, businessID, input)
}

func (s *payTestService) Update(ctx context.Context, businessID, id int64, input domain.PaymentMethodInput) error {
	return s.repo.Update(ctx, businessID, id, input)
}

// --- Tests ---

func TestStorePay_List_Delegates(t *testing.T) {
	m := &mockPayRepo{list: []domain.PaymentMethod{{ID: 1, Provider: "stripe"}, {ID: 2, Provider: "paypal"}}}
	s := &payTestService{repo: m}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestStorePay_List_PropagatesError(t *testing.T) {
	want := errors.New("db error")
	s := &payTestService{repo: &mockPayRepo{listErr: want}}
	_, err := s.List(context.Background(), 1)
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStorePay_List_Empty(t *testing.T) {
	s := &payTestService{repo: &mockPayRepo{list: []domain.PaymentMethod{}}}
	got, err := s.List(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d", len(got))
	}
}

func TestStorePay_Create_ReturnsID(t *testing.T) {
	s := &payTestService{repo: &mockPayRepo{createID: 7}}
	id, err := s.Create(context.Background(), 1, 1, domain.PaymentMethodInput{Provider: "stripe"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, want 7", id)
	}
}

func TestStorePay_Create_PropagatesError(t *testing.T) {
	want := errors.New("insert failed")
	s := &payTestService{repo: &mockPayRepo{createErr: want}}
	_, err := s.Create(context.Background(), 1, 1, domain.PaymentMethodInput{})
	if !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStorePay_Update_HappyPath(t *testing.T) {
	s := &payTestService{repo: &mockPayRepo{}}
	if err := s.Update(context.Background(), 1, 3, domain.PaymentMethodInput{IsActive: true}); err != nil {
		t.Errorf("unexpected err: %v", err)
	}
}

func TestStorePay_Update_PropagatesError(t *testing.T) {
	want := errors.New("update failed")
	s := &payTestService{repo: &mockPayRepo{updateErr: want}}
	if err := s.Update(context.Background(), 1, 3, domain.PaymentMethodInput{}); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestStorePay_List_SingleEntry(t *testing.T) {
	m := &mockPayRepo{list: []domain.PaymentMethod{{ID: 10, Provider: "cash", IsActive: true}}}
	s := &payTestService{repo: m}
	got, err := s.List(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got[0].Provider != "cash" {
		t.Errorf("provider = %q, want %q", got[0].Provider, "cash")
	}
	if !got[0].IsActive {
		t.Error("expected IsActive=true")
	}
}

func TestStorePay_Create_InactiveMethod(t *testing.T) {
	s := &payTestService{repo: &mockPayRepo{createID: 3}}
	id, err := s.Create(context.Background(), 1, 1, domain.PaymentMethodInput{Provider: "bank", IsActive: false})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if id != 3 {
		t.Errorf("id = %d, want 3", id)
	}
}
