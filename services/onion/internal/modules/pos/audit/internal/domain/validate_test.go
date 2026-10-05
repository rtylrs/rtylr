package domain

import "testing"

func TestValidateAuditLogInput_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateAuditLogInput(AuditLogInput{BranchID: 1, EmployeeID: 2, Action: "login", EntityType: "session"}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateAuditLogInput_RejectsMissingBranchID(t *testing.T) {
	if err := ValidateAuditLogInput(AuditLogInput{BusinessID: 1, EmployeeID: 2, Action: "login", EntityType: "session"}); err == nil {
		t.Fatal("expected error for missing branch id")
	}
}

func TestValidateAuditLogInput_RejectsMissingEmployeeID(t *testing.T) {
	if err := ValidateAuditLogInput(AuditLogInput{BusinessID: 1, BranchID: 2, Action: "login", EntityType: "session"}); err == nil {
		t.Fatal("expected error for missing employee id")
	}
}

func TestValidateAuditLogInput_RejectsBlankAction(t *testing.T) {
	if err := ValidateAuditLogInput(AuditLogInput{BusinessID: 1, BranchID: 2, EmployeeID: 3, EntityType: "session"}); err == nil {
		t.Fatal("expected error for blank action")
	}
}

func TestValidateAuditLogInput_RejectsBlankEntityType(t *testing.T) {
	if err := ValidateAuditLogInput(AuditLogInput{BusinessID: 1, BranchID: 2, EmployeeID: 3, Action: "login"}); err == nil {
		t.Fatal("expected error for blank entity type")
	}
}

func TestValidateAuditLogInput_AcceptsValid(t *testing.T) {
	if err := ValidateAuditLogInput(AuditLogInput{
		BusinessID: 1, BranchID: 2, EmployeeID: 3, Action: "login", EntityType: "session",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAuditLogFilter_RejectsMissingBusinessID(t *testing.T) {
	if err := ValidateAuditLogFilter(AuditLogFilter{}); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestValidateAuditLogFilter_AcceptsValid(t *testing.T) {
	if err := ValidateAuditLogFilter(AuditLogFilter{BusinessID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
