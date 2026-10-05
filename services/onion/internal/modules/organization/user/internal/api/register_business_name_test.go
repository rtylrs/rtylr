package api

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/organization/user/internal/application"
	moduleinfra "github.com/voxire/onion/internal/modules/organization/user/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/testutil"
	commonauthv1 "github.com/voxire/voxie/gen/go/common/auth/v1"
)

// TestRegisterWithTakenBusinessNameDoesNotPoisonTheEmail covers B11-R06: a
// business name that is already taken must fail registration before the user
// row exists, so the same email can still complete signup afterwards.
// Previously the user row was created first and the duplicate business insert
// failed after it, leaving an account the signup form could never finish while
// a retry hit "user already exists".
func TestRegisterWithTakenBusinessNameDoesNotPoisonTheEmail(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	service := application.NewService(moduleinfra.NewRepository(db), "")
	service.JWT = jwtutils.NewJWTHelperWithDuration(rand.Text(), "onion-platform", time.Hour)
	a := NewAuthAPI(service, nil, nil, nil, nil)

	firstEmail := "register-taken-a-" + strings.ToLower(rand.Text()) + "@example.com"
	secondEmail := "register-taken-b-" + strings.ToLower(rand.Text()) + "@example.com"
	t.Cleanup(func() {
		cleanupRegisteredUser(t, db, firstEmail)
		cleanupRegisteredUser(t, db, secondEmail)
	})

	sharedName := "Registration Taken " + rand.Text()
	password := "Aa9!" + rand.Text()
	if _, err := a.Register(ctx, &commonauthv1.RegisterRequest{
		Email: firstEmail, Password: password, Name: "First Owner", BusinessName: &sharedName,
	}); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	// The second signup with the same business name must be a conflict and must
	// not leave a user row behind for its email.
	_, err := a.Register(ctx, &commonauthv1.RegisterRequest{
		Email: secondEmail, Password: password, Name: "Second Owner", BusinessName: &sharedName,
	})
	var httpErr *onekithttp.Error
	if !errors.As(err, &httpErr) || httpErr.HTTPStatusCode() != http.StatusConflict {
		t.Fatalf("second Register error = %v, want a 409 conflict", err)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = ?", secondEmail).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("second Register created %d user row(s) despite failing", count)
	}

	// The address is still usable: a retry with a free business name succeeds.
	freeName := "Registration Free " + rand.Text()
	if _, err := a.Register(ctx, &commonauthv1.RegisterRequest{
		Email: secondEmail, Password: password, Name: "Second Owner", BusinessName: &freeName,
	}); err != nil {
		t.Fatalf("retry with a free business name: %v", err)
	}
}

func cleanupRegisteredUser(t *testing.T, db *sql.DB, email string) {
	t.Helper()
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, "DELETE FROM email_jobs WHERE to_email = ?", email); err != nil {
		t.Error(err)
	}
	var id int64
	err := db.QueryRowContext(ctx, "SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return
	}
	if err != nil {
		t.Error(err)
		return
	}
	for _, query := range []string{
		"DELETE FROM authorization_business_members WHERE user_id = ?",
		"DELETE FROM hr_employees WHERE user_id = ?",
		"DELETE FROM subscriptions WHERE owner_id = ?",
		"DELETE FROM businesses WHERE owner_id = ?",
		"DELETE FROM users WHERE id = ?",
	} {
		if _, err := db.ExecContext(ctx, query, id); err != nil {
			t.Error(err)
		}
	}
}
