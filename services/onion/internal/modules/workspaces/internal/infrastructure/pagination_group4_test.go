package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_GetTeamWorkload_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "WorkloadPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	// GetTeamWorkload inner-joins flow_workspace_members to users, so members
	// need a real users row to show up at all.
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("workload-page-%d@example.com", i)
		res, err := db.ExecContext(ctx,
			"INSERT INTO users (email, password_hash, full_name) VALUES (?, ?, ?)",
			email, "", fmt.Sprintf("Workload Page User %d", i),
		)
		if err != nil {
			t.Fatalf("insert user %d: %v", i, err)
		}
		userID, err := res.LastInsertId()
		if err != nil {
			t.Fatalf("last insert id %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() {
			return func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM users WHERE id = ?", id) }
		}(userID))

		if err := repo.AddWorkspaceMember(ctx, domain.AddWorkspaceMemberInput{
			WorkspaceID: wsID, UserID: userID, Role: "member",
		}); err != nil {
			t.Fatalf("AddWorkspaceMember %d: %v", i, err)
		}
	}

	page1, total1, err := repo.GetTeamWorkload(ctx, wsID, 2, 0)
	if err != nil {
		t.Fatalf("GetTeamWorkload page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 members on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetTeamWorkload(ctx, wsID, 2, 2)
	if err != nil {
		t.Fatalf("GetTeamWorkload page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 members on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("member %s appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetTeamWorkload(ctx, wsID, 2, 4)
	if err != nil {
		t.Fatalf("GetTeamWorkload page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder member on page3, got %d", len(page3))
	}
}

func TestWorkspaces_GetRoadmap_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "RoadmapPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	for i := 0; i < 5; i++ {
		projID, err := repo.CreateProject(ctx, domain.ProjectInput{
			WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: fmt.Sprintf("RoadmapPage Project %d", i), CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateProject %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() {
			return func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_projects WHERE id = ?", id) }
		}(projID))
	}

	page1, total1, err := repo.GetRoadmap(ctx, wsID, 2, 0)
	if err != nil {
		t.Fatalf("GetRoadmap page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 projects on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetRoadmap(ctx, wsID, 2, 2)
	if err != nil {
		t.Fatalf("GetRoadmap page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 projects on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("project %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetRoadmap(ctx, wsID, 2, 4)
	if err != nil {
		t.Fatalf("GetRoadmap page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder project on page3, got %d", len(page3))
	}
}
