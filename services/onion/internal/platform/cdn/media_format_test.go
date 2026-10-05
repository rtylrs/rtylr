package cdn

import "testing"

func TestExtensionForFormat(t *testing.T) {
	cases := map[string]string{
		"jpeg": ".jpg",
		"jpg":  ".jpg",
		"JPG":  ".jpg",
		"png":  ".png",
		"webp": ".webp",
		"avif": ".avif",
		"gif":  ".gif",
	}
	for in, want := range cases {
		if got := extensionForFormat(in); got != want {
			t.Errorf("extensionForFormat(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtensionForFormat_DefaultBranch(t *testing.T) {
	// Formats without an explicit case fall through to "." + format. tiff and
	// mov reach here (they have no dedicated case) and must still be sensible.
	cases := map[string]string{
		"tiff":    ".tiff",
		"mov":     ".mov",
		"unknown": ".unknown",
	}
	for in, want := range cases {
		if got := extensionForFormat(in); got != want {
			t.Errorf("extensionForFormat(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMimeForFormat(t *testing.T) {
	cases := map[string]string{
		"jpeg":    "image/jpeg",
		"jpg":     "image/jpeg",
		"png":     "image/png",
		"webp":    "image/webp",
		"avif":    "image/avif",
		"gif":     "image/gif",
		"unknown": "application/octet-stream",
	}
	for in, want := range cases {
		if got := mimeForFormat(in); got != want {
			t.Errorf("mimeForFormat(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMimeTypeToFormat(t *testing.T) {
	cases := map[string]string{
		"image/jpeg":      "jpg",
		"image/png":       "png",
		"image/webp":      "webp",
		"image/gif":       "gif",
		"image/avif":      "avif",
		"image/tiff":      "tiff",
		"video/mp4":       "mp4",
		"video/webm":      "webm",
		"video/quicktime": "mov",
	}
	for in, want := range cases {
		if got := mimeTypeToFormat(in); got != want {
			t.Errorf("mimeTypeToFormat(%q) = %q, want %q", in, got, want)
		}
	}
}
