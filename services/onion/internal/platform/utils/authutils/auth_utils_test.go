package authutils

import (
	"context"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func TestGetAuthUserFromContext_Empty(t *testing.T) {
	if got := GetAuthUserFromContext(context.Background()); got != nil {
		t.Errorf("expected nil for empty context, got %+v", got)
	}
}

func TestGetAuthUserFromContext_Populated(t *testing.T) {
	user := &authPkg.AuthUser{UserID: 42, BusinessID: 7, Role: "admin"}
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, user)

	got := GetAuthUserFromContext(ctx)
	if got == nil || got.UserID != 42 || got.BusinessID != 7 || got.Role != "admin" {
		t.Errorf("expected user 42/7/admin, got %+v", got)
	}
}

func TestIsAuthenticated(t *testing.T) {
	if IsAuthenticated(context.Background()) {
		t.Error("empty context should report unauthenticated")
	}

	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, &authPkg.AuthUser{UserID: 1})
	if !IsAuthenticated(ctx) {
		t.Error("populated context should report authenticated")
	}
}

func TestIsAuthenticated_WrongTypeIsNotAuthenticated(t *testing.T) {
	// Some other middleware planting a bogus value should not authenticate.
	ctx := context.WithValue(context.Background(), authPkg.AuthUserKey, "not-a-user")
	if IsAuthenticated(ctx) {
		t.Error("non-AuthUser value should not authenticate")
	}
}
