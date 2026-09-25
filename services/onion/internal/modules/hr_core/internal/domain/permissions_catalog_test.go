package domain_test

import (
	"testing"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

func TestAllPermissions_KeysAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range domain.AllPermissions() {
		if seen[p.Key] {
			t.Errorf("duplicate permission key: %s", p.Key)
		}
		seen[p.Key] = true
	}
}

func TestAllPermissions_EveryEntryHasLabelAndCategory(t *testing.T) {
	for _, p := range domain.AllPermissions() {
		if p.Label == "" {
			t.Errorf("permission %s missing Label", p.Key)
		}
		if p.Category == "" {
			t.Errorf("permission %s missing Category", p.Key)
		}
	}
}

func TestAllPermissions_IncludesManageEmployees(t *testing.T) {
	for _, p := range domain.AllPermissions() {
		if p.Key == domain.PermissionManageEmployees {
			return
		}
	}
	t.Error("expected catalog to include manage_employees")
}

func TestAllPermissions_CoversEveryConstant(t *testing.T) {
	all := map[string]bool{}
	for _, p := range domain.AllPermissions() {
		all[p.Key] = true
	}
	constants := []string{
		domain.PermissionViewOwnProfile, domain.PermissionViewOwnSchedule, domain.PermissionRequestTimeOff,
		domain.PermissionViewTeam, domain.PermissionApproveTimeOff, domain.PermissionManageSchedules,
		domain.PermissionManageEmployees, domain.PermissionManageRoles, domain.PermissionManageDepartments,
		domain.PermissionManageLocations, domain.PermissionManageHRConfig, domain.PermissionViewReports,
		domain.PermissionManagePayroll, domain.PermissionManageBenefits, domain.PermissionManageApprovalChains,
	}
	for _, c := range constants {
		if !all[c] {
			t.Errorf("constant %s is not present in AllPermissions()", c)
		}
	}
}
