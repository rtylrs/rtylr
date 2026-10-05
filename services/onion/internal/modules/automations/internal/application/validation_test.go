package application

import (
	"encoding/json"
	"testing"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
)

func TestNormalizeAutomationActions_UsesCanonicalObjectTypes(t *testing.T) {
	got, err := NormalizeAutomationActions([]byte(`["send_email",{"type":"tag_add","tag":"vip"}]`))
	if err != nil {
		t.Fatalf("NormalizeAutomationActions: %v", err)
	}

	var actions []map[string]any
	if err := json.Unmarshal(got, &actions); err != nil {
		t.Fatalf("decode normalized actions: %v", err)
	}
	if len(actions) != 2 || actions[0]["type"] != domain.ActionSendEmail || actions[1]["type"] != domain.ActionTagAdd {
		t.Fatalf("unexpected normalized actions: %#v", actions)
	}
}

func TestValidateAutomationDefinition_RejectsPresentationValues(t *testing.T) {
	if err := ValidateAutomationDefinition("New customer", []byte(`[{"type":"Send email"}]`)); err == nil {
		t.Fatal("expected presentation labels to be rejected")
	}
	if err := ValidateAutomationDefinition(domain.TriggerCustomerSignup, []byte(`[{"type":"send_email"}]`)); err != nil {
		t.Fatalf("expected canonical definition to pass: %v", err)
	}
}
