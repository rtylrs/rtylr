package api_test

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/api"
	"github.com/voxire/onion/internal/testutil"
	projectv2 "github.com/voxire/voxie/gen/go/flow/project/v2"
	taskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func TestImportTasksRefusesBadRowsAndWritesNothing(t *testing.T) {
	f := newRoleFixture(t)
	handler := api.NewTaskAPIV2(f.service)
	ctx := ctxAs(roleOwnerID)
	col, err := f.projAPI.CreateProjectColumn(ctx, &projectv2.CreateProjectColumnRequest{ProjectId: testutil.Int64Ptr(f.projID), Name: "Ready", Color: "#6366f1"})
	if err != nil {
		t.Fatal(err)
	}
	due, badDue, badPriority := "2026-12-01", "01/12/2026", "asap"
	other := int64(999999)

	refused, err := handler.ImportTasks(ctx, &taskv2.ImportTasksRequest{ProjectId: f.projID, Rows: []*taskv2.ImportTaskRow{
		{Title: "Fine", DueDate: &due},
		{Title: "  "},
		{Title: "Bad date", DueDate: &badDue},
		{Title: "Bad priority", Priority: &badPriority},
		{Title: "Wrong section", SectionId: &other},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if refused.Created != 0 || len(refused.Errors) != 4 || refused.Errors[0].Row != 2 || refused.Errors[3].Row != 5 {
		t.Fatalf("refused import = %+v", refused)
	}

	effort := int32(90)
	high := "High"
	done, err := handler.ImportTasks(ctx, &taskv2.ImportTasksRequest{ProjectId: f.projID, Rows: []*taskv2.ImportTaskRow{
		{Title: "Imported one", DueDate: &due, Priority: &high, EffortMinutes: &effort, SectionId: &col.Column.Id, Subtasks: []string{"A", " ", "B"}},
		{Title: "Imported two"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if done.Created != 2 || len(done.TaskIds) != 2 || len(done.Errors) != 0 {
		t.Fatalf("import = %+v", done)
	}
	for _, id := range done.TaskIds {
		taskID := id
		t.Cleanup(func() { _ = f.repo.DeleteTask(ctx, taskID, testutil.BusinessID) })
	}
	detail, err := handler.GetTaskDetail(ctx, &taskv2.GetTaskDetailRequest{Id: done.TaskIds[0]})
	if err != nil {
		t.Fatal(err)
	}
	task := detail.Task
	if task.GetPriority() != "high" || task.GetEffortMinutes() != 90 || task.GetSectionId() != col.Column.Id || task.GetDueDate() != "2026-12-01" || len(detail.Subtasks) != 2 {
		t.Fatalf("imported task = %+v with %d subtasks", task, len(detail.Subtasks))
	}

	dup, err := handler.DuplicateTask(ctx, &taskv2.DuplicateTaskRequest{Id: done.TaskIds[0]})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.repo.DeleteTask(ctx, dup.Task.Id, testutil.BusinessID) })
	if dup.Task.GetTitle() != "Imported one" || dup.Task.Id == done.TaskIds[0] || dup.Task.GetPriority() != "high" {
		t.Fatalf("duplicate = %+v", dup.Task)
	}
}
