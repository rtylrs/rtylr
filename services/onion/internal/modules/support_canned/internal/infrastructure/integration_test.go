package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/support_canned/internal/domain"
	"github.com/voxire/onion/internal/modules/support_canned/internal/infrastructure"
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
