package application

import (
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
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

func TestValidateRating_NumericBoundaryExactMin(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 10}
	if err := validateRating(cfg, "1"); err != nil {
		t.Errorf("exact min should be valid, got %v", err)
	}
}

func TestValidateRating_NumericBoundaryExactMax(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 10}
	if err := validateRating(cfg, "10"); err != nil {
		t.Errorf("exact max should be valid, got %v", err)
	}
}

func TestValidateRating_NumericJustBelowMin(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 2, RatingMax: 5}
	if err := validateRating(cfg, "1"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("value below min should fail")
	}
}

func TestValidateRating_NumericJustAboveMax(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 5}
	if err := validateRating(cfg, "6"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("value above max should fail")
	}
}

func TestValidateRating_DescriptiveAllLabelsAccepted(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{
		RatingScaleType: "descriptive",
		RatingLabels:    `["poor","fair","good","excellent"]`,
	}
	for _, label := range []string{"poor", "fair", "good", "excellent"} {
		if err := validateRating(cfg, label); err != nil {
			t.Errorf("label %q should be valid, got %v", label, err)
		}
	}
}

func TestValidateRating_EmptyScaleType(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: ""}
	if err := validateRating(cfg, "3"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("empty scale type should fail")
	}
}

func TestDefaultPerformanceConfig_HasExpectedDefaults(t *testing.T) {
	cfg := domain.DefaultPerformanceConfig()
	if cfg.RatingScaleType != "numeric" {
		t.Errorf("expected numeric, got %q", cfg.RatingScaleType)
	}
	if cfg.RatingMin != 1 || cfg.RatingMax != 5 {
		t.Errorf("expected 1-5 range, got %d-%d", cfg.RatingMin, cfg.RatingMax)
	}
	if cfg.PeerFeedbackEnabled {
		t.Error("peer feedback should be disabled by default")
	}
}

func TestValidateRating_NumericNegative(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{RatingScaleType: "numeric", RatingMin: 1, RatingMax: 5}
	if err := validateRating(cfg, "-3"); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("negative number should fail")
	}
}

func TestValidateRating_DescriptiveEmptyValue(t *testing.T) {
	cfg := &domain.PerformanceConfigRecord{
		RatingScaleType: "descriptive",
		RatingLabels:    `["good","excellent"]`,
	}
	if err := validateRating(cfg, ""); !errors.Is(err, domain.ErrInvalidRating) {
		t.Errorf("empty string should fail for descriptive scale")
	}
}
