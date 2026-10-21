package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/testutil"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
)

func TestRegisterCreatesRequestedBusinessBeforeIssuingToken(t *testing.T) {
	t.Setenv("EMAIL_VERIFICATION_SECRET", rand.Text())
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := application.NewService(moduleinfra.NewRepository(db), "")
	service.JWT = jwtutils.NewJWTHelperWithDuration(rand.Text(), "onion-platform", time.Hour)
	a := NewAuthAPI(service, nil, nil, nil, nil)
	for _, name := range []string{"Registration Business " + rand.Text(), "   "} {
		t.Run(strings.TrimSpace(name), func(t *testing.T) {
			email := "register-business-" + strings.ToLower(rand.Text()) + "@example.com"
			res, err := a.Register(ctx, &commonauthv1.RegisterRequest{
				Email: email, Password: "Aa9!" + rand.Text(), Name: "Registration Owner", BusinessName: &name,
			})
			if err != nil {
				t.Fatalf("Register: %v", err)
			}
			var businessID sql.NullInt64
			if err := db.QueryRowContext(ctx, "SELECT business_id FROM users WHERE id = ?", res.UserId).Scan(&businessID); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = ?", email); err != nil {
					t.Error(err)
				}
				for _, query := range []string{
					"DELETE FROM authorization_business_members WHERE user_id = ?",
					"DELETE FROM user_provisioning_states WHERE user_id = ?",
					"DELETE FROM user_email_verification_tokens WHERE user_id = ?",
					"DELETE FROM hr_employees WHERE user_id = ?",
					"DELETE FROM subscriptions WHERE owner_id = ?",
					"DELETE FROM businesses WHERE owner_id = ?",
					"DELETE FROM users WHERE id = ?",
				} {
					if _, err := db.ExecContext(ctx, query, res.UserId); err != nil {
						t.Error(err)
					}
				}
			})
			claims := &application.AccessTokenClaims{}
			if err := service.JWT.ValidateToken(ctx, res.Token, claims); err != nil {
				t.Fatal(err)
			}
			// The account can load /users/me but nothing else until the code
			// mailed at signup is used.
			var mode, state string
			var provisionedBusinessID int64
			var emailVerified bool
			if err := db.QueryRowContext(ctx, "SELECT mode, state, business_id, email_verified FROM user_provisioning_states WHERE user_id = ?", res.UserId).
				Scan(&mode, &state, &provisionedBusinessID, &emailVerified); err != nil {
				t.Fatalf("signup provisioning row: %v", err)
			}
			if mode != "self_signup" || state != "active" || emailVerified || provisionedBusinessID != businessID.Int64 {
				t.Fatalf("provisioning = %s/%s business=%d verified=%t", mode, state, provisionedBusinessID, emailVerified)
			}
			var pendingCodes int
			if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM user_email_verification_tokens WHERE user_id = ? AND consumed_at IS NULL", res.UserId).Scan(&pendingCodes); err != nil {
				t.Fatal(err)
			}
			if pendingCodes != 1 {
				t.Fatalf("pending verification codes = %d, want 1", pendingCodes)
			}
			if strings.TrimSpace(name) == "" {
				if businessID.Int64 != 0 || claims.BusinessID != 0 {
					t.Fatal("blank business name created a business")
				}
				return
			}
			if businessID.Int64 <= 0 || claims.BusinessID != businessID.Int64 || claims.IsConsoleAdmin {
				t.Fatalf("business=%v, token business=%d console_admin=%t", businessID, claims.BusinessID, claims.IsConsoleAdmin)
			}
			var storedName string
			if err := db.QueryRowContext(ctx, "SELECT name FROM businesses WHERE id = ? AND owner_id = ?", businessID.Int64, res.UserId).Scan(&storedName); err != nil {
				t.Fatal(err)
			}
			if storedName != strings.TrimSpace(name) {
				t.Fatalf("business name = %q", storedName)
			}
		})
	}
}
