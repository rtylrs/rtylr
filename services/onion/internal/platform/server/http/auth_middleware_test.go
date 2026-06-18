package http

import "testing"

func TestExtractToken(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"bare token", "abcdef", "abcdef"},
		{"bearer token", "Bearer abcdef", "abcdef"},
		{"bearer token with extra spaces preserved", "Bearer  abc", " abc"},
		{"lowercase bearer is treated as bare", "bearer abcdef", "bearer abcdef"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractToken(tc.in); got != tc.want {
				t.Fatalf("extractToken(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestAuthSuccessLogsDisabledByDefault(t *testing.T) {
	t.Setenv("AUTH_SUCCESS_LOGS", "")
	t.Setenv("AUTH_REQUEST_LOGS", "")

	if authSuccessLogsEnabled() {
		t.Fatal("expected auth success logs to be disabled")
	}
}

func TestAuthSuccessLogsCanBeEnabled(t *testing.T) {
	t.Setenv("AUTH_SUCCESS_LOGS", "true")
	t.Setenv("AUTH_REQUEST_LOGS", "")

	if !authSuccessLogsEnabled() {
		t.Fatal("expected auth success logs to be enabled")
	}
}
