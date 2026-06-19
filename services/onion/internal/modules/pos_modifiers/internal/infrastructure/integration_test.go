package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_modifiers/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestModifiers_List_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	mods, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	_ = mods
}

func TestModifiers_List_ReturnsNonNilSlice(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	mods, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if mods == nil {
		t.Fatal("expected non-nil slice, got nil")
	}
}

func TestModifiers_List_WrongOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	mods, err := repo.List(context.Background(), 999999)
	if err != nil {
		t.Fatalf("List wrong owner: %v", err)
	}
	if len(mods) != 0 {
		t.Fatalf("expected 0 modifiers for unknown owner, got %d", len(mods))
	}
}

func TestModifiers_List_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	mods, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, m := range mods {
		if m.BusinessID != testutil.BusinessID {
			t.Errorf("mods[%d].BusinessID = %d, want %d", i, m.BusinessID, testutil.BusinessID)
		}
	}
}

func TestModifiers_List_AllHaveNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	mods, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, m := range mods {
		if m.ID == 0 {
			t.Errorf("mods[%d].ID is zero", i)
		}
	}
}

func TestModifiers_List_AllHaveNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	mods, err := repo.List(context.Background(), testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for i, m := range mods {
		if m.Name == "" {
			t.Errorf("mods[%d].Name is empty", i)
		}
	}
}

func TestModifiers_Create_ThenListContainsNew(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.ModifierInput{
		Name:       "Extra Cheese",
		PriceDelta: 1.5,
		GroupName:  "Add-ons",
		IsRequired: false,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == 0 {
		t.Fatal("Create returned ID=0")
	}

	mods, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after create: %v", err)
	}
	found := false
	for _, m := range mods {
		if m.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("new modifier id=%d not found in list", id)
	}
}

func TestModifiers_Update_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.ModifierInput{
		Name:       "Before Update",
		PriceDelta: 0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Update(ctx, testutil.BusinessID, id, domain.ModifierInput{
		Name:       "After Update",
		PriceDelta: 2.0,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	mods, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range mods {
		if m.ID == id {
			if m.Name != "After Update" {
				t.Errorf("Name = %q, want 'After Update'", m.Name)
			}
			return
		}
	}
	t.Errorf("modifier id=%d not found after update", id)
}

func TestModifiers_Update_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, testutil.BusinessID, 9999999, domain.ModifierInput{Name: "Ghost"})
	if !errors.Is(err, domain.ErrModifierNotFound) {
		t.Fatalf("expected ErrModifierNotFound, got %v", err)
	}
}

func TestModifiers_Delete_ThenNotInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.ModifierInput{
		Name:       "To Be Deleted",
		PriceDelta: 0,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Delete(ctx, testutil.BusinessID, id)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	mods, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, m := range mods {
		if m.ID == id {
			t.Errorf("deleted modifier id=%d still found in list", id)
		}
	}
}

func TestModifiers_Delete_UnknownIDReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, testutil.BusinessID, 9999999)
	if !errors.Is(err, domain.ErrModifierNotFound) {
		t.Fatalf("expected ErrModifierNotFound, got %v", err)
	}
}

func TestModifiers_Create_PriceDeltaStoredCorrectly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, domain.ModifierInput{
		Name:       "Price Check",
		PriceDelta: 3.75,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	mods, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, m := range mods {
		if m.ID == id {
			if m.PriceDelta != 3.75 {
				t.Errorf("PriceDelta = %f, want 3.75", m.PriceDelta)
			}
			return
		}
	}
	t.Errorf("modifier id=%d not found", id)
}
