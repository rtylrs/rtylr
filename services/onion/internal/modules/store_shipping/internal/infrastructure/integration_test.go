package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_shipping/internal/domain"
	"github.com/voxire/onion/internal/modules/store_shipping/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreShipping_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreShipping_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range items {
		if m.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, m.OwnerID)
		}
	}
}

func TestStoreShipping_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name:          "Standard Shipping",
		Price:         5.99,
		EstimatedDays: 5,
		IsActive:      true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
}

func TestStoreShipping_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name: "Express", Price: 19.99, EstimatedDays: 1, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, m := range items {
		if m.ID == id {
			found = true
			if m.Name != "Express" {
				t.Errorf("expected name Express, got %s", m.Name)
			}
		}
	}
	if !found {
		t.Errorf("created shipping method id=%d not found in list", id)
	}
}

func TestStoreShipping_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name: "OldName", Price: 3.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.ShippingMethodInput{
		Name: "NewName", Price: 3.0, IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, m := range items {
		if m.ID == id && m.Name != "NewName" {
			t.Errorf("expected name NewName, got %s", m.Name)
		}
	}
}

func TestStoreShipping_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name: "ToDelete", Price: 0.0, IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, m := range items {
		if m.ID == id {
			t.Errorf("deleted shipping method id=%d still in list", id)
		}
	}
}

func TestStoreShipping_List_UnknownOwner_Empty(t *testing.T) {
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

func TestStoreShipping_Update_Price(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name: "PriceTest", Price: 1.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.ShippingMethodInput{
		Name: "PriceTest", Price: 9.99, IsActive: true,
	}); err != nil {
		t.Fatalf("Update price: %v", err)
	}
}

func TestStoreShipping_Update_EstimatedDays(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name: "DaysTest", Price: 5.0, EstimatedDays: 3, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.ShippingMethodInput{
		Name: "DaysTest", Price: 5.0, EstimatedDays: 7, IsActive: true,
	}); err != nil {
		t.Fatalf("Update estimated days: %v", err)
	}
}

func TestStoreShipping_Create_FreeShipping(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ShippingMethodInput{
		Name: "Free Shipping", Price: 0.0, EstimatedDays: 10, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create free shipping: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
