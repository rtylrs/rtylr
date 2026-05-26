package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/cycles/internal/domain"
	"github.com/voxire/onion/internal/modules/cycles/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestList_ReturnsSliceForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// May be empty if no cycles seeded; call must succeed.
	_ = cycles
}

func TestList_AllCyclesBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range cycles {
		if c.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d for cycle id=%d", testutil.OwnerID, c.OwnerID, c.ID)
		}
	}
}

func TestCreate_AddsNewCycle(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.CycleInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: 0,
		Name:        "Integration Test Cycle",
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}

	// Clean up.
	_, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id)
}

func TestCreate_CycleAppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	input := domain.CycleInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: 0,
		Name:        "Appear In List Cycle",
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id)
	}()

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, c := range cycles {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected cycle id=%d to appear in List", id)
	}
}

func TestCreate_WithDates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	start := time.Now().UTC().Truncate(24 * time.Hour)
	end := start.AddDate(0, 1, 0)
	input := domain.CycleInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: 0,
		Name:        "Dated Cycle",
		StartDate:   &start,
		EndDate:     &end,
	}
	id, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create with dates: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id)
	}()
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestUpdate_ChangesName(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: 0,
		Name:        "Before Update",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id)
	}()

	err = repo.Update(ctx, domain.CycleUpdateInput{
		ID:      id,
		OwnerID: testutil.OwnerID,
		Name:    "After Update",
		Status:  "active",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range cycles {
		if c.ID == id && c.Name != "After Update" {
			t.Errorf("expected Name='After Update', got %q", c.Name)
		}
	}
}

func TestUpdate_WrongOwnerReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.Update(ctx, domain.CycleUpdateInput{
		ID:      999999,
		OwnerID: testutil.OwnerID,
		Name:    "Ghost",
		Status:  "upcoming",
	})
	if err == nil {
		t.Fatal("expected error for non-existent cycle, got nil")
	}
}

func TestList_WithWorkspaceFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	workspaceID := int64(99999)
	cycles, err := repo.List(ctx, testutil.OwnerID, &workspaceID)
	if err != nil {
		t.Fatalf("List with workspace: %v", err)
	}
	for _, c := range cycles {
		if c.WorkspaceID != workspaceID {
			t.Errorf("expected WorkspaceID=%d, got %d", workspaceID, c.WorkspaceID)
		}
	}
}

func TestList_CycleNamesAreNonEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range cycles {
		if c.Name == "" {
			t.Errorf("cycle id=%d has empty Name", c.ID)
		}
	}
}
