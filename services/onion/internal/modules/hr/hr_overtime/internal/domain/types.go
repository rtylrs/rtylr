package domain

import "time"

type OvertimeRequestRecord struct {
	ID                   int64
	OwnerID              int64
	BusinessID           int64
	EmployeeID           int64
	Date                 string
	PlannedExtraMinutes  int32
	Reason               string
	Status               string
	CurrentApprovalLevel int32
	RejectionReason      string
	ApprovedBy           int64
	MultiplierApplied    float64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type OvertimeConfigRecord struct {
	ID                    int64
	OwnerID               int64
	BusinessID            int64
	DailyThresholdMinutes int32
	DefaultMultiplier     float64
	AutoDetectEnabled     bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type OvertimeDetectionCandidate struct {
	EmployeeID   int64
	Date         string
	ExtraMinutes int32
}

type ListOvertimeRequestsParams struct {
	EmployeeID *int64
	Status     *string
	Limit      int32
	Offset     int32
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
