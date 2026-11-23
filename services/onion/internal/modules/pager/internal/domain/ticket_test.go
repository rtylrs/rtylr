package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
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
			if !errors.Is(err, ErrInvalidLabel) {
				t.Errorf("NormalizeNumber(%q) = %q, %v; want ErrInvalidLabel", raw, got, err)
			}
			continue
		}
		if err != nil || got != want {
			t.Errorf("NormalizeNumber(%q) = %q, %v; want %q", raw, got, err, want)
		}
	}
}

func TestParseLabelFollowsTheVenuesMode(t *testing.T) {
	for _, c := range []struct {
		mode, raw, label, kind string
	}{
		{LabelModeEither, " 42 ", "42", LabelNumber},
		{LabelModeEither, "a17", "A17", LabelNumber},
		{LabelModeEither, "٤٢", "42", LabelNumber},
		{LabelModeEither, "  Sara   Haddad ", "Sara Haddad", LabelName},
		{LabelModeEither, "سارة", "سارة", LabelName},
		// A digit alone does not make a number: this is a name with a digit.
		{LabelModeEither, "Table 4", "Table 4", LabelName},
		{LabelModeNumber, "a17", "A17", LabelNumber},
		{LabelModeName, "42", "42", LabelName},
	} {
		label, kind, err := ParseLabel(c.raw, c.mode)
		if err != nil || label != c.label || kind != c.kind {
			t.Errorf("ParseLabel(%q, %s) = %q %s %v; want %q %s", c.raw, c.mode, label, kind, err, c.label, c.kind)
		}
	}
	for _, c := range []struct{ mode, raw string }{
		{LabelModeNumber, "Sara"},
		{LabelModeNumber, "4 2"},
		{LabelModeName, "   "},
		{LabelModeName, strings.Repeat("س", 41)},
		{LabelModeEither, ""},
		{LabelModeEither, "Sara\x00"},
	} {
		if _, _, err := ParseLabel(c.raw, c.mode); !errors.Is(err, ErrInvalidLabel) {
			t.Errorf("ParseLabel(%q, %s) err = %v; want ErrInvalidLabel", c.raw, c.mode, err)
		}
	}
	if label, _, err := ParseLabel(strings.Repeat("س", 40), LabelModeName); err != nil || label == "" {
		t.Errorf("a 40 character name was refused: %v", err)
	}
}

func TestUnlabelledTicketsExpireSoonerThanLabelledOnes(t *testing.T) {
	created := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	unlabelled := Ticket{Status: StatusWaiting, CreatedAt: created}
	labelled := Ticket{Status: StatusWaiting, CreatedAt: created, Label: "42", LabelKind: LabelNumber}

	if got := unlabelled.StatusAt(created.Add(UnlabelledLifetime)); got != StatusWaiting {
		t.Fatalf("unlabelled at exactly 15 minutes = %s", got)
	}
	if got := unlabelled.StatusAt(created.Add(UnlabelledLifetime + time.Millisecond)); got != StatusExpired {
		t.Fatalf("unlabelled past 15 minutes = %s", got)
	}
	if got := labelled.StatusAt(created.Add(UnlabelledLifetime + 15*time.Minute)); got != StatusWaiting {
		t.Fatalf("labelled after half an hour = %s", got)
	}
	if got := labelled.StatusAt(created.Add(OpenLifetime)); got != StatusWaiting {
		t.Fatalf("labelled at exactly one hour = %s", got)
	}
	if got := labelled.StatusAt(created.Add(OpenLifetime + time.Millisecond)); got != StatusExpired {
		t.Fatalf("labelled past one hour = %s", got)
	}
}

func TestShownCodeKeepsNumbersAndCapitalisesOldCodes(t *testing.T) {
	for code, want := range map[string]string{"427": "427", "1003": "1003", "panda2": "PANDA2", "A7": "A7"} {
		if got := ShownCode(code); got != want {
			t.Errorf("ShownCode(%q) = %q, want %q", code, got, want)
		}
	}
}
