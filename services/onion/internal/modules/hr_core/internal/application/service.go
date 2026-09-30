package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	hrCoreScope = "hr_core"
	hrCoreTTL   = 5 * time.Minute
)

func invalidateHRCoreScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrCoreScope, ownerID))
}

type Service struct {
	repo               *infrastructure.Repository
	terminationRevoker TerminationRevoker
}

func NewService(_ context.Context, db *sql.DB) *Service {
	return &Service{repo: infrastructure.NewRepository(db)}
}

// --- Employees ---

func (s *Service) ListEmployees(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.EmployeeRecord, int64, error) {
	return s.ListEmployeesWithFilters(ctx, ownerID, domain.ListEmployeesParams{Limit: limit, Offset: offset})
}

// ListEmployeesWithFilters lists a bounded page of employees for a business.
// The optional filters are part of the employee contract and are included in
// the cache key so a filtered response can never be served to an unfiltered
// request (or to a different filter combination).
func (s *Service) ListEmployeesWithFilters(ctx context.Context, ownerID int64, params domain.ListEmployeesParams) ([]domain.EmployeeRecord, int64, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}
	params.Limit = limit
	params.Offset = offset

	departmentID := "*"
	if params.DepartmentID != nil {
		departmentID = fmt.Sprintf("%d", *params.DepartmentID)
	}
	locationID := "*"
	if params.LocationID != nil {
		locationID = fmt.Sprintf("%d", *params.LocationID)
	}
	status := "*"
	if params.Status != nil {
		status = fmt.Sprintf("%q", *params.Status)
	}

	type lr struct {
		Items []domain.EmployeeRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrCoreScope, ownerID, fmt.Sprintf("emps|l=%d|o=%d|d=%s|loc=%s|s=%s", limit, offset, departmentID, locationID, status))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListEmployees(ctx, ownerID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) GetEmployee(ctx context.Context, ownerID, employeeID int64) (*domain.EmployeeRecord, error) {
	key := cacheKeys.Scoped(hrCoreScope, ownerID, fmt.Sprintf("emp=%d", employeeID))
	return cacher.Of[*domain.EmployeeRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.EmployeeRecord, error) {
			return s.repo.GetEmployee(ctx, ownerID, employeeID)
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
}

func (s *Service) GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error) {
	return s.repo.GetEmployeeIDByUserID(ctx, ownerID, userID)
}

func (s *Service) CreateEmployee(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) (*domain.EmployeeRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	if err := s.SeedDefaultRoles(ctx, ownerID); err != nil {
		logger.Ctx(ctx).Error("failed to seed default roles during employee creation", logger.Any("owner_id", ownerID), logger.Error(err))
	}

	if emp.Status == "" {
		emp.Status = "active"
	}
	if emp.EmploymentType == "" {
		emp.EmploymentType = "full_time"
	}

	leaveTypeIDs, err := s.repo.ListActiveLeaveTypeIDs(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateEmployeeWithBalances(ctx, ownerID, emp, leaveTypeIDs)
}

func (s *Service) UpdateEmployee(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) (*domain.EmployeeRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	updated, err := s.repo.UpdateEmployee(ctx, ownerID, emp)
	if err != nil {
		return updated, err
	}
	// Employment ended: drop the user's app permissions for this business
	// immediately. Failure here is logged, not fatal; team's daily
	// reconciliation job is the durability guarantee (#834).
	s.revokeOnTermination(ctx, ownerID, emp)
	return updated, nil
}

// DeactivateEmployee changes only the employment state and end date. The
// existing record is loaded first so termination revocation still has the
// linked user id without passing a zero-filled employee through UpdateEmployee.
func (s *Service) DeactivateEmployee(ctx context.Context, ownerID, employeeID int64) error {
	existing, err := s.repo.GetEmployee(ctx, ownerID, employeeID)
	if err != nil {
		return err
	}
	if err := s.repo.DeactivateEmployee(ctx, ownerID, employeeID, time.Now().UTC()); err != nil {
		return err
	}
	existing.Status = "terminated"
	s.revokeOnTermination(ctx, ownerID, *existing)
	invalidateHRCoreScope(ctx, ownerID)
	return nil
}

// --- Invites ---

func (s *Service) InviteEmployee(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) (string, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	emp.Status = "invited"
	if emp.EmploymentType == "" {
		emp.EmploymentType = "full_time"
	}

	created, err := s.repo.CreateEmployee(ctx, ownerID, emp)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(72 * time.Hour)
	token, err := s.repo.CreateInviteToken(ctx, ownerID, created.ID, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) VerifyInviteToken(ctx context.Context, token string) (*domain.InviteRecord, error) {
	invite, err := s.repo.GetInviteByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if time.Now().After(invite.ExpiresAt) {
		return nil, domain.ErrInviteTokenExpired
	}
	return invite, nil
}

// AcceptInviteForCaller links the accepting account to the invited employee
// row. callerUserID is the caller's own id, read from the auth context by the
// handler, and the invite names the email it was sent to - so a token that
// leaked to the inviter or a log cannot bind an account of the holder's
// choosing. The recipient is checked before the token is consumed, so a wrong
// account cannot burn the invite for the person it was sent to.
func (s *Service) AcceptInviteForCaller(ctx context.Context, token string, callerUserID int64) error {
	invite, err := s.VerifyInviteToken(ctx, token)
	if err != nil {
		return err
	}

	invited, err := s.repo.GetEmployee(ctx, invite.OwnerID, invite.EmployeeID)
	if err != nil {
		return err
	}
	callerEmail, err := s.repo.GetAccountEmail(ctx, callerUserID)
	if err != nil {
		return err
	}
	if !domain.SameEmailAddress(invited.Email, callerEmail) {
		return domain.ErrInviteRecipientMismatch
	}

	if err := s.repo.AcceptInvite(ctx, invite.OwnerID, invite.EmployeeID, callerUserID, token); err != nil {
		return err
	}
	invalidateHRCoreScope(ctx, invite.OwnerID)
	return nil
}

// --- Locations ---

func (s *Service) ListLocations(ctx context.Context, ownerID int64) ([]domain.LocationRecord, error) {
	return cacher.Of[[]domain.LocationRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrCoreScope, ownerID, "locations"),
		func(ctx context.Context) ([]domain.LocationRecord, error) {
			return s.repo.ListLocations(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
}

func (s *Service) CreateLocation(ctx context.Context, ownerID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.CreateLocation(ctx, ownerID, loc)
}

func (s *Service) UpdateLocation(ctx context.Context, ownerID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.UpdateLocation(ctx, ownerID, loc)
}

// --- Departments ---

func (s *Service) ListDepartments(ctx context.Context, ownerID, locationID int64) ([]domain.DepartmentRecord, error) {
	key := cacheKeys.Scoped(hrCoreScope, ownerID, fmt.Sprintf("depts|l=%d", locationID))
	return cacher.Of[[]domain.DepartmentRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.DepartmentRecord, error) {
			return s.repo.ListDepartments(ctx, ownerID, locationID)
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
}

func (s *Service) CreateDepartment(ctx context.Context, ownerID int64, dept domain.DepartmentRecord) (*domain.DepartmentRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.CreateDepartment(ctx, ownerID, dept)
}

func (s *Service) UpdateDepartment(ctx context.Context, ownerID int64, dept domain.DepartmentRecord) (*domain.DepartmentRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.UpdateDepartment(ctx, ownerID, dept)
}

// --- Work Schedules ---

func (s *Service) ListWorkSchedules(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.WorkScheduleRecord, error) {
	return cacher.Of[[]domain.WorkScheduleRecord]().GetOrComputeSingle(ctx,
		// The page is part of the key. Caching a paginated list under a
		// page-independent key serves page one's rows for every other page.
		cacheKeys.Scoped(hrCoreScope, ownerID, fmt.Sprintf("work_schedules:%d:%d", limit, offset)),
		func(ctx context.Context) ([]domain.WorkScheduleRecord, error) {
			return s.repo.ListWorkSchedules(ctx, ownerID, limit, offset)
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
}

// CountWorkSchedules is deliberately uncached: it is one COUNT behind a
// tenant-filtered index, and a stale total beside a fresh page is worse than
// a cheap query.
func (s *Service) CountWorkSchedules(ctx context.Context, ownerID int64) (int64, error) {
	return s.repo.CountWorkSchedules(ctx, ownerID)
}
func (s *Service) CreateWorkSchedule(ctx context.Context, ownerID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.CreateWorkSchedule(ctx, ownerID, ws)
}

func (s *Service) UpdateWorkSchedule(ctx context.Context, ownerID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.UpdateWorkSchedule(ctx, ownerID, ws)
}

// --- Roles & Permissions ---

func (s *Service) ListRoles(ctx context.Context, ownerID int64) ([]domain.RoleRecord, error) {
	return cacher.Of[[]domain.RoleRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrCoreScope, ownerID, "roles"),
		func(ctx context.Context) ([]domain.RoleRecord, error) {
			return s.repo.ListRoles(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
}

func (s *Service) CreateRole(ctx context.Context, ownerID int64, role domain.RoleRecord) (*domain.RoleRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.CreateRole(ctx, ownerID, role)
}

func (s *Service) UpdateRole(ctx context.Context, ownerID int64, role domain.RoleRecord) (*domain.RoleRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.UpdateRole(ctx, ownerID, role)
}

func (s *Service) AssignRole(ctx context.Context, ownerID, employeeID, roleID int64) error {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.AssignRole(ctx, ownerID, employeeID, roleID)
}

func (s *Service) AssignDefaultOwnerRole(ctx context.Context, ownerID, userID int64) error {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.AssignDefaultOwnerRole(ctx, ownerID, userID)
}

func (s *Service) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return s.repo.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

func (s *Service) HasRole(ctx context.Context, ownerID, employeeID, roleID int64) (bool, error) {
	return s.repo.HasRole(ctx, ownerID, employeeID, roleID)
}

// view_directory is granted to every default role: most businesses want
// every employee to see the company directory, and a business that wants to
// restrict it has no working way to do so yet - the Roles/Permissions admin
// UI collects that choice during setup and silently discards it (see the
// setup wizard's roles step). Restricting this by role is real, separate
// follow-up work, not something to fake here.
func (s *Service) SeedDefaultRoles(ctx context.Context, ownerID int64) error {
	defaultRoles := []domain.RoleRecord{
		{Name: "Employee", Description: "Standard employee with basic access", IsSystem: true, Permissions: mustJSON(map[string]bool{
			domain.PermissionViewOwnProfile: true, domain.PermissionViewOwnSchedule: true, domain.PermissionRequestTimeOff: true,
			domain.PermissionViewDirectory: true,
		})},
		{Name: "Manager", Description: "Team manager with approval rights", IsSystem: true, Permissions: mustJSON(map[string]bool{
			domain.PermissionViewOwnProfile: true, domain.PermissionViewOwnSchedule: true, domain.PermissionRequestTimeOff: true,
			domain.PermissionViewTeam: true, domain.PermissionApproveTimeOff: true, domain.PermissionApproveOvertime: true, domain.PermissionManageSchedules: true,
			domain.PermissionViewDirectory: true,
		})},
		{Name: "HR Admin", Description: "Full HR administration access", IsSystem: true, Permissions: mustJSON(map[string]bool{
			domain.PermissionViewOwnProfile: true, domain.PermissionViewOwnSchedule: true, domain.PermissionRequestTimeOff: true,
			domain.PermissionViewTeam: true, domain.PermissionApproveTimeOff: true, domain.PermissionApproveOvertime: true, domain.PermissionManageSchedules: true,
			domain.PermissionManageEmployees: true, domain.PermissionManageRoles: true, domain.PermissionManageDepartments: true,
			domain.PermissionManageLocations: true, domain.PermissionManageHRConfig: true, domain.PermissionViewReports: true,
			domain.PermissionViewDirectory: true,
		})},
		{Name: "Finance/Payroll Manager", Description: "Payroll and finance operations", IsSystem: true, Permissions: mustJSON(map[string]bool{
			domain.PermissionViewOwnProfile: true, domain.PermissionViewTeam: true, domain.PermissionManagePayroll: true,
			domain.PermissionViewReports: true, domain.PermissionManageBenefits: true,
			domain.PermissionViewDirectory: true,
		})},
		{Name: "Executive", Description: "Executive-level oversight", IsSystem: true, Permissions: mustJSON(map[string]bool{
			domain.PermissionViewOwnProfile: true, domain.PermissionViewTeam: true, domain.PermissionApproveTimeOff: true, domain.PermissionApproveOvertime: true,
			domain.PermissionManageEmployees: true, domain.PermissionManageRoles: true, domain.PermissionManageDepartments: true,
			domain.PermissionManageLocations: true, domain.PermissionManageHRConfig: true, domain.PermissionManagePayroll: true,
			domain.PermissionViewReports: true, domain.PermissionManageBenefits: true, domain.PermissionManageApprovalChains: true,
			domain.PermissionViewDirectory: true,
		})},
	}
	return s.repo.SeedDefaultRoles(ctx, ownerID, defaultRoles)
}

// ListPermissionCatalog returns the full set of fine-grained HR permission
// keys a role can be granted. It is a fixed, in-memory catalog — no DB call.
func (s *Service) ListPermissionCatalog() []domain.PermissionRecord {
	return domain.AllPermissions()
}

// --- Setup Wizard ---

func (s *Service) GetSetupStatus(ctx context.Context, ownerID int64) ([]domain.SetupStepRecord, error) {
	return s.repo.GetSetupSteps(ctx, ownerID)
}

func (s *Service) CompleteSetupStep(ctx context.Context, ownerID int64, stepKey string) (*domain.SetupStepRecord, error) {
	return s.repo.CompleteSetupStep(ctx, ownerID, stepKey)
}

func (s *Service) GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRCompanyConfigRecord, error) {
	return cacher.Of[*domain.HRCompanyConfigRecord]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped(hrCoreScope, ownerID, "hr_config"),
		func(ctx context.Context) (*domain.HRCompanyConfigRecord, error) {
			return s.repo.GetHRConfig(ctx, ownerID)
		},
		cacher.WithExpiry(ctx, 10*time.Minute),
	)
}

func (s *Service) UpsertHRConfig(ctx context.Context, ownerID int64, cfg domain.HRCompanyConfigRecord) (*domain.HRCompanyConfigRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.UpsertHRConfig(ctx, ownerID, cfg)
}

// --- Approval Chains ---

func (s *Service) GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*domain.ApprovalChainRecord, error) {
	key := cacheKeys.Scoped(hrCoreScope, ownerID, fmt.Sprintf("approval|t=%s", requestType))
	return cacher.Of[*domain.ApprovalChainRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (*domain.ApprovalChainRecord, error) {
			return s.repo.GetApprovalChain(ctx, ownerID, requestType)
		},
		cacher.WithExpiry(ctx, hrCoreTTL),
	)
}

func (s *Service) UpsertApprovalChain(ctx context.Context, ownerID int64, chain domain.ApprovalChainRecord) (*domain.ApprovalChainRecord, error) {
	defer invalidateHRCoreScope(ctx, ownerID)
	return s.repo.UpsertApprovalChain(ctx, ownerID, chain)
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
