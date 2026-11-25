package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestDuplicateTaskCopiesWorkNotHistory(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "duplicate", domain.VisibilityPrivate, v2UserA)
	field, err := svc.CreateProjectField(ctx, v2UserA, v2Business, domain.CustomFieldInput{ProjectID: p.ID, Name: "Area", Type: domain.FieldTypeText})
	if err != nil {
		t.Fatalf("CreateProjectField: %v", err)
	}
	milestone, err := svc.CreateMilestoneV2(ctx, v2Business, p.ID, "Phase one", "", "2026-12-01")
	if err != nil {
		t.Fatalf("CreateMilestoneV2: %v", err)
	}
	effort := int32(60)
	assignee := v2UserA
	src, err := svc.CreateTask(ctx, domain.TaskInput{
		ProjectID: p.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA,
		Title: "IH screen", Description: "## WHAT AND WHY", Priority: "high", AssigneeID: &assignee,
		EffortMinutes: &effort, TaskType: domain.TaskTypeFeature,
		FieldValues: []domain.FieldValueInput{{FieldID: field.ID, Value: domain.FieldValue{Value: "Payments"}}},
	})
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := svc.SetTaskMilestone(ctx, src.ID, &milestone.ID); err != nil {
		t.Fatalf("SetTaskMilestone: %v", err)
	}
	for _, title := range []string{"Design", "Build"} {
		parent := src.ID
		if _, err := svc.CreateTask(ctx, domain.TaskInput{ProjectID: p.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA, ParentTaskID: &parent, Title: title}); err != nil {
			t.Fatalf("CreateTask(subtask): %v", err)
		}
	}
	if err := svc.UpdateTaskStatus(ctx, domain.UpdateTaskStatusInput{ID: src.ID, OwnerID: v2UserA, BusinessID: v2Business, Status: domain.TaskStatusDone}, v2UserA); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}

	copied, err := svc.DuplicateTask(ctx, v2Business, v2UserA, src.ID, "IH screen (copy)", true)
	if err != nil {
		t.Fatalf("DuplicateTask: %v", err)
	}
	if copied.ID == src.ID || copied.Title != "IH screen (copy)" || copied.Description != "## WHAT AND WHY" ||
		copied.Priority != "high" || copied.TaskType != domain.TaskTypeFeature {
		t.Fatalf("copy = %+v", copied)
	}
	if copied.Status == domain.TaskStatusDone {
		t.Fatal("the copy should start as new work, not done")
	}
	if copied.AssigneeID == nil || *copied.AssigneeID != v2UserA || copied.EffortMinutes == nil || *copied.EffortMinutes != 60 {
		t.Fatalf("assignee %v effort %v", copied.AssigneeID, copied.EffortMinutes)
	}
	if copied.MilestoneID == nil || *copied.MilestoneID != milestone.ID {
		t.Fatalf("milestone = %v, want %d", copied.MilestoneID, milestone.ID)
	}
	detail, err := svc.GetTaskDetail(ctx, copied.ID)
	if err != nil || len(detail.Subtasks) != 2 {
		t.Fatalf("copy subtasks = %d, %v", len(detail.Subtasks), err)
	}
	withFields := []domain.Task{detail.Task}
	if err := svc.attachFieldValues(ctx, v2Business, withFields); err != nil || withFields[0].FieldValues[field.ID].Value != "Payments" {
		t.Fatalf("copied field values = %+v, %v", withFields[0].FieldValues, err)
	}
	groups, err := svc.ListTaskRelations(ctx, copied.ID, v2UserA, v2Business)
	if err != nil || len(groups) != 1 || groups[0].Type != domain.RelationClones || len(groups[0].Relations) != 1 {
		t.Fatalf("clone link = %+v, %v", groups, err)
	}

	unassigned, err := svc.DuplicateTask(ctx, v2Business, v2UserA, src.ID, "", false)
	if err != nil || unassigned.AssigneeID != nil || unassigned.Title != "IH screen" {
		t.Fatalf("copy without assign permission: assignee %v title %q, %v", unassigned.AssigneeID, unassigned.Title, err)
	}
	if _, err := svc.DuplicateTask(ctx, v2Business+1, v2UserA, src.ID, "", true); err == nil {
		t.Fatal("another business must not duplicate the task")
	}
}

func TestImportTasksCreatesRowsWithSubtasks(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "import", domain.VisibilityPrivate, v2UserA)
	base := domain.TaskInput{ProjectID: p.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA}
	first, second := base, base
	first.Title, first.Priority = "Login page", "high"
	second.Title = "Logout"
	ids, err := svc.ImportTasks(ctx, []ImportTaskRow{
		{Input: first, Subtasks: []string{"Form", "Errors"}},
		{Input: second},
	})
	if err != nil || len(ids) != 2 {
		t.Fatalf("ImportTasks = %v, %v", ids, err)
	}
	detail, err := svc.GetTaskDetail(ctx, ids[0])
	if err != nil || detail.Task.Priority != "high" || len(detail.Subtasks) != 2 {
		t.Fatalf("first row: priority %q, %d subtasks, %v", detail.Task.Priority, len(detail.Subtasks), err)
	}

	if err := svc.CheckImportProject(ctx, v2Business, p.ID); err != nil {
		t.Fatalf("CheckImportProject with no required fields: %v", err)
	}
	if _, err := svc.CreateProjectField(ctx, v2UserA, v2Business, domain.CustomFieldInput{ProjectID: p.ID, Name: "Client", Type: domain.FieldTypeText, Required: true}); err != nil {
		t.Fatalf("CreateProjectField: %v", err)
	}
	if err := svc.CheckImportProject(ctx, v2Business, p.ID); err == nil {
		t.Fatal("a required field with no default should refuse the import up front")
	}
}
