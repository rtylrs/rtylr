package cdn

import "testing"

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

func TestResolveUploadURLSource_BareIDRedirectsToUploadedFile(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.com"}

	source, err := c.resolveUploadURLSource("4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b", "")
	if err != nil {
		t.Fatal(err)
	}

	if source.redirectURL != "https://cdn.voxire.com/files/4c85c397-9fc3-4bd8-adf5-9d1ef64f5e8b" {
		t.Fatalf("redirectURL = %q", source.redirectURL)
	}
}
