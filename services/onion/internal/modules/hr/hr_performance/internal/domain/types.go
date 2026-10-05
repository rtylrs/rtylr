package domain

import "time"

type ReviewCycleRecord struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	Name        string
	CycleType   string
	PeriodStart string
	PeriodEnd   string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PerformanceReviewRecord struct {
	ID                   int64
	OwnerID              int64
	BusinessID           int64
	EmployeeID           int64
	ReviewCycleID        int64
	ReviewerEmployeeID   int64
	ReviewType           string
	OverallRating        string
	Narrative            string
	Status               string
	SharedWithEmployeeAt time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type ReviewGoalRecord struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	EmployeeID      int64
	ReviewCycleID   int64
	Title           string
	Description     string
	Weight          float64
	SuccessCriteria string
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PerformanceConfigRecord struct {
	ID                  int64
	OwnerID             int64
	BusinessID          int64
	RatingScaleType     string
	RatingMin           int32
	RatingMax           int32
	RatingLabels        string
	PeerFeedbackEnabled bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func DefaultPerformanceConfig() *PerformanceConfigRecord {
	return &PerformanceConfigRecord{
		RatingScaleType:     "numeric",
		RatingMin:           1,
		RatingMax:           5,
		PeerFeedbackEnabled: false,
	}
}

type CreateReviewCycleParams struct {
	Name        string
	CycleType   string
	PeriodStart string
	PeriodEnd   string
}

type CreateGoalParams struct {
	EmployeeID      int64
	ReviewCycleID   int64
	Title           string
	Description     string
	Weight          float64
	SuccessCriteria string
}

type UpdateGoalParams struct {
	Title           string
	Description     string
	Weight          float64
	SuccessCriteria string
	Status          string
}

type UpsertConfigParams struct {
	RatingScaleType     string
	RatingMin           int32
	RatingMax           int32
	RatingLabels        string
	PeerFeedbackEnabled bool
}

type ListCyclesParams struct {
	Limit  int32
	Offset int32
}

type ListReviewsParams struct {
	CycleID    int64
	EmployeeID int64
	Limit      int32
	Offset     int32
}

type ListGoalsParams struct {
	EmployeeID    int64
	ReviewCycleID int64
	Limit         int32
	Offset        int32
}
