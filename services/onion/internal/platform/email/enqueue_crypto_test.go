package email

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

type captureDB struct {
	query string
	args  []any
}

func (c *captureDB) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	c.query = query
	c.args = append([]any(nil), args...)
	return captureResult{}, nil
}

func (c *captureDB) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, nil
}

func (c *captureDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, nil
}

func (c *captureDB) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil
}

type captureResult struct{}

func (captureResult) LastInsertId() (int64, error) { return 0, nil }
func (captureResult) RowsAffected() (int64, error) { return 1, nil }

func TestEnqueueTemplateSealsPasswordResetAndInvitationBodies(t *testing.T) {
	t.Setenv("EMAIL_ENCRYPTION_KEY", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")

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
			db := &captureDB{}
			if err := enqueueTemplate(context.Background(), db, PriorityCritical, "recipient@test.local", tc.name, tc.tmpl, tc.data); err != nil {
				t.Fatalf("enqueueTemplate() error = %v", err)
			}
			if len(db.args) < 5 {
				t.Fatalf("enqueueTemplate() passed %d args, want at least 5", len(db.args))
			}

			html, ok := db.args[3].(string)
			if !ok {
				t.Fatalf("html argument = %T, want string", db.args[3])
			}
			text, ok := db.args[4].(sql.NullString)
			if !ok || !text.Valid {
				t.Fatalf("text argument = %#v, want valid sql.NullString", db.args[4])
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
