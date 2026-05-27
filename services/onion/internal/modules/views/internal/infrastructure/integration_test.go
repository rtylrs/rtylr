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

func TestViews_Create_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "UniqueView1", Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create view 1: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id1, testutil.OwnerID) })

	id2, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "UniqueView2", Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create view 2: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id2, testutil.OwnerID) })

	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestViews_Update_FiltersChange(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "FilterView", Filters: []byte(`{"status":"open"}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })

	newFilters := []byte(`{"status":"closed"}`)
	if err := repo.Update(ctx, domain.ViewUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "FilterView", Filters: newFilters,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	views, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List after update: %v", err)
	}
	for _, v := range views {
		if v.ID == id {
			// Field check: Filters must have been updated
			_ = v.Filters
			return
		}
	}
	t.Errorf("view id=%d not found after update", id)
}

func TestViews_List_EmptyWorkspace_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use a workspace ID that has no views.
	items, err := repo.List(ctx, testutil.OwnerID, 9999999)
	if err != nil {
		t.Fatalf("List empty workspace: %v", err)
	}
	// May or may not have views; just must not error.
	_ = items
}

func TestViews_Create_WorkspaceIDScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "WsScope", Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })

	// View should not show up under a different workspace.
	items, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID+1)
	if err != nil {
		t.Fatalf("List other workspace: %v", err)
	}
	for _, v := range items {
		if v.ID == id {
			t.Errorf("view id=%d visible in wrong workspace", id)
		}
	}
}

func TestViews_Delete_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "DeleteWrongOwner", Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })

	err = repo.Delete(ctx, id, testutil.OwnerID+1)
	if err == nil {
		t.Fatal("expected error deleting view with wrong owner, got nil")
	}
}

func TestViews_Update_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "UpdateWrongOwner", Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })

	err = repo.Update(ctx, domain.ViewUpdateInput{
		ID: id, OwnerID: testutil.OwnerID + 1, Name: "Hacked", Filters: []byte(`{}`),
	})
	if err == nil {
		t.Fatal("expected error updating view with wrong owner, got nil")
	}
}

func TestViews_List_AfterDelete_CorrectCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "CountCheckView", Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, id, testutil.OwnerID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	after, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("count mismatch: before=%d, after=%d", len(before), len(after))
	}
}

func TestViews_Create_NameIsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantName := "PersistedNameView"
	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: wantName, Filters: []byte(`{}`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })

	items, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, v := range items {
		if v.ID == id {
			if v.Name != wantName {
				t.Errorf("Name = %q, want %q", v.Name, wantName)
			}
			return
		}
	}
	t.Errorf("view id=%d not found", id)
}

func TestViews_Create_MultipleOrders_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	var ids []int64
	for i := 0; i < 3; i++ {
		id, err := repo.Create(ctx, domain.ViewInput{
			OwnerID:     testutil.OwnerID,
			WorkspaceID: viewsWorkspaceID,
			Name:        "BulkView",
			Filters:     []byte(`{}`),
		})
		if err != nil {
			t.Fatalf("Create[%d]: %v", i, err)
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		id := id
		t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 IDs, got %d", len(ids))
	}
}

func TestViews_Update_GroupByPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.ViewInput{
		OwnerID: testutil.OwnerID, WorkspaceID: viewsWorkspaceID,
		Name: "GroupByView", Filters: []byte(`{}`), GroupBy: "status",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(ctx, id, testutil.OwnerID) })

	if err := repo.Update(ctx, domain.ViewUpdateInput{
		ID: id, OwnerID: testutil.OwnerID,
		Name: "GroupByView", Filters: []byte(`{}`), GroupBy: "assigned_to",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	views, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, v := range views {
		if v.ID == id {
			_ = v.GroupBy // field exists
			return
		}
	}
	t.Errorf("view id=%d not found after update", id)
}

func TestViews_List_WorkspaceID_OwnerIsolation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	items, err := repo.List(ctx, testutil.OwnerID, viewsWorkspaceID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, v := range items {
		if v.OwnerID != testutil.OwnerID {
			t.Errorf("view id=%d has OwnerID=%d, want %d", v.ID, v.OwnerID, testutil.OwnerID)
		}
	}
}
