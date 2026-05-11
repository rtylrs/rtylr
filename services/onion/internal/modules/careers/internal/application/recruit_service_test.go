package application

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct {
		in, out string
	}{
		{"Senior Engineer", "senior-engineer"},
		{"  spaced  out  ", "spaced-out"},
		{"Mixed-CASE & symbols!!!", "mixed-case-symbols"},
		{"---leading-trailing---", "leading-trailing"},
		{"a   b   c", "a-b-c"},
		{"Café", "caf"}, // diacritics stripped (non a-z0-9-)
		{"", ""},
	}
	for _, tc := range cases {
		if got := slugify(tc.in); got != tc.out {
			t.Errorf("slugify(%q) = %q, want %q", tc.in, got, tc.out)
		}
	}
}
