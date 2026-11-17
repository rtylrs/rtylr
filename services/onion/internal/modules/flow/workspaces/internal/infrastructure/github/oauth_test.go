package github

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func oauthClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	_, keyPEM := testKey(t)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c, err := NewClient(Config{
		AppID: "4242", Slug: "rtylr-flow", PrivateKeyPEM: keyPEM,
		ClientID: "Iv1.testclient", ClientSecret: "test-client-secret",
		APIURL: server.URL, WebURL: server.URL,
		HTTPClient: server.Client(), Now: func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLinkingNeedsBothClientCredentials(t *testing.T) {
	_, keyPEM := testKey(t)
	c, err := NewClient(Config{AppID: "1", Slug: "s", PrivateKeyPEM: keyPEM, ClientID: "Iv1.only-id"})
	if err != nil {
		t.Fatal(err)
	}
	if c.LinkingEnabled() {
		t.Fatal("linking enabled without a client secret")
	}
}

func TestAuthorizeURLCarriesStateAndPKCEChallenge(t *testing.T) {
	c := oauthClient(t, http.NotFoundHandler())
	raw := c.AuthorizeURL("link.1.2.3.ab.sig", "challenge-value")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Path != "/login/oauth/authorize" || q.Get("client_id") != "Iv1.testclient" ||
		q.Get("state") != "link.1.2.3.ab.sig" || q.Get("code_challenge") != "challenge-value" ||
		q.Get("code_challenge_method") != "S256" {
		t.Fatalf("authorize URL = %s", raw)
	}
	if q.Has("client_secret") || q.Has("redirect_uri") {
		t.Fatalf("authorize URL leaks the secret or pins a redirect: %s", raw)
	}
}

func TestExchangeCodePostsTheVerifierAndReadsTheToken(t *testing.T) {
	c := oauthClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/login/oauth/access_token" {
			t.Errorf("request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		_ = r.ParseForm()
		if r.PostForm.Get("client_id") != "Iv1.testclient" || r.PostForm.Get("client_secret") != "test-client-secret" ||
			r.PostForm.Get("code") != "code-1" || r.PostForm.Get("code_verifier") != "verifier-1" {
			t.Errorf("form = %v", r.PostForm)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"ghu_test","token_type":"bearer","scope":""}`))
	}))
	token, err := c.ExchangeCode(context.Background(), "code-1", "verifier-1")
	if err != nil || token != "ghu_test" {
		t.Fatalf("ExchangeCode = %q, %v", token, err)
	}
}

func TestExchangeCodeMapsGitHubErrors(t *testing.T) {
	// GitHub reports a bad or expired code as 200 with an error field.
	c := oauthClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":"bad_verification_code","error_description":"The code passed is incorrect or expired."}`))
	}))
	if _, err := c.ExchangeCode(context.Background(), "stale", "v"); !errors.Is(err, domain.ErrGitHubAuthorizationFailed) {
		t.Fatalf("bad code err = %v, want ErrGitHubAuthorizationFailed", err)
	}
	down := oauthClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	if _, err := down.ExchangeCode(context.Background(), "c", "v"); !errors.Is(err, domain.ErrGitHubUnavailable) {
		t.Fatalf("502 err = %v, want ErrGitHubUnavailable", err)
	}
}

func TestGetUserUsesTheUserToken(t *testing.T) {
	c := oauthClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer ghu_test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"id":583231,"login":"octocat","avatar_url":"https://avatars.example/u/583231"}`))
	}))
	user, err := c.GetUser(context.Background(), "ghu_test")
	if err != nil || user.ID != 583231 || user.Login != "octocat" || !strings.HasSuffix(user.AvatarURL, "/583231") {
		t.Fatalf("GetUser = %+v, %v", user, err)
	}
	if _, err := c.GetUser(context.Background(), "ghu_revoked"); !errors.Is(err, domain.ErrGitHubAuthorizationFailed) {
		t.Fatalf("revoked token err = %v, want ErrGitHubAuthorizationFailed", err)
	}
}
