package domain_test

import (
	"errors"
	"math"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/domain"
)

func TestIsValidCycleStatus_AllValid(t *testing.T) {
	for _, s := range []string{
		domain.CycleStatusDraft,
		domain.CycleStatusOpen,
		domain.CycleStatusClosed,
		domain.CycleStatusArchived,
	} {
		if !domain.IsValidCycleStatus(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
}

func TestIsValidCycleStatus_Invalid(t *testing.T) {
	if domain.IsValidCycleStatus("active") {
		t.Error("expected 'active' to be invalid")
	}
}

func TestIsValidCycleType_AllValid(t *testing.T) {
	for _, ct := range []string{
		domain.CycleTypeQuarterly,
		domain.CycleTypeAnnual,
		domain.CycleTypeMidYear,
		domain.CycleTypeCustom,
		domain.CycleTypeProbation,
		domain.CycleTypeAdHoc,
	} {
		if !domain.IsValidCycleType(ct) {
			t.Errorf("expected %q to be valid", ct)
		}
	}
}

func TestIsValidCycleType_Invalid(t *testing.T) {
	if domain.IsValidCycleType("monthly") {
		t.Error("expected 'monthly' to be invalid")
	}
}

func TestIsValidReviewType_AllValid(t *testing.T) {
	for _, rt := range []string{
		domain.ReviewTypeSelf,
		domain.ReviewTypeManager,
		domain.ReviewTypePeer,
	} {
		if !domain.IsValidReviewType(rt) {
			t.Errorf("expected %q to be valid", rt)
		}
	}
}

func TestIsValidReviewType_Invalid(t *testing.T) {
	if domain.IsValidReviewType("360") {
		t.Error("expected '360' to be invalid")
	}
}

func TestIsValidGoalStatus_AllValid(t *testing.T) {
	for _, gs := range []string{
		domain.GoalStatusActive,
		domain.GoalStatusCompleted,
		domain.GoalStatusCanceled,
	} {
		if !domain.IsValidGoalStatus(gs) {
			t.Errorf("expected %q to be valid", gs)
		}
	}
}

func TestIsValidRatingScaleType_AllValid(t *testing.T) {
	for _, rs := range []string{
		domain.RatingScaleNumeric,
		domain.RatingScaleStar,
		domain.RatingScaleLabel,
	} {
		if !domain.IsValidRatingScaleType(rs) {
			t.Errorf("expected %q to be valid", rs)
		}
	}
}

func TestValidateReviewCycle_Valid(t *testing.T) {
	p := domain.CreateReviewCycleParams{
		Name:      "Q3 Review",
		CycleType: domain.CycleTypeQuarterly,
	}
	if err := domain.ValidateReviewCycle(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateReviewCycle_EmptyName(t *testing.T) {
	p := domain.CreateReviewCycleParams{Name: "", CycleType: domain.CycleTypeAnnual}
	if err := domain.ValidateReviewCycle(p); !errors.Is(err, domain.ErrEmptyCycleName) {
		t.Errorf("expected ErrEmptyCycleName, got %v", err)
	}
}

func TestValidateReviewCycle_InvalidCycleType(t *testing.T) {
	p := domain.CreateReviewCycleParams{Name: "Test", CycleType: "daily"}
	if err := domain.ValidateReviewCycle(p); !errors.Is(err, domain.ErrInvalidCycleType) {
		t.Errorf("expected ErrInvalidCycleType, got %v", err)
	}
}

func TestValidateGoal_Valid(t *testing.T) {
	p := domain.CreateGoalParams{Title: "Improve velocity", Weight: 1.0}
	if err := domain.ValidateGoal(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateGoal_EmptyTitle(t *testing.T) {
	p := domain.CreateGoalParams{Title: "", Weight: 1.0}
	if err := domain.ValidateGoal(p); !errors.Is(err, domain.ErrEmptyGoalTitle) {
		t.Errorf("expected ErrEmptyGoalTitle, got %v", err)
	}
}

func TestValidateGoal_NegativeWeight(t *testing.T) {
	p := domain.CreateGoalParams{Title: "Goal", Weight: -0.5}
	if err := domain.ValidateGoal(p); !errors.Is(err, domain.ErrNegativeGoalWeight) {
		t.Errorf("expected ErrNegativeGoalWeight, got %v", err)
	}
}

func TestValidateGoal_WeightAboveRange(t *testing.T) {
	p := domain.CreateGoalParams{Title: "Goal", Weight: 100.01}
	if err := domain.ValidateGoal(p); !errors.Is(err, domain.ErrInvalidGoalWeight) {
		t.Errorf("expected ErrInvalidGoalWeight, got %v", err)
	}
}

func TestValidateGoal_NonFiniteWeight(t *testing.T) {
	for _, weight := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if err := domain.ValidateGoal(domain.CreateGoalParams{Title: "Goal", Weight: weight}); !errors.Is(err, domain.ErrInvalidGoalWeight) {
			t.Errorf("weight %v error = %v, want ErrInvalidGoalWeight", weight, err)
		}
	}
}

func TestValidatePerformanceConfig_Valid(t *testing.T) {
	p := domain.UpsertConfigParams{
		RatingScaleType: domain.RatingScaleNumeric,
		RatingMin:       1,
		RatingMax:       5,
	}
	if err := domain.ValidatePerformanceConfig(p); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePerformanceConfig_InvalidScale(t *testing.T) {
	p := domain.UpsertConfigParams{RatingScaleType: "emoji", RatingMin: 1, RatingMax: 5}
	if err := domain.ValidatePerformanceConfig(p); !errors.Is(err, domain.ErrInvalidRatingScale) {
		t.Errorf("expected ErrInvalidRatingScale, got %v", err)
	}
}

func TestValidatePerformanceConfig_MinNotLessThanMax(t *testing.T) {
	p := domain.UpsertConfigParams{RatingScaleType: domain.RatingScaleNumeric, RatingMin: 5, RatingMax: 5}
	if err := domain.ValidatePerformanceConfig(p); !errors.Is(err, domain.ErrInvalidRatingRange) {
		t.Errorf("expected ErrInvalidRatingRange when min==max, got %v", err)
	}
}

func TestDefaultPerformanceConfig_IsValid(t *testing.T) {
	cfg := domain.DefaultPerformanceConfig()
	if cfg == nil {
		t.Fatal("DefaultPerformanceConfig returned nil")
	}
	if cfg.RatingMin >= cfg.RatingMax {
		t.Errorf("default config has invalid range: min=%d max=%d", cfg.RatingMin, cfg.RatingMax)
	}
}
