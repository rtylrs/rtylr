package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNormalizeNumber(t *testing.T) {
	for raw, want := range map[string]string{
		"42":                    "42",
		"  42\t":                "42",
		"a17b":                  "A17B",
		"٤٢":                    "42",
		"۴۲":                    "42",
		"007":                   "007",
		strings.Repeat("9", 12): strings.Repeat("9", 12),
		"":                      "",
		"   ":                   "",
		"4 2":                   "",
		"42!":                   "",
		"١٢-٣":                  "",
		"é":                     "",
		strings.Repeat("9", 13): "",
		strings.Repeat("٩", 13): "",
	} {
		got, err := NormalizeNumber(raw)
		if want == "" {
			if !errors.Is(err, ErrInvalidNumber) {
				t.Errorf("NormalizeNumber(%q) = %q, %v; want ErrInvalidNumber", raw, got, err)
			}
			continue
		}
		if err != nil || got != want {
			t.Errorf("NormalizeNumber(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
}
