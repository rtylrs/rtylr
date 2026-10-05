package domain

import "context"

type ApprovalChainProvider interface {
	GetApprovalChain(ctx context.Context, businessID int64, requestType string) (*ApprovalChainInfo, error)
}

// PermissionProvider resolves the HR employee behind an authenticated user and
// checks that employee's role permissions and approval-chain role assignment.
type PermissionProvider interface {
	GetEmployeeIDByUserID(ctx context.Context, businessID, userID int64) (int64, error)
	GetEmployeeBusinessID(ctx context.Context, businessID, employeeID int64) (int64, error)
	HasPermission(ctx context.Context, businessID, employeeID int64, permissionKey string) (bool, error)
	HasRole(ctx context.Context, businessID, employeeID, roleID int64) (bool, error)
}
