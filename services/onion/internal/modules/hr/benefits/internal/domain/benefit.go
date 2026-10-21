package domain

import (
	"errors"
	"time"
)

var ErrBenefitPlanNotFound = errors.New("benefit plan not found")

type BenefitPlan struct {
	ID           int64
	OwnerID      int64
	BusinessID   int64
	Name         string
	Type         string
	Provider     string
	CostEmployer float64
	CostEmployee float64
	IsActive     bool
	CreatedAt    time.Time
}
type BenefitPlanInput struct {
	Name         string
	Type         string
	Provider     string
	CostEmployer float64
	CostEmployee float64
	IsActive     bool
}
