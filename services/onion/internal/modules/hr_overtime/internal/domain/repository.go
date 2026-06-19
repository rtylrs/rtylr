package domain

import "context"

type Repository interface {
	ListOvertimeRequests(ctx context.Context, businessID int64, params ListOvertimeRequestsParams) ([]OvertimeRequestRecord, int64, error)
	GetOvertimeRequest(ctx context.Context, businessID, requestID int64) (*OvertimeRequestRecord, error)
	CreateOvertimeRequest(ctx context.Context, businessID int64, req OvertimeRequestRecord) (*OvertimeRequestRecord, error)
	ApproveOvertimeRequest(ctx context.Context, businessID, requestID, approverID int64, multiplier float64) error
	AdvanceOvertimeApprovalLevel(ctx context.Context, businessID, requestID int64) error
	RejectOvertimeRequest(ctx context.Context, businessID, requestID int64) error
	CreateFlaggedOvertimeRequest(ctx context.Context, businessID, employeeID int64, date string, extraMinutes int32) error
	ListFlaggedForDate(ctx context.Context, businessID int64, date string) ([]OvertimeRequestRecord, error)

	GetOvertimeConfig(ctx context.Context, businessID int64) (*OvertimeConfigRecord, error)
	UpsertOvertimeConfig(ctx context.Context, businessID int64, cfg OvertimeConfigRecord) error
	ListAllConfiguredOwnerIDs(ctx context.Context) ([]int64, error)
}

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, businessID int64, requestType string) (*ApprovalChainInfo, error)
}
