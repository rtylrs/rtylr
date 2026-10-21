package cdn

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// multipartUpload builds a multipart/form-data body with one part.
func multipartUpload(t *testing.T, field, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}

// asTenant attaches the resolved-tenant principal apiKeyAuth would have
// attached, so handler tests exercise the authenticated path directly.
func asTenant(r *http.Request, businessID int64) *http.Request {
	return r.WithContext(withPrincipal(r.Context(), principal{userID: 1, businessID: businessID}))
}

// asService attaches the shared CDN_API_KEY principal.
func asService(r *http.Request) *http.Request {
	return r.WithContext(withPrincipal(r.Context(), principal{service: true}))
}

func TestUpload_MissingFileField(t *testing.T) {
	c := &CDN{maxUpload: 1 << 20}
	body, ct := multipartUpload(t, "notfile", "x.png", []byte("data"))
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	c.upload(rec, asTenant(req, 4242))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing file field, got %d", rec.Code)
	}
}

func TestUpload_EmptyFileRejected(t *testing.T) {
	c := &CDN{maxUpload: 1 << 20}
	body, ct := multipartUpload(t, "file", "x.png", nil)
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	c.upload(rec, asTenant(req, 4242))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty file, got %d", rec.Code)
	}
}

// Uploads are the one write that cannot be recorded unattributed, so a
// credential that resolves to no business is refused before any byte is read.
func TestUpload_RequiresTenantScope(t *testing.T) {
	c := &CDN{maxUpload: 1 << 20}
	body, ct := multipartUpload(t, "file", "x.png", []byte("data"))
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()

	c.upload(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without a tenant, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/upload", nil)
	c.upload(rec, asService(req))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for the unscoped service credential, got %d: %s", rec.Code, rec.Body.String())
	}
}

// Visibility is opt-in: an upload that does not ask for public delivery must
// not become anonymously readable.
func TestUploadVisibilityDefaultsPrivate(t *testing.T) {
	req := httptest.NewRequest("POST", "/upload", nil)
	if got := uploadVisibility(req); got != visibilityPrivate {
		t.Fatalf("default visibility = %q, want %q", got, visibilityPrivate)
	}
	if visibilityLabel(false) != visibilityPrivate {
		t.Fatal("private label mismatch")
	}

	req = httptest.NewRequest("POST", "/upload", strings.NewReader("visibility=public"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := uploadVisibility(req); got != visibilityPublic {
		t.Fatalf("explicit visibility = %q, want %q", got, visibilityPublic)
	}
	if visibilityLabel(true) != visibilityPublic {
		t.Fatal("public label mismatch")
	}
}
