package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/pos/kds/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/kds/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestKDS_ListActiveOrders_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	_ = orders
}

func TestKDS_ListActiveOrders_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	if orders == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestKDS_ListActiveOrders_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), 999999)
	if err != nil {
		t.Fatalf("ListActiveOrders wrong owner: %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected 0 orders for unknown owner, got %d", len(orders))
	}
}

func TestKDS_ListActiveOrders_OrdersHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		if o.ID == 0 {
			t.Errorf("order[%d].ID is zero", i)
		}
	}
}

func TestKDS_ListActiveOrders_OrdersHaveNonEmptyStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		if o.Status == "" {
			t.Errorf("order[%d].Status is empty", i)
		}
	}
}

func TestKDS_ListActiveOrders_OrdersHaveNonEmptyOrderNumber(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		if o.OrderNumber == "" {
			t.Errorf("order[%d].OrderNumber is empty", i)
		}
	}
}

func TestKDS_ListActiveOrders_ItemsSliceNonNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		if o.Items == nil {
			t.Errorf("order[%d].Items is nil", i)
		}
	}
}

func TestKDS_ListActiveOrders_EachItemLinkedToOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		for j, item := range o.Items {
			if item.OrderID != o.ID {
				t.Errorf("order[%d].Items[%d].OrderID = %d, want %d", i, j, item.OrderID, o.ID)
			}
		}
	}
}

func TestKDS_ListActiveOrders_StatusOnlyActiveValues(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	validStatuses := map[string]bool{"confirmed": true, "preparing": true, "ready": true}

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		if !validStatuses[o.Status] {
			t.Errorf("order[%d].Status = %q, want one of confirmed/preparing/ready", i, o.Status)
		}
	}
}

func TestKDS_ListActiveOrders_EachItemHasNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		for j, item := range o.Items {
			if item.Name == "" {
				t.Errorf("order[%d].Items[%d].Name is empty", i, j)
			}
		}
	}
}

func TestKDS_UpdateItemStatus_NonExistentIsCallable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	// Non-existent IDs -- should not panic; no rows affected is acceptable.
	err := repo.UpdateItemStatus(context.Background(), domain.UpdateItemStatusInput{
		OwnerID: testutil.OwnerID,
		OrderID: 9999999,
		ItemID:  9999999,
		Status:  "ready",
	})
	_ = err
}

func TestKDS_UpdateItemStatusByItemID_NonExistentIsCallable(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	err := repo.UpdateItemStatusByItemID(context.Background(), testutil.OwnerID, 9999999, "ready")
	_ = err
}

func TestKDS_ListActiveOrders_EachItemHasPositiveOrZeroQuantity(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	orders, err := repo.ListActiveOrders(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListActiveOrders: %v", err)
	}
	for i, o := range orders {
		for j, item := range o.Items {
			if item.Quantity < 0 {
				t.Errorf("order[%d].Items[%d].Quantity = %f, want >= 0", i, j, item.Quantity)
			}
		}
	}
}
