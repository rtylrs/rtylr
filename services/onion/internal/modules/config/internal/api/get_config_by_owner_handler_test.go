package api

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestResolveConfigBusinessIDFailsClosedWithoutTenantMapping(t *testing.T) {
	_, sErr := resolveConfigBusinessID(
		context.Background(),
		&authPkg.AuthUser{UserID: 5, BusinessID: 42, Role: "owner"},
		&GetConfigByOwnerRequest{Owner: "not-a-number"},
	)
	if oneKitStatus(sErr) != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", oneKitStatus(sErr), http.StatusForbidden)
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
	if oneKitStatus(sErr) != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", oneKitStatus(sErr), http.StatusForbidden)
	}
}
