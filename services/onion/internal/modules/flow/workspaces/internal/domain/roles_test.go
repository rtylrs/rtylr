package domain_test

import (
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// The two tiers task mutations gate on: writes need member, deletes need manager.
func TestRoleAllows_EveryRoleAgainstBothTiers(t *testing.T) {
	cases := []struct {
		role       string
		writeTier  bool // may create/update tasks, comments, attachments, time logs
		deleteTier bool // may delete them
	}{
		{role: domain.FlowRoleViewer, writeTier: false, deleteTier: false},
		{role: domain.FlowRoleMember, writeTier: true, deleteTier: false},
		{role: domain.FlowRoleAdmin, writeTier: true, deleteTier: true},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			if got := domain.RoleAllows(tc.role, domain.FlowRoleMember); got != tc.writeTier {
				t.Errorf("RoleAllows(%q, member) = %v, want %v", tc.role, got, tc.writeTier)
			}
			if got := domain.RoleAllows(tc.role, domain.FlowRoleAdmin); got != tc.deleteTier {
				t.Errorf("RoleAllows(%q, admin) = %v, want %v", tc.role, got, tc.deleteTier)
			}
		})
	}
}

// The API layer validates roles as manager|member|viewer while the column is
// ENUM('admin','member','viewer'), so both spellings of the top tier must rank
// the same. See RankFlowRole's doc comment.
func TestRoleAllows_ManagerIsAliasForAdmin(t *testing.T) {
	if !domain.RoleAllows(domain.FlowRoleManager, domain.FlowRoleAdmin) {
		t.Error("manager should satisfy the admin tier")
	}
	if !domain.RoleAllows(domain.FlowRoleAdmin, domain.FlowRoleManager) {
		t.Error("admin should satisfy the manager tier")
	}
	if domain.RankFlowRole(domain.FlowRoleManager) != domain.RankFlowRole(domain.FlowRoleAdmin) {
		t.Error("manager and admin must rank equally")
	}
}

// An unrecognised role must never be treated as sufficient, and an unrecognised
// requirement must never be treated as satisfiable — both directions fail closed.
func TestRoleAllows_UnknownRolesFailClosed(t *testing.T) {
	unknown := []string{"", "owner", "ADMIN", "superuser"}

	for _, role := range unknown {
		if domain.RoleAllows(role, domain.FlowRoleMember) {
			t.Errorf("RoleAllows(%q, member) = true, want false", role)
		}
		if domain.RoleAllows(domain.FlowRoleAdmin, role) {
			t.Errorf("RoleAllows(admin, %q) = true, want false", role)
		}
		if domain.RankFlowRole(role) != 0 {
			t.Errorf("RankFlowRole(%q) = %d, want 0", role, domain.RankFlowRole(role))
		}
	}
}

func TestRoleAllows_ViewerTierAdmitsEveryKnownRole(t *testing.T) {
	for _, role := range []string{domain.FlowRoleViewer, domain.FlowRoleMember, domain.FlowRoleAdmin} {
		if !domain.RoleAllows(role, domain.FlowRoleViewer) {
			t.Errorf("RoleAllows(%q, viewer) = false, want true", role)
		}
	}
}

func TestCanonicalFlowMemberRole(t *testing.T) {
	cases := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: domain.FlowRoleAdmin, want: domain.FlowRoleAdmin, ok: true},
		{input: domain.FlowRoleManager, want: domain.FlowRoleAdmin, ok: true},
		{input: domain.FlowRoleMember, want: domain.FlowRoleMember, ok: true},
		{input: domain.FlowRoleViewer, want: domain.FlowRoleViewer, ok: true},
		{input: "owner", ok: false},
	}

	for _, tc := range cases {
		got, ok := domain.CanonicalFlowMemberRole(tc.input)
		if got != tc.want || ok != tc.ok {
			t.Errorf("CanonicalFlowMemberRole(%q) = (%q, %v), want (%q, %v)", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}
