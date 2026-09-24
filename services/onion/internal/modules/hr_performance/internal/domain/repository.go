package domain

import "context"

type Repository interface {
	// Config
	GetPerformanceConfig(ctx context.Context, ownerID int64) (*PerformanceConfigRecord, error)
	UpsertPerformanceConfig(ctx context.Context, ownerID int64, params UpsertConfigParams) (*PerformanceConfigRecord, error)

	// Review Cycles
	CreateReviewCycle(ctx context.Context, ownerID int64, params CreateReviewCycleParams) (*ReviewCycleRecord, error)
	GetReviewCycle(ctx context.Context, ownerID, id int64) (*ReviewCycleRecord, error)
	ListReviewCycles(ctx context.Context, ownerID int64, limit, offset int32) ([]ReviewCycleRecord, int64, error)
	UpdateReviewCycleStatus(ctx context.Context, ownerID, id int64, status string) error
	OpenReviewCycle(ctx context.Context, ownerID, id int64) error

	// Reviews
	CreatePerformanceReview(ctx context.Context, ownerID, employeeID, cycleID, reviewerID int64, reviewType string) error
	GetPerformanceReview(ctx context.Context, ownerID, id int64) (*PerformanceReviewRecord, error)
	ListPerformanceReviews(ctx context.Context, ownerID, cycleID int64, limit, offset int32) ([]PerformanceReviewRecord, int64, error)
	ListPerformanceReviewsByEmployee(ctx context.Context, ownerID, cycleID, employeeID int64, limit, offset int32) ([]PerformanceReviewRecord, int64, error)
	UpdateReviewSubmission(ctx context.Context, ownerID, id int64, rating, narrative string) error
	UpdateReviewStatus(ctx context.Context, ownerID, id int64, status string) error
	ShareReview(ctx context.Context, ownerID, id int64) error

	// Goals
	CreateGoal(ctx context.Context, ownerID int64, params CreateGoalParams) (*ReviewGoalRecord, error)
	GetGoal(ctx context.Context, ownerID, id int64) (*ReviewGoalRecord, error)
	UpdateGoal(ctx context.Context, ownerID, id int64, params UpdateGoalParams) error
	ListGoals(ctx context.Context, ownerID, employeeID, cycleID int64, limit, offset int32) ([]ReviewGoalRecord, int64, error)

	// Employees (cross-schema reads for subject and cycle initialization)
	ListActiveEmployeeIDs(ctx context.Context, ownerID int64) ([]int64, error)
	EmployeeBelongsToBusiness(ctx context.Context, businessID, employeeID int64) (bool, error)
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
}
