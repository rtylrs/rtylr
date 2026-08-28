package application

import (
	"context"
	"database/sql"
	"fmt"
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
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
	}
}

// --- Time Entries ---

func (s *Service) CreateTimeEntry(ctx context.Context, businessID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	if entry.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}

	ts, err := s.repo.GetTimesheetForPeriod(ctx, businessID, entry.EmployeeID, entry.Date[:7]+"-01")
	if err != nil {
		return nil, err
	}
	if ts != nil && ts.Status == "locked" {
		return nil, domain.ErrDateInLockedPeriod
	}

	if entry.Status == "" {
		entry.Status = "draft"
	}
	if entry.Source == "" {
		entry.Source = "manual"
	}

	rec, err := s.repo.CreateTimeEntry(ctx, businessID, entry)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) UpdateTimeEntry(ctx context.Context, businessID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	if entry.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}

	// A locked timesheet must freeze its entries: check both the period the
	// entry currently lives in and the one a new date would move it into.
	existing, err := s.repo.GetTimeEntry(ctx, businessID, entry.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrTimeEntryNotFound
	}
	locked, err := s.periodLocked(ctx, businessID, existing.EmployeeID, existing.Date)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, domain.ErrDateInLockedPeriod
	}
	if entry.Date != "" && entry.Date != existing.Date {
		locked, err := s.periodLocked(ctx, businessID, existing.EmployeeID, entry.Date)
		if err != nil {
			return nil, err
		}
		if locked {
			return nil, domain.ErrDateInLockedPeriod
		}
	}

	rec, err := s.repo.UpdateTimeEntry(ctx, businessID, entry)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) DeleteTimeEntry(ctx context.Context, businessID, entryID int64) error {
	existing, err := s.repo.GetTimeEntry(ctx, businessID, entryID)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrTimeEntryNotFound
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

func (s *Service) ListTimeEntries(ctx context.Context, businessID int64, params domain.ListTimeEntriesParams) ([]domain.TimeEntryRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	type lr struct {
		Items []domain.TimeEntryRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrTimeScope, businessID, fmt.Sprintf("entries|l=%d|o=%d|e=%d", params.Limit, params.Offset, params.EmployeeID))
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
		Status:      "draft",
	})
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
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
	key := cacheKeys.Scoped(hrTimeScope, businessID, fmt.Sprintf("ts_list|e=%d|l=%d|o=%d", params.EmployeeID, params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListTimesheets(ctx, businessID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrTimeTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) SubmitTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status != "draft" && ts.Status != "rejected" {
		return nil, domain.ErrTimesheetAlreadySubmitted
	}

	rec, err := s.repo.SubmitTimesheet(ctx, businessID, timesheetID)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) ApproveTimesheet(ctx context.Context, businessID, timesheetID, approverID int64) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheet(ctx, businessID, timesheetID)
	if err != nil {
		return nil, err
	}

	chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, businessID, "timesheet")
	if chainErr != nil {
		return nil, chainErr
	}
	if chain == nil {
		return nil, domain.ErrApprovalChainNotConfigured
	}

	if ts.CurrentApprovalLevel >= chain.TotalLevels {
		if err := s.repo.LockTimesheet(ctx, businessID, timesheetID); err != nil {
			return nil, err
		}

		lockErr := s.repo.CreateApprovalAction(ctx, businessID, domain.ApprovalActionRecord{
			TimesheetID: timesheetID,
			ApproverID:  approverID,
			Action:      "final_approved",
			Level:       ts.CurrentApprovalLevel,
		})
		if lockErr != nil {
			logger.Ctx(ctx).Error("failed to record final approval action", logger.Any("business_id", businessID), logger.Error(lockErr))
		}

		rec, err := s.repo.GetTimesheet(ctx, businessID, timesheetID)
		invalidateHRTimeScope(ctx, businessID)
		return rec, err
	}

	rec, err := s.repo.AdvanceApprovalLevel(ctx, businessID, timesheetID, approverID)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) RejectTimesheet(ctx context.Context, businessID, timesheetID int64, reason string) (*domain.TimesheetRecord, error) {
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
	result.MarkedCount = int32(marked)                       // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	result.SkippedCount = int32(len(entryIDs) - int(marked)) // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	return result, nil
}
