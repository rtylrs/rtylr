package cdn

import "testing"

func TestIsAllowedMIMEType(t *testing.T) {
	allowed := []string{
		"image/jpeg", "image/png", "image/webp", "image/gif",
		"image/avif", "image/tiff", "video/mp4", "video/webm", "video/quicktime",
		"application/pdf", "application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/csv", "text/plain",
	}
	for _, m := range allowed {
		if !isAllowedMIMEType(m) {
			t.Errorf("expected %q to be allowed", m)
		}
	}

	rejected := []string{
		"application/octet-stream", "text/html", "image/svg+xml",
		"video/x-msvideo", "",
	}
	for _, m := range rejected {
		if isAllowedMIMEType(m) {
			t.Errorf("expected %q to be rejected", m)
		}
	}
}

func TestIsAllowedMIMEType_CaseInsensitive(t *testing.T) {
	for _, m := range []string{"IMAGE/JPEG", "Image/Png", "VIDEO/MP4"} {
		if !isAllowedMIMEType(m) {
			t.Errorf("expected %q (mixed case) to be allowed", m)
		}
	}
}
