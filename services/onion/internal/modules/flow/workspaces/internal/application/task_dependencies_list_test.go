package application

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestAttachTaskDependenciesForTimelineArrows(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "timeline", domain.VisibilityPrivate, v2UserA)
	newTask := func(title string) int64 {
		task, err := svc.CreateTask(ctx, domain.TaskInput{ProjectID: p.ID, BusinessID: v2Business, OwnerID: v2UserA, ReporterID: v2UserA, Title: title})
		if err != nil {
			t.Fatalf("CreateTask(%s): %v", title, err)
		}
		return task.ID
	}
	design, build, related := newTask("Design"), newTask("Build"), newTask("Notes")
	if _, err := svc.AddTaskRelation(ctx, design, build, v2UserA, v2Business, domain.RelationBlocks, domain.RelationOutward); err != nil {
		t.Fatalf("design blocks build: %v", err)
	}
	if _, err := svc.AddTaskRelation(ctx, related, build, v2UserA, v2Business, domain.RelationRelates, domain.RelationOutward); err != nil {
		t.Fatalf("relates: %v", err)
	}

	tasks := []domain.Task{{ID: design}, {ID: build}, {ID: related}}
	if err := svc.AttachTaskDependencies(ctx, v2Business, tasks); err != nil {
		t.Fatalf("AttachTaskDependencies: %v", err)
	}
	if len(tasks[0].Dependencies) != 0 || len(tasks[2].Dependencies) != 0 {
		t.Fatalf("only Build depends on anything: %v %v", tasks[0].Dependencies, tasks[2].Dependencies)
	}
	if len(tasks[1].Dependencies) != 1 || tasks[1].Dependencies[0] != design {
		t.Fatalf("Build depends on %v, want only Design (a relates link is not a dependency)", tasks[1].Dependencies)
	}
	other := []domain.Task{{ID: build}}
	if err := svc.AttachTaskDependencies(ctx, v2Business+1, other); err != nil || len(other[0].Dependencies) != 0 {
		t.Fatalf("another business saw %v, %v", other[0].Dependencies, err)
	}
}
