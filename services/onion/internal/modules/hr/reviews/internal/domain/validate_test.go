package domain

import (
	"errors"
	"math"
	"testing"
)

func TestNormalizeStatus(t *testing.T) {
	cases := map[string]string{"": "draft", "Submitted": "submitted", " DRAFT ": "draft", "acknowledged": "acknowledged"}
	for in, want := range cases {
		got, err := NormalizeStatus(in)
		if err != nil || got != want {
			t.Errorf("NormalizeStatus(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := NormalizeStatus("Completed"); !errors.Is(err, ErrInvalidReview) {
		t.Errorf("expected ErrInvalidReview for an unknown status, got %v", err)
	}
}

func TestValidateRatings(t *testing.T) {
	if err := ValidateRatings(0, 2.5, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	for _, bad := range []float64{-0.1, 5.1, 99.9, math.NaN()} {
		if err := ValidateRatings(4, bad); !errors.Is(err, ErrInvalidReview) {
			t.Errorf("ValidateRatings(%v) = %v, want ErrInvalidReview", bad, err)
		}
	}
}
