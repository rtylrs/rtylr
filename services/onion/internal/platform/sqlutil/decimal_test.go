package sqlutil

import "testing"

func TestFloatToDecimal(t *testing.T) {
	for _, test := range []struct {
		value float64
		want  string
	}{
		{value: 12, want: "12.00"},
		{value: 12.345, want: "12.35"},
		{value: -0.5, want: "-0.50"},
	} {
		if got := FloatToDecimal(test.value); got != test.want {
			t.Errorf("FloatToDecimal(%v) = %q, want %q", test.value, got, test.want)
		}
	}
}
