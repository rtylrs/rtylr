package domain

import (
	"context"
	"time"
)

type Repository interface {
	// Employees
	ListEmployees(ctx context.Context, ownerID int64, params ListEmployeesParams) ([]EmployeeRecord, int64, error)
	GetEmployee(ctx context.Context, ownerID, employeeID int64) (*EmployeeRecord, error)
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
	CreateEmployee(ctx context.Context, ownerID int64, emp EmployeeRecord) (*EmployeeRecord, error)
	UpdateEmployee(ctx context.Context, ownerID int64, emp EmployeeRecord) (*EmployeeRecord, error)
	DeactivateEmployee(ctx context.Context, ownerID, employeeID int64, endDate time.Time) error
	LinkInviteAcceptance(ctx context.Context, ownerID, employeeID, userID int64) error
	CreateEmployeeWithBalances(ctx context.Context, ownerID int64, emp EmployeeRecord, leaveTypeIDs []int64) (*EmployeeRecord, error)

	// Invite
	CreateInviteToken(ctx context.Context, ownerID, employeeID int64, expiresAt time.Time) (string, error)
	GetInviteByToken(ctx context.Context, token string) (*InviteRecord, error)
	AcceptInvite(ctx context.Context, ownerID, employeeID, userID int64, token string) error

	// Locations
	ListLocations(ctx context.Context, ownerID int64) ([]LocationRecord, error)
	GetLocation(ctx context.Context, ownerID, locationID int64) (*LocationRecord, error)
	CreateLocation(ctx context.Context, ownerID int64, loc LocationRecord) (*LocationRecord, error)
	UpdateLocation(ctx context.Context, ownerID int64, loc LocationRecord) (*LocationRecord, error)

	// Departments
	ListDepartments(ctx context.Context, ownerID, locationID int64) ([]DepartmentRecord, error)
	CreateDepartment(ctx context.Context, ownerID int64, dept DepartmentRecord) (*DepartmentRecord, error)
	UpdateDepartment(ctx context.Context, ownerID int64, dept DepartmentRecord) (*DepartmentRecord, error)

	// Work schedules
	ListWorkSchedules(ctx context.Context, ownerID int64) ([]WorkScheduleRecord, error)
	GetWorkScheduleForEmployee(ctx context.Context, ownerID, employeeID int64) (*WorkScheduleRecord, error)
	CreateWorkSchedule(ctx context.Context, ownerID int64, ws WorkScheduleRecord) (*WorkScheduleRecord, error)
	UpdateWorkSchedule(ctx context.Context, ownerID int64, ws WorkScheduleRecord) (*WorkScheduleRecord, error)

	// Roles & permissions
	ListRoles(ctx context.Context, ownerID int64) ([]RoleRecord, error)
	CreateRole(ctx context.Context, ownerID int64, role RoleRecord) (*RoleRecord, error)
	UpdateRole(ctx context.Context, ownerID int64, role RoleRecord) (*RoleRecord, error)
	AssignRole(ctx context.Context, ownerID, employeeID, roleID int64) error
	HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error)
	HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error)

	// Setup wizard
	GetSetupSteps(ctx context.Context, ownerID int64) ([]SetupStepRecord, error)
	CompleteSetupStep(ctx context.Context, ownerID int64, stepKey string) (*SetupStepRecord, error)
	GetHRConfig(ctx context.Context, ownerID int64) (*HRCompanyConfigRecord, error)
	UpsertHRConfig(ctx context.Context, ownerID int64, cfg HRCompanyConfigRecord) (*HRCompanyConfigRecord, error)

	// Approval chains
	GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*ApprovalChainRecord, error)
	UpsertApprovalChain(ctx context.Context, ownerID int64, chain ApprovalChainRecord) (*ApprovalChainRecord, error)
}
