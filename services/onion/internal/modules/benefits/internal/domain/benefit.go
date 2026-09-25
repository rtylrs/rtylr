package domain

import (
	"errors"
	"time"
)

var ErrBenefitPlanNotFound = errors.New("benefit plan not found")

type BenefitPlan struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"owner_id"`
	BusinessID   int64     `json:"business_id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Provider     string    `json:"provider"`
	CostEmployer float64   `json:"cost_employer"`
	CostEmployee float64   `json:"cost_employee"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
type BenefitPlanInput struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Provider     string  `json:"provider"`
	CostEmployer float64 `json:"cost_employer"`
	CostEmployee float64 `json:"cost_employee"`
	IsActive     bool    `json:"is_active"`
}
