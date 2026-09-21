package application

import (
	"context"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/user/internal/domain"
	"github.com/voxire/onion/internal/platform/utils/jwtutils"
	"github.com/voxire/onion/internal/testutil"
)

// Google signups used to store the literal "invalid IP" — a zero netip.Addr's
// String() — and no device info at all. The handler now passes the request's
// device context and the application layer writes the parsed address, leaving
// the column NULL when the address is unusable.
func TestLinkGoogleUser_RecordsSignupContext(t *testing.T) {
	db := testutil.OpenDB(t)

	ctx := context.Background()
	ctx = context.WithValue(ctx, jwtutils.JWTSecretKey, "test-secret")
	ctx = context.WithValue(ctx, jwtutils.JWTDurationKey, time.Hour)

	service := NewService(ctx, db, "")
	if !service.Initialize(ctx) {
		t.Fatal("failed to initialize user service jwt helper")
	}

	newSignup := func(signupIP netip.Addr, deviceInfo string) *domain.User {
		t.Helper()

		email := fmt.Sprintf("google-signup+%d@example.com", time.Now().UnixNano())
		result, err := service.linkGoogleUser(ctx, &GoogleUser{
			Email: email, Name: "Google Signup", EmailVerified: true,
		}, "", signupIP, deviceInfo)
		if err != nil {
			t.Fatalf("linkGoogleUser: %v", err)
		}
		t.Cleanup(func() {
			_, _ = db.ExecContext(context.Background(), "DELETE FROM subscriptions WHERE owner_id = ?", result.UserID)
			_ = service.repo.DeleteUser(context.Background(), result.UserID)
		})

		stored, err := service.repo.GetUserByEmail(ctx, email)
		if err != nil {
			t.Fatalf("GetUserByEmail: %v", err)
		}
		return stored
	}

	stored := newSignup(netip.MustParseAddr("203.0.113.7"), "desktop")
	if stored.SignupIPAddress != "203.0.113.7" {
		t.Errorf("SignupIPAddress = %q, want the parsed client address", stored.SignupIPAddress)
	}
	if stored.SignupDeviceInfo != "desktop" {
		t.Errorf("SignupDeviceInfo = %q, want %q", stored.SignupDeviceInfo, "desktop")
	}

	// An unparsable client address must leave the column empty, not store the
	// zero Addr's "invalid IP".
	stored = newSignup(netip.Addr{}, "")
	if stored.SignupIPAddress != "" {
		t.Errorf("SignupIPAddress = %q, want empty for an unusable address", stored.SignupIPAddress)
	}
}
