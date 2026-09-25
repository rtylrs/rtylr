package api_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func TestTaskV2UpdateRoundTripsTodayFields(t *testing.T) {
	f := newRoleFixture(t)
	taskID := f.newTask(t)
	db := testutil.OpenDB(t)
	if _, err := db.ExecContext(context.Background(), "UPDATE flow_tasks SET assignee_id = ? WHERE id = ?", roleMemberID, taskID); err != nil {
		t.Fatalf("assign task: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM flow_task_focus_pins WHERE task_id = ?", taskID)
	})
	apiV2 := api.NewTaskAPIV2(f.service)
	minutes := int32(45)
	pinned := true

	if _, err := apiV2.UpdateTask(ctxAs(roleMemberID), &flowtaskv2.UpdateTaskRequest{
		Id: taskID,
		Body: &flowtaskv2.UpdateTaskBody{
			EffortMinutes: &minutes,
			FocusPin:      &pinned,
		},
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	response, err := apiV2.GetTaskDetail(ctxAs(roleMemberID), &flowtaskv2.GetTaskDetailRequest{Id: taskID})
	if err != nil {
		t.Fatalf("GetTaskDetail: %v", err)
	}
	if response.Task == nil {
		t.Fatal("GetTaskDetail returned no task")
	}
	if response.Task.EffortMinutes == nil || *response.Task.EffortMinutes != minutes {
		t.Fatalf("effort_minutes = %v, want %d", response.Task.EffortMinutes, minutes)
	}
	if !response.Task.FocusPin {
		t.Fatal("focus_pin = false, want true")
	}

	list, err := apiV2.GetMyTasks(ctxAs(roleMemberID), &flowtaskv2.GetMyTasksRequest{})
	if err != nil {
		t.Fatalf("GetMyTasks: %v", err)
	}
	found := false
	for _, task := range list.Tasks {
		if task != nil && task.Id == taskID {
			found = true
			if !task.FocusPin || task.EffortMinutes == nil || *task.EffortMinutes != minutes {
				t.Fatalf("GetMyTasks projection = %+v, want pinned effort", task)
			}
		}
	}
	if !found {
		t.Fatal("updated task missing from GetMyTasks")
	}
}
