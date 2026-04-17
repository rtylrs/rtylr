package application

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
	flowReader       domain.FlowTaskReader
}

func NewService(ctx context.Context, approvalProvider domain.ApprovalChainProvider, flowReader domain.FlowTaskReader) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(ctx),
		approvalProvider: approvalProvider,
		flowReader:       flowReader,
	}
}

// --- Time Entries ---

func (s *Service) CreateTimeEntry(ctx context.Context, ownerID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	if entry.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}

	ts, err := s.repo.GetTimesheetForPeriod(ctx, ownerID, entry.EmployeeID, entry.Date[:7]+"-01")
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

	return s.repo.CreateTimeEntry(ctx, ownerID, entry)
}

func (s *Service) UpdateTimeEntry(ctx context.Context, ownerID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error) {
	if entry.DurationMinutes <= 0 {
		return nil, domain.ErrInvalidDuration
	}
	return s.repo.UpdateTimeEntry(ctx, ownerID, entry)
}

func (s *Service) DeleteTimeEntry(ctx context.Context, ownerID, entryID int64) error {
	return s.repo.DeleteTimeEntry(ctx, ownerID, entryID)
}

func (s *Service) ListTimeEntries(ctx context.Context, ownerID int64, params domain.ListTimeEntriesParams) ([]domain.TimeEntryRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListTimeEntries(ctx, ownerID, params)
}

// --- Timesheets ---

func (s *Service) GetOrCreateTimesheet(ctx context.Context, ownerID, employeeID int64, periodStart, periodEnd string) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheetForPeriod(ctx, ownerID, employeeID, periodStart)
	if err != nil {
		return nil, err
	}
	if ts != nil {
		return ts, nil
	}

	return s.repo.UpsertTimesheet(ctx, ownerID, domain.TimesheetRecord{
		EmployeeID:  employeeID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      "draft",
	})
}

func (s *Service) GetTimesheetForPeriod(ctx context.Context, ownerID, employeeID int64, periodStart string) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheetForPeriod(ctx, ownerID, employeeID, periodStart)
	if err != nil {
		return nil, err
	}
	if ts == nil {
		return nil, domain.ErrTimesheetNotFound
	}
	return ts, nil
}

func (s *Service) GetTimesheet(ctx context.Context, ownerID, timesheetID int64) (*domain.TimesheetRecord, error) {
	return s.repo.GetTimesheet(ctx, ownerID, timesheetID)
}

func (s *Service) ListTimesheets(ctx context.Context, ownerID int64, params domain.ListTimesheetsParams) ([]domain.TimesheetRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListTimesheets(ctx, ownerID, params)
}

func (s *Service) SubmitTimesheet(ctx context.Context, ownerID, timesheetID int64) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheet(ctx, ownerID, timesheetID)
	if err != nil {
		return nil, err
	}
	if ts.Status != "draft" && ts.Status != "rejected" {
		return nil, domain.ErrTimesheetAlreadySubmitted
	}

	return s.repo.SubmitTimesheet(ctx, ownerID, timesheetID)
}

func (s *Service) ApproveTimesheet(ctx context.Context, ownerID, timesheetID, approverID int64) (*domain.TimesheetRecord, error) {
	ts, err := s.repo.GetTimesheet(ctx, ownerID, timesheetID)
	if err != nil {
		return nil, err
	}

	chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, ownerID, "timesheet")
	if chainErr != nil {
		return nil, chainErr
	}
	if chain == nil {
		return nil, domain.ErrApprovalChainNotConfigured
	}

	if ts.CurrentApprovalLevel >= chain.TotalLevels {
		if err := s.repo.LockTimesheet(ctx, ownerID, timesheetID); err != nil {
			return nil, err
		}

		lockErr := s.repo.CreateApprovalAction(ctx, ownerID, domain.ApprovalActionRecord{
			TimesheetID: timesheetID,
			ApproverID:  approverID,
			Action:      "final_approved",
			Level:       ts.CurrentApprovalLevel,
		})
		if lockErr != nil {
			logger.Ctx(ctx).Error("failed to record final approval action", logger.Any("owner_id", ownerID), logger.Error(lockErr))
		}

		return s.repo.GetTimesheet(ctx, ownerID, timesheetID)
	}

	return s.repo.AdvanceApprovalLevel(ctx, ownerID, timesheetID, approverID)
}

func (s *Service) RejectTimesheet(ctx context.Context, ownerID, timesheetID int64, reason string) (*domain.TimesheetRecord, error) {
	return s.repo.RejectTimesheet(ctx, ownerID, timesheetID, reason)
}

// --- Flow Import ---

func (s *Service) ImportFlowTasks(ctx context.Context, ownerID, employeeID int64, periodStart, periodEnd string) ([]domain.TimeEntryRecord, error) {
	tasks, err := s.flowReader.GetCompletedTasksForEmployee(ctx, ownerID, employeeID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	var imported []domain.TimeEntryRecord
	for _, task := range tasks {
		existing, err := s.repo.GetTimeEntryBySourceAndTask(ctx, ownerID, employeeID, "flow_import", task.TaskID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
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
		created, createErr := s.repo.CreateTimeEntry(ctx, ownerID, entry)
		if createErr != nil {
			logger.Ctx(ctx).Error("failed to import flow task as time entry",
				logger.Any("owner_id", ownerID), logger.Any("task_id", task.TaskID), logger.Error(createErr))
			continue
		}
		imported = append(imported, *created)
	}

	return imported, nil
}

// --- Billable Export ---

func (s *Service) ExportBillableEntries(ctx context.Context, ownerID int64) ([]domain.BillableExportItem, int64, error) {
	entries, err := s.repo.ListApprovedBillableEntries(ctx, ownerID)
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
		if markErr := s.repo.MarkEntryExported(ctx, ownerID, e.ID, batchID); markErr != nil {
			logger.Ctx(ctx).Error("failed to mark entry as exported",
				logger.Any("owner_id", ownerID), logger.Any("entry_id", e.ID), logger.Error(markErr))
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

func (s *Service) CheckUnsubmittedTimesheets(ctx context.Context) error {
	logger.Ctx(ctx).Info("running unsubmitted timesheet check")

	// This is a placeholder - in production this would iterate all active owners
	// and check for unsubmitted timesheets for the previous period.
	// The actual owner iteration requires a cross-module query to list active businesses.
	_ = errors.New("not implemented")

	return nil
}
