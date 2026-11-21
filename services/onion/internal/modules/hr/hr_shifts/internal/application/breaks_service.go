package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/domain"
)

const (
	defaultBreaksLimit int32 = 100
	maxBreaksLimit     int32 = 200
)

type employeeNameProvider interface {
	GetEmployeeDisplayName(ctx context.Context, businessID, employeeID int64) (string, error)
}

func (s *Service) resolveCallerEmployeeID(ctx context.Context, businessID, callerUserID int64) int64 {
	if s.permissions == nil || businessID <= 0 || callerUserID <= 0 {
		return 0
	}
	id, err := s.permissions.GetEmployeeIDByUserID(ctx, businessID, callerUserID)
	if err != nil || id <= 0 {
		return 0
	}
	return id
}

func (s *Service) resolveEmployeeName(ctx context.Context, businessID, employeeID int64) string {
	if employeeID <= 0 || s.permissions == nil {
		return ""
	}
	namer, ok := s.permissions.(employeeNameProvider)
	if !ok {
		return ""
	}
	name, err := namer.GetEmployeeDisplayName(ctx, businessID, employeeID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(name)
}

func (s *Service) ListBreaks(ctx context.Context, businessID, callerUserID int64, canManage bool, limit, offset int32) ([]domain.BreakRecord, error) {
	if limit <= 0 {
		limit = defaultBreaksLimit
	}
	if limit > maxBreaksLimit {
		limit = maxBreaksLimit
	}
	if offset < 0 {
		offset = 0
	}

	params := domain.ListBreaksParams{
		Limit:  limit,
		Offset: offset,
	}
	if !canManage {
		callerEmpID := s.resolveCallerEmployeeID(ctx, businessID, callerUserID)
		if callerEmpID <= 0 {
			return nil, domain.ErrPermissionDenied
		}
		params.EmployeeID = &callerEmpID
	}

	return s.repo.ListBreaks(ctx, businessID, params)
}

func (s *Service) StartBreak(ctx context.Context, businessID, callerUserID int64, canManage bool, input domain.StartBreakInput) (*domain.BreakRecord, error) {
	breakType, err := domain.NormalizeBreakType(input.BreakType)
	if err != nil {
		return nil, err
	}
	input.BreakType = breakType
	input.StaffName = strings.TrimSpace(input.StaffName)

	callerEmpID := s.resolveCallerEmployeeID(ctx, businessID, callerUserID)
	var callerStaffID int64
	if s.staffResolver != nil && callerUserID > 0 {
		if sid, sidErr := s.staffResolver.StaffIDByUserID(ctx, businessID, callerUserID); sidErr == nil {
			callerStaffID = sid
		}
	}

	if !canManage {
		if callerEmpID <= 0 {
			return nil, domain.ErrPermissionDenied
		}
		if input.EmployeeID > 0 && input.EmployeeID != callerEmpID {
			return nil, domain.ErrPermissionDenied
		}
		if input.StaffID > 0 && callerStaffID > 0 && input.StaffID != callerStaffID {
			return nil, domain.ErrPermissionDenied
		}
		input.EmployeeID = callerEmpID
		if callerStaffID > 0 {
			input.StaffID = callerStaffID
		} else {
			input.StaffID = callerEmpID
		}
	} else {
		if input.EmployeeID == 0 && input.StaffID == 0 {
			input.EmployeeID = callerEmpID
			if callerStaffID > 0 {
				input.StaffID = callerStaffID
			} else if callerEmpID > 0 {
				input.StaffID = callerEmpID
			} else {
				input.StaffID = callerUserID
			}
		} else if input.EmployeeID == 0 && input.StaffID > 0 && callerStaffID > 0 && input.StaffID == callerStaffID {
			input.EmployeeID = callerEmpID
		} else if input.StaffID == 0 && input.EmployeeID > 0 {
			if s.staffResolver != nil {
				if sid, sidErr := s.staffResolver.StaffIDByEmployeeID(ctx, businessID, input.EmployeeID); sidErr == nil && sid > 0 {
					input.StaffID = sid
				}
			}
			if input.StaffID == 0 {
				input.StaffID = input.EmployeeID
			}
		}
	}

	if input.StaffName == "" && input.EmployeeID > 0 {
		input.StaffName = s.resolveEmployeeName(ctx, businessID, input.EmployeeID)
	}
	if input.StaffName == "" {
		if input.StaffID > 0 {
			input.StaffName = fmt.Sprintf("Staff #%d", input.StaffID)
		} else {
			input.StaffName = "Team member"
		}
	}

	if input.EmployeeID > 0 {
		open, openErr := s.repo.GetOpenBreakByEmployee(ctx, businessID, input.EmployeeID)
		if openErr != nil {
			return nil, openErr
		}
		if open != nil {
			return nil, domain.ErrBreakAlreadyOpen
		}
	} else if input.StaffID > 0 {
		open, openErr := s.repo.GetOpenBreakByStaff(ctx, businessID, input.StaffID)
		if openErr != nil {
			return nil, openErr
		}
		if open != nil {
			return nil, domain.ErrBreakAlreadyOpen
		}
	}

	rec, err := s.repo.CreateBreak(ctx, businessID, input)
	if err != nil {
		return nil, err
	}
	invalidateHRShiftsScope(ctx, businessID)
	return rec, nil
}

func (s *Service) EndBreak(ctx context.Context, businessID, callerUserID, breakID int64, canManage bool) (*domain.BreakRecord, error) {
	if breakID <= 0 {
		return nil, domain.ErrBreakNotFound
	}
	existing, err := s.repo.GetBreak(ctx, businessID, breakID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		callerEmpID := s.resolveCallerEmployeeID(ctx, businessID, callerUserID)
		if callerEmpID <= 0 || existing.EmployeeID != callerEmpID {
			return nil, domain.ErrPermissionDenied
		}
	}
	if existing.EndedAt != nil {
		return nil, domain.ErrBreakAlreadyEnded
	}
	rec, err := s.repo.EndBreak(ctx, businessID, breakID)
	if err != nil {
		if errors.Is(err, domain.ErrBreakAlreadyEnded) || errors.Is(err, domain.ErrBreakNotFound) {
			return nil, err
		}
		return nil, err
	}
	invalidateHRShiftsScope(ctx, businessID)
	return rec, nil
}
