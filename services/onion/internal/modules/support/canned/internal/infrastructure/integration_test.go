package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support/canned/internal/domain"
	"github.com/voxire/onion/internal/modules/support/canned/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

const testBusinessID int64 = 1

func TestSupportCanned_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestSupportCanned_List_BusinessScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.BusinessID != testBusinessID {
			t.Errorf("expected business_id %d, got %d", testBusinessID, c.BusinessID)
		}
	}
}

func TestSupportCanned_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID,
		Name:       "Greeting",
		Body:       "Hello! How can I help you today?",
		Category:   "general",
		Shortcut:   "/greet",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })
}

func TestSupportCanned_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID,
		Name:       "ListTest",
		Body:       "This is a test response.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	items, err := repo.List(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, c := range items {
		if c.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created canned response id=%d not found in list", id)
	}
}

func TestSupportCanned_Update_ChangesBody(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "UpdateTest", Body: "Old body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	if err := repo.Update(context.Background(), testBusinessID, id, domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "UpdateTest", Body: "New body",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID == id && c.Body != "New body" {
			t.Errorf("expected body 'New body', got %s", c.Body)
		}
	}
}

func TestSupportCanned_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "DeleteMe", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testBusinessID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID == id {
			t.Errorf("deleted canned response id=%d still in list", id)
		}
	}
}

func TestSupportCanned_List_UnknownBusiness_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for unknown business, got %d", len(items))
	}
}

func TestSupportCanned_Create_WithShortcut(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID,
		Name:       "Shortcut Test",
		Body:       "Thanks for reaching out!",
		Shortcut:   "/thanks",
	})
	if err != nil {
		t.Fatalf("Create with shortcut: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID == id && c.Shortcut != "/thanks" {
			t.Errorf("expected shortcut /thanks, got %s", c.Shortcut)
		}
	}
}

func TestSupportCanned_Update_Category(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "CatTest", Body: "body", Category: "billing",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	if err := repo.Update(context.Background(), testBusinessID, id, domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "CatTest", Body: "body", Category: "technical",
	}); err != nil {
		t.Fatalf("Update category: %v", err)
	}
}

func TestSupportCanned_Update_Name(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "OriginalName", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	if err := repo.Update(context.Background(), testBusinessID, id, domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "UpdatedName", Body: "body",
	}); err != nil {
		t.Fatalf("Update name: %v", err)
	}

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID == id && c.Name != "UpdatedName" {
			t.Errorf("expected name 'UpdatedName', got %s", c.Name)
		}
	}
}

func TestSupportCanned_Update_Shortcut(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "ShortcutUpdate", Body: "body", Shortcut: "/old",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	if err := repo.Update(context.Background(), testBusinessID, id, domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "ShortcutUpdate", Body: "body", Shortcut: "/new",
	}); err != nil {
		t.Fatalf("Update shortcut: %v", err)
	}

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID == id && c.Shortcut != "/new" {
			t.Errorf("expected shortcut '/new', got %s", c.Shortcut)
		}
	}
}

func TestSupportCanned_Create_FieldsPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID,
		Name:       "FieldCheck",
		Body:       "Persistent body content",
		Category:   "support",
		Shortcut:   "/fcheck",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	items, err := repo.List(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range items {
		if c.ID == id {
			if c.Name != "FieldCheck" {
				t.Errorf("expected name 'FieldCheck', got %s", c.Name)
			}
			if c.Body != "Persistent body content" {
				t.Errorf("expected body 'Persistent body content', got %s", c.Body)
			}
			if c.Category != "support" {
				t.Errorf("expected category 'support', got %s", c.Category)
			}
			if c.Shortcut != "/fcheck" {
				t.Errorf("expected shortcut '/fcheck', got %s", c.Shortcut)
			}
			return
		}
	}
	t.Errorf("created canned response id=%d not found in list", id)
}

func TestSupportCanned_Delete_WrongBusiness_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// Deleting a non-existent row with a different business ID should not error.
	err := repo.Delete(context.Background(), 9999999, 9999999)
	if err != nil {
		t.Fatalf("Delete non-existent: %v", err)
	}
}

func TestSupportCanned_List_ReturnsCannedResponseSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestSupportCanned_Create_Multiple_AllInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	names := []string{"Multi-A", "Multi-B", "Multi-C"}
	ids := make([]int64, 0, len(names))
	for _, name := range names {
		id, err := repo.Create(context.Background(), domain.CannedResponseInput{
			BusinessID: testBusinessID, Name: name, Body: "body",
		})
		if err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
		ids = append(ids, id)
	}
	t.Cleanup(func() {
		for _, id := range ids {
			_ = repo.Delete(context.Background(), testBusinessID, id)
		}
	})

	items, err := repo.List(context.Background(), testBusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := make(map[int64]bool)
	for _, c := range items {
		found[c.ID] = true
	}
	for _, id := range ids {
		if !found[id] {
			t.Errorf("created canned response id=%d not found in list", id)
		}
	}
}

func TestSupportCanned_Update_WrongBusiness_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	// Update on wrong business or missing ID should silently succeed (0 rows updated).
	err := repo.Update(context.Background(), 9999999, 9999999, domain.CannedResponseInput{
		BusinessID: 9999999, Name: "ghost", Body: "ghost",
	})
	if err != nil {
		t.Fatalf("Update on missing row: %v", err)
	}
}

func TestSupportCanned_Create_BusinessIDStoredCorrectly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "BizCheck", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID == id && c.BusinessID != testBusinessID {
			t.Errorf("expected business_id %d, got %d", testBusinessID, c.BusinessID)
		}
	}
}

func TestSupportCanned_List_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.CannedResponseInput{
		BusinessID: testBusinessID, Name: "IDCheck", Body: "body",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testBusinessID, id) })

	items, _ := repo.List(context.Background(), testBusinessID)
	for _, c := range items {
		if c.ID <= 0 {
			t.Errorf("canned response has non-positive ID: %d", c.ID)
		}
	}
}
