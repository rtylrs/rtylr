package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store_themes/internal/domain"
	"github.com/voxire/onion/internal/modules/store_themes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreThemes_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreThemes_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, th := range items {
		if th.OwnerID != testutil.OwnerID {
			t.Errorf("expected owner_id %d, got %d", testutil.OwnerID, th.OwnerID)
		}
	}
}

func TestStoreThemes_Create_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{
		Name:         "Integration Theme",
		PrimaryColor: "#FF5733",
		Font:         "Roboto",
		HeaderStyle:  "sticky",
		FooterStyle:  "minimal",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive ID, got %d", id)
	}
}

func TestStoreThemes_Create_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{Name: "ListTestTheme"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, th := range items {
		if th.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("created theme id=%d not found in list", id)
	}
}

func TestStoreThemes_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{Name: "ThemeOld"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.ThemeInput{Name: "ThemeNew"}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, _ := repo.List(context.Background(), testutil.OwnerID)
	for _, th := range items {
		if th.ID == id && th.Name != "ThemeNew" {
			t.Errorf("expected name ThemeNew, got %s", th.Name)
		}
	}
}

func TestStoreThemes_Update_PrimaryColor(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{
		Name: "ColorTheme", PrimaryColor: "#000000",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.ThemeInput{
		Name: "ColorTheme", PrimaryColor: "#FFFFFF",
	}); err != nil {
		t.Fatalf("Update color: %v", err)
	}
}

func TestStoreThemes_List_UnknownOwner_Empty(t *testing.T) {
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

func TestStoreThemes_Create_WithCustomCSS(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{
		Name:      "CSSTheme",
		CustomCSS: "body { background: #eee; }",
	})
	if err != nil {
		t.Fatalf("Create with CSS: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestStoreThemes_Create_MultipleThemes(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	names := []string{"Dark Theme", "Light Theme", "Minimal Theme"}
	for _, name := range names {
		id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{Name: name})
		if err != nil {
			t.Fatalf("Create %s: %v", name, err)
		}
		if id <= 0 {
			t.Errorf("theme %s: expected positive ID", name)
		}
	}
}

func TestStoreThemes_Update_AllFields(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, domain.ThemeInput{Name: "FullUpdate"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.OwnerID, id, domain.ThemeInput{
		Name:         "FullUpdate",
		PrimaryColor: "#3498db",
		Font:         "Open Sans",
		HeaderStyle:  "transparent",
		FooterStyle:  "full",
		CustomCSS:    "h1 { color: red; }",
	}); err != nil {
		t.Fatalf("Update all fields: %v", err)
	}
}
