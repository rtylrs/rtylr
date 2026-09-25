package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/store_orders/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreOrders_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreOrders_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, o := range items {
		if o.BusinessID != testutil.BusinessID {
			t.Errorf("expected business_id %d, got %d", testutil.BusinessID, o.BusinessID)
		}
	}
}

func TestStoreOrders_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: "customer@example.com",
		CustomerName:  "Test Customer",
		Status:        "pending",
		Subtotal:      99.0,
		Tax:           9.9,
		ShippingCost:  5.0,
		Total:         113.9,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestStoreOrders_Create_UsesCustomerSync(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("store-orders-sync-%d@example.com", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), "DELETE FROM crm_customers WHERE business_id = ? AND email = ?", testutil.BusinessID, email); err != nil {
			t.Errorf("cleanup synced customer: %v", err)
		}
	})

	firstID, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName:  "  Checkout Customer  ",
		CustomerEmail: "  " + email + "  ",
		CustomerPhone: " +1 555 0100 ",
		Status:        "pending",
		Total:         10.50,
	})
	if err != nil {
		t.Fatalf("Create first order: %v", err)
	}
	if firstID <= 0 {
		t.Fatalf("expected positive first order ID, got %d", firstID)
	}

	secondID, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName:  "Updated Customer",
		CustomerEmail: email,
		CustomerPhone: "+1 555 0100",
		Status:        "pending",
		Total:         4.25,
	})
	if err != nil {
		t.Fatalf("Create second order: %v", err)
	}
	if secondID <= 0 {
		t.Fatalf("expected positive second order ID, got %d", secondID)
	}

	var firstName, lastName, normalizedEmail, normalizedPhone, totalSpent string
	var visitCount int32
	var businessID int64
	err = db.QueryRowContext(context.Background(), `
		SELECT first_name, COALESCE(last_name, ''), email, COALESCE(phone, ''), total_spent, visit_count, business_id
		FROM crm_customers
		WHERE business_id = ? AND email = ?
		ORDER BY id DESC
		LIMIT 1`, testutil.BusinessID, email).Scan(
		&firstName, &lastName, &normalizedEmail, &normalizedPhone, &totalSpent, &visitCount, &businessID,
	)
	if err != nil {
		t.Fatalf("read synced customer: %v", err)
	}
	if firstName != "Updated" || lastName != "Customer" {
		t.Fatalf("expected canonical updated name, got %q %q", firstName, lastName)
	}
	if normalizedEmail != email || normalizedPhone != "+1 555 0100" {
		t.Fatalf("expected normalized contact, got email=%q phone=%q", normalizedEmail, normalizedPhone)
	}
	if totalSpent != "14.75" || visitCount != 2 {
		t.Fatalf("expected accumulated spend and visits, got total=%q visits=%d", totalSpent, visitCount)
	}
	if businessID != testutil.BusinessID {
		t.Fatalf("expected customer business_id %d, got %d", testutil.BusinessID, businessID)
	}
}

func TestStoreOrders_Create_WithItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	input := domain.StoreOrderInput{
		CustomerEmail: "items@example.com",
		Status:        "pending",
		Subtotal:      50.0,
		Tax:           5.0,
		Total:         55.0,
		Items: []domain.OrderItemInput{
			{Name: "Widget", Quantity: 2, Price: 25.0, Total: 50.0},
		},
	}
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create with items: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreOrders_GetDetail_ReturnsOrder(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerName: "Detail Test",
		Status:       "pending",
		Subtotal:     10.0,
		Tax:          1.0,
		Total:        11.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.ID != id {
		t.Errorf("expected order ID %d, got %d", id, detail.Order.ID)
	}
}

func TestStoreOrders_GetDetail_IncludesItems(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status:   "pending",
		Subtotal: 30.0,
		Total:    30.0,
		Items: []domain.OrderItemInput{
			{Name: "Gadget", Quantity: 1, Price: 30.0, Total: 30.0},
		},
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if len(detail.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(detail.Items))
	}
	if detail.Items[0].Name != "Gadget" {
		t.Errorf("expected item name Gadget, got %s", detail.Items[0].Name)
	}
}

func TestStoreOrders_Update_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Subtotal: 5.0, Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.StoreOrderInput{
		Status: "confirmed", Subtotal: 5.0, Total: 5.0,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail after Update: %v", err)
	}
	if detail.Order.Status != "confirmed" {
		t.Errorf("expected status confirmed, got %s", detail.Order.Status)
	}
}

func TestStoreOrders_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(items))
	}
}

func TestStoreOrders_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: "appears@example.com",
		Status:        "pending",
		Total:         15.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, o := range items {
		if o.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created order id=%d not found in list", id)
	}
}

func TestStoreOrders_GetDetail_CrossOwner_Fails(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status: "pending", Total: 5.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = repo.GetDetail(context.Background(), 9999999, id)
	if err == nil {
		t.Error("expected error when fetching order with wrong owner")
	}
}

func TestStoreOrders_Create_WithShippingAddress(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		Status:          "pending",
		Subtotal:        25.0,
		Total:           25.0,
		ShippingAddress: `{"street":"123 Main St","city":"Testville"}`,
	})
	if err != nil {
		t.Fatalf("Create with shipping address: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreOrders_List_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestStoreOrders_Create_CustomerEmailStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "email-stored@example.com"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.StoreOrderInput{
		CustomerEmail: want, Status: "pending", Total: 10.0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := repo.GetDetail(context.Background(), testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("GetDetail: %v", err)
	}
	if detail.Order.CustomerEmail != want {
		t.Errorf("CustomerEmail = %q, want %q", detail.Order.CustomerEmail, want)
	}
}
