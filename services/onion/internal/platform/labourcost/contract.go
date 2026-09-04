// Package labourcost defines the platform boundary for project labour costs.
// Consumers receive only aggregate money and coverage data; hourly rates stay
// inside HR so they cannot leak through Flow, Finance, or Insights.
package labourcost

import "context"

// ProjectCost is the live labour-cost aggregate for one tenant project.
// CostCents is calculated from billable time with a configured hourly rate.
// UncostedMinutes identifies billable time whose employee has no hourly rate;
// it is deliberately not folded into CostCents as zero-cost work.
type ProjectCost struct {
	BusinessID            int64
	ProjectID             int64
	CostCents             int64
	CostedMinutes         int64
	UncostedMinutes       int64
	UncostedEmployeeCount int64
}

// Provider returns tenant-scoped, live project labour aggregates.
// Implementations must not expose employee rates or other compensation data.
type Provider interface {
	GetProjectLabourCost(ctx context.Context, businessID, projectID int64) (ProjectCost, error)
}
