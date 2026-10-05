package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

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
	res, err := cacher.Of[result]().GetOrComputeSingle(ctx, key+"|domain=v2",
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
	return cacher.Of[[]domain.LeaveRequestRecord]().GetOrComputeSingle(ctx, key+"|domain=v2",
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
