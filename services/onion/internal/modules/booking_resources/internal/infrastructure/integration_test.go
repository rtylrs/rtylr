package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/booking_resources/internal/domain"
	"github.com/voxire/onion/internal/modules/booking_resources/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListResources_ZeroOwner_ReturnsEmpty(t *testing.T) {
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

func TestListResources_IsNotNil(t *testing.T) {
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

func TestCreateResource_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "Room A",
		Type:     "room",
		Capacity: 5,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })
}

func TestCreateResource_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "Chair 1",
		Type:     "station",
		Capacity: 1,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.Name != "Chair 1" {
				t.Errorf("expected name Chair 1, got %s", r.Name)
			}
			return
		}
	}
	t.Errorf("resource id=%d not found in list", id)
}

func TestCreateResource_TypeIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "Studio B",
		Type:     "room",
		Capacity: 10,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.Type != "room" {
				t.Errorf("expected type room, got %s", r.Type)
			}
			return
		}
	}
	t.Errorf("resource id=%d not found in list", id)
}

func TestCreateResource_CapacityIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "Conference Room",
		Type:     "room",
		Capacity: 20,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if r.Capacity != 20 {
				t.Errorf("expected capacity 20, got %d", r.Capacity)
			}
			return
		}
	}
	t.Errorf("resource id=%d not found in list", id)
}

func TestCreateResource_IsActiveIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "Active Room",
		Type:     "room",
		Capacity: 4,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			if !r.IsActive {
				t.Errorf("expected IsActive=true, got false")
			}
			return
		}
	}
	t.Errorf("resource id=%d not found in list", id)
}

func TestUpdateResource_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "OldRoomName",
		Type:     "room",
		Capacity: 3,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.ResourceInput{
		Name:     "NewRoomName",
		Type:     "room",
		Capacity: 3,
		IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id && r.Name != "NewRoomName" {
			t.Errorf("expected name NewRoomName, got %s", r.Name)
		}
	}
}

func TestUpdateResource_ChangesCapacity(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "CapRoom",
		Type:     "room",
		Capacity: 5,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	if err := repo.Update(ctx, testutil.OwnerID, id, domain.ResourceInput{
		Name:     "CapRoom",
		Type:     "room",
		Capacity: 15,
		IsActive: true,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id && r.Capacity != 15 {
			t.Errorf("expected capacity 15, got %d", r.Capacity)
		}
	}
}

func TestDeleteResource_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "DeleteRoom",
		Type:     "room",
		Capacity: 2,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			t.Errorf("deleted resource id=%d still in list", id)
		}
	}
}

func TestDeleteResource_NonExistentIsNoop(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	if err := repo.Delete(ctx, testutil.OwnerID, 99999999); err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestCreate_ThenList_FindsResource(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "FindMeRoom",
		Type:     "room",
		Capacity: 8,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, r := range items {
		if r.ID == id {
			return
		}
	}
	t.Errorf("resource id=%d not found after create", id)
}

func TestListResources_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
		Name:     "CountRoom",
		Type:     "room",
		Capacity: 6,
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, testutil.OwnerID, id) })

	after, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) <= len(before) {
		t.Errorf("expected count to increase: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateResource_MultipleResources_AllReturned(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	names := []string{"MultiRoom1", "MultiRoom2", "MultiRoom3"}
	var ids []int64
	for _, name := range names {
		id, err := repo.Create(ctx, testutil.OwnerID, domain.ResourceInput{
			Name:     name,
			Type:     "room",
			Capacity: 4,
			IsActive: true,
		})
		if err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = repo.Delete(ctx, testutil.OwnerID, id)
		}
	})

	items, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := 0
	for _, item := range items {
		for _, id := range ids {
			if item.ID == id {
				found++
			}
		}
	}
	if found != len(ids) {
		t.Errorf("expected %d resources in list, found %d", len(ids), found)
	}
}
