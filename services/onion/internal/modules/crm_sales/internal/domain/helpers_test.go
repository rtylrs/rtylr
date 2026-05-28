package domain

import "testing"

func TestIsOpen_WhenCold(t *testing.T) {
	l := Lead{Stage: LeadStageCold}
	if !l.IsOpen() {
		t.Error("expected IsOpen=true for cold lead")
	}
}

func TestIsOpen_WhenContracted(t *testing.T) {
	l := Lead{Stage: LeadStageContracted}
	if l.IsOpen() {
		t.Error("expected IsOpen=false for contracted lead")
	}
}

func TestIsOpen_WhenLost(t *testing.T) {
	l := Lead{Stage: LeadStageLost}
	if l.IsOpen() {
		t.Error("expected IsOpen=false for lost lead")
	}
}

func TestIsWon_WhenContracted(t *testing.T) {
	l := Lead{Stage: LeadStageContracted}
	if !l.IsWon() {
		t.Error("expected IsWon=true for contracted lead")
	}
}

func TestIsWon_WhenCold(t *testing.T) {
	l := Lead{Stage: LeadStageCold}
	if l.IsWon() {
		t.Error("expected IsWon=false for cold lead")
	}
}

func TestIsLost_WhenLost(t *testing.T) {
	l := Lead{Stage: LeadStageLost}
	if !l.IsLost() {
		t.Error("expected IsLost=true for lost lead")
	}
}

func TestIsLost_WhenContacted(t *testing.T) {
	l := Lead{Stage: LeadStageContacted}
	if l.IsLost() {
		t.Error("expected IsLost=false for contacted lead")
	}
}

func TestHasMRR_WhenSet(t *testing.T) {
	mrr := 999.0
	l := Lead{MRRPotential: &mrr}
	if !l.HasMRR() {
		t.Error("expected HasMRR=true when MRRPotential is set")
	}
}

func TestHasMRR_WhenNil(t *testing.T) {
	l := Lead{MRRPotential: nil}
	if l.HasMRR() {
		t.Error("expected HasMRR=false when MRRPotential is nil")
	}
}

func TestHasMRR_WhenZero(t *testing.T) {
	mrr := 0.0
	l := Lead{MRRPotential: &mrr}
	if l.HasMRR() {
		t.Error("expected HasMRR=false when MRRPotential is 0")
	}
}

func TestTotalPipelineMRR_Sum(t *testing.T) {
	summaries := []PipelineSummary{
		{Stage: "cold", Count: 3, TotalMRR: 500},
		{Stage: "contacted", Count: 2, TotalMRR: 1000},
		{Stage: "contracted", Count: 1, TotalMRR: 2000},
	}
	if got := TotalPipelineMRR(summaries); got != 3500 {
		t.Errorf("TotalPipelineMRR = %f, want 3500", got)
	}
}

func TestTotalPipelineMRR_Empty(t *testing.T) {
	if got := TotalPipelineMRR(nil); got != 0 {
		t.Errorf("expected 0 for empty summaries, got %f", got)
	}
}

func TestTotalPipelineCount_Sum(t *testing.T) {
	summaries := []PipelineSummary{
		{Stage: "cold", Count: 5},
		{Stage: "contacted", Count: 3},
		{Stage: "contracted", Count: 1},
	}
	if got := TotalPipelineCount(summaries); got != 9 {
		t.Errorf("TotalPipelineCount = %d, want 9", got)
	}
}

func TestTotalPipelineCount_Empty(t *testing.T) {
	if got := TotalPipelineCount(nil); got != 0 {
		t.Errorf("expected 0 for empty summaries, got %d", got)
	}
}

func TestValidateCreateLeadRequest_EmptyCompanyName(t *testing.T) {
	msg := ValidateCreateLeadRequest(CreateLeadRequest{CompanyName: ""})
	if msg == "" {
		t.Error("expected validation error for empty company name")
	}
}

func TestValidateCreateLeadRequest_Valid(t *testing.T) {
	msg := ValidateCreateLeadRequest(CreateLeadRequest{CompanyName: "Acme Corp"})
	if msg != "" {
		t.Errorf("expected no validation error, got %q", msg)
	}
}

func TestAllStages_ReturnsAllSeven(t *testing.T) {
	stages := AllStages()
	if len(stages) != 7 {
		t.Errorf("expected 7 stages, got %d", len(stages))
	}
}
