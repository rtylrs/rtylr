package handlers

import (
	"context"
	"testing"
)

func TestCreateAutomationHandler_Validate_MissingName(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "", TriggerType: "customer_signup"}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestCreateAutomationHandler_Validate_MissingTriggerType(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "Welcome", TriggerType: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing trigger_type")
	}
}

func TestCreateAutomationHandler_Validate_Valid(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "Welcome", TriggerType: "customer_signup", Actions: []byte(`[{"type":"send_email"}]`)}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreateAutomationHandler_Validate_RejectsUnsupportedContract(t *testing.T) {
	h := &CreateAutomationHandler{}
	req := &CreateAutomationRequest{Name: "Welcome", TriggerType: "New customer", Actions: []byte(`[{"type":"send_email"}]`)}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected unsupported trigger error")
	}
}

func TestUpdateAutomationHandler_Validate_MissingName(t *testing.T) {
	h := &UpdateAutomationHandler{}
	req := &UpdateAutomationRequest{Name: ""}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestUpdateAutomationHandler_Validate_Valid(t *testing.T) {
	h := &UpdateAutomationHandler{}
	req := &UpdateAutomationRequest{Name: "Updated Automation", TriggerType: "customer_signup"}
	if err := h.Validate(context.Background(), req); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateAutomationHandler_Validate_RejectsUnsupportedAction(t *testing.T) {
	h := &UpdateAutomationHandler{}
	req := &UpdateAutomationRequest{
		Name:        "Updated Automation",
		TriggerType: "customer_signup",
		Actions:     []byte(`[{"type":"send_push_notification"}]`),
	}
	if err := h.Validate(context.Background(), req); err == nil {
		t.Fatal("expected unsupported action error")
	}
}
