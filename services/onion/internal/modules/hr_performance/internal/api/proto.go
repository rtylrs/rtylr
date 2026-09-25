package api

import (
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

func cycleToProto(c domain.ReviewCycleRecord) *hrperformancev1.ReviewCyclePayload {
	return &hrperformancev1.ReviewCyclePayload{
		Id:          &c.ID,
		OwnerId:     c.OwnerID,
		Name:        c.Name,
		Type:        c.CycleType,
		PeriodStart: c.PeriodStart,
		PeriodEnd:   c.PeriodEnd,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt.Unix(),
		UpdatedAt:   c.UpdatedAt.Unix(),
	}
}

func reviewToProto(r domain.PerformanceReviewRecord) *hrperformancev1.PerformanceReviewPayload {
	var rating *string
	if r.OverallRating != "" {
		rating = &r.OverallRating
	}
	var narrative *string
	if r.Narrative != "" {
		narrative = &r.Narrative
	}
	var sharedAt *int64
	if !r.SharedWithEmployeeAt.IsZero() {
		v := r.SharedWithEmployeeAt.Unix()
		sharedAt = &v
	}
	return &hrperformancev1.PerformanceReviewPayload{
		Id:                   &r.ID,
		OwnerId:              r.OwnerID,
		EmployeeId:           r.EmployeeID,
		ReviewCycleId:        r.ReviewCycleID,
		ReviewerEmployeeId:   r.ReviewerEmployeeID,
		ReviewType:           r.ReviewType,
		OverallRating:        rating,
		Narrative:            narrative,
		Status:               r.Status,
		SharedWithEmployeeAt: sharedAt,
		CreatedAt:            r.CreatedAt.Unix(),
		UpdatedAt:            r.UpdatedAt.Unix(),
	}
}

func goalToProto(g domain.ReviewGoalRecord) *hrperformancev1.ReviewGoalPayload {
	var weight *float64
	if g.Weight != 0 {
		weight = &g.Weight
	}
	var criteria *string
	if g.SuccessCriteria != "" {
		criteria = &g.SuccessCriteria
	}
	return &hrperformancev1.ReviewGoalPayload{
		Id:              g.ID,
		OwnerId:         g.OwnerID,
		EmployeeId:      g.EmployeeID,
		ReviewCycleId:   g.ReviewCycleID,
		Title:           g.Title,
		Description:     g.Description,
		Weight:          weight,
		SuccessCriteria: criteria,
		Status:          g.Status,
		CreatedAt:       g.CreatedAt.Unix(),
		UpdatedAt:       g.UpdatedAt.Unix(),
	}
}
