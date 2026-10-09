package http

import (
	"testing"
	"time"
)

func TestParseRouteRateLimit(t *testing.T) {
	for spec, want := range map[string]struct {
		limit  int
		window time.Duration
	}{"10/m": {10, time.Minute}, " 60/s ": {60, time.Second}, "5/h": {5, time.Hour}} {
		limit, window, err := ParseRouteRateLimit(spec)
		if err != nil || limit != want.limit || window != want.window {
			t.Errorf("ParseRouteRateLimit(%q) = %d, %s, %v", spec, limit, window, err)
		}
	}
	for _, bad := range []string{"", "10", "0/m", "-1/m", "ten/m", "10/d"} {
		if _, _, err := ParseRouteRateLimit(bad); err == nil {
			t.Errorf("ParseRouteRateLimit(%q) succeeded, want an error", bad)
		}
	}
}
