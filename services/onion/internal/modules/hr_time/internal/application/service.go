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
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(hrTimeScope, businessID))
	}
}

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
	flowReader       domain.FlowTaskReader
}

func NewService(_ context.Context, db *sql.DB, approvalProvider domain.ApprovalChainProvider, flowReader domain.FlowTaskReader) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(db),
		approvalProvider: approvalProvider,
		flowReader:       flowReader,
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
	rec, err := s.repo.UpdateTimeEntry(ctx, businessID, entry)
	if err == nil {
		invalidateHRTimeScope(ctx, businessID)
	}
	return rec, err
}

func (s *Service) DeleteTimeEntry(ctx context.Context, businessID, entryID int64) error {
	if err := s.repo.DeleteTimeEntry(ctx, businessID, entryID); err != nil {
		return err
	}
	invalidateHRTimeScope(ctx, businessID)
	return nil
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

// --- Flow Import ---

func (s *Service) ImportFlowTasks(ctx context.Context, businessID, employeeID int64, periodStart, periodEnd string) ([]domain.TimeEntryRecord, error) {
	tasks, err := s.flowReader.GetCompletedTasksForEmployee(ctx, businessID, employeeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	// Batch the per-task existence check into one query instead of issuing
	// GetTimeEntryBySourceAndTask once per task (N+1).
	taskIDs := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		taskIDs = append(taskIDs, task.TaskID)
	}
	existingTaskIDs, err := s.repo.GetExistingTaskIDsBySource(ctx, businessID, employeeID, "flow_import", taskIDs)
	if err != nil {
		return nil, err
	}

	var imported []domain.TimeEntryRecord
	for _, task := range tasks {
		if _, ok := existingTaskIDs[task.TaskID]; ok {
			continue
		}

		entry := domain.TimeEntryRecord{
			EmployeeID:      employeeID,
			TaskID:          task.TaskID,
			Date:            task.CompletedAt,
			DurationMinutes: task.DurationMinutes,
			Description:     task.Title,
			Status:          "draft",
			Source:          "flow_import",
		}
		created, createErr := s.repo.CreateTimeEntry(ctx, businessID, entry)
		if createErr != nil {
			logger.Ctx(ctx).Error("failed to import flow task as time entry",
				logger.Any("business_id", businessID), logger.Any("task_id", task.TaskID), logger.Error(createErr))
			continue
		}
		imported = append(imported, *created)
	}

	return imported, nil
}

// --- Billable Export ---

func (s *Service) ExportBillableEntries(ctx context.Context, businessID int64) ([]domain.BillableExportItem, int64, error) {
	entries, err := s.repo.ListApprovedBillableEntries(ctx, businessID)
	if err != nil {
		return nil, 0, err
	}

	if len(entries) == 0 {
		return nil, 0, nil
	}

	batchID := time.Now().UnixMilli()

	items := make([]domain.BillableExportItem, 0, len(entries))
	var totalMinutes int32
	for _, e := range entries {
		if markErr := s.repo.MarkEntryExported(ctx, businessID, e.ID, batchID); markErr != nil {
			logger.Ctx(ctx).Error("failed to mark entry as exported",
				logger.Any("business_id", businessID), logger.Any("entry_id", e.ID), logger.Error(markErr))
			continue
		}
		items = append(items, domain.BillableExportItem{
			EntryID:         e.ID,
			EmployeeID:      e.EmployeeID,
			ProjectID:       e.ProjectID,
			TaskID:          e.TaskID,
			Date:            e.Date,
			DurationMinutes: e.DurationMinutes,
			Description:     e.Description,
			IsBillable:      e.IsBillable,
		})
		totalMinutes += e.DurationMinutes
	}

	return items, batchID, nil
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
