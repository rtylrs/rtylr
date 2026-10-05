package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// Ruling P8: the project listing is curated for EVERYONE, bypass roles included.
// "Bypass grants ACCESS everywhere, never listing."
//
// These assert the repository honours both sides of includeAll, which is what
// makes the service-level decision meaningful: ListProjectsV2 now passes false
// unconditionally, so a hub Owner/Admin/Operator gets the same curated list as
// anybody else. Their access is unaffected, and RequireProjectRole still admits
// them; only what appears in a list changes.
func TestListProjectsV2_CuratedListingExcludesOthersProjects(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := testsupport.NewRepository(db)
	ctx := context.Background()

	const stranger int64 = 970801
	const otherOwner int64 = 970802

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		OwnerID: otherOwner, BusinessID: testutil.BusinessID,
		Name: "P8 Someone Elses Project", CreatedBy: otherOwner,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID)
	})

	// Curated (includeAll=false): the stranger does not see it. This is the path
	// ListProjectsV2 now takes for every caller, admins included.
	curated, _, err := repo.ListProjectsV2(ctx, testutil.BusinessID, stranger, false, nil, "", "", "", 200, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 curated: %v", err)
	}
	for _, p := range curated {
		if p.ID == projID {
			t.Fatalf("curated listing leaked project %d belonging to another user", projID)
		}
	}

	// includeAll=true still returns it. Asserting this too is what proves the
	// curated result above is the filter working rather than the seed missing:
	// without it, a broken query returning nothing would pass.
	all, _, err := repo.ListProjectsV2(ctx, testutil.BusinessID, stranger, true, nil, "", "", "", 200, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 includeAll: %v", err)
	}
	found := false
	for _, p := range all {
		if p.ID == projID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("includeAll=true should return the project; the seed or the query is wrong")
	}
}
