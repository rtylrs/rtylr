package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/store/themes/internal/domain"
	"github.com/voxire/onion/internal/modules/store/themes/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestStoreThemes_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestStoreThemes_List_OwnerScoped(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "ListTestTheme"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "ThemeOld"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ThemeInput{Name: "ThemeNew"}); err != nil {
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name: "ColorTheme", PrimaryColor: "#000000",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ThemeInput{
		Name: "ColorTheme", PrimaryColor: "#FFFFFF",
	}); err != nil {
		t.Fatalf("Update color: %v", err)
	}
}

func TestStoreThemes_List_UnknownOwner_Empty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 9999999) // unknown business ID
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
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
		id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: name})
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
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "FullUpdate"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ThemeInput{
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

func TestStoreThemes_List_IsNotNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestStoreThemes_Create_FontIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "Helvetica"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name: "FontTheme", Font: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, th := range items {
		if th.ID == id {
			if th.Font != want {
				t.Errorf("Font = %q, want %q", th.Font, want)
			}
			return
		}
	}
	t.Errorf("theme id=%d not found", id)
}

func TestStoreThemes_Create_HeaderStyleIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "fixed"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name: "HeaderTheme", HeaderStyle: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, th := range items {
		if th.ID == id {
			if th.HeaderStyle != want {
				t.Errorf("HeaderStyle = %q, want %q", th.HeaderStyle, want)
			}
			return
		}
	}
	t.Errorf("theme id=%d not found", id)
}

func TestStoreThemes_Create_FooterStyleIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "compact"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name: "FooterTheme", FooterStyle: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, th := range items {
		if th.ID == id {
			if th.FooterStyle != want {
				t.Errorf("FooterStyle = %q, want %q", th.FooterStyle, want)
			}
			return
		}
	}
	t.Errorf("theme id=%d not found", id)
}

func TestStoreThemes_List_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	before, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	_, err = repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "CountTheme"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	after, err := repo.List(context.Background(), testutil.OwnerID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count +1: before=%d after=%d", len(before), len(after))
	}
}

func TestStoreThemes_Create_IDsAreUnique(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id1, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "UniqueThemeA"})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}
	id2, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "UniqueThemeB"})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}
	if id1 == id2 {
		t.Errorf("expected unique IDs, both got %d", id1)
	}
}

func TestStoreThemes_Update_Font(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name: "UpdateFontTheme", Font: "Arial",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ThemeInput{
		Name: "UpdateFontTheme", Font: "Georgia",
	}); err != nil {
		t.Fatalf("Update font: %v", err)
	}
}

func TestStoreThemes_Update_CustomCSS(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name: "UpdateCSSTheme",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ThemeInput{
		Name:      "UpdateCSSTheme",
		CustomCSS: "p { font-size: 16px; }",
	}); err != nil {
		t.Fatalf("Update CSS: %v", err)
	}
}

func TestStoreThemes_Update_ReflectedInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "ReflectTheme"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Update(context.Background(), testutil.BusinessID, id, domain.ThemeInput{
		Name: "ReflectThemeUpdated",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, th := range items {
		if th.ID == id {
			if th.Name != "ReflectThemeUpdated" {
				t.Errorf("Name = %q, want ReflectThemeUpdated", th.Name)
			}
			return
		}
	}
	t.Errorf("theme id=%d not found after update", id)
}

func TestStoreThemes_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	items, err := repo.List(context.Background(), 0) // zero business ID
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 for owner_id=0, got %d", len(items))
	}
}

func TestStoreThemes_List_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	_, _ = repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{Name: "IDCheckTheme"})

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, th := range items {
		if th.ID == 0 {
			t.Errorf("items[%d].ID is zero", i)
		}
	}
}

func TestStoreThemes_Create_PrimaryColorIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	const want = "#ABCDEF"
	id, err := repo.Create(context.Background(), testutil.OwnerID, testutil.BusinessID, domain.ThemeInput{
		Name:         "PrimaryColorTheme",
		PrimaryColor: want,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, th := range items {
		if th.ID == id {
			if th.PrimaryColor != want {
				t.Errorf("PrimaryColor = %q, want %q", th.PrimaryColor, want)
			}
			return
		}
	}
	t.Errorf("theme id=%d not found", id)
}
