package application

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

func TestGoogleAuthorizationURLContainsUniqueState(t *testing.T) {
	provider := NewGoogleOAuthProvider("client-id", "client-secret", "https://auth.example.com/callback")
	first, err := provider.GetAuthorizationURL()
	if err != nil {
		t.Fatalf("first URL: %v", err)
	}
	second, err := provider.GetAuthorizationURL()
	if err != nil {
		t.Fatalf("second URL: %v", err)
	}
	firstURL, err := url.Parse(first)
	if err != nil {
		t.Fatalf("parse first URL: %v", err)
	}
	secondURL, err := url.Parse(second)
	if err != nil {
		t.Fatalf("parse second URL: %v", err)
	}
	if firstURL.Host != "accounts.google.com" || firstURL.Query().Get("client_id") != "client-id" || firstURL.Query().Get("redirect_uri") != "https://auth.example.com/callback" {
		t.Fatalf("unexpected Google authorization URL: %s", firstURL)
	}
	firstState, secondState := firstURL.Query().Get("state"), secondURL.Query().Get("state")
	if len(firstState) < 32 || firstState == secondState {
		t.Fatalf("state values must be unpredictable and distinct: first=%q second=%q", firstState, secondState)
	}
}

func TestGoogleCodeExchangeReadsVerifiedIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			if err := r.ParseForm(); err != nil {
				t.Errorf("parse token request: %v", err)
			}
			if got := r.Form.Get("code"); got != "one-time-code" {
				t.Errorf("authorization code = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			})
		case "/userinfo":
			if got := r.Header.Get("Authorization"); got != "Bearer test-access-token" {
				t.Errorf("userinfo authorization = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sub":            "google-user-id",
				"email":          "person@example.com",
				"name":           "Google Person",
				"email_verified": true,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	provider := NewGoogleOAuthProvider("client-id", "client-secret", server.URL+"/callback")
	provider.config.Endpoint = oauth2.Endpoint{AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token"}
	provider.userInfoURL = server.URL + "/userinfo"
	user, err := provider.AuthenticateWithCode(context.Background(), "one-time-code")
	if err != nil {
		t.Fatalf("AuthenticateWithCode: %v", err)
	}
	if user.Email != "person@example.com" || user.Name != "Google Person" || !user.EmailVerified {
		t.Fatalf("Google user = %+v", user)
	}
}

func TestGoogleOAuthUnavailableClassifiesProviderAndClientErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "network error", err: context.DeadlineExceeded, want: true},
		{name: "provider failure", err: googleHTTPError{status: http.StatusBadGateway}, want: true},
		{name: "invalid code", err: googleHTTPError{status: http.StatusBadRequest}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := googleOAuthUnavailable(test.err); got != test.want {
				t.Fatalf("googleOAuthUnavailable() = %t, want %t", got, test.want)
			}
		})
	}
}
