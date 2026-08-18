package domain

import (
	"testing"

	"github.com/voxire/onion/internal/platform/money"
)

func TestHasNotes_WhenSet(t *testing.T) {
	l := Lead{Notes: "Follow up Tuesday"}
	if !l.HasNotes() {
		t.Error("expected HasNotes=true when notes are set")
	}
}

func TestHasNotes_WhenEmpty(t *testing.T) {
	l := Lead{Notes: ""}
	if l.HasNotes() {
		t.Error("expected HasNotes=false when notes are empty")
	}
}

func TestHasAssignee_WhenSet(t *testing.T) {
	id := int64(7)
	l := Lead{AssignedTo: &id}
	if !l.HasAssignee() {
		t.Error("expected HasAssignee=true when assigned_to is set")
	}
}

func TestHasAssignee_WhenNil(t *testing.T) {
	l := Lead{AssignedTo: nil}
	if l.HasAssignee() {
		t.Error("expected HasAssignee=false when assigned_to is nil")
	}
}

func TestValidateCreateContractRequest_Valid(t *testing.T) {
	msg := ValidateCreateContractRequest(CreateContractRequest{LeadID: 1, Plan: "pro"})
	if msg != "" {
		t.Errorf("expected no error, got %q", msg)
	}
}

func TestValidateCreateContractRequest_ZeroLeadID(t *testing.T) {
	msg := ValidateCreateContractRequest(CreateContractRequest{LeadID: 0, Plan: "pro"})
	if msg == "" {
		t.Error("expected error for zero lead_id")
	}
}

func TestValidateCreateContractRequest_EmptyPlan(t *testing.T) {
	msg := ValidateCreateContractRequest(CreateContractRequest{LeadID: 1, Plan: ""})
	if msg == "" {
		t.Error("expected error for empty plan")
	}
}

func TestHasMRRValue_WhenPositive(t *testing.T) {
	c := Contract{MRR: money.Amount(50000)}
	if !c.HasMRRValue() {
		t.Error("expected HasMRRValue=true when MRR > 0")
	}
}

func TestHasMRRValue_WhenZero(t *testing.T) {
	c := Contract{MRR: money.Amount(0)}
	if c.HasMRRValue() {
		t.Error("expected HasMRRValue=false when MRR = 0")
	}
}

func TestHasSignedBy_WhenSet(t *testing.T) {
	c := Contract{SignedByName: "John Doe"}
	if !c.HasSignedBy() {
		t.Error("expected HasSignedBy=true when SignedByName is set")
	}
}

func TestHasSignedBy_WhenEmpty(t *testing.T) {
	c := Contract{SignedByName: ""}
	if c.HasSignedBy() {
		t.Error("expected HasSignedBy=false when SignedByName is empty")
	}
}

func TestHasMRR_WhenSet_AndIsWon(t *testing.T) {
	mrr := money.Amount(99900)
	l := Lead{MRRPotential: &mrr, Stage: LeadStageContracted}
	if !l.HasMRR() {
		t.Error("expected HasMRR=true")
	}
	if !l.IsWon() {
		t.Error("expected IsWon=true for contracted lead")
	}
}

func TestAllStages_ContainsCold(t *testing.T) {
	stages := AllStages()
	for _, s := range stages {
		if s == LeadStageCold {
			return
		}
	}
	t.Error("AllStages should contain LeadStageCold")
}

func TestAllStages_ContainsLost(t *testing.T) {
	stages := AllStages()
	for _, s := range stages {
		if s == LeadStageLost {
			return
		}
	}
	t.Error("AllStages should contain LeadStageLost")
}
