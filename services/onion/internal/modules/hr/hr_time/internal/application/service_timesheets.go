package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	"github.com/voxire/onion/internal/platform/logger"
)

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
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
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
	if len(entryIDs) == 0 {
		return result, nil
	}
	if s.invoiceExistence == nil {
		return result, ErrInvoiceExistenceUnavailable
	}
	exists, err := s.invoiceExistence.InvoiceExists(ctx, businessID, invoiceID)
	if err != nil {
		return result, err
	}
	if !exists {
		return result, domain.ErrInvoiceNotFound
	}
	marked, err := s.repo.MarkEntriesInvoiced(ctx, businessID, invoiceID, entryIDs)
	if err != nil {
		return result, err
	}
	invalidateHRTimeScope(ctx, businessID)
	result.MarkedCount = int32(marked)                       // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	result.SkippedCount = int32(len(entryIDs) - int(marked)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	return result, nil
}

func (s *Service) ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int) ([]domain.TimeEntryRecord, int64, int64, error) {
	if limit <= 0 {
		limit = 1
	}
	if limit > 501 {
		limit = 501
	}
	if offset < 0 {
		offset = 0
	}
	if offset > 10000 {
		offset = 10000
	}
	return s.repo.ListUnbilledBillableTimeEntries(ctx, businessID, int32(limit), int32(offset))
}
