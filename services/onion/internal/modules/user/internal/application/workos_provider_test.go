package application

import (
	"net/url"
	"testing"
)

func TestWorkOSProviderID(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "GoogleOAuth", false}, // default
		{"google", "GoogleOAuth", false},
		{"Google", "GoogleOAuth", false},    // case-insensitive
		{"  github ", "GitHubOAuth", false}, // trimmed
		{"github", "GitHubOAuth", false},
		{"microsoft", "MicrosoftOAuth", false},
		{"apple", "AppleOAuth", false},
		{"facebook", "", true}, // unsupported
		{"twitter", "", true},
	}
	for _, tc := range cases {
		got, err := workOSProviderID(tc.in)
		if (err != nil) != tc.wantErr {
			t.Errorf("workOSProviderID(%q) err=%v wantErr=%v", tc.in, err, tc.wantErr)
			continue
		}
		if got != tc.want {
			t.Errorf("workOSProviderID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestWorkOSAuthorizationURLIncludesUnpredictableState(t *testing.T) {
	provider := NewWorkOSProvider("test-key", "client-id", "https://app.example.com/oauth/callback")
	first, err := provider.GetAuthorizationURL("google")
	if err != nil {
		t.Fatalf("first URL: %v", err)
	}
	second, err := provider.GetAuthorizationURL("google")
	if err != nil {
		t.Fatalf("second URL: %v", err)
	}
	firstURL, err := url.Parse(first)
	if err != nil {
		t.Fatalf("parse first URL: %v", err)
	}
	secondURL, err := url.Parse(second)
	if err != nil {
		t.Fatalf("parse second URL: %v", err)
	}
	firstState := firstURL.Query().Get("state")
	secondState := secondURL.Query().Get("state")
	if len(firstState) < 32 || firstState == secondState {
		t.Fatalf("state values must be unpredictable and distinct: first=%q second=%q", firstState, secondState)
	}
}
