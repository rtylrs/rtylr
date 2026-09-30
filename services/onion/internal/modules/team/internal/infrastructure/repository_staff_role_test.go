package infrastructure

import (
	"database/sql"
	"testing"

	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure/persistence/sqlc"
)

// The team list used to be sourced from user_app_permissions alone, so any
// member without a legacy row was invisible. The owner never has one, which is
// how the owner ended up unable to see themselves in their own team.
func TestStaffRolePrefersNormalizedThenLegacyThenMember(t *testing.T) {
	for _, tc := range []struct {
		name     string
		row      sqlc.ListStaffUsersByBusinessIDRow
		expected domain.Role
	}{
		{
			name:     "normalized owner with no legacy row",
			row:      sqlc.ListStaffUsersByBusinessIDRow{DashRoleKey: "owner"},
			expected: domain.RoleOwner,
		},
		{
			name: "normalized admin wins over a legacy row",
			row: sqlc.ListStaffUsersByBusinessIDRow{
				DashRoleKey: "dash.admin",
				LegacyRole:  legacyRole("staff"),
			},
			expected: domain.RoleAdmin,
		},
		{
			name:     "legacy only falls back to the legacy role",
			row:      sqlc.ListStaffUsersByBusinessIDRow{LegacyRole: legacyRole("staff")},
			expected: domain.RoleStaff,
		},
		{
			name:     "neither source reports member",
			row:      sqlc.ListStaffUsersByBusinessIDRow{},
			expected: domain.Role("member"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := staffRole(tc.row); got != tc.expected {
				t.Fatalf("staffRole = %q, want %q", got, tc.expected)
			}
		})
	}
}

// A member revoked through the normalized path must not be reported active just
// because a stale legacy row is still flagged on.
func TestStaffIsActivePrefersNormalizedStatus(t *testing.T) {
	if !staffIsActive(sqlc.ListStaffUsersByBusinessIDRow{MemberStatus: memberStatus("active")}) {
		t.Fatal("an active normalized membership must be reported active")
	}
	revokedButLegacyOn := sqlc.ListStaffUsersByBusinessIDRow{
		MemberStatus:   memberStatus("revoked"),
		LegacyIsActive: sql.NullBool{Bool: true, Valid: true},
	}
	if staffIsActive(revokedButLegacyOn) {
		t.Fatal("a revoked normalized membership must not be reported active")
	}
	if !staffIsActive(sqlc.ListStaffUsersByBusinessIDRow{LegacyIsActive: sql.NullBool{Bool: true, Valid: true}}) {
		t.Fatal("a legacy-only active member must be reported active")
	}
	if staffIsActive(sqlc.ListStaffUsersByBusinessIDRow{}) {
		t.Fatal("a member with no status must not be reported active")
	}
}

func legacyRole(role string) sqlc.NullUserAppPermissionsRole {
	return sqlc.NullUserAppPermissionsRole{
		UserAppPermissionsRole: sqlc.UserAppPermissionsRole(role),
		Valid:                  true,
	}
}

func memberStatus(status string) sqlc.NullAuthorizationBusinessMembersStatus {
	return sqlc.NullAuthorizationBusinessMembersStatus{
		AuthorizationBusinessMembersStatus: sqlc.AuthorizationBusinessMembersStatus(status),
		Valid:                              true,
	}
}
