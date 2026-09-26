package emailtest

import (
	"context"
	stdErrors "errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/config"
)

func oneKitStatus(err error) int {
	var oneKitErr *onekithttp.Error
	if stdErrors.As(err, &oneKitErr) {
		return int(oneKitErr.GetStatusCode())
	}
	return 0
}

func TestEmailTestHandlersRejectAnonymous(t *testing.T) {
	withTestEnv(t, "development")

	_, err := (&ListTemplatesHandler{}).Serve(context.Background(), &ListTemplatesRequest{})
	if err == nil || oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("expected anonymous list to be forbidden, got %+v", err)
	}

	_, err = (&TestEmailHandler{}).Serve(context.Background(), &TestEmailRequest{})
	if err == nil || oneKitStatus(err) != http.StatusForbidden {
		t.Fatalf("expected anonymous send to be forbidden, got %+v", err)
	}
}

func TestEmailTestHandlersAreUnavailableInProduction(t *testing.T) {
	withTestEnv(t, "production")

	ctx := authenticatedContext()
	_, err := (&ListTemplatesHandler{}).Serve(ctx, &ListTemplatesRequest{})
	if err == nil || oneKitStatus(err) != http.StatusNotFound {
		t.Fatalf("expected production list to be unavailable, got %+v", err)
	}

	_, err = (&TestEmailHandler{}).Serve(ctx, &TestEmailRequest{})
	if err == nil || oneKitStatus(err) != http.StatusNotFound {
		t.Fatalf("expected production send to be unavailable, got %+v", err)
	}
}

func TestEmailTestAnonymousRoutesStayEmpty(t *testing.T) {
	if got := (&Module{}).GetAnonymousRoutes(); len(got) != 0 {
		t.Fatalf("expected no anonymous email test routes, got %v", got)
	}
}

func authenticatedContext() context.Context {
	device := authPkg.Device{IsAnonymous: false, User: &authPkg.AuthUser{UserID: 1, Role: "admin"}}
	return device.Ctx(context.Background())
}

func withTestEnv(t *testing.T, env string) {
	t.Helper()
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", env)
}
