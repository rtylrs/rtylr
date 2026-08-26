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

func TestDecimalToFloat(t *testing.T) {
	for _, test := range []struct {
		name  string
		value any
		want  float64
	}{
		{name: "string", value: "12.35", want: 12.35},
		{name: "bytes", value: []byte("7.5"), want: 7.5},
		{name: "empty", value: "", want: 0},
		{name: "malformed", value: "not-a-number", want: 0},
		{name: "unsupported", value: int64(7), want: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := DecimalToFloat(test.value); got != test.want {
				t.Errorf("DecimalToFloat(%v) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}
