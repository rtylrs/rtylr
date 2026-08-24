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
	cycles, total, err := repo.ListCyclesV2(ctx, projID, testutil.BusinessID, 50, 0)
	if err != nil {
		t.Fatalf("ListCyclesV2: %v", err)
	}
	if total != 1 || len(cycles) != 1 || cycles[0].Name != "Sprint 1" {
		t.Fatalf("expected the created cycle back, got total=%d cycles=%+v", total, cycles)
	}
}
