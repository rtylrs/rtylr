package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_receipts/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_receipts/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestReceipts_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = receipts
}

func TestReceipts_List_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if receipts == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestReceipts_List_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), 999999, "", "")
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(receipts) != 0 {
		t.Fatalf("expected 0 receipts for unknown owner, got %d", len(receipts))
	}
}

func TestReceipts_List_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, r := range receipts {
		if r.ID == 0 {
			t.Errorf("receipts[%d].ID is zero", i)
		}
	}
}

func TestReceipts_List_AllHaveNonEmptyOrderNumber(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, r := range receipts {
		if r.OrderNumber == "" {
			t.Errorf("receipts[%d].OrderNumber is empty", i)
		}
	}
}

func TestReceipts_Get_FixtureOrder3301(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipt, err := repo.Get(context.Background(), testutil.OwnerID, 3301)
	if err != nil {
		t.Fatalf("Get order 3301: %v", err)
	}
	if receipt == nil {
		t.Fatal("expected non-nil receipt for order 3301")
	}
	if receipt.ID != 3301 {
		t.Errorf("receipt.ID = %d, want 3301", receipt.ID)
	}
}

func TestReceipts_Get_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.Get(context.Background(), testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrReceiptNotFound) {
		t.Fatalf("expected ErrReceiptNotFound, got %v", err)
	}
}

func TestReceipts_Get_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	_, err := repo.Get(context.Background(), 999999, 3301)
	if !errors.Is(err, domain.ErrReceiptNotFound) {
		t.Fatalf("expected ErrReceiptNotFound for wrong owner, got %v", err)
	}
}

func TestReceipts_Get_TotalNonNegative(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipt, err := repo.Get(context.Background(), testutil.OwnerID, 3301)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if receipt.Total < 0 {
		t.Errorf("receipt.Total = %f, want >= 0", receipt.Total)
	}
}

func TestReceipts_List_WithDateFrom_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "2020-01-01", "")
	if err != nil {
		t.Fatalf("List with dateFrom: %v", err)
	}
	_ = receipts
}

func TestReceipts_List_WithDateRange_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "2020-01-01", "2030-12-31")
	if err != nil {
		t.Fatalf("List with date range: %v", err)
	}
	_ = receipts
}

func TestReceipts_List_FutureDateRangeReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "2099-01-01", "2099-12-31")
	if err != nil {
		t.Fatalf("List future date range: %v", err)
	}
	if len(receipts) != 0 {
		t.Fatalf("expected 0 receipts for future date range, got %d", len(receipts))
	}
}

func TestReceipts_Get_HasNonEmptyPaymentMethod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipt, err := repo.Get(context.Background(), testutil.OwnerID, 3301)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if receipt.PaymentMethod == "" {
		t.Error("expected non-empty PaymentMethod for order 3301")
	}
}

func TestReceipts_List_AllHaveNonEmptyPaymentMethod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	receipts, err := repo.List(context.Background(), testutil.OwnerID, "", "")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, r := range receipts {
		if r.PaymentMethod == "" {
			t.Errorf("receipts[%d].PaymentMethod is empty", i)
		}
	}
}
