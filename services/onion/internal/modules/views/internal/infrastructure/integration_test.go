package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/views/internal/domain"
	"github.com/voxire/onion/internal/modules/views/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const viewsWorkspaceID int64 = 1

func TestViews_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestViews_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, v := range items {
		if v.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, v.OwnerID)
		}
	}
}

func TestViews_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ViewInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: viewsWorkspaceID,
		Name:        "My Board View",
		Filters:     []byte(`{"status":"open"}`),
		IsDefault:   false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), id, testutil.OwnerID) })
}

func TestViews_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ViewInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: viewsWorkspaceID,
		Name:        "ListTestView",
		Filters:     []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), id, testutil.OwnerID) })

	items, err := repo.List(context.Background(), testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, v := range items {
		if v.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created view id=%d not found in list", id)
	}
}

func TestViews_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ViewInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: viewsWorkspaceID,
		Name:        "OldViewName",
		Filters:     []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), id, testutil.OwnerID) })

	if err := repo.Update(context.Background(), domain.ViewUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "NewViewName", Filters: []byte(`{}`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID, viewsWorkspaceID)
	for _, v := range items {
		if v.ID == id && v.Name != "NewViewName" {
			t.Errorf("expected name NewViewName, got %s", v.Name)
		}
	}
}

func TestViews_Update_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Update(context.Background(), domain.ViewUpdateInput{
		ID: 9999999, OwnerID: testutil.OwnerID, Name: "Missing", Filters: []byte(`{}`),
	})
	if err == nil {
		t.Fatal("expected error for missing view")
	}
}

func TestViews_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ViewInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: viewsWorkspaceID,
		Name:        "DeleteMe",
		Filters:     []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), id, testutil.OwnerID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID, viewsWorkspaceID)
	for _, v := range items {
		if v.ID == id {
			t.Errorf("deleted view id=%d still in list", id)
		}
	}
}

func TestViews_Delete_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Delete(context.Background(), 9999999, testutil.OwnerID)
	if err == nil {
		t.Fatal("expected error for missing view")
	}
}

func TestViews_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown owner, got %d", len(items))
	}
}

func TestViews_Create_WithSortBy(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ViewInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: viewsWorkspaceID,
		Name:        "SortedView",
		Filters:     []byte(`{}`),
		SortBy:      "created_at",
		GroupBy:     "status",
	})
	if err != nil {
		t.Fatalf("Create with sort/group: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), id, testutil.OwnerID) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestViews_Create_IsDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.ViewInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: viewsWorkspaceID,
		Name:        "DefaultView",
		Filters:     []byte(`{}`),
		IsDefault:   true,
	})
	if err != nil {
		t.Fatalf("Create default view: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), id, testutil.OwnerID) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}
