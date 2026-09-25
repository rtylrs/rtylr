package application

import (
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/testutil"
)

// Synthetic tenant kept clear of script fixtures so the provisioning path
// (personal workspace + My Tasks project + columns) always executes below.
const (
	scopeUser     = int64(990001)
	scopeBusiness = int64(4242)
)

// EnsureMyTasksProject used to invalidate the workspaces cache with the
// caller's userID where every other write path invalidates with the business
// id - and ListWorkspaces caches under the business id. Unless the two numbers
// coincide, the eviction targeted an empty scope and the workspaces list kept
// serving pre-provisioning rows until the TTL expired. This pins the fix: the
// caller's business-scope entries are evicted while unrelated scopes survive.
func TestBusinessMyTasksCacheIsolatedByBusiness(t *testing.T) {
	ctx := initCache(t)
	const userID = int64(990011)
	const page = int64(1)
	const limit = int64(50)
	businessA := int64(7001)
	businessB := int64(7002)

	key := func(businessID int64) string {
		return cacheKeys.Scoped(
			"user",
			userID,
			fmt.Sprintf("business_tasks:business=%d:workspace_tasks:page=%d:limit=%d", businessID, page, limit),
		)
	}
	cacher.Of[myTasksPage]().Put(ctx, key(businessA), myTasksPage{
		Tasks: []domain.Task{{ID: 101}},
		Total: 1,
	}, cacher.WithExpiry(ctx, workspacesTTL))
	cacher.Of[myTasksPage]().Put(ctx, key(businessB), myTasksPage{
		Tasks: []domain.Task{{ID: 202}},
		Total: 1,
	}, cacher.WithExpiry(ctx, workspacesTTL))

	svc := &Service{}
	fromA, _, err := svc.GetMyTasksForBusiness(ctx, userID, businessA, page, limit)
	if err != nil {
		t.Fatalf("business A cache read: %v", err)
	}
	fromB, _, err := svc.GetMyTasksForBusiness(ctx, userID, businessB, page, limit)
	if err != nil {
		t.Fatalf("business B cache read: %v", err)
	}
	if len(fromA) != 1 || fromA[0].ID != 101 || len(fromB) != 1 || fromB[0].ID != 202 {
		t.Fatalf("business cache isolation failed: A=%+v B=%+v", fromA, fromB)
	}
}

func TestEnsureMyTasksProjectEvictsBusinessScope(t *testing.T) {
	ctx := initCache(t)
	db := testutil.OpenDB(t)
	svc := NewService(ctx, db)

	// Drop anything a previous run of this test provisioned, so the call below
	// always walks the create/rename paths that perform the eviction.
	cleanScope := func() {
		t.Helper()
		var wsIDs []int64
		rows, err := db.Query(
			"SELECT id FROM flow_workspaces WHERE owner_id = ? AND business_id = ?",
			scopeUser, scopeBusiness)
		if err != nil {
			t.Fatalf("query workspaces: %v", err)
		}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				t.Fatalf("scan workspace id: %v", err)
			}
			wsIDs = append(wsIDs, id)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("iterate workspaces: %v", err)
		}
		rows.Close()
		for _, id := range wsIDs {
			for _, stmt := range []string{
				"DELETE FROM flow_project_columns WHERE workspace_id = ?",
				"DELETE FROM flow_projects WHERE workspace_id = ?",
				"DELETE FROM flow_workspaces WHERE id = ?",
			} {
				if _, err := db.Exec(stmt, id); err != nil {
					t.Fatalf("cleanup (%s): %v", stmt, err)
				}
			}
		}
	}
	cleanScope()
	t.Cleanup(cleanScope)

	listKey := func(scope int64) string {
		return cacheKeys.Scoped(workspacesScope, scope,
			fmt.Sprintf("list:page=%d:limit=%d", 1, 50))
	}

	// A stale workspaces-list page cached under the real business id - the key
	// shape ListWorkspaces reads - and one under the caller's user scope that
	// this path must leave alone.
	businessKey := listKey(scopeBusiness)
	userKey := listKey(scopeUser)
	cacher.Of[workspacesPage]().Put(ctx, businessKey,
		workspacesPage{Total: 7}, cacher.WithExpiry(ctx, workspacesTTL))
	cacher.Of[workspacesPage]().Put(ctx, userKey,
		workspacesPage{Total: 3}, cacher.WithExpiry(ctx, workspacesTTL))

	if _, _, err := svc.EnsureMyTasksProject(ctx, scopeUser, scopeBusiness, 0); err != nil {
		t.Fatalf("EnsureMyTasksProject: %v", err)
	}

	if _, found := cacher.Of[workspacesPage]().Get(ctx, businessKey); found {
		t.Errorf("workspaces cache entry for business %d survived provisioning; "+
			"EnsureMyTasksProject likely invalidated with userID %d instead of businessID",
			scopeBusiness, scopeUser)
	}
	if _, found := cacher.Of[workspacesPage]().Get(ctx, userKey); !found {
		t.Errorf("unrelated user-scope entry for %d was evicted; invalidation is broader than intended",
			scopeUser)
	}
}
