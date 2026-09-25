package stringutils

import "testing"

func TestContainsOneOf(t *testing.T) {
	cases := []struct {
		text     string
		patterns []string
		want     bool
	}{
		{"the quick brown fox", []string{"slow", "fast", "quick"}, true},
		{"hello world", []string{"hello"}, true},
		{"hello world", []string{"goodbye"}, false},
		{"empty patterns", []string{}, false},
		{"nil patterns", nil, false},
		{"", []string{"anything"}, false},
		{"", []string{""}, true}, // strings.Contains(_, "") is always true
	}
	for _, tc := range cases {
		if got := ContainsOneOf(tc.text, tc.patterns); got != tc.want {
			t.Errorf("ContainsOneOf(%q, %v) = %v, want %v", tc.text, tc.patterns, got, tc.want)
		}
	}
}
