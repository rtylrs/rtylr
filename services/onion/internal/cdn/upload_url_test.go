package cdn

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveUploadURLSource_CanonicalPathRedirectsToBaseURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("/files/abc-123?w=1", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.redirectURL != "https://cdn.voxire.com/files/abc-123?w=1" {
		t.Fatalf("redirectURL = %q", source.redirectURL)
	}
}

func TestResolveUploadURLSource_CanonicalPathWithoutSlashRedirectsToBaseURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("files/abc-123", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.redirectURL != "https://cdn.voxire.com/files/abc-123" {
		t.Fatalf("redirectURL = %q", source.redirectURL)
	}
}

func TestResolveUploadURLSource_FileUUIDPathRedirectsToWebPImage(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("/files/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.redirectURL != "https://cdn.voxire.com/images/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b?t=f_webp" {
		t.Fatalf("redirectURL = %q", source.redirectURL)
	}
}

func TestResolveUploadURLSource_FileUUIDPathCanRequestOriginal(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("/files/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b?format=original", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.redirectURL != "https://cdn.voxire.com/files/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b" {
		t.Fatalf("redirectURL = %q", source.redirectURL)
	}
}

func TestResolveUploadURLSource_LegacyCDNPathBecomesStorageKey(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("https://cdn.voxire.dev/web_content/rtylr/logo.webp", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.storageKey != "web_content/rtylr/logo.webp" {
		t.Fatalf("storageKey = %q", source.storageKey)
	}
}

func TestResolveUploadURLSource_PublicRemoteURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("https://example.com/image.png", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.remoteURL == nil || source.remoteURL.String() != "https://example.com/image.png" {
		t.Fatalf("remoteURL = %#v", source.remoteURL)
	}
}

func TestResolveUploadURLSource_RejectsPrivateRemoteURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	if _, err := c.resolveUploadURLSource("http://127.0.0.1/image.png", ""); err == nil {
		t.Fatal("expected private remote URL to be rejected")
	}
}

func TestResolveUploadURLSource_RejectsNestedUploadURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	if _, err := c.resolveUploadURLSource("https://cdn.voxire.com/upload?url=https://example.com/x.png", ""); err == nil {
		t.Fatal("expected nested upload URL to be rejected")
	}
}

func TestResolveUploadURLSource_BareIDRedirectsToWebPImage(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.redirectURL != "https://cdn.voxire.com/images/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b?t=f_webp" {
		t.Fatalf("redirectURL = %q", source.redirectURL)
	}
}

func TestUploadURLWantsOriginal(t *testing.T) {
	if !uploadURLWantsOriginal(mapValues("format", "original")) {
		t.Fatal("expected format=original to request passthrough")
	}
	if !uploadURLWantsOriginal(mapValues("format", "source")) {
		t.Fatal("expected format=source to request passthrough")
	}
	if uploadURLWantsOriginal(mapValues("format", "webp")) {
		t.Fatal("did not expect format=webp to request passthrough")
	}
}

func TestUploadURLWebPCacheKeyUsesSourceHash(t *testing.T) {
	a := uploadURLWebPCacheKey("remote:https://example.com/a.png", "hash-a")
	b := uploadURLWebPCacheKey("remote:https://example.com/a.png", "hash-b")
	if a == b {
		t.Fatal("expected source hash to affect upload-url webp cache key")
	}
}

func TestStageUploadURLImageRejectsOversizedSource(t *testing.T) {
	_, err := stageUploadURLImage(strings.NewReader("abc"), 2)
	if !errors.Is(err, errUploadURLImageTooLarge) {
		t.Fatalf("expected oversized error, got %v", err)
	}
}

func TestServeUploadURLWebPUsesMediaEngineOutput(t *testing.T) {
	engine := fakeUploadURLMediaEngine(t)
	c := &CDN{
		maxUpload:   1 << 20,
		mediaEngine: engine,
		engineSem:   make(chan struct{}, 1),
	}

	req := httptest.NewRequest(http.MethodGet, "/upload?url=https%3A%2F%2Fexample.com%2Fx.png", nil)
	rec := httptest.NewRecorder()

	c.serveUploadURLWebP(rec, req, uploadURLImageSource{
		body:      bytes.NewReader([]byte("source png bytes")),
		mimeType:  "image/png",
		cacheSeed: "remote:https://example.com/x.png",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "image/webp" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Body.String(); got != "fake-webp" {
		t.Fatalf("body = %q", got)
	}
}

func fakeUploadURLMediaEngine(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "media-engine")
	script := `#!/bin/sh
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--output" ]; then
    shift
    out="$1"
  fi
 shift
done
printf fake-webp > "$out"
`
	if err := os.WriteFile(path, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	// The stub is exec'd by serveUploadURLWebP, so it needs the executable bit.
	// Set it via Chmod rather than WriteFile to keep gosec's G306 satisfied.
	if err := os.Chmod(path, 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func mapValues(key, value string) map[string][]string {
	return map[string][]string{key: {value}}
}
