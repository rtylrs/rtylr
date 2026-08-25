package domain

import "time"

type OvertimeRequestRecord struct {
	ID                   int64     `json:"id"`
	OwnerID              int64     `json:"owner_id"`
	BusinessID           int64     `json:"business_id"`
	EmployeeID           int64     `json:"employee_id"`
	Date                 string    `json:"date"`
	PlannedExtraMinutes  int32     `json:"planned_extra_minutes"`
	Reason               string    `json:"reason"`
	Status               string    `json:"status"`
	CurrentApprovalLevel int32     `json:"current_approval_level"`
	ApprovedBy           int64     `json:"approved_by"`
	MultiplierApplied    float64   `json:"multiplier_applied"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type OvertimeConfigRecord struct {
	ID                    int64     `json:"id"`
	OwnerID               int64     `json:"owner_id"`
	BusinessID            int64     `json:"business_id"`
	DailyThresholdMinutes int32     `json:"daily_threshold_minutes"`
	DefaultMultiplier     float64   `json:"default_multiplier"`
	AutoDetectEnabled     bool      `json:"auto_detect_enabled"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type OvertimeDetectionCandidate struct {
	EmployeeID   int64
	Date         string
	ExtraMinutes int32
}

type ListOvertimeRequestsParams struct {
	Limit  int32
	Offset int32
}

type ApprovalChainInfo struct {
	TotalLevels int32
	Levels      []ApprovalLevelInfo
}

type ApprovalLevelInfo struct {
	LevelOrder         int32
	ApproverType       string
	ApproverRoleID     int64
	ApproverEmployeeID int64
}
