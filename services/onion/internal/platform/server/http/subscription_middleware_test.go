package http

import (
	"testing"
	"time"
)

func TestSubscriptionAccessAllowed(t *testing.T) {
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		access subscriptionAccess
		want   bool
	}{
		{name: "active", access: subscriptionAccess{Status: "active"}, want: true},
		{name: "unexpired trial", access: subscriptionAccess{Status: "trial", TrialEnd: now.Add(time.Hour).Format(time.RFC3339Nano)}, want: true},
		{name: "expired trial", access: subscriptionAccess{Status: "trial", TrialEnd: now.Add(-time.Hour).Format(time.RFC3339Nano)}, want: false},
		{name: "manual grant", access: subscriptionAccess{Status: "cancelled", Granted: true}, want: true},
		{name: "cancelled", access: subscriptionAccess{Status: "cancelled"}, want: false},
		{name: "expired", access: subscriptionAccess{Status: "expired"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := subscriptionAccessAllowed(tt.access, now); got != tt.want {
				t.Fatalf("subscriptionAccessAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
