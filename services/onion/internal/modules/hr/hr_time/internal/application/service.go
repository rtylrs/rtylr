package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_time/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const (
	hrTimeScope = "hr_time"
	hrTimeTTL   = 2 * time.Minute
)

func invalidateHRTimeScope(ctx context.Context, businessID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(hrTimeScope, businessID))
}

type Service struct {
	repo             repository
	approvalProvider domain.ApprovalChainProvider
	permissions      domain.PermissionProvider
	invoiceExistence InvoiceExistenceProvider
}

func NewService(_ context.Context, repo repository, approvalProvider domain.ApprovalChainProvider, permissions domain.PermissionProvider) *Service {
	return &Service{
		repo:             repo,
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

type repository interface {
	AdvanceApprovalLevel(ctx context.Context, businessID, timesheetID int64, expectedLevel int32, approverID int64) (*domain.TimesheetRecord, error)
	CreateFlowWorkdayEntry(ctx context.Context, input domain.FlowWorkdayEntryInput) (*domain.FlowWorkdayEntry, error)
	CreateTimeEntry(ctx context.Context, businessID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error)
	DeleteFlowWorkdayEntry(ctx context.Context, id, ownerID, businessID int64) (bool, error)
	DeleteTimeEntry(ctx context.Context, businessID, entryID int64) error
	FlowWorkdayEntryOwner(ctx context.Context, id, ownerID, businessID int64) (domain.FlowWorkdayEntryOwner, bool, error)
	GetTimeEntry(ctx context.Context, businessID, entryID int64) (*domain.TimeEntryRecord, error)
	GetTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error)
	GetTimesheetForPeriod(ctx context.Context, businessID, employeeID int64, periodStart string) (*domain.TimesheetRecord, error)
	ListTimeEntries(ctx context.Context, businessID int64, params domain.ListTimeEntriesParams) ([]domain.TimeEntryRecord, int64, error)
	ListFlowWorkdayEntries(ctx context.Context, ownerID, businessID, workDayID int64) ([]domain.FlowWorkdayEntry, int32, error)
	ListUnbilledBillableTimeEntries(ctx context.Context, businessID int64, limit, offset int32) ([]domain.TimeEntryRecord, int64, int64, error)
	ListTimesheetBusinessIDs(ctx context.Context) ([]int64, error)
	ListTimesheets(ctx context.Context, businessID int64, params domain.ListTimesheetsParams) ([]domain.TimesheetRecord, int64, error)
	ListUnsubmittedEmployees(ctx context.Context, businessID int64, periodStart, periodEnd string) ([]int64, error)
	LockTimesheetWithApprovalAction(ctx context.Context, businessID, timesheetID, approverID int64, level int32) (*domain.TimesheetRecord, error)
	MarkEntriesInvoiced(ctx context.Context, businessID, invoiceID int64, entryIDs []int64) (marked int64, err error)
	RejectTimesheet(ctx context.Context, businessID, timesheetID int64, reason string) (*domain.TimesheetRecord, error)
	SubmitTimesheet(ctx context.Context, businessID, timesheetID int64) (*domain.TimesheetRecord, error)
	SumFlowWorkdayEntryMinutesByDayIDs(ctx context.Context, businessID int64, workDayIDs []int64) (map[int64]int32, error)
	SumFlowWorkdayEntryMinutesByMemberAndTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) ([]domain.FlowWorkdayMemberMinutes, error)
	SumFlowWorkdayEntryMinutesByTaskIDs(ctx context.Context, businessID int64, taskIDs []int64) (map[int64]int64, error)
	UpdateTimeEntry(ctx context.Context, businessID int64, entry domain.TimeEntryRecord) (*domain.TimeEntryRecord, error)
	UpdateFlowWorkdayEntry(ctx context.Context, input domain.FlowWorkdayEntryInput) (*domain.FlowWorkdayEntry, error)
	UpsertTimesheet(ctx context.Context, businessID int64, ts domain.TimesheetRecord) (*domain.TimesheetRecord, error)
}
