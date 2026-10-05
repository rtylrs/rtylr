package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/warehouses/internal/infrastructure"
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

func TestWarehouses_CreateWarehouse_UniqueIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "WH-Unique-1", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse 1: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id1) })

	id2, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "WH-Unique-2", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse 2: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id2) })

	if id1 == id2 {
		t.Errorf("expected distinct IDs, both got %d", id1)
	}
}

func TestWarehouses_UpdateWarehouse_Country(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name: "Country WH", Country: "US", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	if err := repo.UpdateWarehouse(ctx, testutil.OwnerID, id, domain.WarehouseInput{
		Name: "Country WH", Country: "CA", IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateWarehouse country: %v", err)
	}

	warehouses, _ := repo.ListWarehouses(ctx, testutil.OwnerID)
	for _, w := range warehouses {
		if w.ID == id && w.Country != "CA" {
			t.Errorf("expected Country=CA, got %s", w.Country)
		}
	}
}

func TestWarehouses_UpdateWarehouse_Capacity(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name: "Cap WH", Capacity: 100, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	if err := repo.UpdateWarehouse(ctx, testutil.OwnerID, id, domain.WarehouseInput{
		Name: "Cap WH", Capacity: 9999, IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateWarehouse capacity: %v", err)
	}
}

func TestWarehouses_CreateWarehouse_Inactive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name: "Inactive WH", IsActive: false,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse inactive: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestWarehouses_ListWarehouses_NamesAreNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "Named WH", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	warehouses, err := repo.ListWarehouses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWarehouses: %v", err)
	}
	for _, w := range warehouses {
		if w.Name == "" {
			t.Errorf("warehouse id=%d has empty Name", w.ID)
		}
	}
}

func TestWarehouses_UpdateWarehouse_Address(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name: "Addr WH", Address: "Old Address", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	if err := repo.UpdateWarehouse(ctx, testutil.OwnerID, id, domain.WarehouseInput{
		Name: "Addr WH", Address: "New Address 123", IsActive: true,
	}); err != nil {
		t.Fatalf("UpdateWarehouse address: %v", err)
	}
}

func TestWarehouses_CreateWarehouse_ManagerName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name: "Managed WH", ManagerName: "John Manager", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse with manager: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	warehouses, _ := repo.ListWarehouses(ctx, testutil.OwnerID)
	for _, w := range warehouses {
		if w.ID == id && w.ManagerName != "John Manager" {
			t.Errorf("expected ManagerName='John Manager', got %q", w.ManagerName)
		}
	}
}

func TestWarehouses_DeleteWarehouse_WrongOwnerFails(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name: "Protected WH", IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	// Attempt to delete with wrong owner -- should silently not delete it.
	_ = repo.DeleteWarehouse(ctx, 9999999, id)

	warehouses, _ := repo.ListWarehouses(ctx, testutil.OwnerID)
	found := false
	for _, w := range warehouses {
		if w.ID == id {
			found = true
		}
	}
	if !found {
		t.Error("warehouse was deleted by wrong owner -- owner isolation violated")
	}
}

func TestWarehouses_ListWarehouses_IDsArePositive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "ID Check WH", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	warehouses, err := repo.ListWarehouses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListWarehouses: %v", err)
	}
	for _, w := range warehouses {
		if w.ID <= 0 {
			t.Errorf("expected positive ID, got %d", w.ID)
		}
	}
}

func TestWarehouses_CreateWarehouse_FullAddress(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name:     "Full Addr WH",
		Address:  "456 Industrial Blvd",
		City:     "Metropolis",
		Country:  "US",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse full address: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	warehouses, _ := repo.ListWarehouses(ctx, testutil.OwnerID)
	for _, w := range warehouses {
		if w.ID == id {
			if w.City != "Metropolis" {
				t.Errorf("expected City=Metropolis, got %s", w.City)
			}
			return
		}
	}
	t.Errorf("warehouse id=%d not found", id)
}

func TestWarehouses_ListWarehouses_Idempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	a, err := repo.ListWarehouses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	b, err := repo.ListWarehouses(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(a) != len(b) {
		t.Errorf("idempotency: got %d then %d", len(a), len(b))
	}
}

func TestWarehouses_CountWarehouses_All(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "Count WH", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	count, err := repo.CountWarehouses(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("CountWarehouses: %v", err)
	}
	if count <= 0 {
		t.Errorf("expected at least 1 warehouse, got %d", count)
	}
}

func TestWarehouses_CountWarehouses_ActiveOnly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "Active Count WH", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	count, err := repo.CountWarehouses(ctx, testutil.OwnerID, true)
	if err != nil {
		t.Fatalf("CountWarehouses(activeOnly): %v", err)
	}
	if count <= 0 {
		t.Errorf("expected at least 1 active warehouse, got %d", count)
	}
}

func TestWarehouses_CountWarehouses_ZeroOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	count, err := repo.CountWarehouses(ctx, 0, false)
	if err != nil {
		t.Fatalf("CountWarehouses(ownerID=0): %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for ownerID=0, got %d", count)
	}
}

func TestWarehouses_CountWarehouses_AfterCreate_Increases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.CountWarehouses(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("CountWarehouses before: %v", err)
	}

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{Name: "New WH", IsActive: true})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	after, err := repo.CountWarehouses(ctx, testutil.OwnerID, false)
	if err != nil {
		t.Fatalf("CountWarehouses after: %v", err)
	}
	if after != before+1 {
		t.Errorf("expected count to increase by 1, before=%d after=%d", before, after)
	}
}

func TestWarehouses_GetWarehouseByID_Created(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name:     "GetByID Test WH",
		City:     "TestCity",
		Country:  "US",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	w, err := repo.GetWarehouseByID(ctx, testutil.OwnerID, id)
	if err != nil {
		t.Fatalf("GetWarehouseByID: %v", err)
	}
	if w.ID != id {
		t.Errorf("expected ID=%d, got %d", id, w.ID)
	}
	if w.Name != "GetByID Test WH" {
		t.Errorf("expected Name='GetByID Test WH', got %q", w.Name)
	}
}

func TestWarehouses_GetWarehouseByID_NotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetWarehouseByID(ctx, testutil.OwnerID, 9999999)
	if !errors.Is(err, domain.ErrWarehouseNotFound) {
		t.Fatalf("expected ErrWarehouseNotFound, got %v", err)
	}
}

func TestWarehouses_GetWarehouseByID_WrongOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.CreateWarehouse(ctx, testutil.OwnerID, domain.WarehouseInput{
		Name:     "Owner Scope WH",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateWarehouse: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWarehouse(ctx, testutil.OwnerID, id) })

	_, err = repo.GetWarehouseByID(ctx, 9999999, id)
	if !errors.Is(err, domain.ErrWarehouseNotFound) {
		t.Fatalf("expected ErrWarehouseNotFound for wrong owner, got %v", err)
	}
}
