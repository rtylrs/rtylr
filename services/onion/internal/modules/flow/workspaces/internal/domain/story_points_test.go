package domain

import "testing"

func TestIsValidStoryPoints(t *testing.T) {
	cases := []struct {
		scale  string
		points int32
		want   bool
	}{
		{PointsScaleFibonacci, 8, true},
		{PointsScaleFibonacci, 4, false},
		{PointsScaleLinear, 4, true},
		{PointsScaleLinear, 13, false},
		{PointsScaleTShirt, 5, true},
		{PointsScaleTShirt, 13, false},
		{"", 13, true},
		{"", 4, false},
	}
	for _, c := range cases {
		if got := IsValidStoryPoints(c.scale, c.points); got != c.want {
			t.Errorf("IsValidStoryPoints(%q, %d) = %v, want %v", c.scale, c.points, got, c.want)
		}
	}
	if IsValidPointsScale("hours") {
		t.Error("an unknown scale must be rejected")
	}
}
