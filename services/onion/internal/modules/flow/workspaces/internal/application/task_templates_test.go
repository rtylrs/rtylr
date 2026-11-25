package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func TestTaskTemplatesPerProjectAndType(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	p := createV2Project(ctx, t, svc, "templates", domain.VisibilityPrivate, v2UserA)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{ProjectID: p.ID, UserID: v2UserB, Role: domain.FlowRoleMember, AddedBy: v2UserA}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}

	saved, err := svc.SetTaskTemplate(ctx, v2Business, p.ID, v2UserA, domain.TaskTypeBug, "## Steps\n1.")
	if err != nil || saved == nil || saved.Body != "## Steps\n1." {
		t.Fatalf("SetTaskTemplate = %+v, %v", saved, err)
	}
	if _, err := svc.SetTaskTemplate(ctx, v2Business, p.ID, v2UserA, domain.TaskTypeBug, "## Steps v2"); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if _, err := svc.SetTaskTemplate(ctx, v2Business, p.ID, v2UserA, domain.TaskTypeSpike, "## Question"); err != nil {
		t.Fatalf("spike template: %v", err)
	}
	if _, err := svc.SetTaskTemplate(ctx, v2Business, p.ID, v2UserA, "story", "x"); !errors.Is(err, domain.ErrInvalidTaskType) {
		t.Fatalf("unknown type: err = %v", err)
	}
	if _, err := svc.SetTaskTemplate(ctx, v2Business, p.ID, v2UserB, domain.TaskTypeBug, "member edit"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("member saving a template: err = %v", err)
	}

	// A member reads them for the New task form; one per type.
	list, err := svc.ListTaskTemplates(ctx, v2Business, p.ID, v2UserB)
	if err != nil || len(list) != 2 || list[0].TaskType != domain.TaskTypeBug || list[0].Body != "## Steps v2" {
		t.Fatalf("ListTaskTemplates = %+v, %v", list, err)
	}
	if other, err := svc.ListTaskTemplates(ctx, v2Business+1, p.ID, v2UserA); err == nil && len(other) > 0 {
		t.Fatalf("another business read %d templates", len(other))
	}

	if removed, err := svc.SetTaskTemplate(ctx, v2Business, p.ID, v2UserA, domain.TaskTypeBug, "  "); err != nil || removed != nil {
		t.Fatalf("blank body: %+v, %v", removed, err)
	}
	if list, _ := svc.ListTaskTemplates(ctx, v2Business, p.ID, v2UserA); len(list) != 1 || list[0].TaskType != domain.TaskTypeSpike {
		t.Fatalf("after removing the bug template: %+v", list)
	}
}

func TestChoreAndSpikeAreTaskTypes(t *testing.T) {
	for _, taskType := range []string{domain.TaskTypeChore, domain.TaskTypeSpike} {
		if !domain.IsValidTaskType(taskType) {
			t.Errorf("%s should be a valid task type", taskType)
		}
	}
	if domain.IsValidTaskType("story") {
		t.Error("story is not a task type")
	}
}
