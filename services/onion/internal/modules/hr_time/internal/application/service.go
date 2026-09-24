package application

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	hrTimeScope = "hr_time"
	hrTimeTTL   = 2 * time.Minute
)

func invalidateHRTimeScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrTimeScope, businessID))
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

func (s *Service) callerEmployeeID(ctx context.Context, businessID, callerUserID int64) (int64, error) {
	if s.permissions == nil || businessID <= 0 || callerUserID <= 0 {
		return 0, domain.ErrTimeEntryAccessDenied
	}
	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, businessID, callerUserID)
	if err != nil || employeeID <= 0 {
		return 0, domain.ErrTimeEntryAccessDenied
	}
	return employeeID, nil
}

func (s *Service) authorizeTimeEntryTarget(ctx context.Context, businessID, callerUserID, targetEmployeeID int64) error {
	callerEmployeeID, err := s.callerEmployeeID(ctx, businessID, callerUserID)
	if err != nil {
		return err
	}
	if targetEmployeeID == callerEmployeeID {
		return nil
	}

	allowed, err := s.permissions.HasPermission(ctx, businessID, callerEmployeeID, domain.PermissionManageEmployees)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrTimeEntryAccessDenied
	}
	targetBusinessID, err := s.permissions.GetEmployeeBusinessID(ctx, businessID, targetEmployeeID)
	if err != nil || targetBusinessID != businessID {
		return domain.ErrTimeEntryAccessDenied
	}
	return nil
}

// --- Time Entries ---

func (s *Service) CreateTimeEntry(ctx context.Context, businessID, callerUserID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	if entry.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	if _, err := time.Parse(time.DateOnly, entry.Date); err != nil {
		return nil, err
	}
	if err := s.authorizeTimeEntryTarget(ctx, businessID, callerUserID, entry.EmployeeID); err != nil {
		return nil, err
	}

	ts, err := s.repo.GetTimesheetForPeriod(ctx, businessID, entry.EmployeeID, entry.Date[:7]+"-01")
	if err != nil {
		return nil, err
	}
	if ts != nil && ts.Status == domain.TimesheetStatusLocked {
		return nil, domain.ErrDateInLockedPeriod
	}

	if entry.Status == "" {
		entry.Status = domain.EntryStatusDraft
	}
	if entry.Source == "" {
		entry.Source = domain.EntrySourceManual
	}

	rec, err := s.repo.CreateTimeEntry(ctx, businessID, entry)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func applyTimeEntryPatch(entry *domain.TimeEntryRecord, patch domain.TimeEntryPatch) {
	if patch.EmployeeID != nil {
		entry.EmployeeID = *patch.EmployeeID
	}
	if patch.ProjectID != nil {
		entry.ProjectID = *patch.ProjectID
	}
	if patch.TaskID != nil {
		entry.TaskID = *patch.TaskID
	}
	if patch.CategoryID != nil {
		entry.CategoryID = *patch.CategoryID
	}
	if patch.Date != nil {
		entry.Date = *patch.Date
	}
	if patch.DurationMinutes != nil {
		entry.DurationMinutes = *patch.DurationMinutes
	}
	if patch.StartTime != nil {
		entry.StartTime = *patch.StartTime
	}
	if patch.EndTime != nil {
		entry.EndTime = *patch.EndTime
	}
	if patch.Description != nil {
		entry.Description = *patch.Description
	}
	if patch.IsBillable != nil {
		entry.IsBillable = *patch.IsBillable
	}
	if patch.Source != nil {
		entry.Source = *patch.Source
	}
}

func (s *Service) UpdateTimeEntry(ctx context.Context, businessID, callerUserID, entryID int64, patch domain.TimeEntryPatch) (*domain.TimeEntryRecord, error) {
	if patch.DurationMinutes != nil && *patch.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	if patch.EmployeeID != nil && *patch.EmployeeID <= 0 {
		return nil, domain.ErrTimeEntryAccessDenied
	}
	if patch.Date != nil {
		if _, err := time.Parse(time.DateOnly, *patch.Date); err != nil {
			return nil, err
		}
	}

	existing, err := s.repo.GetTimeEntry(ctx, businessID, entryID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrTimeEntryNotFound
	}
	if err := s.authorizeTimeEntryTarget(ctx, businessID, callerUserID, existing.EmployeeID); err != nil {
		return nil, err
	}

	merged := *existing
	applyTimeEntryPatch(&merged, patch)
	if merged.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	if merged.EmployeeID != existing.EmployeeID {
		if err := s.authorizeTimeEntryTarget(ctx, businessID, callerUserID, merged.EmployeeID); err != nil {
			return nil, err
		}
	}
	if merged.EmployeeID <= 0 {
		return nil, domain.ErrTimeEntryAccessDenied
	}
	if _, err := time.Parse(time.DateOnly, merged.Date); err != nil {
		return nil, err
	}

	locked, err := s.periodLocked(ctx, businessID, existing.EmployeeID, existing.Date)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, domain.ErrDateInLockedPeriod
	}
	if merged.EmployeeID != existing.EmployeeID || merged.Date != existing.Date {
		locked, err := s.periodLocked(ctx, businessID, merged.EmployeeID, merged.Date)
		if err != nil {
			return nil, err
		}
		if locked {
			return nil, domain.ErrDateInLockedPeriod
		}
	}

	rec, err := s.repo.UpdateTimeEntry(ctx, businessID, merged)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

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
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
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
	return cacher.Of[*domain.TimesheetRecord]().GetOrComputeSingle(ctx, key,
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

func (s *Service) ListTimesheets(ctx context.Context, businessID int64, params domain.ListTimesheetsParams) ([]domain.TimesheetRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.TimesheetRecord
		Total int64
	}
	key := listTimesheetsCacheKey(businessID, params)
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListTimesheets(ctx, businessID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrTimeTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) ListTimesheetsForCaller(ctx context.Context, businessID, callerUserID int64, params domain.ListTimesheetsParams) ([]domain.TimesheetRecord, int64, error) {
	employeeID, err := s.callerTimesheetEmployeeID(ctx, businessID, callerUserID)
	if err != nil {
		return nil, 0, err
	}
	allowed, err := s.canViewAllTimesheets(ctx, businessID, employeeID)
	if err != nil {
		return nil, 0, err
	}
	if !allowed {
		params.EmployeeID = &employeeID
	}
	return s.ListTimesheets(ctx, businessID, params)
}

func (s *Service) SubmitTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TimesheetStatusDraft && ts.Status != domain.TimesheetStatusRejected {
		return nil, domain.ErrTimesheetAlreadySubmitted
	}

	rec, err := s.repo.SubmitTimesheet(ctx, businessID, timesheetID)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) resolveTimesheetApprover(ctx context.Context, businessID, approverUserID int64) (int64, error) {
	employeeID, err := s.callerTimesheetEmployeeID(ctx, businessID, approverUserID)
	if err != nil {
		return 0, err
	}
	allowed, err := s.permissions.HasPermission(ctx, businessID, employeeID, domain.PermissionApproveTimeOff)
	if err != nil {
		return 0, err
	}
	if !allowed {
		return 0, domain.ErrTimesheetApproverNotAllowed
	}
	return employeeID, nil
}

func (s *Service) authorizeTimesheetDecision(ctx context.Context, businessID int64, ts *domain.TimesheetRecord, approverEmployeeID int64) (*domain.ApprovalChainInfo, error) {
	if ts.EmployeeID == approverEmployeeID {
		return nil, domain.ErrTimesheetSelfApproval
	}
	if s.approvalProvider == nil {
		return nil, domain.ErrApprovalChainNotConfigured
	}
	chain, err := s.approvalProvider.GetApprovalChain(ctx, businessID, "timesheet")
	if err != nil {
		return nil, err
	}
	if chain == nil || chain.TotalLevels <= 0 || len(chain.Levels) == 0 || int(chain.TotalLevels) != len(chain.Levels) {
		return nil, domain.ErrApprovalChainNotConfigured
	}
	if ts.CurrentApprovalLevel < 1 || ts.CurrentApprovalLevel > chain.TotalLevels {
		return nil, domain.ErrTimesheetApproverNotAllowed
	}

	levelIndex := -1
	for index, level := range chain.Levels {
		if level.LevelOrder == ts.CurrentApprovalLevel {
			levelIndex = index
			break
		}
	}
	if levelIndex < 0 {
		levelIndex = int(ts.CurrentApprovalLevel - 1)
	}
	if levelIndex < 0 || levelIndex >= len(chain.Levels) {
		return nil, domain.ErrTimesheetApproverNotAllowed
	}

	level := chain.Levels[levelIndex]
	if level.ApproverEmployeeID > 0 {
		if level.ApproverEmployeeID != approverEmployeeID {
			return nil, domain.ErrTimesheetApproverNotAllowed
		}
		return chain, nil
	}
	if level.ApproverRoleID <= 0 {
		return nil, domain.ErrTimesheetApproverNotAllowed
	}
	allowed, err := s.permissions.HasRole(ctx, businessID, approverEmployeeID, level.ApproverRoleID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domain.ErrTimesheetApproverNotAllowed
	}
	return chain, nil
}

func (s *Service) ApproveTimesheet(ctx context.Context, businessID, timesheetID, approverUserID int64) (*domain.TimesheetRecord, error) {
	approverEmployeeID, err := s.resolveTimesheetApprover(ctx, businessID, approverUserID)
	if err != nil {
		return nil, err
	}
	ts, err := s.repo.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TimesheetStatusSubmitted {
		return nil, domain.ErrTimesheetNotPending
	}
	chain, err := s.authorizeTimesheetDecision(ctx, businessID, ts, approverEmployeeID)
	if err != nil {
		return nil, err
	}

	if ts.CurrentApprovalLevel >= chain.TotalLevels {
		rec, err := s.repo.LockTimesheetWithApprovalAction(ctx, businessID, timesheetID, approverUserID, ts.CurrentApprovalLevel)
		if err == nil {
			invalidateHRTimeScope(ctx, businessID)
		}
		return rec, err
	}

	rec, err := s.repo.AdvanceApprovalLevel(ctx, businessID, timesheetID, ts.CurrentApprovalLevel, approverUserID)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) RejectTimesheet(ctx context.Context, businessID, timesheetID, approverUserID int64, reason string) (*domain.TimesheetRecord, error) {
	approverEmployeeID, err := s.resolveTimesheetApprover(ctx, businessID, approverUserID)
	if err != nil {
		return nil, err
	}
	ts, err := s.repo.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status != domain.TimesheetStatusSubmitted {
		return nil, domain.ErrTimesheetNotPending
	}
	if _, err := s.authorizeTimesheetDecision(ctx, businessID, ts, approverEmployeeID); err != nil {
		return nil, err
	}
	rec, err := s.repo.RejectTimesheet(ctx, businessID, timesheetID, reason)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

// --- Background Job ---

func (s *Service) RunUnsubmittedTimesheetReminder(ctx context.Context) error {
	periodStart, periodEnd := previousWeekRange(time.Now().UTC())

	businessIDs, err := s.repo.ListTimesheetBusinessIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("timesheet reminders: failed to list business IDs", logger.Error(err))
		return err
	}

	var candidateCount int
	for _, businessID := range businessIDs {
		employeeIDs, scanErr := s.repo.ListUnsubmittedEmployees(ctx, businessID, periodStart, periodEnd)
		if scanErr != nil {
			logger.Ctx(ctx).Error("timesheet reminders: failed to list unsubmitted employees",
				logger.Any("business_id", businessID),
				logger.Any("period_start", periodStart),
				logger.Any("period_end", periodEnd),
				logger.Error(scanErr),
			)
			continue
		}
		if len(employeeIDs) == 0 {
			continue
		}

		candidateCount += len(employeeIDs)
		logger.Ctx(ctx).Info("timesheet reminders: unsubmitted employee candidates",
			logger.Any("business_id", businessID),
			logger.Any("period_start", periodStart),
			logger.Any("period_end", periodEnd),
			logger.Any("employee_ids", employeeIDs),
			logger.Int("employee_count", len(employeeIDs)),
		)
	}

	logger.Ctx(ctx).Info("timesheet reminders: scan completed",
		logger.Int("business_count", len(businessIDs)),
		logger.Int("candidate_count", candidateCount),
		logger.Any("period_start", periodStart),
		logger.Any("period_end", periodEnd),
	)

	return nil
}

func previousWeekRange(now time.Time) (string, string) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	thisMonday := today.AddDate(0, 0, 1-weekday)
	previousMonday := thisMonday.AddDate(0, 0, -7)
	previousSunday := previousMonday.AddDate(0, 0, 6)
	return previousMonday.Format("2006-01-02"), previousSunday.Format("2006-01-02")
}

// --- Mark invoiced ---

// MarkEntriesInvoicedResult reports one batch outcome.
type MarkEntriesInvoicedResult struct {
	MarkedCount  int32
	SkippedCount int32
}

// MarkEntriesInvoiced records that one finance invoice settles the listed
// approved billable entries (#836/#791). The repository runs the batch in one
// transaction so either every eligible entry is stamped or nothing changes;
// per-entry skips (already invoiced, foreign, ineligible) are counted rather
// than fatal, and the response separates them so the caller knows when the
// loop is only partially closed.
func (s *Service) MarkEntriesInvoiced(ctx context.Context, businessID, invoiceID int64, entryIDs []int64) (MarkEntriesInvoicedResult, error) {
	var result MarkEntriesInvoicedResult
	marked, err := s.repo.MarkEntriesInvoiced(ctx, businessID, invoiceID, entryIDs)
	if err != nil {
		return result, err
	}
	invalidateHRTimeScope(ctx, businessID)
	result.MarkedCount = int32(marked)                       // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	result.SkippedCount = int32(len(entryIDs) - int(marked)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	return result, nil
}
