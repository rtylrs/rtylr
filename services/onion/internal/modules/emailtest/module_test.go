package emailtest

import (
	"context"
	stdErrors "errors"
	"net/http"
	"os"
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

// The regression this guards: the endpoints render real branded templates
// through the platform's sending domain and post them to a caller-supplied
// address, so an authenticated user reaching them is a phishing primitive. Only
// an explicitly declared development environment may.
func TestEmailTestHandlersRefuseAnythingButDeclaredDevelopment(t *testing.T) {
	cases := []struct {
		name string
		env  string
	}{
		{name: "unset", env: ""},
		{name: "whitespace", env: "   "},
		{name: "nonsense", env: "banana"},
		{name: "misspelled development", env: "developement"},
		{name: "padded production", env: " production "},
		{name: "padded prod", env: " prod "},
		{name: "production-eu", env: "production-eu"},
		{name: "staging", env: "staging"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withTestEnv(t, tc.env)

			ctx := authenticatedContext()
			if _, err := (&ListTemplatesHandler{}).Serve(ctx, &ListTemplatesRequest{}); err == nil || oneKitStatus(err) != http.StatusNotFound {
				t.Errorf("APP_ENV=%q: expected template listing to be unavailable, got %+v", tc.env, err)
			}
			if _, err := (&TestEmailHandler{}).Serve(ctx, &TestEmailRequest{}); err == nil || oneKitStatus(err) != http.StatusNotFound {
				t.Errorf("APP_ENV=%q: expected send to be unavailable, got %+v", tc.env, err)
			}
		})
	}
}

// An unset APP_ENV is the state the exploit used: os.Getenv returns "" and the
// old equality check answered "not production" for it.
func TestEmailTestHandlersRefuseWhenAppEnvIsUnset(t *testing.T) {
	previous := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = previous })
	t.Setenv("APP_ENV", "placeholder")
	if err := os.Unsetenv("APP_ENV"); err != nil {
		t.Fatalf("unset APP_ENV: %v", err)
	}

	ctx := authenticatedContext()
	if _, err := (&TestEmailHandler{}).Serve(ctx, &TestEmailRequest{
		Template: "auth/password_reset",
		To:       "ceo@target-bank.example",
		Subject:  "Action required",
		Data:     map[string]any{"ResetLink": "https://attacker.example/reset?t=..."},
	}); err == nil || oneKitStatus(err) != http.StatusNotFound {
		t.Fatalf("expected send to be unavailable with APP_ENV unset, got %+v", err)
	}
}

// The declared development labels are the only ones that open the endpoints;
// the allowlist must keep covering every value the repo actually ships.
func TestEmailTestHandlersRemainAvailableInDeclaredDevelopment(t *testing.T) {
	for _, env := range []string{"dev", "development", "local", "test", "development ", "DEVELOPMENT"} {
		t.Run(env, func(t *testing.T) {
			withTestEnv(t, env)

			// Authenticated and not production: the request must get past the
			// environment gate and reach the ordinary per-request validation.
			_, err := (&ListTemplatesHandler{}).Serve(authenticatedContext(), &ListTemplatesRequest{})
			if err != nil {
				t.Fatalf("APP_ENV=%q: expected listing to be served, got %+v", env, err)
			}
		})
	}
}

func TestEmailTestAnonymousRoutesStayEmpty(t *testing.T) {
	if got := (&Module{}).GetAnonymousRoutes(); len(got) != 0 {
		t.Fatalf("expected no anonymous email test routes, got %v", got)
	}
}

func authenticatedContext() context.Context {
	device := authPkg.Device{IsAnonymous: false, User: &authPkg.AuthUser{UserID: 1}}
	return device.Ctx(context.Background())
}

func withTestEnv(t *testing.T, env string) {
	t.Helper()
	old := config.GlobalConfig
	config.GlobalConfig = nil
	t.Cleanup(func() { config.GlobalConfig = old })
	t.Setenv("APP_ENV", env)
}
