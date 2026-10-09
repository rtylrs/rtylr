package cdn

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/testutil"
)

// TestPurgeUploadURLErrorStatus pins the URL purge status mapping: a missing
// file is a 404 like the by-id routes, caller-fixable input is a 400, and
// infrastructure failures are a generic 500 that never echoes the raw error.
func TestPurgeUploadURLErrorStatus(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		want     int
		wantBody string
	}{
		{
			name:     "missing file",
			err:      sql.ErrNoRows,
			want:     http.StatusNotFound,
			wantBody: "file not found",
		},
		{
			name:     "bad input",
			err:      fmt.Errorf("%w: invalid url", errBadPurgeInput),
			want:     http.StatusBadRequest,
			wantBody: "invalid purge request",
		},
		{
			name:     "storage failure",
			err:      errors.New("deleting prefix on bucket onion-media: access denied"),
			want:     http.StatusInternalServerError,
			wantBody: "purge failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/purge", nil)
			writePurgeUploadURLError(rec, req, tt.err)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
			body := strings.TrimSpace(rec.Body.String())
			if body != tt.wantBody {
				t.Fatalf("body = %q, want %q", body, tt.wantBody)
			}
			if strings.Contains(body, "bucket") || strings.Contains(body, "access denied") {
				t.Fatalf("response leaked internal error detail: %q", body)
			}
		})
	}
}

func TestPurgeUploadURLRejectsNestedUploadURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}
	target := "https://cdn.voxire.com/upload?url=" + url.QueryEscape("https://example.com/a.png")
	req := httptest.NewRequest("PURGE", "/upload?url="+url.QueryEscape(target), nil)
	rec := httptest.NewRecorder()

	c.purgeUploadURL(rec, asService(req))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "invalid purge request" {
		t.Fatalf("body = %q, want %q", body, "invalid purge request")
	}
}

// A purge with no principal at all is refused before any target is inspected.
func TestPurgeRequiresAPrincipal(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com", db: testutil.OpenDB(t)}
	req := httptest.NewRequest("PURGE", "/files/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b", nil)
	rec := httptest.NewRecorder()

	c.purgeFileVariants(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

// TestPurgeUploadURLMissingFileIsNotFound covers the URL form of the by-id
// 404: a canonical CDN URL whose file UUID has no cdn_files row must not be
// answered with 400 "sql: no rows in result set".
func TestPurgeUploadURLMissingFileIsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	c := &CDN{db: db, baseURL: "https://cdn.voxire.com"}

	target := "https://cdn.voxire.com/files/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b"
	req := httptest.NewRequest("PURGE", "/upload?url="+url.QueryEscape(target), nil)
	rec := httptest.NewRecorder()

	c.purgeUploadURL(rec, asService(req))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d (body %q)", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "file not found" {
		t.Fatalf("body = %q, want %q", body, "file not found")
	}
}
