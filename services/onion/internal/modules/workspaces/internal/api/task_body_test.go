package api_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// statusOf pulls the HTTP status out of a handler error. Asserting on the code
// rather than the message keeps the test pinned to the contract, so rewording
// the message cannot turn the assertion into a tautology.
func statusOf(t *testing.T, err error, what string) int {
	t.Helper()

	var apiErr *onekithttp.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("%s: expected a *onekithttp.Error, got %T: %v", what, err, err)
	}
	return apiErr.HTTPStatusCode()
}

// newLabelledTask seeds a task carrying labels. Labels are the field a rejected
// update can actually destroy: see the clobber note on the second test.
func (f *roleFixture) newLabelledTask(t *testing.T, labels []int64) int64 {
	t.Helper()

	taskID, err := f.repo.CreateTask(context.Background(), domain.TaskInput{
		ProjectID: f.projID, WorkspaceID: f.wsID, OwnerID: roleOwnerID,
		BusinessID: testutil.BusinessID, Title: "Nil Body Guard Task",
		ReporterID: roleOwnerID, Labels: labels,
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = f.repo.DeleteTask(context.Background(), taskID, testutil.BusinessID) })
	return taskID
}

// A request whose JSON never included the "body" key decodes into a nil Body.
// UpdateTask reads AssigneeId and Archived as raw pointer fields rather than
// through the nil-safe GetX() accessors. That is deliberate, so an absent field
// stays nil for COALESCE, but it means a nil body dereferences and panics. The
// recoverer turns that into a 500, but the request is malformed, not a server
// failure, so the handler owes the caller a 400.
func TestTaskBody_UpdateTaskWithNilBodyReturns400(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)

	res, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{Id: taskID})
	if err == nil {
		t.Fatalf("UpdateTask with a nil body: expected an error, got success (%v)", res)
	}
	if got := statusOf(t, err, "UpdateTask with a nil body"); got != http.StatusBadRequest {
		t.Errorf("UpdateTask with a nil body: expected HTTP %d, got %d (%v)", http.StatusBadRequest, got, err)
	}
}

// Rejecting is not enough: the guard has to fire before the update runs. The
// tempting wrong fix is substituting an empty &UpdateTaskBody{} for the nil
// one, which reports success for a malformed request. Most fields survive that,
// because the repository maps an empty string to a NULL param and COALESCE
// keeps the stored value; labels do not, because marshalLabels turns a nil
// slice into a literal "[]" that overwrites whatever the task had.
func TestTaskBody_UpdateTaskWithNilBodyLeavesTaskUntouched(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newLabelledTask(t, []int64{101, 102})

	before, err := f.taskAPI.GetTaskDetail(ctxAs(roleMemberID), &flowtaskv1.GetTaskDetailRequest{Id: taskID})
	if err != nil {
		t.Fatalf("GetTaskDetail before: %v", err)
	}
	title, labels := before.GetTask().GetTitle(), before.GetTask().GetLabels()
	if title == "" || len(labels) == 0 {
		t.Fatalf("seeded task must have a title and labels to detect a clobber, got title=%q labels=%v", title, labels)
	}

	if _, updateErr := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id: taskID,
	}); updateErr == nil {
		t.Fatal("UpdateTask with a nil body: expected an error, got success")
	}

	after, err := f.taskAPI.GetTaskDetail(ctxAs(roleMemberID), &flowtaskv1.GetTaskDetailRequest{Id: taskID})
	if err != nil {
		t.Fatalf("GetTaskDetail after: %v", err)
	}
	if got := after.GetTask().GetTitle(); got != title {
		t.Errorf("expected the rejected update to leave the title as %q, got %q", title, got)
	}
	if got := after.GetTask().GetLabels(); len(got) != len(labels) {
		t.Errorf("expected the rejected update to leave %d labels, got %d (%v)", len(labels), len(got), got)
	}
}
