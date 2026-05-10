package application

import (
	"context"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_core/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	return &Service{repo: infrastructure.NewRepository(ctx)}
}

// --- Employees ---

func (s *Service) ListEmployees(ctx context.Context, ownerID int64, limit, offset int32) ([]domain.EmployeeRecord, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.ListEmployees(ctx, ownerID, domain.ListEmployeesParams{Limit: limit, Offset: offset})
}

func (s *Service) GetEmployee(ctx context.Context, ownerID, employeeID int64) (*domain.EmployeeRecord, error) {
	return s.repo.GetEmployee(ctx, ownerID, employeeID)
}

func (s *Service) CreateEmployee(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) (*domain.EmployeeRecord, error) {
	if err := s.SeedDefaultRoles(ctx, ownerID); err != nil {
		logger.Ctx(ctx).Error("failed to seed default roles during employee creation", logger.Any("owner_id", ownerID), logger.Error(err))
	}

	if emp.Status == "" {
		emp.Status = "active"
	}
	if emp.EmploymentType == "" {
		emp.EmploymentType = "full_time"
	}

	return s.repo.CreateEmployeeWithBalances(ctx, ownerID, emp, []int64{})
}

func (s *Service) UpdateEmployee(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) (*domain.EmployeeRecord, error) {
	return s.repo.UpdateEmployee(ctx, ownerID, emp)
}

// --- Invites ---

func (s *Service) InviteEmployee(ctx context.Context, ownerID int64, emp domain.EmployeeRecord) (string, error) {
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

func (s *Service) AcceptInvite(ctx context.Context, token string, userID int64) error {
	invite, err := s.VerifyInviteToken(ctx, token)
	if err != nil {
		return err
	}

	if err := s.repo.ConsumeInviteToken(ctx, token); err != nil {
		return err
	}

	_, err = s.repo.UpdateEmployee(ctx, invite.OwnerID, domain.EmployeeRecord{
		ID:     invite.EmployeeID,
		UserID: userID,
		Status: "active",
	})
	return err
}

// --- Locations ---

func (s *Service) ListLocations(ctx context.Context, ownerID int64) ([]domain.LocationRecord, error) {
	return s.repo.ListLocations(ctx, ownerID)
}

func (s *Service) CreateLocation(ctx context.Context, ownerID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	return s.repo.CreateLocation(ctx, ownerID, loc)
}

func (s *Service) UpdateLocation(ctx context.Context, ownerID int64, loc domain.LocationRecord) (*domain.LocationRecord, error) {
	return s.repo.UpdateLocation(ctx, ownerID, loc)
}

// --- Departments ---

func (s *Service) ListDepartments(ctx context.Context, ownerID, locationID int64) ([]domain.DepartmentRecord, error) {
	return s.repo.ListDepartments(ctx, ownerID, locationID)
}

func (s *Service) CreateDepartment(ctx context.Context, ownerID int64, dept domain.DepartmentRecord) (*domain.DepartmentRecord, error) {
	return s.repo.CreateDepartment(ctx, ownerID, dept)
}

func (s *Service) UpdateDepartment(ctx context.Context, ownerID int64, dept domain.DepartmentRecord) (*domain.DepartmentRecord, error) {
	return s.repo.UpdateDepartment(ctx, ownerID, dept)
}

// --- Work Schedules ---

func (s *Service) ListWorkSchedules(ctx context.Context, ownerID int64) ([]domain.WorkScheduleRecord, error) {
	return s.repo.ListWorkSchedules(ctx, ownerID)
}

func (s *Service) GetEffectiveSchedule(ctx context.Context, ownerID, employeeID int64) (*domain.WorkScheduleRecord, error) {
	ws, err := s.repo.GetWorkScheduleForEmployee(ctx, ownerID, employeeID)
	if err != nil {
		return nil, err
	}
	if ws != nil {
		return ws, nil
	}

	emp, err := s.repo.GetEmployee(ctx, ownerID, employeeID)
	if err != nil {
		return nil, err
	}

	if emp.DepartmentID > 0 {
		depts, deptErr := s.repo.ListDepartments(ctx, ownerID, 0)
		if deptErr == nil {
			for _, dept := range depts {
				if dept.ID == emp.DepartmentID && emp.LocationID > 0 {
					break
				}
			}
		}
	}

	cfg, err := s.repo.GetHRConfig(ctx, ownerID)
	if err != nil {
		return nil, nil
	}
	if cfg.DefaultWorkScheduleID > 0 {
		schedules, listErr := s.repo.ListWorkSchedules(ctx, ownerID)
		if listErr != nil {
			return nil, nil
		}
		for i, sched := range schedules {
			if sched.ID == cfg.DefaultWorkScheduleID {
				return &schedules[i], nil
			}
		}
	}

	return nil, nil
}

func (s *Service) CreateWorkSchedule(ctx context.Context, ownerID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	return s.repo.CreateWorkSchedule(ctx, ownerID, ws)
}

func (s *Service) UpdateWorkSchedule(ctx context.Context, ownerID int64, ws domain.WorkScheduleRecord) (*domain.WorkScheduleRecord, error) {
	return s.repo.UpdateWorkSchedule(ctx, ownerID, ws)
}

// --- Roles & Permissions ---

func (s *Service) ListRoles(ctx context.Context, ownerID int64) ([]domain.RoleRecord, error) {
	return s.repo.ListRoles(ctx, ownerID)
}

func (s *Service) CreateRole(ctx context.Context, ownerID int64, role domain.RoleRecord) (*domain.RoleRecord, error) {
	return s.repo.CreateRole(ctx, ownerID, role)
}

func (s *Service) UpdateRole(ctx context.Context, ownerID int64, role domain.RoleRecord) (*domain.RoleRecord, error) {
	return s.repo.UpdateRole(ctx, ownerID, role)
}

func (s *Service) AssignRole(ctx context.Context, ownerID, employeeID, roleID int64) error {
	return s.repo.AssignRole(ctx, ownerID, employeeID, roleID)
}

func (s *Service) HasPermission(ctx context.Context, ownerID, employeeID int64, permissionKey string) (bool, error) {
	return s.repo.HasPermission(ctx, ownerID, employeeID, permissionKey)
}

func (s *Service) SeedDefaultRoles(ctx context.Context, ownerID int64) error {
	defaultRoles := []domain.RoleRecord{
		{Name: "Employee", Description: "Standard employee with basic access", IsSystem: true, Permissions: mustJSON(map[string]bool{
			"view_own_profile": true, "view_own_schedule": true, "request_time_off": true,
		})},
		{Name: "Manager", Description: "Team manager with approval rights", IsSystem: true, Permissions: mustJSON(map[string]bool{
			"view_own_profile": true, "view_own_schedule": true, "request_time_off": true,
			"view_team": true, "approve_time_off": true, "manage_schedules": true,
		})},
		{Name: "HR Admin", Description: "Full HR administration access", IsSystem: true, Permissions: mustJSON(map[string]bool{
			"view_own_profile": true, "view_own_schedule": true, "request_time_off": true,
			"view_team": true, "approve_time_off": true, "manage_schedules": true,
			"manage_employees": true, "manage_roles": true, "manage_departments": true,
			"manage_locations": true, "manage_hr_config": true, "view_reports": true,
		})},
		{Name: "Finance/Payroll Manager", Description: "Payroll and finance operations", IsSystem: true, Permissions: mustJSON(map[string]bool{
			"view_own_profile": true, "view_team": true, "manage_payroll": true,
			"view_reports": true, "manage_benefits": true,
		})},
		{Name: "Executive", Description: "Executive-level oversight", IsSystem: true, Permissions: mustJSON(map[string]bool{
			"view_own_profile": true, "view_team": true, "approve_time_off": true,
			"manage_employees": true, "manage_roles": true, "manage_departments": true,
			"manage_locations": true, "manage_hr_config": true, "manage_payroll": true,
			"view_reports": true, "manage_benefits": true, "manage_approval_chains": true,
		})},
	}
	return s.repo.SeedDefaultRoles(ctx, ownerID, defaultRoles)
}

// --- Setup Wizard ---

func (s *Service) GetSetupStatus(ctx context.Context, ownerID int64) ([]domain.SetupStepRecord, error) {
	return s.repo.GetSetupSteps(ctx, ownerID)
}

func (s *Service) CompleteSetupStep(ctx context.Context, ownerID int64, stepKey string) (*domain.SetupStepRecord, error) {
	return s.repo.CompleteSetupStep(ctx, ownerID, stepKey)
}

func (s *Service) GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRCompanyConfigRecord, error) {
	return s.repo.GetHRConfig(ctx, ownerID)
}

func (s *Service) UpsertHRConfig(ctx context.Context, ownerID int64, cfg domain.HRCompanyConfigRecord) (*domain.HRCompanyConfigRecord, error) {
	return s.repo.UpsertHRConfig(ctx, ownerID, cfg)
}

// --- Approval Chains ---

func (s *Service) GetApprovalChain(ctx context.Context, ownerID int64, requestType string) (*domain.ApprovalChainRecord, error) {
	return s.repo.GetApprovalChain(ctx, ownerID, requestType)
}

func (s *Service) UpsertApprovalChain(ctx context.Context, ownerID int64, chain domain.ApprovalChainRecord) (*domain.ApprovalChainRecord, error) {
	return s.repo.UpsertApprovalChain(ctx, ownerID, chain)
}

func mustJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
