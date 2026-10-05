package application

import (
	"strings"
	"testing"
)

func TestFirstWord(t *testing.T) {
	cases := map[string]string{
		"":               "",
		"Alice":          "Alice",
		"Alice Smith":    "Alice",
		"Alice Mary Doe": "Alice",
		"Alice\tSmith":   "Alice",
		" leading space": "",
		"NoSpaces12345":  "NoSpaces12345",
	}
	for in, want := range cases {
		if got := firstWord(in); got != want {
			t.Errorf("firstWord(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEscape_HandlesAllSpecials(t *testing.T) {
	in := `<script>alert("xss & 'rm -rf /'")</script>`
	got := escape(in)
	for _, must := range []string{"&lt;", "&gt;", "&amp;", "&quot;", "&#39;"} {
		if !strings.Contains(got, must) {
			t.Errorf("escape missing %q: got %q", must, got)
		}
	}
	for _, mustNot := range []string{"<", ">"} {
		if strings.Contains(got, mustNot) {
			t.Errorf("escape leaked unescaped %q: got %q", mustNot, got)
		}
	}
}

func TestEscape_LeavesPlainTextAlone(t *testing.T) {
	in := "Just a normal string"
	if got := escape(in); got != in {
		t.Errorf("escape mutated plain text: %q -> %q", in, got)
	}
}
