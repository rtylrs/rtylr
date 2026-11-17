package github

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

var fixedNow = time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)

func testKey(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return key, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
}

func testClient(t *testing.T, handler http.Handler) (*Client, *rsa.PrivateKey) {
	t.Helper()
	key, keyPEM := testKey(t)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c, err := NewClient(Config{
		AppID: "4242", Slug: "rtylr-flow", PrivateKeyPEM: keyPEM,
		APIURL: server.URL, WebURL: "https://github.example",
		HTTPClient: server.Client(), Now: func() time.Time { return fixedNow },
	})
	if err != nil {
		t.Fatal(err)
	}
	return c, key
}

func assertAppJWT(t *testing.T, r *http.Request, key *rsa.PrivateKey) {
	t.Helper()
	raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(tok *jwt.Token) (any, error) {
		if tok.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("alg %v", tok.Header["alg"])
		}
		return &key.PublicKey, nil
	}, jwt.WithTimeFunc(func() time.Time { return fixedNow }))
	if err != nil {
		t.Fatalf("app JWT does not verify: %v", err)
	}
	if claims.Issuer != "4242" {
		t.Errorf("iss = %q, want app id", claims.Issuer)
	}
	if claims.ExpiresAt.Sub(claims.IssuedAt.Time) > 10*time.Minute {
		t.Errorf("JWT lifetime %v exceeds GitHub's ten minutes", claims.ExpiresAt.Sub(claims.IssuedAt.Time))
	}
}

func TestAppJWTIsRS256SignedWithTheAppKey(t *testing.T) {
	c, key := testClient(t, http.NotFoundHandler())
	token, err := c.AppJWT()
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	assertAppJWT(t, req, key)

	_, otherPEM := testKey(t)
	other, _ := NewClient(Config{AppID: "4242", Slug: "s", PrivateKeyPEM: otherPEM, Now: func() time.Time { return fixedNow }})
	forged, _ := other.AppJWT()
	if _, err := jwt.Parse(forged, func(*jwt.Token) (any, error) { return &key.PublicKey, nil }); err == nil {
		t.Fatal("a JWT signed with another key verified")
	}
}

func TestNewClientRejectsABadKey(t *testing.T) {
	if _, err := NewClient(Config{AppID: "1", Slug: "s", PrivateKeyPEM: []byte("not a key")}); err == nil {
		t.Fatal("expected an error for an unparseable key")
	}
}

func TestInstallURLCarriesTheState(t *testing.T) {
	c, _ := testClient(t, http.NotFoundHandler())
	got := c.InstallURL("1.2.3.ab.sig+/=")
	want := "https://github.example/apps/rtylr-flow/installations/new?state=1.2.3.ab.sig%2B%2F%3D"
	if got != want {
		t.Fatalf("InstallURL = %q, want %q", got, want)
	}
}

func TestGetInstallationMapsHealth(t *testing.T) {
	var key *rsa.PrivateKey
	mux := http.NewServeMux()
	mux.HandleFunc("GET /app/installations/7", func(w http.ResponseWriter, r *http.Request) {
		assertAppJWT(t, r, key)
		_, _ = w.Write([]byte(`{"id":7,"account":{"login":"acme","type":"Organization"},"created_at":"2026-11-17T11:59:00Z","suspended_at":"2026-11-17T12:00:00Z"}`))
	})
	c, k := testClient(t, mux)
	key = k

	info, err := c.GetInstallation(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if info.AccountLogin != "acme" || info.AccountType != "Organization" || info.SuspendedAt == nil {
		t.Fatalf("info = %+v", info)
	}
	if _, err := c.GetInstallation(context.Background(), 8); !errors.Is(err, domain.ErrGitHubInstallationNotFound) {
		t.Fatalf("missing installation err = %v, want ErrGitHubInstallationNotFound", err)
	}
}

func TestListInstallationRepositoriesUsesAnInstallationTokenAndPages(t *testing.T) {
	var key *rsa.PrivateKey
	mux := http.NewServeMux()
	mux.HandleFunc("POST /app/installations/7/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		assertAppJWT(t, r, key)
		_, _ = w.Write([]byte(`{"token":"ghs_installation","expires_at":"2026-11-17T13:00:00Z"}`))
	})
	mux.HandleFunc("GET /installation/repositories", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ghs_installation" {
			t.Errorf("listing used %q, want the installation token", r.Header.Get("Authorization"))
		}
		type repo struct {
			ID       int64  `json:"id"`
			FullName string `json:"full_name"`
			Private  bool   `json:"private"`
		}
		var page []repo
		switch r.URL.Query().Get("page") {
		case "1":
			for i := 0; i < perPage; i++ {
				page = append(page, repo{ID: int64(i + 1), FullName: fmt.Sprintf("acme/r%d", i+1)})
			}
		case "2":
			page = []repo{{ID: 101, FullName: "acme/secret", Private: true}}
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"total_count": perPage + 1, "repositories": page})
	})
	c, k := testClient(t, mux)
	key = k

	repos, err := c.ListInstallationRepositories(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != perPage+1 {
		t.Fatalf("got %d repositories, want %d", len(repos), perPage+1)
	}
	if last := repos[len(repos)-1]; last.ID != 101 || last.FullName != "acme/secret" || !last.Private {
		t.Fatalf("last repository = %+v", last)
	}
}

func TestUpstreamFailuresAreUnavailable(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /app/installations/7", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("DELETE /app/installations/9", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	c, _ := testClient(t, mux)
	if err := c.DeleteInstallation(context.Background(), 7); !errors.Is(err, domain.ErrGitHubUnavailable) {
		t.Fatalf("500 err = %v, want ErrGitHubUnavailable", err)
	}
	if err := c.DeleteInstallation(context.Background(), 9); err != nil {
		t.Fatalf("204 err = %v", err)
	}
}
