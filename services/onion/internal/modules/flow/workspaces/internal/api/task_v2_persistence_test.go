package api_test

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
	cyclev2 "github.com/voxire/voxie/gen/go/flow/cycle/v2"
	projectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
	taskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
	viewv2 "github.com/voxire/voxie/gen/go/flow/view/v2"
)

func TestTaskV2CreationPersistsSelectedFields(t *testing.T) {
	f := newRoleFixture(t)
	handler := api.NewTaskAPIV2(f.service)
	ctx := ctxAs(roleOwnerID)
	col, err := f.projAPI.CreateProjectColumn(ctx, &projectv2.CreateProjectColumnRequest{ProjectId: testutil.Int64Ptr(f.projID), Name: "Verification", Color: "#6366f1"})
	if err != nil {
		t.Fatal(err)
	}
	minutes := int32(60)
	hours := float64(1)
	response, err := handler.CreateTask(ctx, &taskv2.CreateTaskRequest{ProjectId: f.projID, SectionId: &col.Column.Id, Title: "Persist selected values", Description: "Details", Status: "done", Priority: "high", EffortMinutes: &minutes, EstimatedHours: &hours, DueDate: "2026-10-05", AssigneeId: testutil.Int64Ptr(roleOwnerID)})
	if err != nil {
		t.Fatal(err)
	}
	id := response.Task.Id
	t.Cleanup(func() { _ = f.repo.DeleteTask(ctx, id, testutil.BusinessID) })
	detail, err := handler.GetTaskDetail(ctx, &taskv2.GetTaskDetailRequest{Id: id})
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range []*taskv2.FlowTask{response.Task, detail.Task} {
		if task.GetStatus() != "done" || task.CompletedAt == nil || task.GetEffortMinutes() != 60 || task.GetEstimatedHours() != 1 || task.GetSectionId() != col.Column.Id || task.GetDueDate() != "2026-10-05" || task.CreatedAt == "" {
			t.Fatalf("creation lost selected fields: %+v", task)
		}
	}
	if _, err := handler.CreateTask(ctx, &taskv2.CreateTaskRequest{Title: "Wrong project", SectionId: &col.Column.Id}); err == nil {
		t.Fatal("accepted a column outside the selected project")
	}
}

func TestTaskV2PartialUpdateCanClearFields(t *testing.T) {
	f := newRoleFixture(t)
	id := f.newPopulatedTask(t)
	handler := api.NewTaskAPIV2(f.service)
	ctx := ctxAs(roleOwnerID)
	waiting, note := "Approval", "Keep the original plan"
	if _, err := handler.UpdateTask(ctx, &taskv2.UpdateTaskRequest{Id: id, Body: &taskv2.UpdateTaskBody{AssigneeId: testutil.Int64Ptr(roleOwnerID), WaitingOn: &waiting, DecisionNote: &note}}); err != nil {
		t.Fatal(err)
	}
	if _, err := handler.UpdateTask(ctx, &taskv2.UpdateTaskRequest{Id: id, Body: &taskv2.UpdateTaskBody{Title: "Only title changed"}}); err != nil {
		t.Fatal(err)
	}
	detail, err := handler.GetTaskDetail(ctx, &taskv2.GetTaskDetailRequest{Id: id})
	if err != nil {
		t.Fatal(err)
	}
	if detail.Task.GetDescription() == "" || detail.Task.DueDate == nil || detail.Task.StartDate == nil || detail.Task.GetWaitingOn() != waiting || detail.Task.GetDecisionNote() != note {
		t.Fatalf("omitted fields changed: %+v", detail.Task)
	}
	empty := ""
	zero := int64(0)
	if _, err := handler.UpdateTask(ctx, &taskv2.UpdateTaskRequest{Id: id, Body: &taskv2.UpdateTaskBody{Description: &empty, DueDate: &empty, StartDate: &empty, AssigneeId: &zero, WaitingOn: &empty, DecisionNote: &empty}}); err != nil {
		t.Fatal(err)
	}
	detail, err = handler.GetTaskDetail(ctx, &taskv2.GetTaskDetailRequest{Id: id})
	if err != nil {
		t.Fatal(err)
	}
	task := detail.Task
	if task.Title != "Only title changed" || task.Description != "" || task.DueDate != nil || task.StartDate != nil || task.AssigneeId != nil || task.WaitingOn != "" || task.DecisionNote != "" {
		t.Fatalf("clear did not persist: %+v", task)
	}
}

func TestViewV2CreatesUpdatesAndDeletesWithoutWorkspace(t *testing.T) {
	f := newRoleFixture(t)
	handler := api.NewViewAPIV2(f.service)
	ctx := ctxAs(roleOwnerID)
	created, err := handler.CreateView(ctx, &viewv2.CreateViewRequest{Name: "Portal verification", Icon: "list", Filters: `{"status":"todo"}`, GroupBy: "status", SortBy: "created_at", SortDir: "asc", Visibility: "private"})
	if err != nil {
		t.Fatal(err)
	}
	id := created.View.Id
	db := testutil.OpenDB(t)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM flow_views WHERE id = ?", id) })
	if _, err := handler.UpdateView(ctx, &viewv2.UpdateViewRequest{Id: id, Name: "Updated view", Icon: "list", Filters: `{"status":"done"}`, GroupBy: "priority", SortBy: "created_at", SortDir: "desc", Visibility: "shared"}); err != nil {
		t.Fatal(err)
	}
	listed, err := handler.ListViews(ctx, &viewv2.ListViewsRequest{Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, v := range listed.Views {
		if v.Id == id {
			found = v.Name == "Updated view" && v.GroupBy == "priority" && v.Visibility == "shared"
		}
	}
	if !found {
		t.Fatal("saved view absent or fields failed to persist")
	}
	if _, err := handler.DeleteView(ctx, &viewv2.DeleteViewRequest{Id: id}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM flow_views WHERE id = ?", id).Scan(&count); err != nil || count != 0 {
		t.Fatalf("view not deleted: %d %v", count, err)
	}
}

func TestTaskV2ReassignmentGrantsAccessAndColumnCanClear(t *testing.T) {
	f := newRoleFixture(t)
	id := f.newTask(t)
	handler := api.NewTaskAPIV2(f.service)
	ctx := ctxAs(roleOwnerID)
	col, err := f.projAPI.CreateProjectColumn(ctx, &projectv2.CreateProjectColumnRequest{ProjectId: testutil.Int64Ptr(f.projID), Name: "Another column", Color: "#6366f1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := handler.UpdateTask(ctx, &taskv2.UpdateTaskRequest{Id: id, Body: &taskv2.UpdateTaskBody{AssigneeId: testutil.Int64Ptr(roleStrangerID), SectionId: &col.Column.Id}}); err != nil {
		t.Fatal(err)
	}
	detail, err := handler.GetTaskDetail(ctxAs(roleStrangerID), &taskv2.GetTaskDetailRequest{Id: id})
	if err != nil {
		t.Fatalf("assignee could not open their task: %v", err)
	}
	if detail.Task.GetSectionId() != col.Column.Id {
		t.Fatal("selected column did not persist")
	}
	zero := int64(0)
	if _, err := handler.UpdateTask(ctx, &taskv2.UpdateTaskRequest{Id: id, Body: &taskv2.UpdateTaskBody{SectionId: &zero}}); err != nil {
		t.Fatal(err)
	}
	detail, err = handler.GetTaskDetail(ctx, &taskv2.GetTaskDetailRequest{Id: id})
	if err != nil {
		t.Fatal(err)
	}
	if detail.Task.SectionId != nil {
		t.Fatal("column clear did not persist")
	}
}

func TestTaskV2ReadsPersistedRelatedCounts(t *testing.T) {
	f := newRoleFixture(t)
	id := f.newTask(t)
	db := testutil.OpenDB(t)
	ctx := ctxAs(roleOwnerID)
	handler := api.NewTaskAPIV2(f.service)
	child, err := db.Exec("INSERT INTO flow_tasks (project_id, workspace_id, owner_id, business_id, parent_task_id, title, status, reporter_id) VALUES (?, 0, ?, ?, ?, 'Completed child', 'done', ?)", f.projID, roleOwnerID, testutil.BusinessID, id, roleOwnerID)
	if err != nil {
		t.Fatal(err)
	}
	childID, _ := child.LastInsertId()
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM flow_tasks WHERE id = ?", childID) })
	if _, err := db.Exec("INSERT INTO flow_task_comments (task_id, user_id, content) VALUES (?, ?, 'Persisted comment')", id, roleOwnerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO flow_task_attachments (task_id, owner_id, business_id, name, url, size, uploaded_by) VALUES (?, ?, ?, 'verification.txt', 'https://example.test/file', 10, ?)", id, roleOwnerID, testutil.BusinessID, roleOwnerID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM flow_task_comments WHERE task_id = ?", id)
		_, _ = db.Exec("DELETE FROM flow_task_attachments WHERE task_id = ?", id)
	})
	detail, err := handler.GetTaskDetail(ctx, &taskv2.GetTaskDetailRequest{Id: id})
	if err != nil {
		t.Fatal(err)
	}
	if detail.Task.SubtaskCount != 1 || detail.Task.CompletedSubtasks != 1 || detail.Task.CommentCount != 1 || detail.Task.AttachmentCount != 1 {
		t.Fatalf("related counts absent: %+v", detail.Task)
	}
	list, err := handler.ListTasks(ctx, &taskv2.ListTasksRequest{ProjectId: &f.projID, Limit: 500})
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range list.Tasks {
		if task.Id == id && task.CommentCount == 1 && task.AttachmentCount == 1 && task.CompletedSubtasks == 1 {
			return
		}
	}
	t.Fatal("list omitted persisted counts")
}

func TestCycleV2ListRetainsCreatedDatesAndProject(t *testing.T) {
	f := newRoleFixture(t)
	ctx := ctxAs(roleOwnerID)
	handler := api.NewCycleAPIV2(f.service)
	created, err := handler.CreateCycle(ctx, &cyclev2.CreateCycleRequest{ProjectId: &f.projID, Name: "Round-trip sprint", StartDate: "2026-12-01", EndDate: "2026-12-15"})
	if err != nil {
		t.Fatal(err)
	}
	db := testutil.OpenDB(t)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM flow_cycles WHERE id = ?", created.Cycle.Id) })
	list, err := handler.ListCycles(ctx, &cyclev2.ListCyclesRequest{ProjectId: &f.projID, Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	for _, cycle := range list.Cycles {
		if cycle.Id == created.Cycle.Id {
			if cycle.ProjectId != f.projID || cycle.StartDate != created.Cycle.StartDate || cycle.EndDate != created.Cycle.EndDate {
				t.Fatalf("cycle lost dates or project: %+v", cycle)
			}
			return
		}
	}
	t.Fatal("created cycle absent after reload")
}
