package application

import "testing"

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
