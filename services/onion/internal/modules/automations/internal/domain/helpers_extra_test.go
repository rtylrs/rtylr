package domain

import (
	"testing"
	"time"
)

func TestHasRuns_WhenPositive(t *testing.T) {
	a := Automation{RunCount: 3}
	if !a.HasRuns() {
		t.Error("expected HasRuns=true when RunCount > 0")
	}
}

func TestHasRuns_WhenZero(t *testing.T) {
	a := Automation{RunCount: 0}
	if a.HasRuns() {
		t.Error("expected HasRuns=false when RunCount = 0")
	}
}

func TestHasLastRun_WhenSet(t *testing.T) {
	now := time.Now()
	a := Automation{LastRunAt: &now}
	if !a.HasLastRun() {
		t.Error("expected HasLastRun=true when LastRunAt is set")
	}
}

func TestHasLastRun_WhenNil(t *testing.T) {
	a := Automation{LastRunAt: nil}
	if a.HasLastRun() {
		t.Error("expected HasLastRun=false when LastRunAt is nil")
	}
}

func TestValidateAutomationUpdateInput_Valid(t *testing.T) {
	msg := ValidateAutomationUpdateInput(AutomationUpdateInput{
		ID:          10,
		OwnerID:     1,
		Name:        "Welcome Flow",
		TriggerType: TriggerCustomerSignup,
	})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateAutomationUpdateInput_ZeroID(t *testing.T) {
	msg := ValidateAutomationUpdateInput(AutomationUpdateInput{
		ID:          0,
		OwnerID:     1,
		Name:        "Welcome Flow",
		TriggerType: TriggerOrderCreated,
	})
	if msg == "" {
		t.Error("expected error for zero id")
	}
}

func TestValidateAutomationUpdateInput_ZeroOwner(t *testing.T) {
	msg := ValidateAutomationUpdateInput(AutomationUpdateInput{
		ID:          5,
		OwnerID:     0,
		Name:        "Welcome Flow",
		TriggerType: TriggerOrderCreated,
	})
	if msg == "" {
		t.Error("expected error for zero owner_id")
	}
}

func TestValidateAutomationUpdateInput_EmptyName(t *testing.T) {
	msg := ValidateAutomationUpdateInput(AutomationUpdateInput{
		ID:          5,
		OwnerID:     1,
		Name:        "",
		TriggerType: TriggerOrderCreated,
	})
	if msg == "" {
		t.Error("expected error for empty name")
	}
}

func TestValidateAutomationUpdateInput_EmptyTrigger(t *testing.T) {
	msg := ValidateAutomationUpdateInput(AutomationUpdateInput{
		ID:          5,
		OwnerID:     1,
		Name:        "My Flow",
		TriggerType: "",
	})
	if msg == "" {
		t.Error("expected error for empty trigger type")
	}
}
