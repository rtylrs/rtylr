package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

// R13: creating a task with no project_id files it in the caller's My Tasks
// bucket. The bucket is lazily provisioned on first call.
func TestCreateTaskV2WithoutProjectLandsInBucket(t *testing.T) {
	svc, ctx, _ := v2Svc(t)

	task, err := svc.CreateTask(ctx, domain.TaskInput{
		BusinessID: v2Business,
		OwnerID:    v2UserA,
		Title:      "orphan capture",
		ReporterID: v2UserA,
	})
	if err != nil {
		t.Fatalf("CreateTask without project: %v", err)
	}
	if task.ProjectID == 0 {
		t.Fatal("task landed nowhere; expected My Tasks bucket")
	}

	bucketID, _, err := svc.EnsureMyTasksProject(ctx, v2UserA, v2Business, 0)
	if err != nil {
		t.Fatalf("EnsureMyTasksProject: %v", err)
	}
	if task.ProjectID != bucketID {
		t.Errorf("task landed in project %d, want bucket %d", task.ProjectID, bucketID)
	}
}

// T1: assigning someone to a task auto-grants membership so they can see it.
func TestAssigneeAutoMembership(t *testing.T) {
	svc, ctx, db := v2Svc(t)

	p := createV2Project(ctx, t, svc, "assign-grant", "", v2UserA)
	assignee := v2UserB

	if _, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID:  p.ID,
		BusinessID: v2Business,
		OwnerID:    v2UserA,
		Title:      "for you",
		ReporterID: v2UserA,
		AssigneeID: &assignee,
	}); err != nil {
		t.Fatalf("CreateTask with assignee: %v", err)
	}

	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM flow_project_members WHERE project_id = ? AND user_id = ?",
		p.ID, assignee).Scan(&count); err != nil {
		t.Fatalf("count members: %v", err)
	}
	if count != 1 {
		t.Errorf("assignee has %d membership rows, want 1", count)
	}

	// The assignee can now list the project they were assigned into.
	got, total, err := svc.ListProjectsV2(ctx, assignee, v2Business, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 as assignee: %v", err)
	}
	found := false
	for _, lp := range got {
		if lp.ID == p.ID {
			found = true
		}
	}
	if !found || total < 1 {
		t.Errorf("assignee cannot see the project they were assigned to (found=%v total=%d)", found, total)
	}
}

// S5: business-wide task listing is visibility-gated row by row. A stranger
// sees zero tasks from a private project even though the query spans the whole
// business.
func TestListTasksV2VisibilityGating(t *testing.T) {
	svc, ctx, _ := v2Svc(t)

	private := createV2Project(ctx, t, svc, "private-tasks", "", v2UserA)
	if _, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID: private.ID, BusinessID: v2Business, OwnerID: v2UserA,
		Title: "secret work", ReporterID: v2UserA,
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	_, strangerTotal, err := svc.ListTasksV2(ctx, v2UserB, v2Business, domain.ListTasksFilter{})
	if err != nil {
		t.Fatalf("ListTasksV2 stranger: %v", err)
	}
	if strangerTotal != 0 {
		t.Errorf("stranger sees %d tasks from a private project, want 0", strangerTotal)
	}

	_, creatorTotal, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{})
	if err != nil {
		t.Fatalf("ListTasksV2 creator: %v", err)
	}
	if creatorTotal < 1 {
		t.Errorf("creator sees %d tasks, want >= 1", creatorTotal)
	}
}

func TestListTasksV2FiltersProjectAndPagination(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	first := createV2Project(ctx, t, svc, "first-project", "", v2UserA)
	second := createV2Project(ctx, t, svc, "second-project", "", v2UserA)
	for _, project := range []domain.Project{first, first, second} {
		if _, err := svc.CreateTask(ctx, domain.TaskInput{ProjectID: project.ID, BusinessID: v2Business, OwnerID: v2UserA, Title: "project task", ReporterID: v2UserA}); err != nil {
			t.Fatal(err)
		}
	}
	for _, offset := range []int64{0, 1} {
		rows, total, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{ProjectID: first.ID, Limit: 1, Offset: offset})
		if err != nil {
			t.Fatal(err)
		}
		if total != 2 || len(rows) != 1 || rows[0].ProjectID != first.ID {
			t.Fatalf("project page %d = %v, total %d", offset, rows, total)
		}
	}
	rows, total, err := svc.ListTasksV2(ctx, v2UserA, v2Business, domain.ListTasksFilter{ProjectID: second.ID, Limit: 10})
	if err != nil || total != 1 || len(rows) != 1 || rows[0].ProjectID != second.ID {
		t.Fatalf("second project = %v, total %d, error %v", rows, total, err)
	}
}
