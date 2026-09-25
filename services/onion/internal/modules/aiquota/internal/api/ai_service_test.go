package api

import "testing"

func TestCanUpdateAiQuotaSettings(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{name: "owner", role: "owner", want: true},
		{name: "admin", role: "admin", want: true},
		{name: "console admin", role: "console_admin", want: true},
		{name: "case insensitive", role: "Console_Admin", want: true},
		{name: "member", role: "member", want: false},
		{name: "empty", role: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canUpdateAiQuotaSettings(tt.role); got != tt.want {
				t.Fatalf("canUpdateAiQuotaSettings(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}
