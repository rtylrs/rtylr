package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// CreateCycleV2 omitted workspace_id from its INSERT while the column is
// NOT NULL with no default, so every v2 cycle creation failed outright with
// "Error 1364: Field 'workspace_id' doesn't have a default value".
//
// Its siblings owner_id and business_id both carry DEFAULT 0; workspace_id does
// not, which is why only this insert broke.
func TestCreateCycleV2_SucceedsWithoutAWorkspace(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const owner int64 = 970601
	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: owner, BusinessID: testutil.BusinessID,
		Name: "Cycle Create Probe", CreatedBy: owner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_cycles WHERE project_id = ?", projID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID)
	})

	id, err := repo.CreateCycleV2(ctx, domain.CycleInput{
		BusinessID: testutil.BusinessID, ProjectID: &projID, Name: "Sprint 1",
	})
	if err != nil {
		t.Fatalf("CreateCycleV2: %v", err)
	}
	if id == 0 {
		t.Fatal("CreateCycleV2 returned id 0, so nothing was inserted")
	}

	// And it is readable back, which proves the row is well formed rather than
	// merely accepted.
	cycles, total, err := repo.ListCyclesV2(ctx, projID, testutil.BusinessID, owner, false, 50, 0)
	if err != nil {
		t.Fatalf("ListCyclesV2: %v", err)
	}
	if total != 1 || len(cycles) != 1 || cycles[0].Name != "Sprint 1" {
		t.Fatalf("expected the created cycle back, got total=%d cycles=%+v", total, cycles)
	}
}

// A business member who is not the creator, owner, or a project member must not
// see the project's cycles. Before the visibility clause, ListCyclesV2 scoped
// only by business_id and project_id, so passing a foreign project id returned
// its cycles: names and date ranges of a project the caller cannot open.
func TestListCyclesV2_HidesCyclesFromNonMembers(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const owner int64 = 970701
	const stranger int64 = 970702

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: owner, BusinessID: testutil.BusinessID,
		Name: "Private Cycle Project", CreatedBy: owner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_cycles WHERE project_id = ?", projID)
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID)
	})

	if _, err := repo.CreateCycleV2(ctx, domain.CycleInput{
		BusinessID: testutil.BusinessID, ProjectID: &projID, Name: "Secret Sprint",
	}); err != nil {
		t.Fatalf("CreateCycleV2: %v", err)
	}

	// The owner sees it. Asserting this too is what makes the absence below
	// meaningful: without it, a query returning nothing at all would pass.
	own, ownTotal, err := repo.ListCyclesV2(ctx, projID, testutil.BusinessID, owner, false, 50, 0)
	if err != nil {
		t.Fatalf("ListCyclesV2 as owner: %v", err)
	}
	if ownTotal != 1 || len(own) != 1 {
		t.Fatalf("owner should see their own cycle, got total=%d len=%d", ownTotal, len(own))
	}

	// The stranger sees nothing, and the count agrees. A total that still said 1
	// would leak the project's shape even with the rows withheld.
	got, total, err := repo.ListCyclesV2(ctx, projID, testutil.BusinessID, stranger, false, 50, 0)
	if err != nil {
		t.Fatalf("ListCyclesV2 as stranger: %v", err)
	}
	if len(got) != 0 || total != 0 {
		t.Fatalf("a non-member must see no cycles, got len=%d total=%d %+v", len(got), total, got)
	}

	// A hub-bypass caller still sees everything.
	all, allTotal, err := repo.ListCyclesV2(ctx, projID, testutil.BusinessID, stranger, true, 50, 0)
	if err != nil {
		t.Fatalf("ListCyclesV2 with bypass: %v", err)
	}
	if allTotal != 1 || len(all) != 1 {
		t.Fatalf("bypass should see the cycle, got total=%d len=%d", allTotal, len(all))
	}
}
