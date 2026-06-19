package infrastructure_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/voxire/onion/internal/modules/store_payments/internal/domain"
	"github.com/voxire/onion/internal/modules/store_payments/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStorePayments_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStorePayments_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, p.OwnerID)
		}
	}
}

func TestStorePayments_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "stripe",
		Config:   `{"key":"pk_test_xxx"}`,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() {
		_ = repo.Update(context.Background(), testutil.OwnerID, id, domain.PaymentMethodInput{Provider: "stripe", IsActive: false})
	})
}

func TestStorePayments_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "paypal",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, p := range items {
		if p.ID == id {
			found = true
			if p.Provider != "paypal" {
				t.Errorf("expected provider paypal, got %s", p.Provider)
			}
		}
	}
	if !found {
		t.Errorf("created payment method id=%d not found in list", id)
	}
}

func TestStorePayments_Update_ChangesProvider(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{Provider: "cash", IsActive: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.PaymentMethodInput{Provider: "bank_transfer", IsActive: true}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id && p.Provider != "bank_transfer" {
			t.Errorf("expected provider bank_transfer, got %s", p.Provider)
		}
	}
}

func TestStorePayments_Update_ActiveFlag(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{Provider: "cod", IsActive: true})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.PaymentMethodInput{Provider: "cod", IsActive: false}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id && p.IsActive {
			t.Error("expected IsActive=false after update")
		}
	}
}

func TestStorePayments_List_UnknownOwner_Empty(t *testing.T) {
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

func TestStorePayments_Create_WithConfig(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	config := `{"publishable_key":"pk_live_xxx","secret_key":"sk_live_xxx"}`
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "stripe",
		Config:   config,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create with config: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, p := range items {
		if p.ID == id {
			var wantMap, gotMap map[string]interface{}
			json.Unmarshal([]byte(config), &wantMap)
			json.Unmarshal([]byte(p.Config), &gotMap)
			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("expected config %s, got %s", config, p.Config)
			}
		}
	}
}

func TestStorePayments_Create_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	providers := []string{"stripe", "paypal", "square"}
	for _, prov := range providers {
		id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{Provider: prov, IsActive: false})
		if err != nil {
			t.Fatalf("Create %s: %v", prov, err)
		}
		if id <= 0 {
			t.Errorf("provider %s: expected positive ID", prov)
		}
	}
}

func TestListPaymentMethods_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for owner 0, got %d", len(items))
	}
}

func TestListPaymentMethods_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestCreatePaymentMethod_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "test_provider",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreatePaymentMethod_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "named_provider",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			if p.Provider != "named_provider" {
				t.Errorf("expected provider named_provider, got %s", p.Provider)
			}
			return
		}
	}
	t.Errorf("payment method id=%d not found in list", id)
}

func TestCreatePaymentMethod_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "type_provider",
		Config:   `{"type":"gateway"}`,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			if p.Provider != "type_provider" {
				t.Errorf("expected provider type_provider, got %s", p.Provider)
			}
			return
		}
	}
	t.Errorf("payment method id=%d not found in list", id)
}

func TestCreatePaymentMethod_IsActiveIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "active_provider",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			if !p.IsActive {
				t.Errorf("expected IsActive=true, got false")
			}
			return
		}
	}
	t.Errorf("payment method id=%d not found in list", id)
}

func TestUpdatePaymentMethod_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "old_provider",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.PaymentMethodInput{
		Provider: "new_provider",
		IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id && p.Provider != "new_provider" {
			t.Errorf("expected provider new_provider, got %s", p.Provider)
		}
	}
}

func TestCreate_ThenList_FindsPaymentMethod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "find_provider",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			return
		}
	}
	t.Errorf("payment method id=%d not found after create", id)
}

func TestListPaymentMethods_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "count_provider",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreatePaymentMethod_TypeCard(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "card",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create card payment: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			if p.Provider != "card" {
				t.Errorf("expected provider card, got %s", p.Provider)
			}
			return
		}
	}
	t.Errorf("card payment method id=%d not found", id)
}

func TestCreatePaymentMethod_TypeCash(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "cash",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create cash payment: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			if p.Provider != "cash" {
				t.Errorf("expected provider cash, got %s", p.Provider)
			}
			return
		}
	}
	t.Errorf("cash payment method id=%d not found", id)
}

func TestCreatePaymentMethod_IsActiveFalse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.OwnerID, domain.PaymentMethodInput{
		Provider: "inactive_provider",
		IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, p := range items {
		if p.ID == id {
			if p.IsActive {
				t.Errorf("expected IsActive=false, got true")
			}
			return
		}
	}
	t.Errorf("payment method id=%d not found in list", id)
}
