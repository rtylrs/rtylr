package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

// CreateSubtask reaches the same CreateTask path as CreateTask/UpdateTask, so
// assigning the subtask to another user takes the same flow:tasks:assign gate.
// The v2 handler only checked flow:tasks:create, so a member with create/edit
// but not assign could hand work to a colleague and grant them project
// membership. The v1 handler carried the same gap.

func TestTaskSubtaskV2_MemberNeedsAssignmentPermissionForAnotherUser(t *testing.T) {
	f := newRoleFixture(t)
	taskV2 := api.NewTaskAPIV2(f.svc())
	other := roleOwnerID

	_, err := taskV2.CreateSubtask(ctxAsWithoutTaskAssignment(roleMemberID), &flowtaskv2.CreateSubtaskRequest{
		ParentTaskId: testutil.Int64Ptr(f.newTask(t)),
		Title:        "assigned to someone else",
		AssigneeId:   &other,
	})

	assertPermissionDenied(t, err, "member CreateSubtask assignment to another user via v2")
}

func TestTaskSubtaskV1_MemberNeedsAssignmentPermissionForAnotherUser(t *testing.T) {
	f := newRoleFixture(t)
	other := roleOwnerID

	_, err := f.taskAPI.CreateSubtask(ctxAsWithoutTaskAssignment(roleMemberID), &flowtaskv1.CreateSubtaskRequest{
		ParentTaskId: testutil.Int64Ptr(f.newTask(t)),
		ProjectId:    testutil.Int64Ptr(f.projID),
		WorkspaceId:  testutil.Int64Ptr(f.wsID),
		Title:        "assigned to someone else",
		AssigneeId:   &other,
	})

	assertPermissionDenied(t, err, "member CreateSubtask assignment to another user via v1")
}

// Self-assignment stays allowed for the same caller, mirroring the CreateTask
// and UpdateTask policy: the gate exempts the caller's own user id.
func TestTaskSubtaskV2_MemberCanSelfAssignWithoutAssignmentPermission(t *testing.T) {
	f := newRoleFixture(t)
	taskV2 := api.NewTaskAPIV2(f.svc())
	self := roleMemberID

	response, err := taskV2.CreateSubtask(ctxAsWithoutTaskAssignment(roleMemberID), &flowtaskv2.CreateSubtaskRequest{
		ParentTaskId: testutil.Int64Ptr(f.newTask(t)),
		Title:        "self assigned subtask",
		AssigneeId:   &self,
	})
	if err != nil {
		t.Fatalf("member CreateSubtask self assignment via v2: expected success, got %v", err)
	}
	t.Cleanup(func() {
		if deleteErr := f.repo.DeleteTask(context.Background(), response.TaskId, testutil.BusinessID); deleteErr != nil {
			t.Errorf("delete self assigned subtask: %v", deleteErr)
		}
	})
}

// Holding flow:tasks:assign still lets a member hand the subtask to a colleague.
func TestTaskSubtaskV2_AssignPermissionAllowsAnotherUser(t *testing.T) {
	f := newRoleFixture(t)
	taskV2 := api.NewTaskAPIV2(f.svc())
	other := roleOwnerID

	response, err := taskV2.CreateSubtask(ctxAs(roleMemberID), &flowtaskv2.CreateSubtaskRequest{
		ParentTaskId: testutil.Int64Ptr(f.newTask(t)),
		Title:        "assigned with permission",
		AssigneeId:   &other,
	})
	if err != nil {
		t.Fatalf("assignee CreateSubtask via v2: expected success, got %v", err)
	}
	t.Cleanup(func() {
		if deleteErr := f.repo.DeleteTask(context.Background(), response.TaskId, testutil.BusinessID); deleteErr != nil {
			t.Errorf("delete assigned subtask: %v", deleteErr)
		}
	})
}
