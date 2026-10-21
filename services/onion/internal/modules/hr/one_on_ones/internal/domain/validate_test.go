package domain

import (
	"errors"
	"testing"
)

func TestNormalizeActionItems(t *testing.T) {
	cases := map[string]string{
		"":                "[]",
		"   ":             "[]",
		`["a","b"]`:       `["a","b"]`,
		"first\nsecond\n": `["first","second"]`,
		"just text":       `["just text"]`,
		`{"a":1}`:         `["{\"a\":1}"]`,
	}
	for in, want := range cases {
		if got := NormalizeActionItems(in); got != want {
			t.Errorf("NormalizeActionItems(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeStatus(t *testing.T) {
	if got, err := NormalizeStatus(""); err != nil || got != "scheduled" {
		t.Errorf("empty status = %q, %v", got, err)
	}
	if got, err := NormalizeStatus("Cancelled"); err != nil || got != "cancelled" {
		t.Errorf("Cancelled = %q, %v", got, err)
	}
	if _, err := NormalizeStatus("done"); !errors.Is(err, ErrInvalidOneOnOne) {
		t.Errorf("expected ErrInvalidOneOnOne, got %v", err)
	}
}
