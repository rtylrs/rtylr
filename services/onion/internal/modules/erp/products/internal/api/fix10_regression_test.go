package api

import (
	"math"
	"testing"
)

func TestNormalizeProductListOffsetBoundsPageMath(t *testing.T) {
	cases := []struct {
		name   string
		offset int64
		want   int64
	}{
		{name: "negative", offset: -1, want: 0},
		{name: "normal", offset: 250, want: 250},
		{name: "max int", offset: math.MaxInt64, want: maxProductListOffset},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeProductListOffset(tc.offset); got != tc.want {
				t.Fatalf("normalizeProductListOffset(%d) = %d, want %d", tc.offset, got, tc.want)
			}
		})
	}
}
