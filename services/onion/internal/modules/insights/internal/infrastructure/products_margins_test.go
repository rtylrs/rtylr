package infrastructure

import "testing"

// TestDecimalToFloatHandlesDriverRepresentations pins B55-R04: DECIMAL
// aggregates come back from the MySQL driver as []byte/string, which the old
// float conversion fell through to 0 for.
func TestDecimalToFloatHandlesDriverRepresentations(t *testing.T) {
	cases := []struct {
		name string
		in   interface{}
		want float64
	}{
		{"driver bytes", []byte("2.500"), 2.5},
		{"driver bytes integer", []byte("3"), 3},
		{"string with three decimals", "1.125", 1.125},
		{"int64", int64(4), 4},
		{"float64", float64(1.5), 1.5},
		{"nil", nil, 0},
		{"empty bytes", []byte(""), 0},
		{"unparseable bytes", []byte("bogus"), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := decimalToFloat(tc.in); got != tc.want {
				t.Errorf("decimalToFloat(%#v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMarginBucketsCoverExpectedBoundaries(t *testing.T) {
	cases := []struct {
		margin float64
		want   string
	}{
		{-1, "<0%"},
		{0, "0-10%"},
		{10, "10-25%"},
		{25, "25-50%"},
		{50, "50-75%"},
		{75, "75%+"},
	}
	for _, tc := range cases {
		var got string
		for _, bucket := range marginBuckets {
			if tc.margin >= bucket.Lower && tc.margin < bucket.Upper {
				got = bucket.Label
				break
			}
		}
		if got != tc.want {
			t.Errorf("margin %.0f assigned to %q, want %q", tc.margin, got, tc.want)
		}
	}
}
