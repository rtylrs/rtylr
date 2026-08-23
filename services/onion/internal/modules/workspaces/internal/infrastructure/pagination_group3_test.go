package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestWorkspaces_ListCycles_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "CyclePageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	start := time.Now()
	for i := 0; i < 5; i++ {
		cycleID, err := repo.CreateCycle(ctx, domain.CycleInput{
			BusinessID: testutil.OwnerID, WorkspaceID: wsID,
			Name:      fmt.Sprintf("CyclePage Cycle %d", i),
			StartDate: start.AddDate(0, 0, i*14),
			EndDate:   start.AddDate(0, 0, i*14+13),
			CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateCycle %d: %v", i, err)
		}
		// No repository DeleteCycle exists yet (production has no way to
		// remove a cycle either), so clean up the fixture row directly.
		t.Cleanup(func(id int64) func() {
			return func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_cycles WHERE id = ?", id) }
		}(cycleID))
	}

	page1, total1, err := repo.ListCycles(ctx, wsID, 2, 0)
	if err != nil {
		t.Fatalf("ListCycles page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 cycles on page1, got %d", len(page1))
	}

	page2, total2, err := repo.ListCycles(ctx, wsID, 2, 2)
	if err != nil {
		t.Fatalf("ListCycles page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 cycles on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("cycle %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.ListCycles(ctx, wsID, 2, 4)
	if err != nil {
		t.Fatalf("ListCycles page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder cycle on page3, got %d", len(page3))
	}
}

func TestWorkspaces_ListMilestones_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "MilestonePageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "MilestonePage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM flow_projects WHERE id = ?", projID) })

	for i := 0; i < 5; i++ {
		milestoneID, err := repo.CreateMilestone(ctx, domain.MilestoneInput{
			ProjectID: projID, WorkspaceID: wsID, BusinessID: testutil.OwnerID,
			Name: fmt.Sprintf("MilestonePage Milestone %d", i), DueDate: time.Now().AddDate(0, 1, i), CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateMilestone %d: %v", i, err)
		}
		// No repository DeleteMilestone exists yet (production has no way to
		// remove a milestone either), so clean up the fixture row directly.
		t.Cleanup(func(id int64) func() {
			return func() { _, _ = db.ExecContext(context.Background(), "DELETE FROM flow_milestones WHERE id = ?", id) }
		}(milestoneID))
	}

	// projectID == nil path: workspace-scoped listing.
	page1, total1, err := repo.ListMilestones(ctx, wsID, nil, 2, 0)
	if err != nil {
		t.Fatalf("ListMilestones (workspace) page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 milestones on page1, got %d", len(page1))
	}

	page2, total2, err := repo.ListMilestones(ctx, wsID, nil, 2, 2)
	if err != nil {
		t.Fatalf("ListMilestones (workspace) page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 milestones on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("milestone %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.ListMilestones(ctx, wsID, nil, 2, 4)
	if err != nil {
		t.Fatalf("ListMilestones (workspace) page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder milestone on page3, got %d", len(page3))
	}

	// projectID != nil path: project-scoped listing, same underlying data.
	projPage1, projTotal1, err := repo.ListMilestones(ctx, wsID, &projID, 2, 0)
	if err != nil {
		t.Fatalf("ListMilestones (project) page1: %v", err)
	}
	if projTotal1 != 5 {
		t.Errorf("expected project-scoped total=5, got %d", projTotal1)
	}
	if len(projPage1) != 2 {
		t.Fatalf("expected 2 milestones on project-scoped page1, got %d", len(projPage1))
	}
}

func TestWorkspaces_ListViews_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, BusinessID: testutil.BusinessID, Name: "ViewPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.BusinessID) })

	for i := 0; i < 5; i++ {
		viewID, err := repo.CreateView(ctx, domain.ViewInput{
			BusinessID: testutil.BusinessID, WorkspaceID: wsID, Name: fmt.Sprintf("ViewPage View %d", i),
			Filters: map[string]interface{}{}, CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateView %d: %v", i, err)
		}
		if i == 0 {
			var storedBusinessID int64
			if err := db.QueryRowContext(ctx, "SELECT business_id FROM flow_views WHERE id = ?", viewID).Scan(&storedBusinessID); err != nil {
				t.Fatalf("read created view business_id: %v", err)
			}
			if storedBusinessID != testutil.BusinessID {
				t.Fatalf("created view business_id = %d, want %d", storedBusinessID, testutil.BusinessID)
			}
		}
		t.Cleanup(func(id int64) func() {
			return func() { _ = repo.DeleteView(ctx, id, testutil.OwnerID, testutil.BusinessID) }
		}(viewID))
	}

	page1, total1, err := repo.ListViews(ctx, testutil.BusinessID, wsID, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("ListViews page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 views on page1, got %d", len(page1))
	}

	page2, total2, err := repo.ListViews(ctx, testutil.BusinessID, wsID, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("ListViews page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 views on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("view %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.ListViews(ctx, testutil.BusinessID, wsID, testutil.OwnerID, 2, 4)
	if err != nil {
		t.Fatalf("ListViews page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder view on page3, got %d", len(page3))
	}
}
