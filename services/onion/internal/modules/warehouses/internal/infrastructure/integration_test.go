package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/warehouses/internal/domain"
	"github.com/voxire/onion/internal/modules/warehouses/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWarehouses_ListWarehouses_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.ListWarehouses(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWarehouses: %v", err)
	}
}

func TestWarehouses_ListWarehouses_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	warehouses, err := repo.ListWarehouses(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWarehouses: %v", err)
	}
	for _, w := range warehouses {
		if w.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, w.OwnerID)
		}
	}
}

func TestWarehouses_CreateWarehouse_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name:     "Main Warehouse",
		Address:  "123 Storage Rd",
		City:     "Testville",
		Country:  "US",
		Capacity: 1000,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id) })
}

func TestWarehouses_CreateWarehouse_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name: "List Test Warehouse", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id) })

	warehouses, err := repo.ListWarehouses(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWarehouses: %v", err)
	}
	found := false
	for _, w := range warehouses {
		if w.ID == id {
			found = true
			if w.Name != "List Test Warehouse" {
				t.Errorf("expected name 'List Test Warehouse', got %s", w.Name)
			}
		}
	}
	if !found {
		t.Errorf("created warehouse id=%d not found in list", id)
	}
}

func TestWarehouses_UpdateWarehouse_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name: "OldWarehouse", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id) })

	if err := repo.UpdateWarehouse(context.Background(), testutil.OwnerID, id, domain.WarehouseInput{
		Name: "NewWarehouse", IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateWarehouse: %v", err)
	}

	warehouses, _ := repo.ListWarehouses(context.Background(), testutil.OwnerID)
	for _, w := range warehouses {
		if w.ID == id && w.Name != "NewWarehouse" {
			t.Errorf("expected name NewWarehouse, got %s", w.Name)
		}
	}
}

func TestWarehouses_DeleteWarehouse_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name: "DeleteMe", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}

	if err := repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("DeleteWarehouse: %v", err)
	}

	warehouses, _ := repo.ListWarehouses(context.Background(), testutil.OwnerID)
	for _, w := range warehouses {
		if w.ID == id {
			t.Errorf("deleted warehouse id=%d still in list", id)
		}
	}
}

func TestWarehouses_ListWarehouses_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	warehouses, err := repo.ListWarehouses(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("ListWarehouses unknown owner: %v", err)
	}
	if len(warehouses) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(warehouses))
	}
}

func TestWarehouses_CreateWarehouse_DefaultWarehouse(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name:      "Default WH",
		IsDefault: true,
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse default: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id) })

	warehouses, _ := repo.ListWarehouses(context.Background(), testutil.OwnerID)
	for _, w := range warehouses {
		if w.ID == id && !w.IsDefault {
			t.Error("expected IsDefault=true")
		}
	}
}

func TestWarehouses_UpdateWarehouse_City(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name: "CityWH", City: "OldCity", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id) })

	if err := repo.UpdateWarehouse(context.Background(), testutil.OwnerID, id, domain.WarehouseInput{
		Name: "CityWH", City: "NewCity", IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateWarehouse city: %v", err)
	}
}

func TestWarehouses_CreateWarehouse_WithCapacity(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.CreateWarehouse(context.Background(), testutil.OwnerID, domain.WarehouseInput{
		Name:     "Large WH",
		Capacity: 5000,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse with capacity: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
