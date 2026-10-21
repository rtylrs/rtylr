package domain

import "context"

type Repository interface {
	// Leave Types
	ListLeaveTypes(ctx context.Context, ownerID int64, params ListLeaveTypesParams) ([]LeaveTypeRecord, int64, error)
	GetLeaveType(ctx context.Context, ownerID, leaveTypeID int64) (*LeaveTypeRecord, error)
	CreateLeaveType(ctx context.Context, ownerID int64, lt LeaveTypeRecord) (*LeaveTypeRecord, error)
	UpdateLeaveType(ctx context.Context, ownerID int64, lt LeaveTypeRecord) error
	UpdateLeaveTypePatch(ctx context.Context, ownerID, leaveTypeID int64, patch LeaveTypePatch) error

	// Leave Requests
	ListLeaveRequests(ctx context.Context, ownerID int64, params ListLeaveRequestsParams) ([]LeaveRequestRecord, int64, error)
	GetLeaveRequest(ctx context.Context, ownerID, requestID int64) (*LeaveRequestRecord, error)
	CreateLeaveRequest(ctx context.Context, ownerID int64, req LeaveRequestRecord) (*LeaveRequestRecord, error)
	CreateLeaveRequestWithBalance(ctx context.Context, ownerID int64, req LeaveRequestRecord, allocations []LeaveDayAllocation) (*LeaveRequestRecord, error)
	ApproveLeaveRequest(ctx context.Context, ownerID, requestID int64) error
	ApproveLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, year int32, days float64) error
	ApproveLeaveRequestWithAllocations(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []LeaveDayAllocation) error
	AdvanceLeaveApprovalLevel(ctx context.Context, ownerID, requestID, currentLevel int32) error
	RejectLeaveRequest(ctx context.Context, ownerID, requestID int64, reason string) (bool, error)
	RejectLeaveRequestWithBalance(
		ctx context.Context,
		ownerID, requestID int64,
		employeeID, leaveTypeID int64,
		allocations []LeaveDayAllocation,
		reason string,
	) error
	CancelLeaveRequest(ctx context.Context, ownerID, requestID int64) (bool, error)
	CancelLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []LeaveDayAllocation) error
	ListOverlappingLeaveRequests(ctx context.Context, ownerID, employeeID int64, startDate, endDate string) ([]LeaveRequestRecord, error)
	ListLeaveCalendar(ctx context.Context, businessID, locationID int64, dateFrom, dateTo string) ([]LeaveRequestRecord, error)

	// Leave Balances
	ListLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]LeaveBalanceRecord, error)
	CountLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) (int64, error)
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
	ApplyYearEndCarryover(ctx context.Context, ownerID int64, previousYear, currentYear int32) error
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainInfo, error)
}

// PermissionProvider resolves the HR employee behind the authenticated
// platform user and checks the role/approval-chain permissions used by leave
// self-service and approval actions.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
	HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error)
}
