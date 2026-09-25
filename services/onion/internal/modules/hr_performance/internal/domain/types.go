package domain

import "time"

type ReviewCycleRecord struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	Name        string    `json:"name"`
	CycleType   string    `json:"type"`
	PeriodStart string    `json:"period_start"`
	PeriodEnd   string    `json:"period_end"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PerformanceReviewRecord struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	EmployeeID           int64     `json:"employee_id"`
	ReviewCycleID        int64     `json:"review_cycle_id"`
	ReviewerEmployeeID   int64     `json:"reviewer_employee_id"`
	ReviewType           string    `json:"review_type"`
	OverallRating        string    `json:"overall_rating"`
	Narrative            string    `json:"narrative"`
	Status               string    `json:"status"`
	SharedWithEmployeeAt time.Time `json:"shared_with_employee_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type ReviewGoalRecord struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	EmployeeID      int64     `json:"employee_id"`
	ReviewCycleID   int64     `json:"review_cycle_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Weight          float64   `json:"weight"`
	SuccessCriteria string    `json:"success_criteria"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PerformanceConfigRecord struct {
	ID                  int64     `json:"id"`
	OwnerID             int64     `json:"owner_id"`
	BusinessID          int64     `json:"business_id"`
	RatingScaleType     string    `json:"rating_scale_type"`
	RatingMin           int32     `json:"rating_min"`
	RatingMax           int32     `json:"rating_max"`
	RatingLabels        string    `json:"rating_labels"`
	PeerFeedbackEnabled bool      `json:"peer_feedback_enabled"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
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
	Name        string `json:"name"`
	CycleType   string `json:"cycle_type"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

type CreateGoalParams struct {
	EmployeeID      int64   `json:"employee_id"`
	ReviewCycleID   int64   `json:"review_cycle_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Weight          float64 `json:"weight"`
	SuccessCriteria string  `json:"success_criteria"`
}

type UpdateGoalParams struct {
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Weight          float64 `json:"weight"`
	SuccessCriteria string  `json:"success_criteria"`
	Status          string  `json:"status"`
}

type UpsertConfigParams struct {
	RatingScaleType     string `json:"rating_scale_type"`
	RatingMin           int32  `json:"rating_min"`
	RatingMax           int32  `json:"rating_max"`
	RatingLabels        string `json:"rating_labels"`
	PeerFeedbackEnabled bool   `json:"peer_feedback_enabled"`
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
