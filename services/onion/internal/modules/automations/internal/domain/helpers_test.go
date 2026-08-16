package domain

import (
	"encoding/json"
	"testing"
)

func TestIsEnabled_WhenActive(t *testing.T) {
	a := Automation{IsActive: true}
	if !a.IsEnabled() {
		t.Error("expected IsEnabled=true when IsActive=true")
	}
}

func TestIsEnabled_WhenInactive(t *testing.T) {
	a := Automation{IsActive: false}
	if a.IsEnabled() {
		t.Error("expected IsEnabled=false when IsActive=false")
	}
}

func TestHasTrigger_WhenSet(t *testing.T) {
	a := Automation{TriggerType: TriggerOrderCreated}
	if !a.HasTrigger() {
		t.Error("expected HasTrigger=true when TriggerType is set")
	}
}

func TestHasTrigger_WhenEmpty(t *testing.T) {
	a := Automation{TriggerType: ""}
	if a.HasTrigger() {
		t.Error("expected HasTrigger=false when TriggerType is empty")
	}
}

func TestValidateAutomationInput_EmptyName(t *testing.T) {
	msg := ValidateAutomationInput(AutomationInput{Name: "", TriggerType: TriggerOrderCreated})
	if msg == "" {
		t.Error("expected validation error for empty name")
	}
}

func TestValidateAutomationInput_EmptyTriggerType(t *testing.T) {
	msg := ValidateAutomationInput(AutomationInput{Name: "Welcome", TriggerType: ""})
	if msg == "" {
		t.Error("expected validation error for empty trigger type")
	}
}

func TestValidateAutomationInput_Valid(t *testing.T) {
	msg := ValidateAutomationInput(AutomationInput{
		Name:        "Welcome Series",
		TriggerType: TriggerCustomerSignup,
	})
	if msg != "" {
		t.Errorf("expected no validation error, got %q", msg)
	}
}

func TestTriggerConstants_AreUnique(t *testing.T) {
	triggers := []string{
		TriggerOrderCreated,
		TriggerCustomerSignup,
		TriggerLoyaltyTierUp,
		TriggerLoyaltyTierDown,
		TriggerDateBirthday,
		TriggerInactivity,
		TriggerTagAdded,
		TriggerPointsThreshold,
	}
	seen := make(map[string]bool)
	for _, tr := range triggers {
		if seen[tr] {
			t.Errorf("duplicate trigger constant: %q", tr)
		}
		seen[tr] = true
	}
}

func TestNormalizeAutomationActions_UsesCanonicalObjectTypes(t *testing.T) {
	got, err := NormalizeAutomationActions([]byte(`["send_email",{"type":"tag_add","tag":"vip"}]`))
	if err != nil {
		t.Fatalf("NormalizeAutomationActions: %v", err)
	}

	var actions []map[string]any
	if err := json.Unmarshal(got, &actions); err != nil {
		t.Fatalf("decode normalized actions: %v", err)
	}
	if len(actions) != 2 || actions[0]["type"] != ActionSendEmail || actions[1]["type"] != ActionTagAdd {
		t.Fatalf("unexpected normalized actions: %#v", actions)
	}
}

func TestValidateAutomationDefinition_RejectsPresentationValues(t *testing.T) {
	if err := ValidateAutomationDefinition("New customer", []byte(`[{"type":"Send email"}]`)); err == nil {
		t.Fatal("expected presentation labels to be rejected")
	}
	if err := ValidateAutomationDefinition(TriggerCustomerSignup, []byte(`[{"type":"send_email"}]`)); err != nil {
		t.Fatalf("expected canonical definition to pass: %v", err)
	}
}

func TestActionConstants_AreUnique(t *testing.T) {
	actions := []string{
		ActionSendEmail,
		ActionSendSMS,
		ActionAwardPts,
		ActionTagAdd,
	}
	seen := make(map[string]bool)
	for _, ac := range actions {
		if seen[ac] {
			t.Errorf("duplicate action constant: %q", ac)
		}
		seen[ac] = true
	}
}
