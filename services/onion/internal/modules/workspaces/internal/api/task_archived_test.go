package api_test

import (
	"testing"

	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// readArchivedAs fetches the task through the same handler the client calls and
// returns the archived flag as it appears on the wire payload.
func (f *roleFixture) readArchivedAs(t *testing.T, userID, taskID int64) bool {
	t.Helper()

	res, err := f.taskAPI.GetTaskDetail(ctxAs(userID), &flowtaskv1.GetTaskDetailRequest{Id: taskID})
	if err != nil {
		t.Fatalf("GetTaskDetail: %v", err)
	}
	if res.GetTask() == nil {
		t.Fatalf("GetTaskDetail returned no task")
	}
	return res.GetTask().GetArchived()
}

// The full API round trip the archive UI depends on: UpdateTask carries the
// flag in, GetTaskDetail carries it back out.
func TestTaskArchived_UpdateTaskArchivesAndDetailReturnsIt(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)

	if f.readArchivedAs(t, roleMemberID, taskID) {
		t.Fatalf("expected a new task to read back as archived=false")
	}

	archived := true
	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Archived: &archived},
	}); err != nil {
		t.Fatalf("UpdateTask archiving task: %v", err)
	}

	if !f.readArchivedAs(t, roleMemberID, taskID) {
		t.Errorf("expected archived=true on the GetTaskDetail payload after archiving")
	}
}

// An update that omits archived must not silently un-archive the task. This is
// the case that breaks if the handler reads GetArchived() instead of the
// pointer, since the accessor reports an absent field as false.
func TestTaskArchived_UpdateOmittingArchivedKeepsItArchived(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)

	archived := true
	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Archived: &archived},
	}); err != nil {
		t.Fatalf("UpdateTask archiving task: %v", err)
	}

	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Title: "Retitled, archived untouched"},
	}); err != nil {
		t.Fatalf("UpdateTask omitting archived: %v", err)
	}

	if !f.readArchivedAs(t, roleMemberID, taskID) {
		t.Errorf("expected archived=true to survive an update that omitted archived")
	}
}

// Restoring must still be possible: an explicit false is a real value, not an
// absent one, so COALESCE must receive it.
func TestTaskArchived_UpdateWithFalseRestoresTask(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)

	archived := true
	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Archived: &archived},
	}); err != nil {
		t.Fatalf("UpdateTask archiving task: %v", err)
	}

	restored := false
	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Archived: &restored},
	}); err != nil {
		t.Fatalf("UpdateTask restoring task: %v", err)
	}

	if f.readArchivedAs(t, roleMemberID, taskID) {
		t.Errorf("expected archived=false after an explicit restore")
	}
}
