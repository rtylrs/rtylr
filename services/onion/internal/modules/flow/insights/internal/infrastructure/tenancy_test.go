package infrastructure_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Distinct ids, well clear of the fixture range, so a failure here cannot be a
// seeded row or another test's leftovers.
const (
	roadmapOwnerA    int64 = 940100
	roadmapBusinessA int64 = 940200
	roadmapOwnerB    int64 = 940300
	roadmapBusinessB int64 = 940400
)

func seedRoadmapProject(t *testing.T, db *sql.DB, ownerID, businessID int64, name string) int64 {
	t.Helper()
	ctx := context.Background()

	res, err := db.ExecContext(ctx,
		`INSERT INTO flow_projects (workspace_id, owner_id, business_id, name, status, start_date, due_date, created_by)
		 VALUES (0, ?, ?, ?, 'active', '2027-01-01', '2027-12-31', ?)`,
		ownerID, businessID, name, ownerID,
	)
	if err != nil {
		t.Fatalf("insert flow_projects(%s): %v", name, err)
	}
	projectID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId(%s): %v", name, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", projectID)
	})

	// A milestone for the same project. GetMilestonesByProjectIDs used to filter
	// on project_id alone, so this row is what proves the tenant filter is real
	// rather than inherited from the caller's project list.
	if _, err := db.ExecContext(ctx,
		`INSERT INTO flow_milestones (owner_id, business_id, project_id, name, target_date)
		 VALUES (?, ?, ?, ?, '2027-06-30')`,
		ownerID, businessID, projectID, name+" milestone",
	); err != nil {
		t.Fatalf("insert flow_milestones(%s): %v", name, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_milestones WHERE project_id = ?", projectID)
	})

	return projectID
}

// GetRoadmap is one of the three readers insights.dashboard.v1's handlers now
// delegate to, so a leak here would surface on an owner's cross-app dashboard.
//
// The existing tests in this package all pass a single tenant id and assert a
// property of the result, so they hold whether or not business_id scopes
// anything. This one seeds two tenants and asserts the second is absent.
//
// Both halves are asserted deliberately: absence alone would pass if the query
// returned nothing at all. (tryrtylr/onion#700 adds a shared helper for exactly
// this; asserted inline here so this branch does not depend on that one.)
func TestGetRoadmap_DoesNotReturnAnotherTenantsProjects(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	mine := seedRoadmapProject(t, db, roadmapOwnerA, roadmapBusinessA, "Tenant A project")
	theirs := seedRoadmapProject(t, db, roadmapOwnerB, roadmapBusinessB, "Tenant B project")

	result, err := repo.GetRoadmap(ctx, roadmapBusinessA)
	if err != nil {
		t.Fatalf("GetRoadmap: %v", err)
	}

	seen := make(map[int64]bool, len(result.Projects))
	for _, p := range result.Projects {
		seen[p.ID] = true
	}

	if !seen[mine] {
		t.Errorf("own project %d missing from the result, so absence below proves nothing", mine)
	}
	if seen[theirs] {
		t.Errorf("cross-tenant leak: project %d belongs to business %d", theirs, roadmapBusinessB)
	}
}

// The milestone query is a separate statement from the project query, and it was
// the one filtering on project_id alone. Asserting the milestone comes back for
// the caller's own project keeps the added business_id filter honest: a filter
// that matched nothing would also produce no leak.
func TestGetRoadmap_StillReturnsOwnMilestones(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	mine := seedRoadmapProject(t, db, roadmapOwnerA, roadmapBusinessA, "Tenant A milestones")

	result, err := repo.GetRoadmap(ctx, roadmapBusinessA)
	if err != nil {
		t.Fatalf("GetRoadmap: %v", err)
	}

	for _, p := range result.Projects {
		if p.ID != mine {
			continue
		}
		if len(p.Milestones) == 0 {
			t.Fatalf("project %d has no milestones; the business_id filter matches nothing", mine)
		}
		return
	}
	t.Fatalf("project %d not returned at all", mine)
}
