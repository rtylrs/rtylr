package http

import "testing"

func TestIsMaliciousPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		// Sub-segments of legitimate API paths must NOT match.
		{"/hr/setup/status", false},
		{"/erp/admin-panel", false},
		{"/console", false},
		{"/console/stats", false},
		{"/console/resources", false},
		{"/upload", false},
		{"/upload/plan", false},
		{"/files/abc-123.jpg", false},
		{"/images/abc-123.png", false},
		{"/uploads/file.png", false},

		// First-segment matches against the pattern list ARE scanner probes.
		{"/admin/settings", true},
		{"/etc/passwd", true},
		{"/proc/self", true},
		{"/wp-admin", true}, // /wp-admin IS in the pattern list

		// Patterns declared with a trailing slash must still match; before the
		// normalisation these entries were dead because a path segment can
		// never contain '/'.
		{"/.git/config", true},
		{"/.git", true},
		{"/debug/pprof", true},
		{"/log/app.log", true},
		{"/lock", true},
		{"/api/v1/probe", true},

		// But the allow-listed routes they would otherwise shadow stay reachable.
		{"/.well-known/rtylr.json", false},

		// A trailing-slash pattern must not become a substring match.
		{"/logger", false},

		// Empty path / root — never malicious.
		{"/", false},
		{"", false},

		// Case insensitivity.
		{"/Admin/Settings", true},
		{"/ETC/passwd", true},
	}

	for _, tc := range cases {
		if got := isMaliciousPath(tc.path); got != tc.want {
			t.Errorf("isMaliciousPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
