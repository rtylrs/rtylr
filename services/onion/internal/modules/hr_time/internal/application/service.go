package application

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/hr_time/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_time/internal/infrastructure"
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
