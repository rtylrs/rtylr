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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Welcome Email",
		Type:       "email",
		Subject:    "Welcome to our service!",
		Body:       "Hello {{name}}, welcome!",
		Variables:  []byte(`["name"]`),
		IsActive:   true,
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
		BusinessID: testutil.BusinessID,
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
		BusinessID: testutil.BusinessID,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "NewName", Type: "sms", Body: "body", Variables: []byte(`[]`),
		BusinessID: testutil.BusinessID,
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
		BusinessID: testutil.BusinessID,
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
		BusinessID: testutil.BusinessID,
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
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "SMS Template",
		Type:       "sms",
		Body:       "Your code is {{code}}",
		Variables:  []byte(`["code"]`),
		IsActive:   true,
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
		BusinessID: testutil.BusinessID,
		Variables:  []byte(`[]`), IsActive: true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "ActiveToggle", Type: "email", Body: "body",
		BusinessID: testutil.BusinessID,
		Variables:  []byte(`[]`), IsActive: false,
	}); err != nil {
		t.Fatalf("Update IsActive: %v", err)
	}
}

func TestTemplates_Create_SubjectStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "SubjectTest",
		Type:       "email",
		Subject:    "Test Subject Line",
		Body:       "body",
		Variables:  []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, tmpl := range items {
		if tmpl.ID == id && tmpl.Subject != "Test Subject Line" {
			t.Errorf("expected subject 'Test Subject Line', got %s", tmpl.Subject)
		}
	}
}

func TestTemplates_Update_Body(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "BodyUpdate", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "old body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "BodyUpdate", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "new body", Variables: []byte(`[]`),
	}); err != nil {
		t.Fatalf("Update body: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, tmpl := range items {
		if tmpl.ID == id && tmpl.Body != "new body" {
			t.Errorf("expected body 'new body', got %s", tmpl.Body)
		}
	}
}

func TestTemplates_Create_PushType(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "Push Notification Template",
		Type:       "push",
		Body:       "You have a new message",
		Variables:  []byte(`[]`),
		IsActive:   true,
	})
	if err != nil {
		t.Fatalf("Create push type: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTemplates_List_EachHasPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "IDCheckTmpl", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tmpl := range items {
		if tmpl.ID <= 0 {
			t.Errorf("template has non-positive ID: %d", tmpl.ID)
		}
	}
}

func TestTemplates_List_EachHasNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "NameCheckTmpl", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tmpl := range items {
		if tmpl.Name == "" {
			t.Errorf("template id=%d has empty name", tmpl.ID)
		}
	}
}

func TestTemplates_Create_Variables_MultipleVars(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	vars := []byte(`["name","email","order_id"]`)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "MultiVar Template",
		Type:       "email",
		Body:       "Hello {{name}}",
		Variables:  vars,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestTemplates_Update_Variables(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "VarUpdateTmpl", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`["v1"]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	if err := repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: testutil.OwnerID, Name: "VarUpdateTmpl", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`["v1","v2"]`),
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
}

func TestTemplates_Create_InactiveByDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID:    testutil.OwnerID,
		BusinessID: testutil.BusinessID,
		Name:       "InactiveTmpl",
		Type:       "email",
		Body:       "body",
		Variables:  []byte(`[]`),
		IsActive:   false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, tmpl := range items {
		if tmpl.ID == id && tmpl.IsActive {
			t.Errorf("expected IsActive=false, got true")
		}
	}
}

func TestTemplates_Update_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "WrongOwnerTmpl", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	err = repo.Update(context.Background(), domain.TemplateUpdateInput{
		ID: id, OwnerID: 9999999, Name: "WrongOwnerTmpl", Type: "email",
		Body: "body", Variables: []byte(`[]`),
	})
	if err == nil {
		t.Fatal("expected error when updating with wrong owner, got nil")
	}
}

func TestTemplates_Delete_WrongOwner_ReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "DeleteWrongOwner", Type: "email",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), testutil.OwnerID, id) })

	err = repo.Delete(context.Background(), 9999999, id)
	if err == nil {
		t.Fatal("expected error when deleting with wrong owner, got nil")
	}
}

func TestTemplates_List_ReturnedSliceNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice from List")
	}
}

func TestTemplates_Create_ThenDeleteThenNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), domain.TemplateInput{
		OwnerID: testutil.OwnerID, Name: "LifecycleTmpl", Type: "sms",
		BusinessID: testutil.BusinessID,
		Body:       "body", Variables: []byte(`[]`),
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(context.Background(), testutil.OwnerID, id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, tmpl := range items {
		if tmpl.ID == id {
			t.Errorf("deleted template id=%d still in list", id)
		}
	}
}
