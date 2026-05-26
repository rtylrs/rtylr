package infrastructure_test

import (
	"context"
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.PaymentMethodInput{
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.PaymentMethodInput{
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.PaymentMethodInput{Provider: "cash", IsActive: true})
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.PaymentMethodInput{Provider: "cod", IsActive: true})
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.PaymentMethodInput{
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
		if p.ID == id && p.Config != config {
			t.Errorf("expected config %s, got %s", config, p.Config)
		}
	}
}

func TestStorePayments_Create_Multiple(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	providers := []string{"stripe", "paypal", "square"}
	for _, prov := range providers {
		id, err := repo.Create(context.Background(), testutil.OwnerID, domain.PaymentMethodInput{Provider: prov, IsActive: false})
		if err != nil {
			t.Fatalf("Create %s: %v", prov, err)
		}
		if id <= 0 {
			t.Errorf("provider %s: expected positive ID", prov)
		}
	}
}
