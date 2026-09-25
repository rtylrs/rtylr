package api_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// taskSnapshot is every field UpdateTask can write, as it reads back on the
// wire. Comparing two snapshots is how these tests assert that a partial update
// left the fields it never mentioned alone.
type taskSnapshot struct {
	Title       string
	Description string
	Status      string
	Priority    string
	DueDate     string
	StartDate   string
	Labels      []int64
}

func (f *roleFixture) readSnapshot(t *testing.T, taskID int64) taskSnapshot {
	t.Helper()

	res, err := f.taskAPI.GetTaskDetail(ctxAs(roleMemberID), &flowtaskv1.GetTaskDetailRequest{Id: taskID})
	if err != nil {
		t.Fatalf("GetTaskDetail: %v", err)
	}
	task := res.GetTask()
	if task == nil {
		t.Fatalf("GetTaskDetail returned no task")
	}
	return taskSnapshot{
		Title:       task.GetTitle(),
		Description: task.GetDescription(),
		Status:      task.GetStatus(),
		Priority:    task.GetPriority(),
		DueDate:     task.GetDueDate(),
		StartDate:   task.GetStartDate(),
		Labels:      task.GetLabels(),
	}
}

// newPopulatedTask seeds a task with a non-zero value in every column a partial
// update could clobber, so a regression shows up as a field reverting to a
// placeholder or an empty list rather than as a no-op.
func (f *roleFixture) newPopulatedTask(t *testing.T) int64 {
	t.Helper()

	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	taskID, err := f.repo.CreateTask(context.Background(), domain.TaskInput{
		ProjectID: f.projID, WorkspaceID: f.wsID, OwnerID: roleOwnerID,
		BusinessID:  testutil.BusinessID,
		Title:       "Ship Q4 billing migration",
		Description: "Real description that must survive a partial update",
		Priority:    "high",
		ReporterID:  roleOwnerID,
		DueDate:     &due,
		StartDate:   &start,
		Labels:      []int64{11, 22},
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	t.Cleanup(func() { _ = f.repo.DeleteTask(context.Background(), taskID, testutil.BusinessID) })
	return taskID
}

// assertOnly compares two snapshots and fails for every field that changed
// except the ones named in changed.
func assertOnly(t *testing.T, before, after taskSnapshot, changed ...string) {
	t.Helper()

	check := func(field, was, now string) {
		if slices.Contains(changed, field) {
			return
		}
		if was != now {
			t.Errorf("%s was overwritten by a partial update that never mentioned it: %q -> %q", field, was, now)
		}
	}
	check("title", before.Title, after.Title)
	check("description", before.Description, after.Description)
	check("status", before.Status, after.Status)
	check("priority", before.Priority, after.Priority)
	check("due_date", before.DueDate, after.DueDate)
	check("start_date", before.StartDate, after.StartDate)

	if !slices.Contains(changed, "labels") && !slices.Equal(before.Labels, after.Labels) {
		t.Errorf("labels were overwritten by a partial update that never mentioned them: %v -> %v", before.Labels, after.Labels)
	}
}

// One partial update per field: set exactly that field, then assert every other
// field held its prior value. This is the coverage that would have caught the
// placeholder clobber, where a priority-only edit rewrote title and description
// to "-", status to "todo", and labels to [0].
func TestTaskPartialUpdate_EachFieldLeavesTheOthersAlone(t *testing.T) {
	cases := []struct {
		name    string
		body    *flowtaskv1.UpdateTaskBody
		changed string
	}{
		{"title", &flowtaskv1.UpdateTaskBody{Title: "Retitled only"}, "title"},
		{"description", &flowtaskv1.UpdateTaskBody{Description: "Redescribed only"}, "description"},
		{"status", &flowtaskv1.UpdateTaskBody{Status: "in_progress"}, "status"},
		{"priority", &flowtaskv1.UpdateTaskBody{Priority: "urgent"}, "priority"},
		{"due_date", &flowtaskv1.UpdateTaskBody{DueDate: "2026-12-25"}, "due_date"},
		{"start_date", &flowtaskv1.UpdateTaskBody{StartDate: "2026-11-01"}, "start_date"},
		{"labels", &flowtaskv1.UpdateTaskBody{Labels: []int64{33}}, "labels"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newRoleFixture(t)
			taskID := f.newPopulatedTask(t)

			before := f.readSnapshot(t, taskID)

			if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
				Id:   taskID,
				Body: tc.body,
			}); err != nil {
				t.Fatalf("UpdateTask setting %s: %v", tc.name, err)
			}

			after := f.readSnapshot(t, taskID)
			assertOnly(t, before, after, tc.changed)
		})
	}
}

// The specific regression this change fixes: labels were rewritten on every
// update regardless of the request, because marshalLabels turns a nil slice
// into "[]" and the guard tested those bytes instead of the input slice.
func TestTaskPartialUpdate_OmittingLabelsKeepsThem(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newPopulatedTask(t)

	before := f.readSnapshot(t, taskID)
	if len(before.Labels) == 0 {
		t.Fatalf("fixture seeded no labels, the assertion below would pass vacuously")
	}

	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Priority: "urgent"},
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	after := f.readSnapshot(t, taskID)
	if !slices.Equal(before.Labels, after.Labels) {
		t.Errorf("expected labels %v to survive a priority-only update, got %v", before.Labels, after.Labels)
	}
}

// The other half of the nil vs empty distinction: an explicit empty slice is a
// real value meaning "clear every label", so it must still be written. A guard
// that skipped anything falsy would turn this into a silent no-op.
func TestTaskPartialUpdate_ExplicitEmptyLabelsClearsThem(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newPopulatedTask(t)

	if _, err := f.taskAPI.UpdateTask(ctxAs(roleMemberID), &flowtaskv1.UpdateTaskRequest{
		Id:   taskID,
		Body: &flowtaskv1.UpdateTaskBody{Labels: []int64{}},
	}); err != nil {
		t.Fatalf("UpdateTask clearing labels: %v", err)
	}

	if after := f.readSnapshot(t, taskID); len(after.Labels) != 0 {
		t.Errorf("expected an explicit empty labels list to clear every label, got %v", after.Labels)
	}
}
