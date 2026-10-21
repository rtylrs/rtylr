package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidReview = errors.New("invalid review")
	ErrStaffNotFound = errors.New("staff member not found in this business")
)

const maxRating = 5.0

// NormalizeStatus maps the accepted spellings onto the hr_reviews status enum.
func NormalizeStatus(status string) (string, error) {
	switch s := strings.ToLower(strings.TrimSpace(status)); s {
	case "":
		return "draft", nil
	case "draft", "submitted", "acknowledged":
		return s, nil
	default:
		return "", fmt.Errorf("%w: unknown status %q", ErrInvalidReview, status)
	}
}

// ValidateRatings requires every rating to sit inside the 0..5 scale the
// decimal(3,1) columns and the UI both assume.
func ValidateRatings(ratings ...float64) error {
	for _, r := range ratings {
		if r < 0 || r > maxRating || r != r {
			return fmt.Errorf("%w: ratings must be between 0 and %.0f", ErrInvalidReview, maxRating)
		}
	}
	return nil
}
