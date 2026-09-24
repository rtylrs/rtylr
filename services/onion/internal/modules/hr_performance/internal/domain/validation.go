package domain

import (
	"errors"
	"math"
	"strings"
)

var (
	ErrEmptyCycleName     = errors.New("review cycle name is required")
	ErrInvalidCycleStatus = errors.New("invalid review cycle status")
	ErrInvalidCycleType   = errors.New("invalid review cycle type")
	ErrInvalidReviewType  = errors.New("invalid review type")
	ErrEmptyGoalTitle     = errors.New("goal title is required")
	ErrNegativeGoalWeight = errors.New("goal weight must be non-negative")
	ErrInvalidGoalWeight  = errors.New("goal weight must be between 0 and 100")
	ErrInvalidGoalStatus  = errors.New("invalid goal status")
	ErrInvalidRatingScale = errors.New("invalid rating scale type")
	ErrInvalidRatingRange = errors.New("rating min must be less than rating max")
)

// IsValidCycleStatus returns true when the status is recognised.
func IsValidCycleStatus(status string) bool {
	switch status {
	case CycleStatusDraft, CycleStatusOpen, CycleStatusClosed, CycleStatusArchived:
		return true
	}
	return false
}

// IsValidCycleType returns true when the cycle type is recognised.
func IsValidCycleType(cycleType string) bool {
	switch cycleType {
	case CycleTypeQuarterly, CycleTypeAnnual, CycleTypeMidYear, CycleTypeCustom, CycleTypeProbation, CycleTypeAdHoc:
		return true
	}
	return false
}

// IsValidReviewType returns true when the review type is recognised.
func IsValidReviewType(reviewType string) bool {
	switch reviewType {
	case ReviewTypeSelf, ReviewTypeManager, ReviewTypePeer:
		return true
	}
	return false
}

// IsValidGoalStatus returns true when the goal status is recognised.
func IsValidGoalStatus(status string) bool {
	switch status {
	case GoalStatusActive, GoalStatusCompleted, GoalStatusCanceled:
		return true
	}
	return false
}

// IsValidRatingScaleType returns true when the rating scale type is recognised.
func IsValidRatingScaleType(scaleType string) bool {
	switch scaleType {
	case RatingScaleNumeric, RatingScaleStar, RatingScaleLabel:
		return true
	}
	return false
}

// ValidateReviewCycle checks required fields on a create review cycle request.
func ValidateReviewCycle(p CreateReviewCycleParams) error {
	if strings.TrimSpace(p.Name) == "" {
		return ErrEmptyCycleName
	}
	if p.CycleType != "" && !IsValidCycleType(p.CycleType) {
		return ErrInvalidCycleType
	}
	return nil
}

// ValidateGoal checks required fields on a create goal request.
func ValidateGoal(p CreateGoalParams) error {
	if strings.TrimSpace(p.Title) == "" {
		return ErrEmptyGoalTitle
	}
	if math.IsNaN(p.Weight) || math.IsInf(p.Weight, 0) || p.Weight > 100 {
		return ErrInvalidGoalWeight
	}
	if p.Weight < 0 {
		return ErrNegativeGoalWeight
	}
	return nil
}

// ValidatePerformanceConfig checks that config values are sensible.
func ValidatePerformanceConfig(p UpsertConfigParams) error {
	if p.RatingScaleType != "" && !IsValidRatingScaleType(p.RatingScaleType) {
		return ErrInvalidRatingScale
	}
	if p.RatingMin >= p.RatingMax {
		return ErrInvalidRatingRange
	}
	return nil
}
