package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
)

func TestValidateRating_NumericInRange(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 5}
	for _, r := range []string{"1", "3", "5"} {
		if err := validateRating(cfg, r); err != nil {
			t.Errorf("expected %q to be valid, got %v", r, err)
		}
	}
}

func TestValidateRating_NumericOutOfRange(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 5}
	for _, r := range []string{"0", "6", "100", "-1"} {
		if err := validateRating(cfg, r); !errors.Is(err, domain.ErrInvalidRating) {
			t.Errorf("expected ErrInvalidRating for %q, got %v", r, err)
		}
	}
}

func TestValidateRating_NumericNotANumber(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 5}
	for _, r := range []string{"", "abc", "3.5", "five"} {
		if err := validateRating(cfg, r); !errors.Is(err, domain.ErrInvalidRating) {
			t.Errorf("expected ErrInvalidRating for %q, got %v", r, err)
		}
	}
}

func TestValidateRating_DescriptiveAcceptsListed(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{
		RatingScaleType: "descriptive",
		RatingLabels:    `["needs_improvement","meets","exceeds"]`,
	}
	if err := validateRating(cfg, "meets"); err != nil {
		t.Errorf("expected meets to be valid, got %v", err)
	}
}

func TestValidateRating_DescriptiveRejectsUnlistedAndMalformed(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{
		RatingScaleType: "descriptive",
		RatingLabels:    `["meets","exceeds"]`,
	}
	if err := validateRating(cfg, "fails"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("unlisted descriptive should fail, got %v", err)
	}

	cfg.RatingLabels = ""
	if err := validateRating(cfg, "anything"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("empty labels should fail")
	}

	cfg.RatingLabels = "{not-json}"
	if err := validateRating(cfg, "anything"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("malformed labels should fail")
	}
}

func TestValidateRating_UnknownScaleType(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "smiley_face"}
	if err := validateRating(cfg, "great"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("unknown scale type should fail")
	}
}
