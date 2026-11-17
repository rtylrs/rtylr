package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

type recordedAudits struct{ events []audit.Event }

func (r *recordedAudits) RecordAudit(_ context.Context, event audit.Event) error {
	r.events = append(r.events, event)
	return nil
}

func TestCreateProjectStoresDeveloperMode(t *testing.T) {
	svc, ctx, _ := v2Svc(t)

	created, err := svc.CreateProject(ctx, domain.ProjectInput{
		BusinessID: v2Business, OwnerID: v2UserA, CreatedBy: v2UserA,
		Name: "software", Description: "v2 test", DeveloperMode: true,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	stored, err := svc.GetProjectV2(ctx, created.ID, v2UserA, v2Business)
	if err != nil {
		t.Fatalf("GetProjectV2: %v", err)
	}
	if !created.DeveloperMode || !stored.DeveloperMode {
		t.Errorf("developer_mode created=%v stored=%v, want true", created.DeveloperMode, stored.DeveloperMode)
	}

	plain := createV2Project(ctx, t, svc, "marketing", "", v2UserA)
	if plain.DeveloperMode {
		t.Errorf("a project created without the switch is a developer project")
	}
}

func TestSetProjectDeveloperMode(t *testing.T) {
	svc, ctx, _ := v2Svc(t)
	audits := &recordedAudits{}
	svc.WithAuditRecorder(audits)

	p := createV2Project(ctx, t, svc, "dev-switch", "", v2UserA)
	if err := svc.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: p.ID, UserID: v2UserB, Role: domain.FlowRoleMember,
	}); err != nil {
		t.Fatalf("AddProjectMember: %v", err)
	}

	if _, err := svc.SetProjectDeveloperMode(ctx, p.ID, v2UserB, v2Business, true); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("non-admin switch err = %v, want ErrUnauthorized", err)
	}

	on, err := svc.SetProjectDeveloperMode(ctx, p.ID, v2UserA, v2Business, true)
	if err != nil || !on.DeveloperMode {
		t.Fatalf("switch on: developer_mode=%v err=%v", on.DeveloperMode, err)
	}
	// Switching to the value it already has changes nothing and logs nothing.
	if _, err := svc.SetProjectDeveloperMode(ctx, p.ID, v2UserA, v2Business, true); err != nil {
		t.Fatalf("repeat switch on: %v", err)
	}
	off, err := svc.SetProjectDeveloperMode(ctx, p.ID, v2UserA, v2Business, false)
	if err != nil || off.DeveloperMode {
		t.Fatalf("switch off: developer_mode=%v err=%v", off.DeveloperMode, err)
	}

	if len(audits.events) != 2 {
		t.Fatalf("audit entries = %d, want 2", len(audits.events))
	}
	if audits.events[0].Action != "developer_mode_enabled" || audits.events[1].Action != "developer_mode_disabled" {
		t.Errorf("audit actions = %q, %q", audits.events[0].Action, audits.events[1].Action)
	}
	if audits.events[0].EntityID != p.ID || audits.events[0].BusinessID != v2Business {
		t.Errorf("audit entry targets project %d business %d", audits.events[0].EntityID, audits.events[0].BusinessID)
	}
}

func TestSetProjectDeveloperModeRejectsMyTasks(t *testing.T) {
	svc, ctx, _ := v2Svc(t)

	bucket, err := svc.CreateProject(ctx, domain.ProjectInput{
		BusinessID: v2Business, OwnerID: v2UserA, CreatedBy: v2UserA,
		Name: "My Tasks", Description: "v2 test", IsPersonal: true, DeveloperMode: true,
	})
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if bucket.DeveloperMode {
		t.Errorf("My Tasks was created as a developer project")
	}
	if _, err := svc.SetProjectDeveloperMode(ctx, bucket.ID, v2UserA, v2Business, true); !errors.Is(err, ErrPersonalProjectDeveloperMode) {
		t.Errorf("switching My Tasks err = %v, want ErrPersonalProjectDeveloperMode", err)
	}
}
