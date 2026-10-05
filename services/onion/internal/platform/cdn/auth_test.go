package cdn

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/voxire/onion/internal/modules/organization/scopes"
	"github.com/voxire/onion/internal/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyAuth(t *testing.T) {
	// next writes 418 so a test can tell "the request passed the middleware"
	// apart from a 401 produced by the middleware itself.
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	t.Run("dev-open when no shared key is configured", func(t *testing.T) {
		t.Setenv("APP_ENV", "development")
		c := &CDN{apiKey: ""}
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, httptest.NewRequest("POST", "/upload", nil))
		if rec.Code != http.StatusTeapot {
			t.Errorf("expected pass-through, got %d", rec.Code)
		}
	})

	t.Run("accepts the shared key", func(t *testing.T) {
		c := &CDN{apiKey: "s3cret"}
		req := httptest.NewRequest("POST", "/upload", nil)
		req.Header.Set("Authorization", "Bearer s3cret")
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, req)
		if rec.Code != http.StatusTeapot {
			t.Errorf("expected accept, got %d", rec.Code)
		}
	})

	t.Run("rejects a missing token", func(t *testing.T) {
		c := &CDN{apiKey: "s3cret"}
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, httptest.NewRequest("POST", "/upload", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("fails closed without a key outside development", func(t *testing.T) {
		t.Setenv("APP_ENV", "production")
		c := &CDN{apiKey: ""}
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, httptest.NewRequest("POST", "/upload", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", rec.Code)
		}
	})
}

func TestValidateSecurityConfig(t *testing.T) {
	if err := ValidateSecurityConfig("production", ""); err == nil {
		t.Fatal("expected production without CDN_API_KEY to fail")
	}
	if err := ValidateSecurityConfig("development", ""); err != nil {
		t.Fatalf("development without CDN_API_KEY should be allowed: %v", err)
	}
}

func TestPersonalUploadCredentialRequiresCurrentScopeAndExpiry(t *testing.T) {
	db := testutil.OpenDB(t)
	const businessID, userID = int64(9968201), int64(9968202)
	testutil.GrantBusinessAccess(t, db, businessID, userID)
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM users WHERE id = ?", userID) })
	const token = "synthetic-upload-credential"
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	res, err := db.Exec("INSERT INTO hub_api_keys (owner_id, business_id, name, key_hash, key_prefix, scopes) VALUES (?, ?, 'Synthetic upload', ?, 'test', JSON_ARRAY('upload'))", userID, businessID, hash)
	if err != nil {
		t.Fatal(err)
	}
	keyID, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec("DELETE FROM hub_api_keys WHERE id = ?", keyID) })
	c := (&CDN{db: db, apiKey: "service-key"}).WithUploadKeyReader(scopes.NewReader(db).UploadKeyScope)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := principalFrom(r.Context())
		if !ok || p.businessID != businessID {
			t.Fatal("upload credential lost its business scope")
		}
		w.WriteHeader(http.StatusTeapot)
	})
	check := func(want int) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/upload", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		c.apiKeyAuth(next).ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("upload status=%d, want %d: %s", rec.Code, want, rec.Body.String())
		}
	}
	check(http.StatusTeapot)
	for _, mutation := range []string{"scopes = JSON_ARRAY('read')", "scopes = JSON_ARRAY('upload'), expires_at = CURRENT_TIMESTAMP - INTERVAL 1 HOUR", "expires_at = NULL, is_active = FALSE"} {
		if _, err := db.ExecContext(context.Background(), "UPDATE hub_api_keys SET "+mutation+" WHERE id = ?", keyID); err != nil {
			t.Fatal(err)
		}
		check(http.StatusUnauthorized)
	}
	if _, err := db.Exec("UPDATE hub_api_keys SET is_active = TRUE WHERE id = ?", keyID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("UPDATE businesses SET is_active = FALSE WHERE id = ?", businessID); err != nil {
		t.Fatal(err)
	}
	check(http.StatusForbidden)
}
