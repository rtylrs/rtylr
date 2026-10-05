package domain

import (
	"errors"
	"testing"
)

func TestValidateLoyaltyPoints_AllowsInt32Boundary(t *testing.T) {
	if err := ValidateLoyaltyPoints(MaxLoyaltyPoints); err != nil {
		t.Fatalf("boundary value rejected: %v", err)
	}
}

func TestValidateLoyaltyPoints_RejectsValuePastInt32(t *testing.T) {
	err := ValidateLoyaltyPoints(MaxLoyaltyPoints + 1)
	if !errors.Is(err, ErrLoyaltyPointsOutOfRange) {
		t.Fatalf("error = %v, want ErrLoyaltyPointsOutOfRange", err)
	}
}

func TestValidateLoyaltyPoints_RejectsNonPositiveValues(t *testing.T) {
	for _, points := range []int64{0, -1} {
		if err := ValidateLoyaltyPoints(points); !errors.Is(err, ErrLoyaltyPointsNotPositive) {
			t.Errorf("ValidateLoyaltyPoints(%d) = %v, want ErrLoyaltyPointsNotPositive", points, err)
		}
	}
}
