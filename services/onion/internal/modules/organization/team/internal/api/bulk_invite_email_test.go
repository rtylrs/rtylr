package api

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/team/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/team/internal/infrastructure"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/config"
	emailPkg "github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/testutil"
	hubteaminvitationsv1 "github.com/voxire/voxie/gen/go/hub/team/invitations/v1"
)

type testEmailNotifications struct{ db *sql.DB }

func (n testEmailNotifications) QueueInvite(ctx context.Context, invite application.InviteEmail) error {
	return emailPkg.EnqueueInvitation(ctx, n.db, invite.To, emailPkg.InvitationData{
		InviterName: invite.InviterName, BusinessName: invite.BusinessName, InviteLink: invite.Link, ValidDays: invite.ValidDays,
	})
}

func (testEmailNotifications) QueueAccessRequest(context.Context, string, string, string, string) error {
	return nil
}

// seedBulkInviteTenant creates a dedicated owner+business pair. The shared
// fixture business cannot be used: it is written to and dropped by other suites
// running in parallel, and businesses.name is unique, so re-seeding under the
// fixture's name silently collides.
func seedBulkInviteTenant(t *testing.T, db *sql.DB) (actorID, businessID int64) {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().UnixNano()

	res, err := db.ExecContext(ctx,
		`INSERT INTO users (email, username, password_hash) VALUES (?, ?, '')`,
		fmt.Sprintf("bulk-invite-actor-%d@test.local", suffix), fmt.Sprintf("bulk-invite-actor-%d", suffix))
	if err != nil {
		t.Fatalf("insert actor: %v", err)
	}
	actorID, err = res.LastInsertId()
	if err != nil {
		t.Fatalf("actor id: %v", err)
	}

	res, err = db.ExecContext(ctx,
		`INSERT INTO businesses (name, owner_id) VALUES (?, ?)`,
		fmt.Sprintf("Bulk Invite Business %d", suffix), actorID)
	if err != nil {
		t.Fatalf("insert business: %v", err)
	}
	businessID, err = res.LastInsertId()
	if err != nil {
		t.Fatalf("business id: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM businesses WHERE id = ?", businessID)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", actorID)
	})
	return actorID, businessID
}

// Bulk-invited members are created without a password, so the queued invitation
// email is their only route into the account. Before the fix BulkInvite answered
// "Bulk invite complete" and never enqueued anything.
func TestBulkInviteEnqueuesInvitationEmailForNewInvitee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	actorID, businessID := seedBulkInviteTenant(t, db)

	invitee := fmt.Sprintf("bulk-invitee-%d@test.local", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = ?", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM invite_tokens WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_license_assignments WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_role_assignments WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_business_members WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM authorization_audit_events WHERE target_user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM user_provisioning_states WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE user_id IN (SELECT id FROM users WHERE email = ?)", invitee)
		_, _ = db.ExecContext(ctx, "DELETE FROM users WHERE email = ?", invitee)
	})

	api := NewInvitationsAPI(application.NewService(ctx, moduleinfra.NewRepository(db)).WithEmailNotifications(testEmailNotifications{db: db}))
	authCtx := authPkg.WithAuthUser(ctx, &authPkg.AuthUser{
		UserID:     actorID,
		BusinessID: businessID,
		EffectiveAccess: &authorization.EffectiveAccess{
			IsOwner:         true,
			DashPermissions: []string{"hub:team:invite"},
		},
	})

	response, err := api.BulkInvite(authCtx, &hubteaminvitationsv1.BulkInviteRequest{
		Emails:     []string{invitee},
		MemberType: "member",
	})
	if err != nil {
		t.Fatalf("BulkInvite: %v", err)
	}
	if len(response.Results) != 1 {
		t.Fatalf("got %d results, want 1", len(response.Results))
	}
	if status := response.Results[0].Status; status != "invited" {
		t.Fatalf("status = %q, want invited (message %q)", status, response.Results[0].Message)
	}

	var emailCount int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM email_jobs WHERE to_email = ?`, invitee).Scan(&emailCount); err != nil {
		t.Fatalf("count invitation emails: %v", err)
	}
	if emailCount != 1 {
		t.Fatalf("invitation email count = %d, want 1", emailCount)
	}

	var token string
	if err := db.QueryRowContext(ctx,
		`SELECT t.token FROM invite_tokens t JOIN users u ON u.id = t.user_id WHERE u.email = ?`, invitee).Scan(&token); err != nil {
		t.Fatalf("read invite token: %v", err)
	}
	var subject, html string
	if err := db.QueryRowContext(ctx,
		`SELECT subject, html FROM email_jobs WHERE to_email = ? ORDER BY id DESC LIMIT 1`, invitee).
		Scan(&subject, &html); err != nil {
		t.Fatalf("read queued email: %v", err)
	}
	// The inviter has no name, so their email stands in; the business is named.
	if !strings.Contains(subject, "invited you to join Bulk Invite Business") || !strings.Contains(subject, "bulk-invite-actor-") {
		t.Fatalf("queued email subject = %q, want the inviter and business named", subject)
	}
	if strings.Contains(subject, "rtylr on rtylr") || !strings.Contains(html, "expires in 7 days") {
		t.Fatalf("queued invitation = %q, want real names and the expiry", subject)
	}
	if !strings.Contains(html, config.AuthAppURL+"/invite/"+token) {
		t.Fatalf("queued invitation does not contain the invite link for the created token")
	}
}
