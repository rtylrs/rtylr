package application

import (
	"context"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

func (s *Service) DeleteTimeEntry(ctx context.Context, businessID, callerUserID, entryID int64) error {
	existing, err := s.repo.GetTimeEntry(ctx, businessID, entryID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrTimeEntryNotFound
	}
	if err := s.authorizeTimeEntryTarget(ctx, businessID, callerUserID, existing.EmployeeID); err != nil {
		return err
	}
	locked, err := s.periodLocked(ctx, businessID, existing.EmployeeID, existing.Date)
	if err != nil {
		return err
	}
	if locked {
		return domain.ErrDateInLockedPeriod
	}
	if err := s.repo.DeleteTimeEntry(ctx, businessID, entryID); err != nil {
		return err
	}
	invalidateHRTimeScope(ctx, businessID)
	return nil
}

// periodLocked reports whether the employee's timesheet covering the month
// of date (format "2006-01-02") is locked.
func (s *Service) periodLocked(ctx context.Context, businessID, employeeID int64, date string) (bool, error) {
	if len(date) < 7 {
		return false, nil
	}
	ts, err := s.repo.GetTimesheetForPeriod(ctx, businessID, employeeID, date[:7]+"-01")
	if err != nil {
		return false, err
	}
	return ts != nil && ts.Status == "locked", nil
}

func normalizeListTimeEntriesParams(params domain.ListTimeEntriesParams) domain.ListTimeEntriesParams {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	return params
}

// cacheKeyValue renders one optional filter for a cache key. The filters are
// pointers so "absent" is distinguishable from a zero value; formatting the
// pointer itself (%d) printed the address, which made every combination
// without that filter collide on one key. "-" is the nil marker and strings
// are quoted so a value cannot masquerade as another field.
func cacheKeyValue[T any](value *T) string {
	if value == nil {
		return "-"
	}
	if str, ok := any(*value).(string); ok {
		return strconv.Quote(str)
	}
	return fmt.Sprint(*value)
}

func listTimeEntriesCacheKey(businessID int64, params domain.ListTimeEntriesParams) string {
	return cacheKeys.Scoped(hrTimeScope, businessID, fmt.Sprintf(
		"entries|l=%d|o=%d|e=%s|from=%s|to=%s|p=%s|billable=%s|status=%s",
		params.Limit,
		params.Offset,
		cacheKeyValue(params.EmployeeID),
		cacheKeyValue(params.FromDate),
		cacheKeyValue(params.ToDate),
		cacheKeyValue(params.ProjectID),
		cacheKeyValue(params.IsBillable),
		cacheKeyValue(params.Status),
	))
}

func listTimesheetsCacheKey(businessID int64, params domain.ListTimesheetsParams) string {
	return cacheKeys.Scoped(hrTimeScope, businessID, fmt.Sprintf(
		"ts_list|e=%s|s=%s|l=%d|o=%d",
		cacheKeyValue(params.EmployeeID),
		cacheKeyValue(params.Status),
		params.Limit,
		params.Offset,
	))
}

func (s *Service) ListTimeEntries(ctx context.Context, businessID int64, params domain.ListTimeEntriesParams) ([]domain.TimeEntryRecord, int64, error) {
	params = normalizeListTimeEntriesParams(params)
	type lr struct {
		Items []domain.TimeEntryRecord
		Total int64
	}
	key := listTimeEntriesCacheKey(businessID, params)
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListTimeEntries(ctx, businessID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrTimeTTL),
	)
	return res.Items, res.Total, err
}

// --- Timesheets ---

func (s *Service) GetOrCreateTimesheet(ctx context.Context, businessID, employeeID int64, periodStart, periodEnd string) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheetForPeriod(ctx, businessID, employeeID, periodStart)
	if err != nil {
		return nil, err
	}
	if ts != nil {
		return ts, nil
	}

	rec, err := s.repo.UpsertTimesheet(ctx, businessID, domain.TimesheetRecord{
		EmployeeID:  employeeID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      domain.TimesheetStatusDraft,
	})
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) GetOrCreateTimesheetForCaller(ctx context.Context, businessID, callerUserID, employeeID int64, periodStart, periodEnd string) (*domain.TimesheetRecord, error) {
	if err := s.authorizeTimeEntryTarget(ctx, businessID, callerUserID, employeeID); err != nil {
		return nil, domain.ErrTimesheetAccessDenied
	}
	return s.GetOrCreateTimesheet(ctx, businessID, employeeID, periodStart, periodEnd)
}

func (s *Service) GetTimesheetForPeriod(ctx context.Context, businessID, employeeID int64, periodStart string) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheetForPeriod(ctx, businessID, employeeID, periodStart)
	if err != nil {
		return nil, err
	}
	if ts == nil {
		return nil, domain.ErrTimesheetNotFound
	}
	return ts, nil
}

func (s *Service) GetTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error) {
	key := cacheKeys.Scoped(hrTimeScope, businessID, fmt.Sprintf("ts=%d", timesheetID))
	return cacher.Of[*domain.TimesheetRecord]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (*domain.TimesheetRecord, error) {
			return s.repo.GetTimesheet(ctx, businessID, timesheetID)
		},
		cacher.WithExpiry(ctx, hrTimeTTL),
	)
}

func (s *Service) callerTimesheetEmployeeID(ctx context.Context, businessID, callerUserID int64) (int64, error) {
	employeeID, err := s.callerEmployeeID(ctx, businessID, callerUserID)
	if err != nil {
		return 0, domain.ErrTimesheetAccessDenied
	}
	return employeeID, nil
}

func (s *Service) canViewAllTimesheets(ctx context.Context, businessID, employeeID int64) (bool, error) {
	allowed, err := s.permissions.HasPermission(ctx, businessID, employeeID, domain.PermissionApproveTimeOff)
	if err != nil || allowed {
		return allowed, err
	}
	return s.permissions.HasPermission(ctx, businessID, employeeID, domain.PermissionManageEmployees)
}

func (s *Service) GetTimesheetForCaller(ctx context.Context, businessID, callerUserID, timesheetID int64) (*domain.TimesheetRecord, error) {
	ts, err := s.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}
	employeeID, err := s.callerTimesheetEmployeeID(ctx, businessID, callerUserID)
	if err != nil {
		return nil, err
	}
	if ts.EmployeeID == employeeID {
		return ts, nil
	}
	allowed, err := s.canViewAllTimesheets(ctx, businessID, employeeID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrTimesheetAccessDenied
	}
	return ts, nil
}
