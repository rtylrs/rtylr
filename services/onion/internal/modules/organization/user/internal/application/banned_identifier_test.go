package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/platform/utils/passwordutils"
	"github.com/voxire/onion/internal/testutil"
)

// TestAuthenticateRejectsBannedUserWithActiveAccess covers B18-R01: console
// BanUser only flips users.is_banned and leaves the membership active, so
// the ban has to be carried by GetUserByIdentifier itself. When it was not, the
// identifier lookup returned IsBanned=false, the active membership passed the
// access check, and login issued a token for a banned account.
func TestAuthenticateRejectsBannedUserWithActiveAccess(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, jwtutils.JWTSecretKey, "test-secret")
	ctx = context.WithValue(ctx, jwtutils.JWTDurationKey, time.Hour)

	service := newTestService(db, "")
	if !service.Initialize(ctx) {
		t.Fatal("failed to initialize user service jwt helper")
	}

	password := "CorrectHorseBatteryStaple!"
	hash, err := passwordutils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	email := fmt.Sprintf("banned-identifier+%d@example.com", time.Now().UnixNano())
	userID, err := service.repo.CreateUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: hash,
		FullName:     "Banned Identifier",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	const businessID = int64(1)
	if _, err := db.ExecContext(ctx, `
INSERT INTO authorization_business_members (business_id, user_id, member_type, status)
VALUES (?, ?, 'member', 'active')`, businessID, userID); err != nil {
		t.Fatalf("insert membership: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM authorization_business_members WHERE user_id = ?", userID)
		_ = service.repo.DeleteUser(context.Background(), userID)
	})

	// The account authenticates while it is neither banned nor deactivated.
	if got, err := service.Authenticate(ctx, email, password); err != nil || got != userID {
		t.Fatalf("active account should authenticate, id=%d err=%v", got, err)
	}

	// BanUser flips only is_banned; the membership stays active.
	if _, err := db.ExecContext(ctx, "UPDATE users SET is_banned = TRUE WHERE id = ?", userID); err != nil {
		t.Fatalf("ban user: %v", err)
	}
	var active bool
	if err := db.QueryRowContext(ctx,
		"SELECT status = 'active' FROM authorization_business_members WHERE user_id = ?", userID).Scan(&active); err != nil {
		t.Fatalf("read membership state: %v", err)
	}
	if !active {
		t.Fatal("membership must stay active so the ban is the only rejection")
	}

	// The identifier lookup used by Authenticate must surface the ban.
	user, err := service.repo.GetUserByIdentifier(ctx, email)
	if err != nil {
		t.Fatalf("GetUserByIdentifier: %v", err)
	}
	if !user.IsBanned {
		t.Fatal("GetUserByIdentifier dropped is_banned")
	}

	if _, err := service.Authenticate(ctx, email, password); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("banned user login error = %v, want ErrInvalidCredentials", err)
	}
}
