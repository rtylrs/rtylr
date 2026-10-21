package domain

import "testing"

func TestMaskCodeKeepsOnlyTheLastFourCharacters(t *testing.T) {
	cases := map[string]string{
		"":                  "****",
		"AB12":              "****",
		"GC-1234-5678-ABCD": "****ABCD",
	}
	for code, want := range cases {
		if got := MaskCode(code); got != want {
			t.Errorf("MaskCode(%q) = %q, want %q", code, got, want)
		}
	}
}
