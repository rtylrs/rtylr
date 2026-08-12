package domain

import "context"

type Repository interface {
	// Leave Types
	ListLeaveTypes(ctx context.Context, ownerID int64, params ListLeaveTypesParams) ([]LeaveTypeRecord, int64, error)
	GetLeaveType(ctx context.Context, ownerID, leaveTypeID int64) (*LeaveTypeRecord, error)
	CreateLeaveType(ctx context.Context, ownerID int64, lt LeaveTypeRecord) (*LeaveTypeRecord, error)
	UpdateLeaveType(ctx context.Context, ownerID int64, lt LeaveTypeRecord) error

	// Leave Requests
	ListLeaveRequests(ctx context.Context, ownerID int64, params ListLeaveRequestsParams) ([]LeaveRequestRecord, int64, error)
	GetLeaveRequest(ctx context.Context, ownerID, requestID int64) (*LeaveRequestRecord, error)
	CreateLeaveRequest(ctx context.Context, ownerID int64, req LeaveRequestRecord) (*LeaveRequestRecord, error)
	ApproveLeaveRequest(ctx context.Context, ownerID, requestID int64) error
	ApproveLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, year int32, days float64) error
	AdvanceLeaveApprovalLevel(ctx context.Context, ownerID, requestID int64) error
	RejectLeaveRequest(ctx context.Context, ownerID, requestID int64, reason string) error
	ListOverlappingLeaveRequests(ctx context.Context, ownerID, employeeID int64, startDate, endDate string) ([]LeaveRequestRecord, error)
	ListLeaveCalendar(ctx context.Context, ownerID int64, dateFrom, dateTo string) ([]LeaveRequestRecord, error)

	// Leave Balances
	ListLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]LeaveBalanceRecord, error)
	GetLeaveBalance(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32) (*LeaveBalanceRecord, error)
	UpsertLeaveBalance(ctx context.Context, ownerID int64, bal LeaveBalanceRecord) error
	AddToUsedDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64) error
	AddToPendingDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64) error
	SubtractPendingDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64) error

	// Accrual
	ListDistinctOwnerIDs(ctx context.Context) ([]int64, error)
	ListAllActiveLeaveTypesForOwner(ctx context.Context, ownerID int64) ([]LeaveTypeRecord, error)
	ListBalancesForAccrual(ctx context.Context, ownerID int64, year, currentMonth int32) ([]LeaveBalanceRecord, error)
	AddToEntitledDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64, month int32) error
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}
