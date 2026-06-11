package handlers

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func TestAuthenticatedUserReadsAuthContext(t *testing.T) {
	want := &authPkg.AuthUser{UserID: 42}
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, want)

	got, ok := authenticatedUser(ctx)
	if !ok {
		t.Fatal("expected authenticated user")
	}
	if got.UserID != want.UserID {
		t.Fatalf("UserID = %d, want %d", got.UserID, want.UserID)
	}
}

func TestAuthenticatedUserFallsBackToDeviceContext(t *testing.T) {
	want := &authPkg.AuthUser{UserID: 42}
	device := authPkg.Device{User: want}
	ctx := device.Ctx(context.Background())

	got, ok := authenticatedUser(ctx)
	if !ok {
		t.Fatal("expected authenticated user")
	}
	if got.UserID != want.UserID {
		t.Fatalf("UserID = %d, want %d", got.UserID, want.UserID)
	}
}

func TestAuthenticatedUserRejectsAnonymousContext(t *testing.T) {
	if user, ok := authenticatedUser(context.Background()); ok {
		t.Fatalf("expected anonymous context, got user %+v", user)
	}
}
