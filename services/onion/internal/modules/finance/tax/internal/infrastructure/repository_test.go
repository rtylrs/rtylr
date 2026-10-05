package infrastructure

import "testing"

func TestParseMinorUnits(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int64
	}{
		{name: "whole amount", value: "12", want: 1200},
		{name: "two decimals", value: "12.34", want: 1234},
		{name: "one decimal", value: "0.1", want: 10},
		{name: "leading decimal", value: ".05", want: 5},
		{name: "negative", value: "-1.25", want: -125},
		{name: "whitespace", value: " 7.50 ", want: 750},
		{name: "empty", value: "", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMinorUnits(tt.value)
			if err != nil {
				t.Fatalf("parseMinorUnits(%q): %v", tt.value, err)
			}
			if got != tt.want {
				t.Fatalf("parseMinorUnits(%q) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestParseMinorUnitsRejectsExcessPrecision(t *testing.T) {
	if _, err := parseMinorUnits("1.234"); err == nil {
		t.Fatal("expected amounts with more than two fractional digits to be rejected")
	}
}

func TestMinorUnitsFromInterfacePreservesDecimalString(t *testing.T) {
	if got := minorUnitsFromInterface("0.10"); got != 10 {
		t.Fatalf("minorUnitsFromInterface(0.10) = %d, want 10", got)
	}
}

func TestFormatRateBasisPoints(t *testing.T) {
	tests := map[int64]string{
		0:    "0.00",
		725:  "7.25",
		2000: "20.00",
	}
	for value, want := range tests {
		if got := formatRateBasisPoints(value); got != want {
			t.Errorf("formatRateBasisPoints(%d) = %q, want %q", value, got, want)
		}
	}
}
