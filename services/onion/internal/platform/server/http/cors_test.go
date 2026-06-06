package http

import "testing"

func TestAllowedOriginsRequireHTTPSForHostedDomains(t *testing.T) {
	cases := []struct {
		origin string
		want   bool
	}{
		{origin: "http://localhost:3000", want: true},
		{origin: "https://rtylr.com", want: true},
		{origin: "https://dash.rtylr.com", want: true},
		{origin: "https://voxire.com", want: true},
		{origin: "https://api.voxire.tech", want: true},
		{origin: "http://rtylr.com", want: false},
		{origin: "http://dash.rtylr.com", want: false},
		{origin: "http://voxire.com", want: false},
		{origin: "http://api.voxire.tech", want: false},
	}

	allowedOrigins := getAllowedOrigins()
	for _, tc := range cases {
		t.Run(tc.origin, func(t *testing.T) {
			got := false
			for _, pattern := range allowedOrigins {
				if pattern.MatchString(tc.origin) {
					got = true
					break
				}
			}
			if got != tc.want {
				t.Fatalf("allowed origin %q = %v, want %v", tc.origin, got, tc.want)
			}
		})
	}
}
