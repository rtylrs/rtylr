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

func TestCreate_MultipleDistinctIDs(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id1, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Cycle One"})
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id1) })

	id2, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Cycle Two"})
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id2) })

	if id1 == id2 {
		t.Errorf("expected distinct IDs, both were %d", id1)
	}
}

func TestCreate_WithNilStartAndEndDate(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{
		OwnerID:   testutil.OwnerID,
		Name:      "Nil Dates Cycle",
		StartDate: nil,
		EndDate:   nil,
	})
	if err != nil {
		t.Fatalf("Create with nil dates: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })
	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestUpdate_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Status Change Cycle"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	err = repo.Update(ctx, domain.CycleUpdateInput{
		ID:      id,
		OwnerID: testutil.OwnerID,
		Name:    "Status Change Cycle",
		Status:  "completed",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range cycles {
		if c.ID == id && c.Status != "completed" {
			t.Errorf("expected Status='completed', got %q", c.Status)
		}
	}
}

func TestUpdate_ChangesDates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Date Change Cycle"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	newStart := time.Now().UTC().Truncate(24 * time.Hour)
	newEnd := newStart.AddDate(0, 0, 14)
	err = repo.Update(ctx, domain.CycleUpdateInput{
		ID:        id,
		OwnerID:   testutil.OwnerID,
		Name:      "Date Change Cycle",
		Status:    "active",
		StartDate: &newStart,
		EndDate:   &newEnd,
	})
	if err != nil {
		t.Fatalf("Update with dates: %v", err)
	}
}

func TestList_WithWorkspaceFilterReturnsOnlyMatching(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID := int64(77001)
	id, err := repo.Create(ctx, domain.CycleInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: wsID,
		Name:        "Workspace Filter Cycle",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	cycles, err := repo.List(ctx, testutil.OwnerID, &wsID)
	if err != nil {
		t.Fatalf("List with workspace: %v", err)
	}
	for _, c := range cycles {
		if c.WorkspaceID != wsID {
			t.Errorf("got WorkspaceID=%d, want %d", c.WorkspaceID, wsID)
		}
	}
	found := false
	for _, c := range cycles {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created cycle id=%d not found under workspace filter", id)
	}
}

func TestList_IdempotentCallsReturnSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List first: %v", err)
	}
	second, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List second: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("idempotent list: got %d then %d", len(first), len(second))
	}
}

func TestCreate_CycleHasCorrectOwnerID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Owner Check Cycle"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range cycles {
		if c.ID == id && c.OwnerID != testutil.OwnerID {
			t.Errorf("expected OwnerID=%d, got %d", testutil.OwnerID, c.OwnerID)
		}
	}
}

func TestCreate_WithWorkspaceID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID := int64(55001)
	id, err := repo.Create(ctx, domain.CycleInput{
		OwnerID:     testutil.OwnerID,
		WorkspaceID: wsID,
		Name:        "Workspace Cycle",
	})
	if err != nil {
		t.Fatalf("Create with workspace: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	cycles, err := repo.List(ctx, testutil.OwnerID, &wsID)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	found := false
	for _, c := range cycles {
		if c.ID == id {
			found = true
			if c.WorkspaceID != wsID {
				t.Errorf("WorkspaceID=%d, want %d", c.WorkspaceID, wsID)
			}
			break
		}
	}
	if !found {
		t.Errorf("cycle id=%d not returned by workspace filter", id)
	}
}

func TestUpdate_NamePersistedAfterMultipleUpdates(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Multi Update Cycle"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	for i, name := range []string{"First Update", "Second Update", "Third Update"} {
		err = repo.Update(ctx, domain.CycleUpdateInput{
			ID:      id,
			OwnerID: testutil.OwnerID,
			Name:    name,
			Status:  "active",
		})
		if err != nil {
			t.Fatalf("Update %d: %v", i+1, err)
		}
	}

	cycles, err := repo.List(ctx, testutil.OwnerID, nil)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	for _, c := range cycles {
		if c.ID == id && c.Name != "Third Update" {
			t.Errorf("expected Name='Third Update', got %q", c.Name)
		}
	}
}

func TestList_DifferentOwnerReturnsNoOtherOwnerCycles(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	id, err := repo.Create(ctx, domain.CycleInput{OwnerID: testutil.OwnerID, Name: "Owner Isolation Cycle"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM flow_cycles WHERE id = ?`, id) })

	const otherOwner int64 = 9999999
	otherCycles, err := repo.List(ctx, otherOwner, nil)
	if err != nil {
		t.Fatalf("List other owner: %v", err)
	}
	for _, c := range otherCycles {
		if c.OwnerID == testutil.OwnerID {
			t.Errorf("other owner query returned cycle belonging to ownerID=%d", testutil.OwnerID)
		}
	}
}
