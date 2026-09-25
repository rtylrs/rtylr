package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	"github.com/voxire/onion/internal/modules/team/internal/domain"
	"github.com/voxire/onion/internal/modules/team/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestInviteViewerGetsReadOnlyHRAndPOSRoles(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := application.NewService(ctx, db)
	repo := infrastructure.NewRepository(db)
	email := fmt.Sprintf("viewer-role-invite-%d@test.local", time.Now().UnixNano())

	userID, _, err := service.InviteUser(ctx, testutil.OwnerID, testutil.BusinessID, domain.InviteUserInput{
		Email: email,
		Apps:  []string{"hr", "pos"},
		Role:  domain.RoleViewer,
	})
	if err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.DeleteInvitedUser(ctx, testutil.BusinessID, testutil.OwnerID, userID); err != nil {
			t.Errorf("clean up invited viewer: %v", err)
		}
	})

	rows, err := db.QueryContext(ctx, `
		SELECT r.app_id, r.role_key
		FROM authorization_role_assignments a
		JOIN authorization_roles r ON r.id = a.role_id
		WHERE a.business_id = ? AND a.user_id = ? AND a.status IN ('active', 'pending')
		ORDER BY r.app_id`, testutil.BusinessID, userID)
	if err != nil {
		t.Fatalf("query viewer role assignments: %v", err)
	}
	assignments := make(map[string]string)
	for rows.Next() {
		var appID, roleKey string
		if err := rows.Scan(&appID, &roleKey); err != nil {
			closeErr := rows.Close()
			if closeErr != nil {
				t.Errorf("close viewer role assignments after scan failure: %v", closeErr)
			}
			t.Fatalf("scan viewer role assignment: %v", err)
		}
		assignments[appID] = roleKey
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("close viewer role assignments: %v", err)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate viewer role assignments: %v", err)
	}

	for appID, wantRoleKey := range map[string]string{
		"hr":  "hr.viewer",
		"pos": "pos.viewer",
	} {
		if got := assignments[appID]; got != wantRoleKey {
			t.Errorf("%s role key = %q, want %q", appID, got, wantRoleKey)
		}
	}

	var viewerRoleCount, permissionCount int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT r.role_key), COUNT(*)
		FROM authorization_role_permissions rp
		JOIN authorization_permission_catalog p ON p.permission_key = rp.permission_key
		JOIN authorization_roles r ON r.id = rp.role_id
		WHERE r.business_id = 0
		  AND r.role_key IN ('hr.viewer', 'pos.viewer')`).Scan(&viewerRoleCount, &permissionCount); err != nil {
		t.Fatalf("query viewer role permission counts: %v", err)
	}
	if viewerRoleCount != 2 || permissionCount == 0 {
		t.Fatalf("viewer roles = %d and permissions = %d, want 2 roles with permissions", viewerRoleCount, permissionCount)
	}

	var nonReadOnlyPermissions int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM authorization_role_permissions rp
		JOIN authorization_permission_catalog p ON p.permission_key = rp.permission_key
		JOIN authorization_roles r ON r.id = rp.role_id
		WHERE r.business_id = 0
		  AND r.role_key IN ('hr.viewer', 'pos.viewer')
		  AND (p.action <> 'view' OR p.is_sensitive = TRUE)`).Scan(&nonReadOnlyPermissions); err != nil {
		t.Fatalf("query viewer role permissions: %v", err)
	}
	if nonReadOnlyPermissions != 0 {
		t.Fatalf("viewer roles have %d non-view or sensitive permissions, want 0", nonReadOnlyPermissions)
	}
}
