package infrastructure

import "testing"

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
