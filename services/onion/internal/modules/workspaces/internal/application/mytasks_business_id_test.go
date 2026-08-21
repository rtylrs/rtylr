package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// EnsureMyTasksProject used to write the caller's user id into the created
// project's business_id column instead of their real business id, since the
// two happened to be passed as the same argument at every call site. This
// pins the fix: the created project must be findable by its real business,
// not by the (numerically different) user id.
func TestEnsureMyTasksProjectUsesRealBusinessID(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	svc := NewService(ctx, db)

	projectID, workspaceID, err := svc.EnsureMyTasksProject(ctx, cacheTestUserA, cacheTestBusiness, 0)
	if err != nil {
		t.Fatalf("EnsureMyTasksProject: %v", err)
	}

	projects, _, err := svc.ListProjects(ctx, workspaceID, cacheTestBusiness, 50, 0)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}

	found := false
	for _, p := range projects {
		if p.ID == projectID {
			found = true
			if p.BusinessID != cacheTestBusiness {
				t.Errorf("BusinessID = %d, want %d (the caller's real business, not their user id %d)",
					p.BusinessID, cacheTestBusiness, cacheTestUserA)
			}
		}
	}
	if !found {
		t.Fatalf("My Tasks project %d not found when listed by its real business id %d - business_id was likely written wrong",
			projectID, cacheTestBusiness)
	}
}
