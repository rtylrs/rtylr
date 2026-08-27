package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
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
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
	permissions      domain.PermissionProvider
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider, permissions domain.PermissionProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
		permissions:      permissions,
	}
}

// --- Leave Types ---

func (s *Service) ListLeaveTypes(ctx context.Context, ownerID int64, params domain.ListLeaveTypesParams) ([]domain.LeaveTypeRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	type lr struct {
		Types []domain.LeaveTypeRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrLeaveScope, ownerID, fmt.Sprintf("types|l=%d|o=%d", params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			types, total, err := s.repo.ListLeaveTypes(ctx, ownerID, params)
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

func (s *Service) CreateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) (*domain.LeaveTypeRecord, error) {
	if lt.Name == "" {
		lt.Name = "Unnamed Leave Type"
	}
	lt.IsActive = true
	rec, err := s.repo.CreateLeaveType(ctx, ownerID, lt)
	if err == nil {
		invalidateHRLeaveScope(ctx, ownerID)
	}
	return rec, err
}

func (s *Service) UpdateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) error {
	if err := s.repo.UpdateLeaveType(ctx, ownerID, lt); err != nil {
		return err
	}
	invalidateHRLeaveScope(ctx, ownerID)
	return nil
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

func (s *Service) ListLeaveRequests(ctx context.Context, ownerID, callerUserID int64, params domain.ListLeaveRequestsParams) ([]domain.LeaveRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	effectiveEmployeeID, err := s.scopeLeaveRequests(ctx, ownerID, callerUserID, params.EmployeeID)
	if err != nil {
		return nil, 0, err
	}
	params.EmployeeID = effectiveEmployeeID
	params.Status = strings.ToLower(strings.TrimSpace(params.Status))
	type lr struct {
		Items []domain.LeaveRequestRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrLeaveScope, ownerID, fmt.Sprintf("reqs|e=%d|s=%s|f=%s|t=%s|l=%d|o=%d", params.EmployeeID, params.Status, derefStr(params.FromDate), derefStr(params.ToDate), params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListLeaveRequests(ctx, ownerID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) ApproveLeave(ctx context.Context, ownerID, requestID, approverUserID int64) error {
	req, lt, err := s.authorizedLeaveDecision(ctx, ownerID, requestID, approverUserID)
	if err != nil {
		return err
	}

	if lt.RequiresApproval {
		chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, ownerID, "leave")
		if chainErr != nil {
			return chainErr
		}
		if chain == nil {
			return domain.ErrApprovalChainNotConfigured
		}

		if req.CurrentApprovalLevel+1 < chain.TotalLevels {
			if err := s.repo.AdvanceLeaveApprovalLevel(ctx, ownerID, requestID, req.CurrentApprovalLevel); err != nil {
				return err
			}
			invalidateHRLeaveScope(ctx, ownerID)
			return nil
		}
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return parseErr
	}

	// Final approval atomically moves the request from pending to approved and
	// moves every year's allocation from pending to used.
	if err := s.repo.ApproveLeaveRequestWithAllocations(ctx, ownerID, requestID, req.EmployeeID, req.LeaveTypeID, leaveDayAllocations(startDate, endDate)); err != nil {
		return err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return nil
}

func (s *Service) RejectLeave(ctx context.Context, ownerID, requestID, approverUserID int64, reason string) error {
	req, _, err := s.authorizedLeaveDecision(ctx, ownerID, requestID, approverUserID)
	if err != nil {
		return err
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return parseErr
	}
	if err := s.repo.RejectLeaveRequestWithBalance(ctx, ownerID, requestID, req.EmployeeID, req.LeaveTypeID, leaveDayAllocations(startDate, endDate), reason); err != nil {
		return err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return nil
}

// CancelLeaveRequest withdraws the caller's own pending leave request.
// Unlike ApproveLeave/RejectLeave, which are an approver's decision over
// someone else's request, this is self-service: only the request's own
// subject may cancel it - the leave approval permission grants no override
// here - and only while it is still pending. Once a request has been
// decided, undoing it can have payroll or scheduling consequences a bare
// status flip should not trigger silently; that path is a request to a
// manager, not a self-service button.
func (s *Service) CancelLeaveRequest(ctx context.Context, ownerID, callerUserID, requestID int64) error {
	if s.permissions == nil || callerUserID <= 0 {
		return domain.ErrLeaveAccessDenied
	}
	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil || employeeID <= 0 {
		return domain.ErrLeaveAccessDenied
	}

	req, err := s.repo.GetLeaveRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}
	if req.EmployeeID != employeeID {
		return domain.ErrLeaveAccessDenied
	}
	if req.Status != domain.LeaveStatusPending {
		return domain.ErrLeaveRequestNotPending
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return parseErr
	}
	if err := s.repo.CancelLeaveRequestWithBalance(ctx, ownerID, requestID, req.EmployeeID, req.LeaveTypeID, leaveDayAllocations(startDate, endDate)); err != nil {
		return err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return nil
}

// scopeLeaveRequests resolves the authenticated user's employee row before
// applying any client-supplied filter. A caller with the leave approval
// permission may request the team feed; every other caller is forced to their
// own employee row, even when employee_id is supplied by the client.
func (s *Service) scopeLeaveRequests(ctx context.Context, ownerID, callerUserID, requestedEmployeeID int64) (int64, error) {
	if s.permissions == nil || callerUserID <= 0 {
		return 0, domain.ErrLeaveAccessDenied
	}

	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil || employeeID <= 0 {
		return 0, domain.ErrLeaveAccessDenied
	}

	canViewTeam, permissionErr := s.permissions.HasPermission(ctx, ownerID, employeeID, permissionApproveTimeOff)
	if permissionErr == nil && canViewTeam {
		return requestedEmployeeID, nil
	}
	return employeeID, nil
}

// authorizeLeaveSubject verifies the caller may act on employeeID's leave -
// either it is their own, or they hold the leave approval permission.
// Unlike scopeLeaveRequests, which narrows an unauthorized list to the
// caller's own view rather than fail the read, a write silently substituting
// a different employee than the one requested would file the wrong person's
// leave without anyone noticing - so this denies outright instead.
func (s *Service) authorizeLeaveSubject(ctx context.Context, ownerID, callerUserID, requestedEmployeeID int64) error {
	if s.permissions == nil || callerUserID <= 0 {
		return domain.ErrLeaveAccessDenied
	}

	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil || employeeID <= 0 {
		return domain.ErrLeaveAccessDenied
	}
	if employeeID == requestedEmployeeID {
		return nil
	}

	canActForOthers, permissionErr := s.permissions.HasPermission(ctx, ownerID, employeeID, permissionApproveTimeOff)
	if permissionErr == nil && canActForOthers {
		return nil
	}
	return domain.ErrLeaveAccessDenied
}

// authorizedLeaveDecision performs the permission and approval-chain checks
// before loading the requested row. This keeps an unauthorised caller from
// using a request id to probe another employee's leave data or mutate it.
func (s *Service) authorizedLeaveDecision(ctx context.Context, ownerID, requestID, approverUserID int64) (*domain.LeaveRequestRecord, *domain.LeaveTypeRecord, error) {
	if s.permissions == nil || approverUserID <= 0 {
		return nil, nil, domain.ErrLeaveAccessDenied
	}

	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, approverUserID)
	if err != nil || employeeID <= 0 {
		return nil, nil, domain.ErrLeaveAccessDenied
	}
	allowed, err := s.permissions.HasPermission(ctx, ownerID, employeeID, permissionApproveTimeOff)
	if err != nil {
		return nil, nil, err
	}
	if !allowed {
		return nil, nil, domain.ErrLeaveApproverNotAllowed
	}

	req, err := s.repo.GetLeaveRequest(ctx, ownerID, requestID)
	if err != nil {
		return nil, nil, err
	}
	if req.EmployeeID == employeeID {
		return nil, nil, domain.ErrLeaveSelfApproval
	}
	lt, err := s.repo.GetLeaveType(ctx, ownerID, req.LeaveTypeID)
	if err != nil {
		return nil, nil, err
	}

	if !lt.RequiresApproval {
		return req, lt, nil
	}
	if s.approvalProvider == nil {
		return nil, nil, domain.ErrApprovalChainNotConfigured
	}
	chain, err := s.approvalProvider.GetApprovalChain(ctx, ownerID, "leave")
	if err != nil {
		return nil, nil, err
	}
	if chain == nil {
		return nil, nil, domain.ErrApprovalChainNotConfigured
	}
	if len(chain.Levels) == 0 {
		return req, lt, nil
	}

	matched, err := s.approverMatchesChain(ctx, ownerID, employeeID, req.CurrentApprovalLevel, chain.Levels)
	if err != nil {
		return nil, nil, err
	}
	if !matched {
		return nil, nil, domain.ErrLeaveApproverNotAllowed
	}
	return req, lt, nil
}

func (s *Service) approverMatchesChain(ctx context.Context, ownerID, employeeID int64, currentLevel int32, levels []domain.ApprovalLevelInfo) (bool, error) {
	if len(levels) == 0 {
		return true, nil
	}

	targetLevel := currentLevel + 1
	levelIndex := -1
	for index, level := range levels {
		if level.LevelOrder == targetLevel {
			levelIndex = index
			break
		}
	}
	// Older approval-chain rows may not have a level_order value. The stored
	// order is still represented by the slice returned from SQL, so use the
	// current zero-based position as a backwards-compatible fallback.
	if levelIndex < 0 && currentLevel >= 0 && int(currentLevel) < len(levels) {
		levelIndex = int(currentLevel)
	}
	if levelIndex < 0 {
		return false, nil
	}

	level := levels[levelIndex]
	if level.ApproverEmployeeID > 0 && level.ApproverEmployeeID == employeeID {
		return true, nil
	}
	if level.ApproverRoleID <= 0 {
		return false, nil
	}
	return s.permissions.HasRole(ctx, ownerID, employeeID, level.ApproverRoleID)
}

// GetLeaveBalances resolves which employee's balances the caller may see the
// same way ListLeaveRequests does: an ordinary caller is narrowed to their
// own balances regardless of the requested employeeID, and only a caller
// holding the approval permission sees another employee's. Balances are
// financial-adjacent data (days entitled/used/remaining), so this was never
// safe to leave open the way it was found.
func (s *Service) GetLeaveBalances(ctx context.Context, ownerID, callerUserID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error) {
	scopedEmployeeID, err := s.scopeLeaveRequests(ctx, ownerID, callerUserID, employeeID)
	if err != nil {
		return nil, err
	}

	key := cacheKeys.Scoped(hrLeaveScope, ownerID, fmt.Sprintf("balances|e=%d|y=%d", scopedEmployeeID, year))
	return cacher.Of[[]domain.LeaveBalanceRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LeaveBalanceRecord, error) {
			return s.fetchBalances(ctx, ownerID, scopedEmployeeID, year)
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
}

func (s *Service) fetchBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error) {
	if year == 0 {
		year = int32(time.Now().UTC().Year())
	}
	return s.repo.ListLeaveBalances(ctx, ownerID, employeeID, year)
}

func (s *Service) GetLeaveCalendar(ctx context.Context, ownerID int64, params domain.LeaveCalendarParams) ([]domain.LeaveRequestRecord, error) {
	key := cacheKeys.Scoped(hrLeaveScope, ownerID, fmt.Sprintf("calendar|f=%s|t=%s", params.DateFrom, params.DateTo))
	return cacher.Of[[]domain.LeaveRequestRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LeaveRequestRecord, error) {
			return s.repo.ListLeaveCalendar(ctx, ownerID, params.DateFrom, params.DateTo)
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
}

// --- Background Job: Monthly Accrual ---

func (s *Service) RunMonthlyAccrual(ctx context.Context) error {
	now := time.Now().UTC()
	currentMonth := int32(now.Month())
	currentYear := int32(now.Year())

	ownerIDs, err := s.repo.ListDistinctOwnerIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("leave accrual: failed to list owner IDs", logger.Error(err))
		return err
	}

	for _, ownerID := range ownerIDs {
		if err := s.accrueForOwner(ctx, ownerID, currentYear, currentMonth); err != nil {
			logger.Ctx(ctx).Error("leave accrual: failed for owner",
				logger.Any("owner_id", ownerID), logger.Error(err))
		}
	}

	return nil
}

func (s *Service) accrueForOwner(ctx context.Context, ownerID int64, year, month int32) error {
	if month == 1 {
		if err := s.repo.ApplyYearEndCarryover(ctx, ownerID, year-1, year); err != nil {
			return err
		}
	}
	leaveTypes, err := s.repo.ListAllActiveLeaveTypesForOwner(ctx, ownerID)
	if err != nil {
		return err
	}

	// The accrual balances do not depend on the leave type, so fetch them once
	// and group by leave type ID instead of re-querying per leave type.
	balances, balErr := s.repo.ListBalancesForAccrual(ctx, ownerID, year, month)
	if balErr != nil {
		logger.Ctx(ctx).Error("leave accrual: failed to list balances",
			logger.Any("owner_id", ownerID), logger.Error(balErr))
		return balErr
	}

	balancesByType := make(map[int64][]domain.LeaveBalanceRecord, len(balances))
	for _, bal := range balances {
		balancesByType[bal.LeaveTypeID] = append(balancesByType[bal.LeaveTypeID], bal)
	}

	for _, lt := range leaveTypes {
		if lt.AccrualRule == "" || lt.AccrualRule == "none" || lt.AccrualAmount <= 0 {
			continue
		}

		if lt.AccrualRule != "monthly" {
			continue
		}

		for _, bal := range balancesByType[lt.ID] {
			addErr := s.repo.AddToEntitledDays(ctx, ownerID, bal.EmployeeID, lt.ID, year, lt.AccrualAmount, month)
			if addErr != nil {
				logger.Ctx(ctx).Error("leave accrual: failed to add entitled days",
					logger.Any("owner_id", ownerID), logger.Any("employee_id", bal.EmployeeID),
					logger.Any("leave_type_id", lt.ID), logger.Error(addErr))
			}
		}
	}

	return nil
}

// leaveDayAllocations computes the authoritative whole-calendar-day count
// from the dates. The current API has no half-day or work-calendar fields, so
// a client-supplied total is deliberately ignored rather than trusted.
func leaveDayAllocations(start, end time.Time) []domain.LeaveDayAllocation {
	allocations := make([]domain.LeaveDayAllocation, 0, 2)
	for current := start; !current.After(end); {
		yearEnd := time.Date(current.Year(), time.December, 31, 0, 0, 0, 0, time.UTC)
		segmentEnd := yearEnd
		if end.Before(segmentEnd) {
			segmentEnd = end
		}
		days := float64(int(segmentEnd.Sub(current).Hours()/24) + 1)
		allocations = append(allocations, domain.LeaveDayAllocation{Year: int32(current.Year()), Days: days})
		current = segmentEnd.AddDate(0, 0, 1)
	}
	return allocations
}

func totalAllocatedDays(allocations []domain.LeaveDayAllocation) float64 {
	var total float64
	for _, allocation := range allocations {
		total += allocation.Days
	}
	return total
}

func derefStr(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
