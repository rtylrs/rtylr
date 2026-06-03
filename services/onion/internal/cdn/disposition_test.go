package cdn

import (
	"strings"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	cases := map[string]string{
		"photo.jpg":         "photo.jpg",
		"my report.pdf":     "my report.pdf",
		"a/b/c.png":         "a_b_c.png",
		`weird".jpg`:        "weird_.jpg",
		"back\\slash.png":   "back_slash.png",
		"":                  "download",
		"   ":               "download",
		"...":               "download",
		"with\r\ninjection": "withinjection",
		"trailing.   ":      "trailing",
	}
	for in, want := range cases {
		if got := sanitizeFilename(in); got != want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizeFilename_NoHeaderBreakingChars(t *testing.T) {
	got := sanitizeFilename("evil\r\nSet-Cookie: x=1\".jpg")
	for _, bad := range []string{"\r", "\n", `"`} {
		if strings.Contains(got, bad) {
			t.Errorf("result %q still contains header-breaking %q", got, bad)
		}
	}
}

func TestSanitizeFilename_CapsLength(t *testing.T) {
	if got := sanitizeFilename(strings.Repeat("a", 500)); len(got) > 200 {
		t.Errorf("result length %d exceeds cap", len(got))
	}
}
