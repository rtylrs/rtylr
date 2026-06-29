package handlers

import (
	"context"
	"net/http"
	"testing"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func TestResolveConfigBusinessID_UsesTokenBusinessID(t *testing.T) {
	got, sErr := resolveConfigBusinessID(
		context.Background(),
		&authPkg.AuthUser{UserID: 5, BusinessID: 42, Role: "owner"},
		&GetConfigByOwnerRequest{Owner: "not-a-number"},
	)
	if sErr != nil {
		t.Fatalf("resolveConfigBusinessID returned error: %v", sErr)
	}
	if got != 42 {
		t.Fatalf("business id = %d, want 42", got)
	}
}

func TestResolveConfigBusinessID_AllowsConsoleAdminRequestedBusiness(t *testing.T) {
	got, sErr := resolveConfigBusinessID(
		context.Background(),
		&authPkg.AuthUser{UserID: 1, Role: "console_admin"},
		&GetConfigByOwnerRequest{Owner: "7"},
	)
	if sErr != nil {
		t.Fatalf("resolveConfigBusinessID returned error: %v", sErr)
	}
	if got != 7 {
		t.Fatalf("business id = %d, want 7", got)
	}
}

func TestResolveConfigBusinessID_DeniesInvalidRequestedBusiness(t *testing.T) {
	_, sErr := resolveConfigBusinessID(
		context.Background(),
		&authPkg.AuthUser{UserID: 1, Role: "console_admin"},
		&GetConfigByOwnerRequest{Owner: "nope"},
	)
	if sErr == nil {
		t.Fatal("expected permission error")
	}
	if sErr.HTTPCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", sErr.HTTPCode, http.StatusForbidden)
	}
}
