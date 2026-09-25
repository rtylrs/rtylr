package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/testutil"
	flowlabelv1 "github.com/voxire/voxie/gen/go/flow/label/v1"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// Distinct from the shared fixture tenant so these rows cannot collide with
// data other integration tests create.
const (
	roleOwnerID    int64 = 900001 // owns the workspace and the project
	roleViewerID   int64 = 900002 // flow_project_members.role = viewer
	roleMemberID   int64 = 900003 // flow_project_members.role = member
	roleManagerID  int64 = 900004 // flow_project_members.role = admin
	roleStrangerID int64 = 900005 // no membership anywhere
)

type roleFixture struct {
	service  *application.Service
	taskAPI  *api.TaskAPI
	labelAPI *api.LabelAPI
	projAPI  *api.ProjectAPIV2
	repo     *infrastructure.Repository
	wsID     int64
	projID   int64
}

// newRoleFixture builds a workspace + project + explicit member rows for the
// three project roles, and returns a TaskAPI wired to a real Service.
func newRoleFixture(t *testing.T) *roleFixture {
	t.Helper()

	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := infrastructure.NewRepository(db)
	svc := application.NewService(ctx, db)

	// v2 handlers resolve the tenant through user_business_access; without a row
	// there they answer 403 regardless of project membership.
	testutil.GrantBusinessAccess(t, db, testutil.BusinessID,
		roleOwnerID, roleViewerID, roleMemberID, roleManagerID, roleStrangerID)

	wsID, err := repo.CreateWorkspace(ctx, domain.WorkspaceInput{
		OwnerID: roleOwnerID, BusinessID: testutil.BusinessID,
		Name: "RoleGateWS", CreatedBy: roleOwnerID,
	}, false)
	if err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}
	t.Cleanup(func() { _ = repo.DeleteWorkspace(ctx, wsID, roleOwnerID) })

	projID, err := repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: wsID, OwnerID: roleOwnerID, BusinessID: testutil.BusinessID,
		Name: "Role Gate Project", CreatedBy: roleOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	// Workspace rows as well as project rows. v1's CreateTask resolves the
	// workspace before it checks any role, so without these a viewer is turned
	// away with "Invalid workspace_id" and the role gate is never reached: the
	// test would report a block it never actually exercised.
	for userID, role := range map[int64]string{
		roleViewerID:  domain.FlowRoleViewer,
		roleMemberID:  domain.FlowRoleMember,
		roleManagerID: domain.FlowRoleAdmin,
	} {
		if addErr := repo.AddWorkspaceMember(ctx, domain.AddWorkspaceMemberInput{
			WorkspaceID: wsID, UserID: userID, Role: role, InvitedBy: roleOwnerID,
		}); addErr != nil {
			t.Fatalf("AddWorkspaceMember(%d, %s): %v", userID, role, addErr)
		}
	}

	for userID, role := range map[int64]string{
		roleViewerID:  domain.FlowRoleViewer,
		roleMemberID:  domain.FlowRoleMember,
		roleManagerID: domain.FlowRoleAdmin,
	} {
		if addErr := repo.AddProjectMember(ctx, domain.AddProjectMemberInput{
			ProjectID: projID, UserID: userID, Role: role, AddedBy: roleOwnerID,
		}); addErr != nil {
			t.Fatalf("AddProjectMember(%d, %s): %v", userID, role, addErr)
		}
	}

	return &roleFixture{
		service:  svc,
		taskAPI:  api.NewTaskAPI(svc),
		labelAPI: api.NewLabelAPI(svc),
		projAPI:  api.NewProjectAPIV2(svc),
		repo:     repo,
		wsID:     wsID,
		projID:   projID,
	}
}

// newTask seeds a task the mutation handlers can be pointed at.
func (f *roleFixture) newTask(t *testing.T) int64 {
	t.Helper()

	taskID, err := f.repo.CreateTask(context.Background(), domain.TaskInput{
		ProjectID: f.projID, WorkspaceID: f.wsID, OwnerID: roleOwnerID,
		BusinessID: testutil.BusinessID, Title: "Role Gate Task", ReporterID: roleOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = f.repo.DeleteTask(context.Background(), taskID, testutil.BusinessID) })
	return taskID
}

// ctxAs returns a context carrying an authenticated caller, the same shape the
// auth middleware installs in production.
func ctxAs(userID int64) context.Context {
	return ctxAsWithTaskAssignment(userID, true)
}

func ctxAsWithoutTaskAssignment(userID int64) context.Context {
	return ctxAsWithTaskAssignment(userID, false)
}

func ctxAsWithTaskAssignment(userID int64, canAssignTasks bool) context.Context {
	permissions := []string{
		"flow:projects:view", "flow:projects:create", "flow:projects:edit",
		"flow:projects:assign", "flow:tasks:view", "flow:tasks:create",
		"flow:tasks:edit", "flow:team:manage", "flow:managers:manage",
		"flow:settings:manage",
	}
	if canAssignTasks {
		permissions = append(permissions, "flow:tasks:assign")
	}

	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		UserID:     userID,
		BusinessID: testutil.BusinessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			BusinessID: testutil.BusinessID,
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppFlow,
				Licensed:    true,
				HasAccess:   true,
				Permissions: permissions,
			}},
			IsOwner: userID == roleOwnerID,
			IsAdmin: userID == roleManagerID,
		},
	})
}

func assertPermissionDenied(t *testing.T, err error, what string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected permission denied, got nil (mutation was allowed)", what)
	}
	if want := onekiterrors.PermissionDenied().Error(); err.Error() != want {
		t.Fatalf("%s: expected %q, got %q", what, want, err.Error())
	}
}

func (f *roleFixture) updateTaskAs(userID, taskID int64) error {
	_, err := f.taskAPI.UpdateTask(ctxAs(userID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Title: "Retitled by test"},
	})
	return err
}

func (f *roleFixture) deleteTaskAs(userID, taskID int64) error {
	_, err := f.taskAPI.DeleteTask(ctxAs(userID), &flowtaskv1.DeleteTaskRequest{Id: taskID})
	return err
}

// The reported bug: a project viewer could mutate tasks by calling the API
// directly, because no handler checked role.
func TestTaskRole_ViewerCannotUpdateTask(t *testing.T) {
	f := newRoleFixture(t)
	assertPermissionDenied(t, f.updateTaskAs(roleViewerID, f.newTask(t)), "viewer UpdateTask")
}

func TestTaskRole_ViewerCannotDeleteTask(t *testing.T) {
	f := newRoleFixture(t)
	assertPermissionDenied(t, f.deleteTaskAs(roleViewerID, f.newTask(t)), "viewer DeleteTask")
}

// member sits above the write bar but below the delete bar.
func TestTaskRole_MemberCanUpdateTask(t *testing.T) {
	f := newRoleFixture(t)
	if err := f.updateTaskAs(roleMemberID, f.newTask(t)); err != nil {
		t.Fatalf("member UpdateTask: expected success, got %v", err)
	}
}

func TestTaskRole_MemberCanCreateSelfAssignedTaskWithoutAssignmentPermission(t *testing.T) {
	f := newRoleFixture(t)
	self := roleMemberID
	response, err := f.taskAPI.CreateTask(ctxAsWithoutTaskAssignment(roleMemberID), &flowtaskv1.CreateTaskRequest{
		ProjectId:   testutil.Int64Ptr(f.projID),
		WorkspaceId: testutil.Int64Ptr(f.wsID),
		Title:       "Self assigned task",
		AssigneeId:  &self,
	})
	if err != nil {
		t.Fatalf("member CreateTask self assignment: expected success, got %v", err)
	}
	if response.Task == nil || response.Task.AssigneeId == nil || *response.Task.AssigneeId != roleMemberID {
		t.Fatalf("created task assignee = %v, want %d", response.Task, roleMemberID)
	}
	t.Cleanup(func() {
		if deleteErr := f.repo.DeleteTask(context.Background(), response.Task.Id, testutil.BusinessID); deleteErr != nil {
			t.Errorf("delete self assigned task: %v", deleteErr)
		}
	})
}

func TestTaskRole_MemberCanUpdateToSelfWithoutAssignmentPermission(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)
	self := roleMemberID
	if _, err := f.taskAPI.UpdateTask(ctxAsWithoutTaskAssignment(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id: taskID,
		Body: &flowtaskv1.UpdateTaskBody{
			AssigneeId: &self,
		},
	}); err != nil {
		t.Fatalf("member UpdateTask self assignment: expected success, got %v", err)
	}

	task, err := f.repo.GetTaskByID(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.AssigneeID == nil || *task.AssigneeID != roleMemberID {
		t.Errorf("updated task assignee = %v, want %d", task.AssigneeID, roleMemberID)
	}
}

func TestTaskRole_MemberStillNeedsAssignmentPermissionForAnotherUser(t *testing.T) {
	f := newRoleFixture(t)
	other := roleOwnerID
	_, err := f.taskAPI.UpdateTask(ctxAsWithoutTaskAssignment(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id: f.newTask(t),
		Body: &flowtaskv1.UpdateTaskBody{
			AssigneeId: &other,
		},
	})
	assertPermissionDenied(t, err, "member assignment to another user")
}

func TestTaskRole_MemberCannotDeleteTask(t *testing.T) {
	f := newRoleFixture(t)
	assertPermissionDenied(t, f.deleteTaskAs(roleMemberID, f.newTask(t)), "member DeleteTask")
}

func TestTaskRole_ManagerCanUpdateTask(t *testing.T) {
	f := newRoleFixture(t)
	if err := f.updateTaskAs(roleManagerID, f.newTask(t)); err != nil {
		t.Fatalf("manager UpdateTask: expected success, got %v", err)
	}
}

func TestTaskRole_ManagerCanDeleteTask(t *testing.T) {
	f := newRoleFixture(t)
	if err := f.deleteTaskAs(roleManagerID, f.newTask(t)); err != nil {
		t.Fatalf("manager DeleteTask: expected success, got %v", err)
	}
}

// A caller with no membership of any kind must not reach the task at all,
// even though the underlying UPDATE is scoped only by business_id.
func TestTaskRole_NonMemberCannotUpdateTask(t *testing.T) {
	f := newRoleFixture(t)
	assertPermissionDenied(t, f.updateTaskAs(roleStrangerID, f.newTask(t)), "stranger UpdateTask")
}

// Regression guard for the fallback chain: a project owner holds no
// flow_project_members row, because CreateProject never writes one. Gating on
// that table alone would lock the owner out of their own project.
func TestTaskRole_ProjectOwnerCanDeleteWithoutMemberRow(t *testing.T) {
	f := newRoleFixture(t)
	ctx := context.Background()

	if _, err := f.repo.GetProjectMemberRole(ctx, f.projID, roleOwnerID); err == nil {
		t.Fatal("expected the project owner to have no flow_project_members row")
	}
	if err := f.deleteTaskAs(roleOwnerID, f.newTask(t)); err != nil {
		t.Fatalf("project owner DeleteTask: expected success, got %v", err)
	}
}

// Second leg of the fallback chain: a workspace member with no project row
// inherits their workspace role, so a workspace-level viewer stays read-only.
func TestTaskRole_WorkspaceViewerCannotUpdateTask(t *testing.T) {
	f := newRoleFixture(t)
	ctx := context.Background()

	if err := f.repo.AddWorkspaceMember(ctx, domain.AddWorkspaceMemberInput{
		WorkspaceID: f.wsID, UserID: roleStrangerID,
		Role: domain.FlowRoleViewer, InvitedBy: roleOwnerID,
	}); err != nil {
		t.Fatalf("AddWorkspaceMember: %v", err)
	}

	assertPermissionDenied(t, f.updateTaskAs(roleStrangerID, f.newTask(t)), "workspace viewer UpdateTask")
}

// Every remaining gated handler, exercised as a project viewer. UpdateTask and
// DeleteTask have dedicated tests above; this covers the rest of the surface so
// a newly added handler cannot quietly skip its gate.
func TestTaskRole_ViewerBlockedFromEveryTaskMutation(t *testing.T) {
	f := newRoleFixture(t)
	ctx := context.Background()
	taskID := f.newTask(t)

	commentID, err := f.repo.CreateComment(ctx, taskID, roleOwnerID, "seed comment")
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	attachmentID, err := f.repo.CreateAttachment(ctx, domain.TaskAttachmentInput{
		TaskID: taskID, OwnerID: roleOwnerID, BusinessID: testutil.BusinessID,
		Name: "seed.pdf", URL: "https://example.invalid/seed.pdf", UploadedBy: roleOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateAttachment: %v", err)
	}
	timeLogID, err := f.repo.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: roleOwnerID, BusinessID: testutil.BusinessID, TaskID: taskID,
		UserID: roleOwnerID, DurationMinutes: 30, Description: "seed log",
	})
	if err != nil {
		t.Fatalf("CreateTimeLog: %v", err)
	}

	viewer := ctxAs(roleViewerID)
	cases := []struct {
		name string
		call func() error
	}{
		{"CreateTask", func() error {
			_, e := f.taskAPI.CreateTask(viewer, &flowtaskv1.CreateTaskRequest{
				ProjectId: testutil.Int64Ptr(f.projID), WorkspaceId: testutil.Int64Ptr(f.wsID), Title: "viewer task",
			})
			return e
		}},
		{"UpdateTaskStatus", func() error {
			_, e := f.taskAPI.UpdateTaskStatus(viewer, &flowtaskv1.UpdateTaskStatusRequest{
				Id: taskID, Body: &flowtaskv1.UpdateTaskStatusBody{Status: "done"},
			})
			return e
		}},
		{"CreateSubtask", func() error {
			_, e := f.taskAPI.CreateSubtask(viewer, &flowtaskv1.CreateSubtaskRequest{
				ParentTaskId: testutil.Int64Ptr(taskID), ProjectId: testutil.Int64Ptr(f.projID), WorkspaceId: testutil.Int64Ptr(f.wsID), Title: "viewer subtask",
			})
			return e
		}},
		{"AddComment", func() error {
			_, e := f.taskAPI.AddComment(viewer, &flowtaskv1.AddCommentRequest{
				TaskId: taskID, Content: "viewer comment",
			})
			return e
		}},
		{"DeleteComment", func() error {
			_, e := f.taskAPI.DeleteComment(viewer, &flowtaskv1.DeleteCommentRequest{Id: commentID})
			return e
		}},
		{"CreateAttachment", func() error {
			_, e := f.taskAPI.CreateAttachment(viewer, &flowtaskv1.CreateAttachmentRequest{
				TaskId: taskID, Name: "viewer.pdf", Url: "https://example.invalid/v.pdf",
			})
			return e
		}},
		{"DeleteAttachment", func() error {
			_, e := f.taskAPI.DeleteAttachment(viewer, &flowtaskv1.DeleteAttachmentRequest{Id: attachmentID})
			return e
		}},
		{"CreateTimeLog", func() error {
			_, e := f.taskAPI.CreateTimeLog(viewer, &flowtaskv1.CreateTimeLogRequest{
				TaskId: taskID, DurationMinutes: testutil.Int64Ptr(15),
			})
			return e
		}},
		{"DeleteTimeLog", func() error {
			_, e := f.taskAPI.DeleteTimeLog(viewer, &flowtaskv1.DeleteTimeLogRequest{Id: timeLogID})
			return e
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertPermissionDenied(t, tc.call(), "viewer "+tc.name)
		})
	}
}

// A member holds the write tier, so the non-delete half of that surface must
// still work — a gate that denied everyone would pass the test above.
func TestTaskRole_MemberAllowedOnWriteTierMutations(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)
	member := ctxAs(roleMemberID)

	if _, err := f.taskAPI.AddComment(member, &flowtaskv1.AddCommentRequest{
		TaskId: taskID, Content: "member comment",
	}); err != nil {
		t.Errorf("member AddComment: expected success, got %v", err)
	}
	if _, err := f.taskAPI.CreateTimeLog(member, &flowtaskv1.CreateTimeLogRequest{
		TaskId: taskID, DurationMinutes: testutil.Int64Ptr(15),
	}); err != nil {
		t.Errorf("member CreateTimeLog: expected success, got %v", err)
	}
	if _, err := f.taskAPI.UpdateTaskStatus(member, &flowtaskv1.UpdateTaskStatusRequest{
		Id: taskID, Body: &flowtaskv1.UpdateTaskStatusBody{Status: "done"},
	}); err != nil {
		t.Errorf("member UpdateTaskStatus: expected success, got %v", err)
	}
}

// v1 predates the business-wide label pool and its insert used to rely on
// flow_labels.business_id's default of zero. The row was therefore invisible to
// v2 and could collide in the recurring dedupe migration with labels belonging
// to unrelated businesses. Exercise the registered v1 handler, then read via
// the current v2 service so the regression covers the full API-to-SQL scope.
func TestV1CreateLabelPersistsBusinessScope(t *testing.T) {
	f := newRoleFixture(t)
	created, err := f.labelAPI.CreateLabel(ctxAs(roleOwnerID), &flowlabelv1.CreateLabelRequest{
		WorkspaceId: testutil.Int64Ptr(f.wsID),
		Name:        "Business-scoped regression label",
		Color:       "#123456",
	})
	if err != nil {
		t.Fatalf("v1 CreateLabel: %v", err)
	}
	if created.GetLabel() == nil || created.GetLabel().GetId() <= 0 {
		t.Fatalf("v1 CreateLabel returned no stored label: %+v", created)
	}
	labelID := created.GetLabel().GetId()
	t.Cleanup(func() {
		_ = f.repo.DeleteLabel(context.Background(), labelID, roleOwnerID, testutil.BusinessID)
	})

	labels, _, err := f.service.ListLabelsV2(context.Background(), testutil.BusinessID, 10_000, 0)
	if err != nil {
		t.Fatalf("v2 ListLabels: %v", err)
	}
	for _, label := range labels {
		if label.ID == labelID {
			return
		}
	}
	t.Fatalf("v1 label %d is absent from the business-scoped v2 list", labelID)
}

// SetTaskLabels lives in LabelAPI but mutates a task. Its SQL guard admits any
// flow_workspace_members row regardless of role, so the exploitable case is a
// workspace viewer — a project viewer is already turned away for lack of a
// workspace row, which is membership, not a role check.
// The label set must differ from the task's current one: the UPDATE reports
// zero changed rows when the value is identical, and requireAffected reads that
// as a denial — which would let these tests pass without proving anything.
func (f *roleFixture) setTaskLabelsAs(userID, taskID int64) error {
	_, err := f.labelAPI.SetTaskLabels(ctxAs(userID), &flowlabelv1.SetTaskLabelsRequest{
		TaskId: testutil.Int64Ptr(taskID), LabelIds: []int64{101, 102},
	})
	return err
}

func (f *roleFixture) addWorkspaceMember(t *testing.T, userID int64, role string) {
	t.Helper()

	if err := f.repo.AddWorkspaceMember(context.Background(), domain.AddWorkspaceMemberInput{
		WorkspaceID: f.wsID, UserID: userID, Role: role, InvitedBy: roleOwnerID,
	}); err != nil {
		t.Fatalf("AddWorkspaceMember(%s): %v", role, err)
	}
}

func TestTaskRole_WorkspaceViewerCannotSetTaskLabels(t *testing.T) {
	f := newRoleFixture(t)
	f.addWorkspaceMember(t, roleStrangerID, domain.FlowRoleViewer)
	assertPermissionDenied(t, f.setTaskLabelsAs(roleStrangerID, f.newTask(t)), "workspace viewer SetTaskLabels")
}

func TestTaskRole_WorkspaceMemberCanSetTaskLabels(t *testing.T) {
	f := newRoleFixture(t)
	f.addWorkspaceMember(t, roleStrangerID, domain.FlowRoleMember)
	if err := f.setTaskLabelsAs(roleStrangerID, f.newTask(t)); err != nil {
		t.Fatalf("workspace member SetTaskLabels: expected success, got %v", err)
	}
}

// ...and a workspace member at member rank may write without a project row.
func TestTaskRole_WorkspaceMemberCanUpdateTask(t *testing.T) {
	f := newRoleFixture(t)
	ctx := context.Background()

	if err := f.repo.AddWorkspaceMember(ctx, domain.AddWorkspaceMemberInput{
		WorkspaceID: f.wsID, UserID: roleStrangerID,
		Role: domain.FlowRoleMember, InvitedBy: roleOwnerID,
	}); err != nil {
		t.Fatalf("AddWorkspaceMember: %v", err)
	}

	if err := f.updateTaskAs(roleStrangerID, f.newTask(t)); err != nil {
		t.Fatalf("workspace member UpdateTask: expected success, got %v", err)
	}
}

// svc exposes the fixture's service so a test can build another handler against
// the same seeded workspace, project and role rows.
func (f *roleFixture) svc() *application.Service { return f.service }
