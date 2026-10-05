package cdn

import "testing"

func TestNegotiateFormat(t *testing.T) {
	cases := []struct {
		accept string
		want   string
	}{
		{"image/avif,image/webp,image/*", "avif"},
		{"image/webp,image/*", "webp"},
		{"image/png,image/*", ""},
		{"*/*", ""},
		{"", ""},
		{"IMAGE/AVIF", "avif"},
		{"text/html,image/webp", "webp"},
	}
	for _, tc := range cases {
		if got := negotiateFormat(tc.accept); got != tc.want {
			t.Errorf("negotiateFormat(%q) = %q, want %q", tc.accept, got, tc.want)
		}
	}
}

// AVIF must win over WebP when both are advertised.
func TestNegotiateFormat_PrefersAVIF(t *testing.T) {
	if got := negotiateFormat("image/webp,image/avif"); got != "avif" {
		t.Errorf("expected avif to win, got %q", got)
	}
}
