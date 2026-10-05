package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrLeaveScope             = "hr_leave"
	hrLeaveTTL               = 2 * time.Minute
	permissionApproveTimeOff = "approve_time_off"
)

func invalidateHRLeaveScope(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrLeaveScope, ownerID))
}

type Service struct {
	repo             repository
	approvalProvider domain.ApprovalChainProvider
	permissions      domain.PermissionProvider
}

func NewService(_ context.Context, repo repository, approvalProvider domain.ApprovalChainProvider, permissions domain.PermissionProvider) *Service {
	return &Service{
		repo:             repo,
		approvalProvider: approvalProvider,
		permissions:      permissions,
	}
}

// --- Leave Types ---

func (s *Service) ListLeaveTypes(ctx context.Context, businessID int64, params domain.ListLeaveTypesParams) ([]domain.LeaveTypeRecord, int64, error) {
	if businessID <= 0 {
		return nil, 0, domain.ErrLeaveAccessDenied
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	type lr struct {
		Types []domain.LeaveTypeRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrLeaveScope, businessID, fmt.Sprintf("types|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (lr, error) {
			types, total, err := s.repo.ListLeaveTypes(ctx, businessID, params)
			if err != nil {
				return lr{}, err
			}
			if total == 0 {
				return lr{}, domain.ErrLeaveNotConfigured
			}
			return lr{Types: types, Total: total}, nil
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
	return res.Types, res.Total, err
}

func (s *Service) CreateLeaveType(ctx context.Context, businessID int64, lt domain.LeaveTypeRecord) (*domain.LeaveTypeRecord, error) {
	if businessID <= 0 {
		return nil, domain.ErrLeaveAccessDenied
	}
	if lt.Name == "" {
		lt.Name = "Unnamed Leave Type"
	}
	lt.AccrualRule = strings.ToLower(strings.TrimSpace(lt.AccrualRule))
	lt.IsActive = true
	if err := domain.ValidateLeaveType(lt); err != nil {
		return nil, err
	}
	rec, err := s.repo.CreateLeaveType(ctx, businessID, lt)
	if err == nil {
		invalidateHRLeaveScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) GetLeaveType(ctx context.Context, businessID, leaveTypeID int64) (*domain.LeaveTypeRecord, error) {
	if businessID <= 0 {
		return nil, domain.ErrLeaveAccessDenied
	}
	return s.repo.GetLeaveType(ctx, businessID, leaveTypeID)
}

func (s *Service) UpdateLeaveType(ctx context.Context, businessID int64, lt domain.LeaveTypeRecord) error {
	if businessID <= 0 {
		return domain.ErrLeaveAccessDenied
	}
	if err := domain.ValidateLeaveType(lt); err != nil {
		return err
	}
	// The repository update is scoped by business and id. The existence probe
	// also turns a foreign or deleted id into the module's not-found error
	// instead of an apparent success.
	if _, err := s.repo.GetLeaveType(ctx, businessID, lt.ID); err != nil {
		return err
	}
	if err := s.repo.UpdateLeaveType(ctx, businessID, lt); err != nil {
		return err
	}
	invalidateHRLeaveScope(ctx, businessID)
	return nil
}

// UpdateLeaveTypePatch applies the generated update contract without turning
// omitted pointer fields into zeroes. The current row is loaded for validation
// and the repository updates only fields whose pointers are present.
func (s *Service) UpdateLeaveTypePatch(ctx context.Context, businessID, leaveTypeID int64, patch domain.LeaveTypePatch) (*domain.LeaveTypeRecord, error) {
	if businessID <= 0 {
		return nil, domain.ErrLeaveAccessDenied
	}
	if leaveTypeID <= 0 || patch.Empty() {
		return nil, domain.ErrEmptyLeaveTypeUpdate
	}
	current, err := s.repo.GetLeaveType(ctx, businessID, leaveTypeID)
	if err != nil {
		return nil, err
	}
	if patch.Name != nil {
		name := strings.TrimSpace(*patch.Name)
		patch.Name = &name
	}
	if patch.AccrualRule != nil {
		rule := strings.ToLower(strings.TrimSpace(*patch.AccrualRule))
		patch.AccrualRule = &rule
	}
	updated := domain.ApplyLeaveTypePatch(*current, patch)
	updated.Name = strings.TrimSpace(updated.Name)
	updated.AccrualRule = strings.ToLower(strings.TrimSpace(updated.AccrualRule))
	if err := domain.ValidateLeaveType(updated); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateLeaveTypePatch(ctx, businessID, leaveTypeID, patch); err != nil {
		return nil, err
	}
	invalidateHRLeaveScope(ctx, businessID)
	return s.repo.GetLeaveType(ctx, businessID, leaveTypeID)
}

// --- Leave Requests ---

func (s *Service) CreateLeaveRequest(ctx context.Context, ownerID, callerUserID int64, req domain.LeaveRequestRecord) (*domain.LeaveRequestRecord, error) {
	if err := s.authorizeLeaveSubject(ctx, ownerID, callerUserID, req.EmployeeID); err != nil {
		return nil, err
	}

	// Validate leave type exists
	leaveTypes, _, err := s.repo.ListLeaveTypes(ctx, ownerID, domain.ListLeaveTypesParams{Limit: 1, Offset: 0})
	if err != nil {
		return nil, err
	}
	if len(leaveTypes) == 0 {
		return nil, domain.ErrLeaveNotConfigured
	}

	lt, err := s.repo.GetLeaveType(ctx, ownerID, req.LeaveTypeID)
	if err != nil {
		return nil, err
	}
	// ListLeaveTypes above only proves that the business has at least one
	// active policy. The selected type is loaded by id, so enforce its own
	// availability before any balance or overlap mutation.
	if !lt.IsActive {
		return nil, domain.ErrLeaveTypeNotFound
	}

	// Validate date range
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, domain.ErrInvalidDateRange
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, domain.ErrInvalidDateRange
	}
	if endDate.Before(startDate) {
		return nil, domain.ErrInvalidDateRange
	}

	// Validate min notice
	if lt.MinNoticeDays > 0 {
		minStart := time.Now().UTC().AddDate(0, 0, int(lt.MinNoticeDays))
		if startDate.Before(minStart) {
			return nil, domain.ErrInsufficientNotice
		}
	}

	allocations := leaveDayAllocations(startDate, endDate)
	req.TotalDays = totalAllocatedDays(allocations)

	req.Status = "pending"
	req.CurrentApprovalLevel = 0

	created, err := s.repo.CreateLeaveRequestWithBalance(ctx, ownerID, req, allocations)
	if err != nil {
		return nil, err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return created, nil
}

type repository interface {
	AddToEntitledDays(ctx context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64, month int32) error
	AdvanceLeaveApprovalLevel(ctx context.Context, ownerID, requestID int64, currentLevel int32) error
	ApplyYearEndCarryover(ctx context.Context, ownerID int64, previousYear, currentYear int32) error
	ApproveLeaveRequestWithAllocations(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation) (err error)
	CancelLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation) (err error)
	CountLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) (int64, error)
	CreateLeaveRequestWithBalance(ctx context.Context, ownerID int64, req domain.LeaveRequestRecord, allocations []domain.LeaveDayAllocation) (created *domain.LeaveRequestRecord, err error)
	CreateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) (*domain.LeaveTypeRecord, error)
	GetLeaveRequest(ctx context.Context, ownerID, requestID int64) (*domain.LeaveRequestRecord, error)
	GetLeaveType(ctx context.Context, ownerID, leaveTypeID int64) (*domain.LeaveTypeRecord, error)
	ListAllActiveLeaveTypesForOwner(ctx context.Context, ownerID int64) ([]domain.LeaveTypeRecord, error)
	ListBalancesForAccrual(ctx context.Context, ownerID int64, year, currentMonth int32) ([]domain.LeaveBalanceRecord, error)
	ListDistinctOwnerIDs(ctx context.Context) ([]int64, error)
	ListLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error)
	ListLeaveCalendar(ctx context.Context, businessID, locationID int64, dateFrom, dateTo string) ([]domain.LeaveRequestRecord, error)
	ListLeaveRequests(ctx context.Context, ownerID int64, params domain.ListLeaveRequestsParams) ([]domain.LeaveRequestRecord, int64, error)
	ListLeaveTypes(ctx context.Context, ownerID int64, params domain.ListLeaveTypesParams) ([]domain.LeaveTypeRecord, int64, error)
	RejectLeaveRequestWithBalance(ctx context.Context, ownerID, requestID, employeeID, leaveTypeID int64, allocations []domain.LeaveDayAllocation, reason string) (err error)
	UpdateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) error
	UpdateLeaveTypePatch(ctx context.Context, ownerID, leaveTypeID int64, patch domain.LeaveTypePatch) error
}
