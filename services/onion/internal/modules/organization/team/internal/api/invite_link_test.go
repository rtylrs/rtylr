package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/config"
	"github.com/voxire/onion/internal/testutil"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
)

func cleanupInvitee(t *testing.T, email string) {
	t.Helper()
	db := testutil.OpenDB(t)
	t.Cleanup(func() {
		ctx := context.Background()
		for _, stmt := range []string{
			"DELETE FROM email_jobs WHERE to_email = ?",
			"DELETE FROM invite_tokens WHERE user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM authorization_license_assignments WHERE user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM authorization_role_assignments WHERE user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM authorization_business_members WHERE user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM authorization_audit_events WHERE target_user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM user_provisioning_states WHERE user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM hr_employees WHERE user_id IN (SELECT id FROM users WHERE email = ?)",
			"DELETE FROM users WHERE email = ?",
		} {
			_, _ = db.ExecContext(ctx, stmt, email)
		}
	})
}

// An invited member has no password and stays pending until they open the
// activation link. With no outbound email that link reached nobody, so the
// inviter now gets it back to share another way (onion#1100).
func TestInviteUserReturnsTheActivationLinkToTheInviter(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	actorID, businessID := seedBulkInviteTenant(t, db)
	invitee := fmt.Sprintf("link-invitee-%d@test.local", time.Now().UnixNano())
	cleanupInvitee(t, invitee)

	api := NewInvitationsAPI(application.NewService(ctx, moduleinfra.NewRepository(db)))
	authCtx := authPkg.WithAuthUser(ctx, &authPkg.AuthUser{
		UserID:     actorID,
		BusinessID: businessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			IsOwner:         true,
			DashPermissions: []string{"hub:team:invite"},
		},
	})

	response, err := api.InviteUser(authCtx, &hubteaminvitationsv1.InviteUserRequest{
		Email: invitee, FullName: "Link Invitee", MemberType: "member",
	})
	if err != nil {
		t.Fatalf("InviteUser: %v", err)
	}

	var token string
	if err := db.QueryRowContext(ctx,
		`SELECT t.token FROM invite_tokens t JOIN users u ON u.id = t.user_id WHERE u.email = ?`, invitee).Scan(&token); err != nil {
		t.Fatalf("read invite token: %v", err)
	}
	want := config.AuthAppURL + "/invite/" + token
	if response.InviteLink == nil || *response.InviteLink != want {
		t.Fatalf("invite link = %v, want the link for the created token", response.InviteLink)
	}
}

// Direct credentials already give the inviter a username and password, and no
// invitation is created, so there is no link to return.
func TestInviteUserReturnsNoLinkForDirectCredentials(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	actorID, businessID := seedBulkInviteTenant(t, db)
	suffix := time.Now().UnixNano()
	invitee := fmt.Sprintf("direct-invitee-%d@test.local", suffix)
	cleanupInvitee(t, invitee)

	api := NewInvitationsAPI(application.NewService(ctx, moduleinfra.NewRepository(db)))
	authCtx := authPkg.WithAuthUser(ctx, &authPkg.AuthUser{
		UserID:     actorID,
		BusinessID: businessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			IsOwner:         true,
			DashPermissions: []string{"hub:team:invite"},
		},
	})

	username := fmt.Sprintf("direct%d", suffix)
	password := "Temp-Password-123!"
	mode := "direct_credentials"
	response, err := api.InviteUser(authCtx, &hubteaminvitationsv1.InviteUserRequest{
		Email: invitee, FullName: "Direct Invitee", MemberType: "member",
		Username: &username, TemporaryPassword: &password, ProvisioningMode: &mode,
	})
	if err != nil {
		t.Fatalf("InviteUser: %v", err)
	}
	if response.InviteLink != nil {
		t.Fatalf("invite link = %q, want none for direct credentials", *response.InviteLink)
	}
}
