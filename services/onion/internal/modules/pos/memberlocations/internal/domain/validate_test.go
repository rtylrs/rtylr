package domain

import "testing"

func TestMemberLocationInput_Validate_RejectsMissingBusinessID(t *testing.T) {
	m := MemberLocationInput{UserID: 1, LocationID: 2}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing business id")
	}
}

func TestMemberLocationInput_Validate_RejectsMissingUserID(t *testing.T) {
	m := MemberLocationInput{BusinessID: 1, LocationID: 2}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing user id")
	}
}

func TestMemberLocationInput_Validate_RejectsMissingLocationID(t *testing.T) {
	m := MemberLocationInput{BusinessID: 1, UserID: 2}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing location id")
	}
}

func TestMemberLocationInput_Validate_AcceptsValid(t *testing.T) {
	m := MemberLocationInput{BusinessID: 1, UserID: 2, LocationID: 3}
	if err := m.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemberLocation_HasManagementPrivileges_PosUsers(t *testing.T) {
	m := MemberLocation{CanManagePOSUsers: true}
	if !m.HasManagementPrivileges() {
		t.Error("expected HasManagementPrivileges=true when CanManagePOSUsers is set")
	}
}

func TestMemberLocation_HasManagementPrivileges_Terminals(t *testing.T) {
	m := MemberLocation{CanManageTerminals: true}
	if !m.HasManagementPrivileges() {
		t.Error("expected HasManagementPrivileges=true when CanManageTerminals is set")
	}
}

func TestMemberLocation_HasManagementPrivileges_None(t *testing.T) {
	m := MemberLocation{}
	if m.HasManagementPrivileges() {
		t.Error("expected HasManagementPrivileges=false with no privileges")
	}
}

func TestMemberLocation_HasManagementPrivileges_ReportsOnlyNoPrivileges(t *testing.T) {
	m := MemberLocation{CanViewReports: true}
	if m.HasManagementPrivileges() {
		t.Error("expected HasManagementPrivileges=false for report-only access")
	}
}
