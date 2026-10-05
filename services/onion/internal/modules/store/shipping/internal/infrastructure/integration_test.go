package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store/shipping/internal/domain"
	"github.com/voxire/onion/internal/modules/store/shipping/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreShipping_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreShipping_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
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
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
}

func TestStoreShipping_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "Express", Price: 19.99, EstimatedDays: 1, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "OldName", Price: 3.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ShippingMethodInput{
		Name: "NewName", Price: 3.0, IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id && m.Name != "NewName" {
			t.Errorf("expected name NewName, got %s", m.Name)
		}
	}
}

func TestStoreShipping_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "ToDelete", Price: 0.0, IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "PriceTest", Price: 1.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ShippingMethodInput{
		Name: "PriceTest", Price: 9.99, IsActive: true,
	}); err != nil {
		t.Fatalf("Update price: %v", err)
	}
}

func TestStoreShipping_Update_EstimatedDays(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "DaysTest", Price: 5.0, EstimatedDays: 3, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ShippingMethodInput{
		Name: "DaysTest", Price: 5.0, EstimatedDays: 7, IsActive: true,
	}); err != nil {
		t.Fatalf("Update estimated days: %v", err)
	}
}

func TestStoreShipping_Create_FreeShipping(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "Free Shipping", Price: 0.0, EstimatedDays: 10, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create free shipping: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestListShippingMethods_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for owner_id=0, got %d", len(items))
	}
}

func TestCreateShippingMethod_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "Persist Test", Price: 4.99, EstimatedDays: 3, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestCreateShippingMethod_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "NameStoredMethod"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: want, Price: 2.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			if m.Name != want {
				t.Errorf("Name = %q, want %q", m.Name, want)
			}
			return
		}
	}
	t.Errorf("created method id=%d not found", id)
}

func TestCreateShippingMethod_PriceIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = 12.50
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "PriceStoredMethod", Price: want, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			if m.Price != want {
				t.Errorf("Price = %v, want %v", m.Price, want)
			}
			return
		}
	}
	t.Errorf("created method id=%d not found", id)
}

func TestCreateShippingMethod_IsActiveIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "ActiveStoredMethod", Price: 1.0, IsActive: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			if m.IsActive {
				t.Errorf("expected IsActive=false, got true")
			}
			return
		}
	}
	t.Errorf("created method id=%d not found", id)
}

func TestUpdateShippingMethod_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "BeforeRename", Price: 5.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ShippingMethodInput{
		Name: "AfterRename", Price: 5.0, IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			if m.Name != "AfterRename" {
				t.Errorf("Name = %q, want AfterRename", m.Name)
			}
			return
		}
	}
	t.Errorf("method id=%d not found after update", id)
}

func TestUpdateShippingMethod_ChangesPrice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "UpdatePriceMethod", Price: 1.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ShippingMethodInput{
		Name: "UpdatePriceMethod", Price: 99.99, IsActive: true,
	}); err != nil {
		t.Fatalf("Update price: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			if m.Price != 99.99 {
				t.Errorf("Price = %v, want 99.99", m.Price)
			}
			return
		}
	}
	t.Errorf("method id=%d not found after price update", id)
}

func TestDeleteShippingMethod_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "DeleteRow", Price: 3.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			t.Errorf("method id=%d still exists after delete", id)
		}
	}
}

func TestDeleteShippingMethod_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Delete(context.Background(), testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestCreate_ThenList_FindsMethod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "FindMeMethod", Price: 7.0, EstimatedDays: 4, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range items {
		if m.ID == id {
			return
		}
	}
	t.Errorf("created method id=%d not found in list", id)
}

func TestListShippingMethods_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	before, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "CountMethod", Price: 6.0, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	after, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d after=%d", len(before), len(after))
	}
}

func TestCreateShippingMethod_EstimatedDaysIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = 14
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ShippingMethodInput{
		Name: "DaysStoredMethod", Price: 3.0, EstimatedDays: want, IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.BusinessID, id) })

	items, _ := repo.List(context.Background(), testutil.BusinessID)
	for _, m := range items {
		if m.ID == id {
			if m.EstimatedDays != want {
				t.Errorf("EstimatedDays = %d, want %d", m.EstimatedDays, want)
			}
			return
		}
	}
	t.Errorf("created method id=%d not found", id)
}

func TestListShippingMethods_IsNotNil(t *testing.T) {
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
