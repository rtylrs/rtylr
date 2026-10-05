package domain_test

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_onboarding/internal/domain"
)

func TestIsValidTriggerType_AllValid(t *testing.T) {
	for _, tt := range []string{
		domain.TriggerTypeHire,
		domain.TriggerTypeTerminate,
		domain.TriggerTypeTermination,
		domain.TriggerTypeTransfer,
		domain.TriggerTypePromotion,
		domain.TriggerTypeOnboarding,
		domain.TriggerTypeOffboarding,
	} {
		if !domain.IsValidTriggerType(tt) {
			t.Errorf("expected %q to be valid", tt)
		}
	}
}

func TestIsValidTriggerType_Invalid(t *testing.T) {
	if domain.IsValidTriggerType("offboard") {
		t.Error("expected 'offboard' to be invalid")
	}
}

func TestIsValidChecklistStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.ChecklistStatusPending,
		domain.ChecklistStatusInProgress,
		domain.ChecklistStatusCompleted,
		domain.ChecklistStatusCanceled,
	} {
		if !domain.IsValidChecklistStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidChecklistStatus_Invalid(t *testing.T) {
	if domain.IsValidChecklistStatus("archived") {
		t.Error("expected 'archived' to be invalid")
	}
}

func TestIsValidItemStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.ItemStatusPending,
		domain.ItemStatusCompleted,
		domain.ItemStatusSkipped,
	} {
		if !domain.IsValidItemStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidItemStatus_Invalid(t *testing.T) {
	if domain.IsValidItemStatus("blocked") {
		t.Error("expected 'blocked' to be invalid")
	}
}

func TestIsValidAssigneeRole_AllValid(t *testing.T) {
	for _, r := range []string{
		domain.AssigneeRoleHR,
		domain.AssigneeRoleManager,
		domain.AssigneeRoleEmployee,
		domain.AssigneeRoleIT,
		domain.AssigneeRoleFinance,
	} {
		if !domain.IsValidAssigneeRole(r) {
			t.Errorf("expected %q to be valid", r)
		}
	}
}

func TestIsValidAssigneeRole_Invalid(t *testing.T) {
	if domain.IsValidAssigneeRole("legal") {
		t.Error("expected 'legal' to be invalid")
	}
}

func TestValidateCreateTemplateParams_Valid(t *testing.T) {
	p := domain.CreateTemplateParams{
		Name:        "New Hire Checklist",
		TriggerType: domain.TriggerTypeHire,
	}
	if err := domain.ValidateCreateTemplateParams(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCreateTemplateParams_EmptyName(t *testing.T) {
	p := domain.CreateTemplateParams{Name: ""}
	if err := domain.ValidateCreateTemplateParams(p); !errors.Is(err, domain.ErrEmptyTemplateName) {
		t.Errorf("expected ErrEmptyTemplateName, got %v", err)
	}
}

func TestValidateCreateTemplateParams_WhitespaceName(t *testing.T) {
	p := domain.CreateTemplateParams{Name: "   "}
	if err := domain.ValidateCreateTemplateParams(p); !errors.Is(err, domain.ErrEmptyTemplateName) {
		t.Errorf("expected ErrEmptyTemplateName for whitespace, got %v", err)
	}
}

func TestValidateCreateTemplateParams_InvalidTriggerType(t *testing.T) {
	p := domain.CreateTemplateParams{Name: "Test", TriggerType: "unknown"}
	if err := domain.ValidateCreateTemplateParams(p); !errors.Is(err, domain.ErrInvalidTriggerType) {
		t.Errorf("expected ErrInvalidTriggerType, got %v", err)
	}
}

func TestValidateCreateTemplateParams_EmptyTriggerAllowed(t *testing.T) {
	p := domain.CreateTemplateParams{Name: "Test", TriggerType: ""}
	if err := domain.ValidateCreateTemplateParams(p); err != nil {
		t.Errorf("empty trigger type should be allowed, got %v", err)
	}
}

func TestValidateCreateTemplateItemParams_Valid(t *testing.T) {
	p := domain.CreateTemplateItemParams{
		Title:        "Sign NDA",
		AssigneeRole: domain.AssigneeRoleHR,
	}
	if err := domain.ValidateCreateTemplateItemParams(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateCreateTemplateItemParams_EmptyTitle(t *testing.T) {
	p := domain.CreateTemplateItemParams{Title: ""}
	if err := domain.ValidateCreateTemplateItemParams(p); !errors.Is(err, domain.ErrEmptyItemTitle) {
		t.Errorf("expected ErrEmptyItemTitle, got %v", err)
	}
}

func TestValidateCreateTemplateItemParams_InvalidAssigneeRole(t *testing.T) {
	p := domain.CreateTemplateItemParams{Title: "Task", AssigneeRole: "cfo"}
	if err := domain.ValidateCreateTemplateItemParams(p); !errors.Is(err, domain.ErrInvalidAssigneeRole) {
		t.Errorf("expected ErrInvalidAssigneeRole, got %v", err)
	}
}

func TestValidateAttachmentURL(t *testing.T) {
	if err := domain.ValidateAttachmentURL("https://cdn.example.test/proof.pdf"); err != nil {
		t.Fatalf("valid attachment URL rejected: %v", err)
	}
	for _, raw := range []string{"", "javascript:alert(1)", "/relative/path"} {
		if err := domain.ValidateAttachmentURL(raw); !errors.Is(err, domain.ErrInvalidAttachmentURL) {
			t.Errorf("attachment URL %q error = %v, want ErrInvalidAttachmentURL", raw, err)
		}
	}
}
