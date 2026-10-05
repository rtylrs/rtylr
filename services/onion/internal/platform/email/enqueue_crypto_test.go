package email

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/voxire/onion/internal/testutil"
)

func TestEnqueueTemplateSealsPasswordResetAndInvitationBodies(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	db := testutil.OpenDB(t)
	ctx := context.Background()

	cases := []struct {
		name  string
		tmpl  Template
		data  any
		token string
	}{
		{
			name:  "password reset",
			tmpl:  TemplatePasswordReset,
			data:  PasswordResetData{Subject: "Reset", ResetLink: "https://auth.test/reset?token=reset-secret"},
			token: "reset-secret",
		},
		{
			name:  "invitation",
			tmpl:  TemplateInvitation,
			data:  InvitationData{Subject: "Invite", InviterName: "Alex", BusinessName: "Acme", InviteLink: "https://auth.test/invite?token=invite-secret"},
			token: "invite-secret",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subject := "encrypt-" + uuid.NewString()
			t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM email_jobs WHERE subject = ?", subject) })
			if err := EnqueueTemplate(ctx, db, PriorityCritical, "recipient@test.local", subject, tc.tmpl, tc.data); err != nil {
				t.Fatalf("EnqueueTemplate() error = %v", err)
			}

			var html string
			var text sql.NullString
			if err := db.QueryRowContext(ctx, "SELECT html, text_body FROM email_jobs WHERE subject = ?", subject).Scan(&html, &text); err != nil {
				t.Fatalf("read queued body: %v", err)
			}
			if !text.Valid {
				t.Fatal("queued text body is NULL")
			}
			for name, stored := range map[string]string{"html": html, "text": text.String} {
				if !strings.HasPrefix(stored, encryptedPrefix) {
					t.Errorf("%s was stored without the encryption prefix", name)
				}
				if strings.Contains(stored, tc.token) {
					t.Errorf("%s contains the live token in plaintext", name)
				}
			}
		})
	}
}
