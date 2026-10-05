package cdn

import "testing"

func TestFileURL(t *testing.T) {
	c := &CDN{baseURL: "https://cdn.voxire.dev"}
	got := c.fileURL("abc-123")
	want := "https://cdn.voxire.dev/files/abc-123"
	if got != want {
		t.Errorf("fileURL = %q, want %q", got, want)
	}
}
