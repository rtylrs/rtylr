package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/templates/internal/domain"
	"github.com/voxire/onion/internal/modules/templates/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestTemplates_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestTemplates_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tmpl := range items {
		if tmpl.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, tmpl.OwnerID)
		}
	}
}

func TestTemplates_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID:   testutil.OwnerID,
		Name:      "Welcome Email",
		Type:      "email",
		Subject:   "Welcome to our service!",
		Body:      "Hello {{name}}, welcome!",
		Variables: []byte(`["name"]`),
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
}

func TestTemplates_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "ListTest", Type: "email", Body: "body", Variables: []byte(`[]`),
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
	for _, tmpl := range items {
		if tmpl.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created template id=%d not found in list", id)
	}
}

func TestTemplates_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "OldName", Type: "sms", Body: "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "NewName", Type: "sms", Body: "body", Variables: []byte(`[]`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, tmpl := range items {
		if tmpl.ID == id && tmpl.Name != "NewName" {
			t.Errorf("expected name NewName, got %s", tmpl.Name)
		}
	}
}

func TestTemplates_Update_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: 9999999, OwnerID: testutil.OwnerID, Name: "Missing", Type: "email", Body: "body", Variables: []byte(`[]`),
	})
	if err == nil {
		t.Fatal("expected error for missing template")
	}
}

func TestTemplates_Delete_RemovesFromList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "DeleteMe", Type: "email", Body: "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, tmpl := range items {
		if tmpl.ID == id {
			t.Errorf("deleted template id=%d still in list", id)
		}
	}
}

func TestTemplates_Delete_NotFound_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	err := repo.Delete(context.Background(), testutil.OwnerID, 9999999)
	if err == nil {
		t.Fatal("expected error for missing template")
	}
}

func TestTemplates_List_UnknownOwner_Empty(t *testing.T) {
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

func TestTemplates_Create_SMSType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID:   testutil.OwnerID,
		Name:      "SMS Template",
		Type:      "sms",
		Body:      "Your code is {{code}}",
		Variables: []byte(`["code"]`),
		IsActive:  true,
	})
	if err != nil {
		t.Fatalf("Create SMS: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTemplates_Update_IsActive(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "ActiveToggle", Type: "email", Body: "body",
		Variables: []byte(`[]`), IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "ActiveToggle", Type: "email", Body: "body",
		Variables: []byte(`[]`), IsActive: false,
	}); err != nil {
		t.Fatalf("Update IsActive: %v", err)
	}
}
