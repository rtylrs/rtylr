package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/pos_combos/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_combos/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListCombos_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
}

func TestListCombos_UnknownBusinessReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	combos, err := repo.List(ctx, 0)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(combos) != 0 {
		t.Errorf("expected empty list for business_id=0, got %d combos", len(combos))
	}
}

func TestListCombos_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	combos, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, c := range combos {
		if c.ID == 0 {
			t.Errorf("combo[%d] has zero ID", i)
		}
	}
}

func TestListCombos_EachBusinessIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	combos, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, c := range combos {
		if c.BusinessID != testutil.BusinessID {
			t.Errorf("combo[%d] has BusinessID=%d, want %d", i, c.BusinessID, testutil.BusinessID)
		}
	}
}

func TestListCombos_EachHasNonEmptyName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	combos, err := repo.List(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for i, c := range combos {
		if c.Name == "" {
			t.Errorf("combo[%d] (id=%d) has empty Name", i, c.ID)
		}
	}
}

func TestCreateCombo_InsertsRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	before, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List before: %v", err)
	}

	input := domain.ComboInput{
		Name:        "Test Combo Integration",
		Description: "burger + soda",
		Price:       15.99,
		Items:       `[{"id":1,"qty":1},{"id":2,"qty":1}]`,
		IsActive:    true,
	}
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_combos WHERE id = ?", id)
	})

	after, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after: %v", err)
	}
	if len(after) != len(before)+1 {
		t.Errorf("expected count to grow by 1: before=%d, after=%d", len(before), len(after))
	}
}

func TestCreateCombo_PricePersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantPrice := 24.99
	input := domain.ComboInput{
		Name:     "Price Test Combo",
		Price:    wantPrice,
		Items:    `[]`,
		IsActive: true,
	}
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_combos WHERE id = ?", id)
	})

	combos, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, c := range combos {
		if c.ID == id {
			if c.Price != wantPrice {
				t.Errorf("Price mismatch: got %g, want %g", c.Price, wantPrice)
			}
			return
		}
	}
	t.Errorf("inserted combo id=%d not found", id)
}

func TestUpdateCombo_UpdatesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ComboInput{
		Name:     "Original Name",
		Price:    10.00,
		Items:    `[]`,
		IsActive: true,
	}
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM pos_combos WHERE id = ?", id)
	})

	updateInput := domain.ComboInput{
		Name:     "Updated Name",
		Price:    10.00,
		Items:    `[]`,
		IsActive: true,
	}
	if err := repo.Update(ctx, testutil.BusinessID, id, updateInput); err != nil {
		t.Fatalf("Update error: %v", err)
	}

	combos, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	for _, c := range combos {
		if c.ID == id {
			if c.Name != "Updated Name" {
				t.Errorf("Name not updated: got %q", c.Name)
			}
			return
		}
	}
	t.Errorf("combo id=%d not found after update", id)
}

func TestDeleteCombo_RemovesRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ComboInput{
		Name:     "To Be Deleted",
		Price:    5.00,
		Items:    `[]`,
		IsActive: true,
	}
	id, err := repo.Create(ctx, testutil.OwnerID, testutil.BusinessID, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if err := repo.Delete(ctx, testutil.BusinessID, id); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	combos, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, c := range combos {
		if c.ID == id {
			t.Errorf("combo id=%d still exists after Delete", id)
		}
	}
}

func TestUpdateCombo_UnknownIDReturnsErrComboNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.ComboInput{Name: "ghost", Price: 1.00, Items: `[]`, IsActive: false}
	err := repo.Update(ctx, testutil.BusinessID, 0, input)
	if !errors.Is(err, domain.ErrComboNotFound) {
		t.Errorf("expected ErrComboNotFound, got %v", err)
	}
}

func TestDeleteCombo_UnknownIDReturnsErrComboNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Delete(ctx, testutil.BusinessID, 0)
	if !errors.Is(err, domain.ErrComboNotFound) {
		t.Errorf("expected ErrComboNotFound, got %v", err)
	}
}

func TestListCombos_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("first List: %v", err)
	}
	second, err := repo.List(ctx, testutil.BusinessID)
	if err != nil {
		t.Fatalf("second List: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed: %d vs %d", len(first), len(second))
	}
}
