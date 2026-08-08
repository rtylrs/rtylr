package infrastructure_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// Sentinel business ID unlikely to collide with fixture/other-test data for
// ListWorkspaces, which is scoped by business_id rather than a workspace this
// test creates itself.
const paginationBusinessID = int64(848484848)

func TestWorkspaces_GetWorkspacesByOwner_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
			OwnerID: paginationBusinessID, BusinessID: paginationBusinessID,
			Name: fmt.Sprintf("WSPage WS %d", i), CreatedBy: paginationBusinessID,
		}, false)
		if err != nil {
			t.Fatalf("CreateWorkspace %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteWorkspace(ctx, id, paginationBusinessID) } }(wsID))
	}

	page1, total1, err := repo.GetWorkspacesByOwner(ctx, paginationBusinessID, 2, 0)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 workspaces on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetWorkspacesByOwner(ctx, paginationBusinessID, 2, 2)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 workspaces on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("workspace %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetWorkspacesByOwner(ctx, paginationBusinessID, 2, 4)
	if err != nil {
		t.Fatalf("GetWorkspacesByOwner page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder workspace on page3, got %d", len(page3))
	}
}

func TestWorkspaces_GetProjectsByWorkspace_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "ProjListPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	for i := 0; i < 5; i++ {
		_, err := repo.CreateProject(ctx, domain.ProjectInput{
			WorkspaceID: wsID, OwnerID: testutil.OwnerID, BusinessID: testutil.OwnerID,
			Name: fmt.Sprintf("ProjListPage Project %d", i), CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateProject %d: %v", i, err)
		}
	}

	page1, total1, err := repo.GetProjectsByWorkspace(ctx, wsID, testutil.OwnerID, 2, 0)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 projects on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetProjectsByWorkspace(ctx, wsID, testutil.OwnerID, 2, 2)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace page2: %v", err)
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

	page3, total3, err := repo.GetProjectsByWorkspace(ctx, wsID, testutil.OwnerID, 2, 4)
	if err != nil {
		t.Fatalf("GetProjectsByWorkspace page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder project on page3, got %d", len(page3))
	}
}

func TestWorkspaces_GetColumnsByProject_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "ColumnPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "ColumnPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, err := repo.CreateColumn(ctx, domain.ProjectColumnInput{
			ProjectID: projID, WorkspaceID: wsID, Name: fmt.Sprintf("Column %d", i),
			Color: "#aabbcc", SortOrder: i, CreatedBy: testutil.OwnerID,
		})
		if err != nil {
			t.Fatalf("CreateColumn %d: %v", i, err)
		}
	}

	page1, total1, err := repo.GetColumnsByProject(ctx, projID, 2, 0)
	if err != nil {
		t.Fatalf("GetColumnsByProject page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 columns on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetColumnsByProject(ctx, projID, 2, 2)
	if err != nil {
		t.Fatalf("GetColumnsByProject page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 columns on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("column %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetColumnsByProject(ctx, projID, 2, 4)
	if err != nil {
		t.Fatalf("GetColumnsByProject page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder column on page3, got %d", len(page3))
	}
}

func TestWorkspaces_ListWorkspaceMembers_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "WSMemberPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	for i := 0; i < 5; i++ {
		userID := int64(700000000 + i)
		if err := repo.AddWorkspaceMember(ctx, domain.AddWorkspaceMemberInput{
			WorkspaceID: wsID, UserID: userID, Role: "member",
		}); err != nil {
			t.Fatalf("AddWorkspaceMember %d: %v", i, err)
		}
	}

	page1, total1, err := repo.ListWorkspaceMembers(ctx, wsID, 2, 0)
	if err != nil {
		t.Fatalf("ListWorkspaceMembers page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 members on page1, got %d", len(page1))
	}

	page2, total2, err := repo.ListWorkspaceMembers(ctx, wsID, 2, 2)
	if err != nil {
		t.Fatalf("ListWorkspaceMembers page2: %v", err)
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
				t.Errorf("member %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.ListWorkspaceMembers(ctx, wsID, 2, 4)
	if err != nil {
		t.Fatalf("ListWorkspaceMembers page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder member on page3, got %d", len(page3))
	}
}

func TestWorkspaces_ListProjectMembers_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "ProjMemberPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: testutil.OwnerID, Name: "ProjMemberPage Project", CreatedBy: testutil.OwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	for i := 0; i < 5; i++ {
		userID := int64(710000000 + i)
		if err := repo.AddProjectMember(ctx, domain.AddProjectMemberInput{
			ProjectID: projID, UserID: userID, Role: "member",
		}); err != nil {
			t.Fatalf("AddProjectMember %d: %v", i, err)
		}
	}

	page1, total1, err := repo.ListProjectMembers(ctx, projID, 2, 0)
	if err != nil {
		t.Fatalf("ListProjectMembers page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 members on page1, got %d", len(page1))
	}

	page2, total2, err := repo.ListProjectMembers(ctx, projID, 2, 2)
	if err != nil {
		t.Fatalf("ListProjectMembers page2: %v", err)
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
				t.Errorf("member %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.ListProjectMembers(ctx, projID, 2, 4)
	if err != nil {
		t.Fatalf("ListProjectMembers page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder member on page3, got %d", len(page3))
	}
}

func TestWorkspaces_GetLabelsByWorkspace_PaginationAndTotal(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: testutil.OwnerID, Name: "LabelPageWS", CreatedBy: testutil.OwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, testutil.OwnerID) })

	for i := 0; i < 5; i++ {
		labelID, err := repo.CreateLabel(ctx, domain.LabelInput{
			WorkspaceID: wsID, Name: fmt.Sprintf("Label %d", i), Color: "#112233",
		})
		if err != nil {
			t.Fatalf("CreateLabel %d: %v", i, err)
		}
		t.Cleanup(func(id int64) func() { return func() { _ = repo.DeleteLabel(ctx, id, testutil.OwnerID) } }(labelID))
	}

	page1, total1, err := repo.GetLabelsByWorkspace(ctx, wsID, 2, 0)
	if err != nil {
		t.Fatalf("GetLabelsByWorkspace page1: %v", err)
	}
	if total1 != 5 {
		t.Errorf("expected total=5, got %d", total1)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 labels on page1, got %d", len(page1))
	}

	page2, total2, err := repo.GetLabelsByWorkspace(ctx, wsID, 2, 2)
	if err != nil {
		t.Fatalf("GetLabelsByWorkspace page2: %v", err)
	}
	if total2 != 5 {
		t.Errorf("expected total=5 on page2, got %d", total2)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 labels on page2, got %d", len(page2))
	}
	for _, p1 := range page1 {
		for _, p2 := range page2 {
			if p1.ID == p2.ID {
				t.Errorf("label %d appeared on both page1 and page2", p1.ID)
			}
		}
	}

	page3, total3, err := repo.GetLabelsByWorkspace(ctx, wsID, 2, 4)
	if err != nil {
		t.Fatalf("GetLabelsByWorkspace page3: %v", err)
	}
	if total3 != 5 {
		t.Errorf("expected total=5 on page3, got %d", total3)
	}
	if len(page3) != 1 {
		t.Errorf("expected 1 remainder label on page3, got %d", len(page3))
	}
}
