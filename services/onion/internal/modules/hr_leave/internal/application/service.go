package application

import (
	"context"
	"database/sql"
	"errors"
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
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
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
func (s *Service) GetLeaveBalances(ctx context.Context, businessID, callerUserID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error) {
	balances, _, err := s.GetLeaveBalancesWithTotal(ctx, businessID, callerUserID, employeeID, year)
	return balances, err
}

func (s *Service) GetLeaveBalancesWithTotal(ctx context.Context, businessID, callerUserID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, int64, error) {
	if businessID <= 0 {
		return nil, 0, domain.ErrLeaveAccessDenied
	}
	scopedEmployeeID, err := s.scopeLeaveRequests(ctx, businessID, callerUserID, employeeID)
	if err != nil {
		return nil, 0, err
	}
	if year == 0 {
		year = int32(time.Now().UTC().Year()) // #nosec G115 -- calendar years returned by the application clock fit the int32 contract field.
	}

	key := cacheKeys.Scoped(hrLeaveScope, businessID, fmt.Sprintf("balances|v2|e=%d|y=%d", scopedEmployeeID, year))
	type result struct {
		Items []domain.LeaveBalanceRecord
		Total int64
	}
	res, err := cacher.Of[result]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (result, error) {
			items, listErr := s.repo.ListLeaveBalances(ctx, businessID, scopedEmployeeID, year)
			if listErr != nil {
				return result{}, listErr
			}
			total, countErr := s.repo.CountLeaveBalances(ctx, businessID, scopedEmployeeID, year)
			if countErr != nil {
				return result{}, countErr
			}
			return result{Items: items, Total: total}, nil
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) GetLeaveCalendar(ctx context.Context, businessID int64, params domain.LeaveCalendarParams) ([]domain.LeaveRequestRecord, error) {
	if businessID <= 0 {
		return nil, domain.ErrLeaveAccessDenied
	}
	normalized, err := normalizeLeaveCalendarParams(params)
	if err != nil {
		return nil, err
	}
	key := leaveCalendarCacheKey(businessID, normalized)
	return cacher.Of[[]domain.LeaveRequestRecord]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) ([]domain.LeaveRequestRecord, error) {
			return s.repo.ListLeaveCalendar(ctx, businessID, normalized.LocationID, normalized.DateFrom, normalized.DateTo)
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
}

func leaveCalendarCacheKey(businessID int64, params domain.LeaveCalendarParams) string {
	return cacheKeys.Scoped(hrLeaveScope, businessID, fmt.Sprintf("calendar|l=%d|f=%s|t=%s", params.LocationID, params.DateFrom, params.DateTo))
}

func normalizeLeaveCalendarParams(params domain.LeaveCalendarParams) (domain.LeaveCalendarParams, error) {
	// Location 0 is the documented unfiltered value; negative values are
	// rejected before they can become a cache key or repository parameter.
	if params.LocationID < 0 {
		return domain.LeaveCalendarParams{}, domain.ErrInvalidCalendarLocation
	}
	from := strings.TrimSpace(params.DateFrom)
	to := strings.TrimSpace(params.DateTo)
	if err := domain.ValidateDateRange(from, to); err != nil {
		return domain.LeaveCalendarParams{}, err
	}
	return domain.LeaveCalendarParams{
		LocationID: params.LocationID,
		DateFrom:   from,
		DateTo:     to,
	}, nil
}

// --- Background Job: Monthly Accrual ---

func (s *Service) RunMonthlyAccrual(ctx context.Context) error {
	now := time.Now().UTC()
	currentMonth := int32(now.Month()) // #nosec G115 -- time.Month is bounded to the twelve calendar months.
	currentYear := int32(now.Year())   // #nosec G115 -- calendar years returned by the application clock fit the int32 contract field.

	ownerIDs, err := s.repo.ListDistinctOwnerIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("leave accrual: failed to list owner IDs", logger.Error(err))
		return err
	}

	var runErrs []error
	for _, ownerID := range ownerIDs {
		if err := s.accrueForOwner(ctx, ownerID, currentYear, currentMonth); err != nil {
			wrappedErr := fmt.Errorf("leave accrual failed for owner %d: %w", ownerID, err)
			runErrs = append(runErrs, wrappedErr)
			logger.Ctx(ctx).Error("leave accrual: failed for owner",
				logger.Any("owner_id", ownerID), logger.Error(wrappedErr))
		}
	}

	return errors.Join(runErrs...)
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

	var accrualErrs []error
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
				wrappedErr := fmt.Errorf(
					"add entitled days for owner %d, employee %d, leave type %d: %w",
					ownerID, bal.EmployeeID, lt.ID, addErr,
				)
				accrualErrs = append(accrualErrs, wrappedErr)
				logger.Ctx(ctx).Error("leave accrual: failed to add entitled days",
					logger.Any("owner_id", ownerID), logger.Any("employee_id", bal.EmployeeID),
					logger.Any("leave_type_id", lt.ID), logger.Error(wrappedErr))
			}
		}
	}

	return errors.Join(accrualErrs...)
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
		allocations = append(allocations, domain.LeaveDayAllocation{Year: int32(current.Year()), Days: days}) // #nosec G115 -- leave dates are validated application dates and fit the int32 contract field.
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
