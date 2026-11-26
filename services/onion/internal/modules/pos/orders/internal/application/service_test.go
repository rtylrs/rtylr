package application

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/settings"
)

// ProcessRefund's input guards run before any repository call, so we can
// exercise them with a zero-value Service and verify they reject malformed
// input before touching the database.

func TestProcessRefund_RejectsMissingOrderID(t *testing.T) {
	s := &Service{}
	if _, err := s.ProcessRefund(context.Background(), 1, 1, domain.RefundInput{Amount: 5}); err == nil {
		t.Fatalf("expected error for missing order id")
	} else if !strings.Contains(err.Error(), "order id") {
		t.Fatalf("expected order id error, got: %v", err)
	}
}

func TestProcessRefund_RejectsNonPositiveAmount(t *testing.T) {
	s := &Service{}
	for _, amount := range []float64{0, -1, -0.01} {
		if _, err := s.ProcessRefund(context.Background(), 1, 1, domain.RefundInput{OrderID: 7, Amount: amount}); err == nil {
			t.Fatalf("expected error for amount %v", amount)
		}
	}
}

// ─── CreateOrder input guards ────────────────────────────────────────────────

func TestCreateOrder_RejectsEmptyItems(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateOrder(context.Background(), 1, 1, 1, domain.OrderInput{}); err == nil {
		t.Error("expected error for empty items")
	} else if !strings.Contains(err.Error(), "item") {
		t.Errorf("expected item-related error, got: %v", err)
	}
}

func TestDefaultOrderTypeForProfile(t *testing.T) {
	if got := defaultOrderTypeForProfile(settings.ProfileRetail); got != domain.OrderTypeTakeaway {
		t.Fatalf("retail default = %q, want %q", got, domain.OrderTypeTakeaway)
	}
	if got := defaultOrderTypeForProfile(settings.ProfileQuickService); got != domain.OrderTypeTakeaway {
		t.Fatalf("quick service default = %q, want %q", got, domain.OrderTypeTakeaway)
	}
	if got := defaultOrderTypeForProfile(settings.ProfileRestaurant); got != domain.OrderTypeDineIn {
		t.Fatalf("restaurant default = %q, want %q", got, domain.OrderTypeDineIn)
	}
}

// ─── CreateLocation input guard ──────────────────────────────────────────────

func TestCreateLocation_RejectsEmptyName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateLocation(context.Background(), 1, domain.LocationInput{}); err == nil {
		t.Error("expected error for empty location name")
	} else if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected name-related error, got: %v", err)
	}
}

// ─── CreateTable input guards ─────────────────────────────────────────────────

func TestCreateTable_RejectsEmptyName(t *testing.T) {
	s := &Service{}
	if _, err := s.CreateTable(context.Background(), 1, domain.TableInput{}); err == nil {
		t.Error("expected error for empty table name")
	}
}

// ─── UpdateTable input guards ─────────────────────────────────────────────────

func TestUpdateTable_RejectsZeroID(t *testing.T) {
	s := &Service{}
	if err := s.UpdateTable(context.Background(), 1, domain.TableInput{Name: "T1"}); err == nil {
		t.Error("expected error for zero table id")
	}
}

func TestUpdateTable_RejectsEmptyName(t *testing.T) {
	s := &Service{}
	if err := s.UpdateTable(context.Background(), 1, domain.TableInput{ID: 1}); err == nil {
		t.Error("expected error for empty name")
	}
}

// ─── UpdateTableStatus / DeleteTable guards ───────────────────────────────────

func TestUpdateTableStatus_RejectsZeroTableID(t *testing.T) {
	s := &Service{}
	if err := s.UpdateTableStatus(context.Background(), 1, 0, domain.TableStatusAvailable); err == nil {
		t.Error("expected error for zero table id")
	}
}

func TestDeleteTable_RejectsZeroTableID(t *testing.T) {
	s := &Service{}
	if err := s.DeleteTable(context.Background(), 1, 0); err == nil {
		t.Error("expected error for zero table id")
	}
}

// ─── OpenSession guard ────────────────────────────────────────────────────────

func TestProcessPayment_RejectsZeroOrderID(t *testing.T) {
	s := &Service{}
	if err := s.ProcessPayment(context.Background(), 1, domain.PaymentInput{OrderID: 0}); err == nil {
		t.Error("expected error for zero order id")
	}
}

// ─── CreatePublicOrder guards ─────────────────────────────────────────────────

func TestCreatePublicOrder_RejectsZeroOwnerID(t *testing.T) {
	s := &Service{}
	if _, _, _, err := s.CreatePublicOrder(context.Background(), 0, PublicOrderInput{
		Items: []PublicOrderItemInput{{Name: "Burger", Quantity: 1}},
	}); err == nil {
		t.Error("expected error for zero owner_id")
	}
}

func TestCreatePublicOrder_RejectsEmptyItems(t *testing.T) {
	s := &Service{}
	if _, _, _, err := s.CreatePublicOrder(context.Background(), 1, PublicOrderInput{}); err == nil {
		t.Error("expected error for empty items")
	}
}

// ─── HandlePaddleTransactionCompleted guard ───────────────────────────────────

func TestHandlePaddleTransactionCompleted_RejectsEmptyTxID(t *testing.T) {
	s := &Service{}
	if err := s.HandlePaddleTransactionCompleted(context.Background(), "", "transaction.completed"); err == nil {
		t.Error("expected error for empty transaction id")
	}
}

// ─── ApplyDiscount guards ─────────────────────────────────────────────────────

func TestApplyDiscount_RejectsZeroOrderID(t *testing.T) {
	s := &Service{}
	if err := s.ApplyDiscount(context.Background(), 1, 1, domain.DiscountInput{DiscountType: "pct", DiscountValue: 10}); err == nil {
		t.Error("expected error for zero order id")
	}
}

func TestApplyDiscount_RejectsBlankType(t *testing.T) {
	s := &Service{}
	if err := s.ApplyDiscount(context.Background(), 1, 1, domain.DiscountInput{OrderID: 1, DiscountValue: 10}); err == nil {
		t.Error("expected error for blank discount type")
	}
}

func TestApplyDiscount_RejectsNegativeValue(t *testing.T) {
	s := &Service{}
	if err := s.ApplyDiscount(context.Background(), 1, 1, domain.DiscountInput{OrderID: 1, DiscountType: "pct", DiscountValue: -5}); err == nil {
		t.Error("expected error for negative discount value")
	}
}

func TestPaginateOrders_ClampsPageSize(t *testing.T) {
	limit, offset := paginateOrders(0, 0)
	if limit != 50 || offset != 0 {
		t.Errorf("got limit=%d offset=%d, want 50,0", limit, offset)
	}
	limit, _ = paginateOrders(1, 10000)
	if limit != 500 {
		t.Errorf("got limit=%d, want 500", limit)
	}
}

func TestPaginateOrders_CapsRunawayOffset(t *testing.T) {
	_, offset := paginateOrders(1_000_000, 200)
	if offset != 10000 {
		t.Errorf("got offset=%d, want capped 10000", offset)
	}
	_, offset = paginateOrders(9223372036854775807, 200)
	if offset < 0 || offset > 10000 {
		t.Errorf("got offset=%d, want within [0,10000]", offset)
	}
}

func TestPaginateOrders_NormalPaging(t *testing.T) {
	limit, offset := paginateOrders(3, 50)
	if limit != 50 || offset != 100 {
		t.Errorf("got limit=%d offset=%d, want 50,100", limit, offset)
	}
}

type fakePublicOrderRepo struct {
	repository
	items        []domain.OrderItemInput
	basePrice    float64
	orderCreated bool
}

func (f *fakePublicOrderRepo) IsOnlineOrderingEnabled(_ context.Context, _ int64) (bool, error) {
	return true, nil
}

func (f *fakePublicOrderRepo) GetPublicMenuItem(_ context.Context, _, menuItemID int64) (domain.OrderItemInput, error) {
	return domain.OrderItemInput{MenuItemID: menuItemID, Name: "Latte", UnitPrice: f.basePrice}, nil
}

func (f *fakePublicOrderRepo) CreatePublicOrder(_ context.Context, _, _, _ int64, _, _, _, _, _ string, items []domain.OrderItemInput) (int64, string, error) {
	f.items = items
	f.orderCreated = true
	return 101, "ORD-101", nil
}

func TestCreatePublicOrder_BranchOverridesTenancyAndCRMOptIn(t *testing.T) {
	repo := &fakePublicOrderRepo{basePrice: 5.00}
	var hookCRMOptIn bool
	var hookTotal float64

	svc := NewService(repo, nil).
		WithPublicOrderTenancyValidator(func(_ context.Context, _, locationID, tableID int64) error {
			if locationID == 999 || tableID == 888 {
				return errors.New("foreign branch or table")
			}
			return nil
		}).
		WithBranchItemPriceResolver(func(_ context.Context, _, _, menuItemID int64) (bool, bool, float64, error) {
			if menuItemID == 20 {
				return false, false, 0, nil
			}
			if menuItemID == 10 {
				return true, true, 7.50, nil
			}
			return true, false, 0, nil
		}).
		WithPublicOrderHook(func(_ context.Context, _, _, _, _ int64, _, _, _, _, _ string, crmOptIn bool, total float64) {
			hookCRMOptIn = crmOptIn
			hookTotal = total
		})

	// Rejects foreign location.
	if _, _, _, err := svc.CreatePublicOrder(context.Background(), 42, PublicOrderInput{
		LocationID: 999,
		Items:      []PublicOrderItemInput{{MenuItemID: 10, Quantity: 1}},
	}); err == nil {
		t.Fatalf("expected error for foreign location")
	}

	// Rejects branch-hidden item.
	if _, _, _, err := svc.CreatePublicOrder(context.Background(), 42, PublicOrderInput{
		LocationID: 1,
		Items:      []PublicOrderItemInput{{MenuItemID: 20, Quantity: 1}},
	}); err == nil || !strings.Contains(err.Error(), "not available at this branch") {
		t.Fatalf("expected hidden branch item error, got %v", err)
	}

	// Applies branch price override and propagates crmOptIn=false.
	if _, _, _, err := svc.CreatePublicOrder(context.Background(), 42, PublicOrderInput{
		LocationID:    1,
		CustomerPhone: "+15550001111",
		CRMOptIn:      false,
		Items:         []PublicOrderItemInput{{MenuItemID: 10, Quantity: 2}},
	}); err != nil {
		t.Fatalf("CreatePublicOrder failed: %v", err)
	}
	if !repo.orderCreated || len(repo.items) != 1 || repo.items[0].UnitPrice != 7.50 {
		t.Fatalf("expected overridden unit price 7.50, got %+v", repo.items)
	}
	if hookCRMOptIn {
		t.Fatalf("expected crmOptIn=false propagated to hook")
	}
	if hookTotal != 15.00 {
		t.Fatalf("expected hook total 15.00, got %.2f", hookTotal)
	}
}
