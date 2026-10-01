package application

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

// Synthetic tenant kept clear of fixtures so runs are deterministic on the
// persistent test database.
const (
	v2UserA    = int64(880001)
	v2UserB    = int64(880002)
	v2Business = int64(5201)
)

func v2Svc(t *testing.T) (*Service, context.Context, *sql.DB) {
	t.Helper()
	db := testutil.OpenDB(t)
	ctx := context.Background()

	// Wipe everything this suite created under its business so prior runs never
	// leak rows into current assertions (persistent DB, shared across suites).
	for _, stmt := range []string{
		"DELETE ct FROM flow_cycle_tasks ct JOIN flow_cycles c ON c.id = ct.cycle_id WHERE c.business_id = ?",
		"DELETE FROM flow_cycles WHERE business_id = ?",
		"DELETE mt FROM flow_milestone_tasks mt JOIN flow_milestones m ON m.id = mt.milestone_id WHERE m.business_id = ?",
		"DELETE FROM flow_milestones WHERE business_id = ?",
		"DELETE a FROM flow_task_attachments a JOIN flow_tasks t ON t.id = a.task_id WHERE t.business_id = ?",
		"DELETE tl FROM flow_time_logs tl JOIN flow_tasks t ON t.id = tl.task_id WHERE t.business_id = ?",
		"DELETE c FROM flow_task_comments c JOIN flow_tasks t ON t.id = c.task_id WHERE t.business_id = ?",
		"DELETE act FROM flow_task_activity act JOIN flow_tasks t ON t.id = act.task_id WHERE t.business_id = ?",
		"DELETE fp FROM flow_task_focus_pins fp JOIN flow_tasks t ON t.id = fp.task_id WHERE t.business_id = ?",
		"DELETE FROM flow_task_dependencies WHERE task_id IN (SELECT id FROM flow_tasks WHERE business_id = ?)",
		"DELETE FROM flow_tasks WHERE business_id = ?",
		"DELETE FROM flow_project_columns WHERE workspace_id IN (SELECT id FROM flow_workspaces WHERE business_id = ?)",
		"DELETE pm FROM flow_project_members pm JOIN flow_projects p ON p.id = pm.project_id WHERE p.business_id = ?",
		"DELETE FROM flow_projects WHERE business_id = ?",
		"DELETE FROM flow_workspaces WHERE business_id = ?",
	} {
		if _, err := db.Exec(stmt, v2Business); err != nil {
			t.Fatalf("v2 cleanup failed: %v", err)
		}
	}

	return NewService(ctx, db), ctx, db
}

func createV2Project(ctx context.Context, t *testing.T, svc *Service, name, visibility string, creator int64) domain.Project {
	t.Helper()
	p, err := svc.CreateProject(ctx, domain.ProjectInput{
		BusinessID:  v2Business,
		OwnerID:     creator,
		Name:        name,
		Description: "v2 test",
		Icon:        "🧪",
		Color:       "#64C5B9",
		Visibility:  visibility,
		CreatedBy:   creator,
	})
	if err != nil {
		t.Fatalf("CreateProject(%s): %v", name, err)
	}
	return p
}

func TestCreateProjectV2VisibilityDefaults(t *testing.T) {
	svc, ctx, _ := v2Svc(t)

	private := createV2Project(ctx, t, svc, "default-private", "", v2UserA)
	if private.Visibility != domain.VisibilityPrivate {
		t.Errorf("empty visibility = %q, want private", private.Visibility)
	}
	public := createV2Project(ctx, t, svc, "explicit-public", domain.VisibilityPublic, v2UserA)
	if public.Visibility != domain.VisibilityPublic || public.IsPersonal {
		t.Errorf("public project visibility=%q personal=%v", public.Visibility, public.IsPersonal)
	}
}

func TestListProjectsV2CurationAndBucketExclusion(t *testing.T) {
	svc, ctx, _ := v2Svc(t)

	mine := createV2Project(ctx, t, svc, "curated-mine", "", v2UserA)

	bucketID, _, err := svc.EnsureMyTasksProject(ctx, v2UserA, v2Business, 0)
	if err != nil {
		t.Fatalf("EnsureMyTasksProject: %v", err)
	}

	got, total, err := svc.ListProjectsV2(ctx, v2UserA, v2Business, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2: %v", err)
	}
	foundMine, foundBucket := false, false
	for _, p := range got {
		switch p.ID {
		case mine.ID:
			foundMine = true
		case bucketID:
			foundBucket = true
		}
	}
	if !foundMine {
		t.Errorf("creator does not see their own project")
	}
	if foundBucket {
		t.Errorf("My Tasks bucket leaked into shared list")
	}
	if total < 1 {
		t.Errorf("total = %d, want >= 1", total)
	}

	_, strangerTotal, err := svc.ListProjectsV2(ctx, v2UserB, v2Business, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 stranger: %v", err)
	}
	if strangerTotal != 0 {
		t.Errorf("stranger total = %d, want 0", strangerTotal)
	}

	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: mine.ID, UserID: v2UserB, Role: domain.FlowRoleMember,
	}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}
	_, memberTotal, err := svc.ListProjectsV2(ctx, v2UserB, v2Business, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 member: %v", err)
	}
	if memberTotal != 1 {
		t.Errorf("member total = %d, want 1", memberTotal)
	}

	_, qTotal, err := svc.ListProjectsV2(ctx, v2UserA, v2Business, "", "", "no-such-name", 50, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 q: %v", err)
	}
	if qTotal != 0 {
		t.Errorf("q-filtered total = %d, want 0", qTotal)
	}
}

func TestArchiveRestorePurgeLifecycle(t *testing.T) {
	svc, ctx, db := v2Svc(t)

	p := createV2Project(ctx, t, svc, "bin-lifecycle", "", v2UserA)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: p.ID, UserID: v2UserB, Role: domain.FlowRoleViewer,
	}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}

	if _, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID: p.ID, BusinessID: v2Business, WorkspaceID: 0,
		Title: "cascade me", ReporterID: v2UserA,
	}); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := svc.DeleteProject(ctx, p.ID, v2Business); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	binned, err := svc.GetProjectV2(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("GetProjectV2 archived: %v", err)
	}
	if binned.ArchivedAt == nil {
		t.Errorf("archived_at not stamped")
	}
	if binned.PurgeAt == nil {
		t.Fatalf("purge_at not stamped")
	}
	wantDeadline := time.Now().UTC().AddDate(0, 0, DefaultBinRetentionDays)
	if diff := binned.PurgeAt.Sub(wantDeadline); diff > 2*time.Minute || diff < -2*time.Minute {
		t.Errorf("purge_at off the retention window by %v", diff)
	}

	activeList, _, err := svc.ListProjectsV2(ctx, v2UserA, v2Business, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("ListProjectsV2 post-archive: %v", err)
	}
	for _, lp := range activeList {
		if lp.ID == p.ID {
			t.Errorf("archived project still listed as active")
		}
	}
	binList, binTotal, err := svc.ListArchivedProjectsV2(ctx, v2UserA, v2Business, 50, 0)
	if err != nil {
		t.Fatalf("ListArchivedProjectsV2: %v", err)
	}
	if binTotal != 1 || len(binList) != 1 || binList[0].ID != p.ID {
		t.Errorf("bin list = (%d rows, total %d), want exactly the archived project", len(binList), binTotal)
	}

	if _, err := svc.RestoreProjectV2(ctx, p.ID, v2UserA, v2Business); err != nil {
		t.Fatalf("RestoreProjectV2: %v", err)
	}
	restored, err := svc.GetProjectV2(ctx, p.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("GetProjectV2 restored: %v", err)
	}
	if restored.ArchivedAt != nil || restored.PurgeAt != nil || restored.Status != "active" {
		t.Errorf("restored project not clean: archived=%v purge=%v status=%q",
			restored.ArchivedAt, restored.PurgeAt, restored.Status)
	}

	if err := svc.DeleteProject(ctx, p.ID, v2Business); err != nil {
		t.Fatalf("re-archive: %v", err)
	}
	if _, err := svc.PurgeProjectV2(ctx, p.ID, v2UserB, v2Business); err == nil {
		t.Errorf("viewer purged a project; gate failed open")
	}
	if _, err := svc.PurgeProjectV2(ctx, p.ID, v2UserA, v2Business); err != nil {
		t.Fatalf("PurgeProjectV2 as creator: %v", err)
	}

	if _, err := svc.GetProjectV2(ctx, p.ID, v2UserA, v2Business); err == nil {
		t.Errorf("project survived purge")
	}
	var count int
	if err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM flow_tasks WHERE project_id = ?", p.ID).Scan(&count); err != nil {
		t.Fatalf("count tasks: %v", err)
	}
	if count != 0 {
		t.Errorf("%d tasks survived the cascade", count)
	}
}
