package application

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo             *infrastructure.Repository
	approvalProvider domain.ApprovalChainProvider
}

func NewService(ctx context.Context, approvalProvider domain.ApprovalChainProvider) *Service {
	return &Service{
		repo:             infrastructure.NewRepository(ctx),
		approvalProvider: approvalProvider,
	}
}

// --- Leave Types ---

func (s *Service) ListLeaveTypes(ctx context.Context, ownerID int64, params domain.ListLeaveTypesParams) ([]domain.LeaveTypeRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	types, total, err := s.repo.ListLeaveTypes(ctx, ownerID, params)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return nil, 0, domain.ErrLeaveNotConfigured
	}

	return types, total, nil
}

func (s *Service) CreateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) (*domain.LeaveTypeRecord, error) {
	if lt.Name == "" {
		lt.Name = "Unnamed Leave Type"
	}
	lt.IsActive = true
	return s.repo.CreateLeaveType(ctx, ownerID, lt)
}

func (s *Service) UpdateLeaveType(ctx context.Context, ownerID int64, lt domain.LeaveTypeRecord) error {
	return s.repo.UpdateLeaveType(ctx, ownerID, lt)
}

// --- Leave Requests ---

func (s *Service) CreateLeaveRequest(ctx context.Context, ownerID int64, req domain.LeaveRequestRecord) (*domain.LeaveRequestRecord, error) {
	// Validate leave type exists
	leaveTypes, _, err := s.repo.ListLeaveTypes(ctx, ownerID, domain.ListLeaveTypesParams{Limit: 1, Offset: 0})
	if err != nil {
		return nil, err
	}
	if len(leaveTypes) == 0 {
		return nil, domain.ErrLeaveNotConfigured
	}

	lt, err := s.repo.GetLeaveType(ctx, ownerID, req.LeaveTypeID)
	if err != nil {
		return nil, err
	}

	// Validate date range
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, domain.ErrInvalidDateRange
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, domain.ErrInvalidDateRange
	}
	if endDate.Before(startDate) {
		return nil, domain.ErrInvalidDateRange
	}

	// Validate min notice
	if lt.MinNoticeDays > 0 {
		minStart := time.Now().AddDate(0, 0, int(lt.MinNoticeDays))
		if startDate.Before(minStart) {
			return nil, domain.ErrInsufficientNotice
		}
	}

	// Check for date overlap with existing approved/pending leave
	overlapping, err := s.repo.ListOverlappingLeaveRequests(ctx, ownerID, req.EmployeeID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}
	if len(overlapping) > 0 {
		return nil, domain.ErrLeaveConflict
	}

	// Check balance
	year := int32(startDate.Year())
	balance, err := s.repo.GetLeaveBalance(ctx, ownerID, req.EmployeeID, req.LeaveTypeID, year)
	if err != nil {
		return nil, err
	}
	remaining := balance.EntitledDays - balance.UsedDays - balance.PendingDays
	if req.TotalDays > remaining {
		return nil, domain.ErrInsufficientLeaveBalance
	}

	req.Status = "pending"
	req.CurrentApprovalLevel = 0

	created, err := s.repo.CreateLeaveRequest(ctx, ownerID, req)
	if err != nil {
		return nil, err
	}

	// Add to pending days
	addErr := s.repo.AddToPendingDays(ctx, ownerID, req.EmployeeID, req.LeaveTypeID, year, req.TotalDays)
	if addErr != nil {
		logger.Ctx(ctx).Error("failed to add pending days after creating leave request",
			logger.Any("owner_id", ownerID), logger.Any("employee_id", req.EmployeeID), logger.Error(addErr))
	}

	return created, nil
}

func (s *Service) ListLeaveRequests(ctx context.Context, ownerID int64, params domain.ListLeaveRequestsParams) ([]domain.LeaveRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	return s.repo.ListLeaveRequests(ctx, ownerID, params)
}

func (s *Service) ApproveLeave(ctx context.Context, ownerID, requestID int64) error {
	req, err := s.repo.GetLeaveRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}

	lt, err := s.repo.GetLeaveType(ctx, ownerID, req.LeaveTypeID)
	if err != nil {
		return err
	}

	if lt.RequiresApproval {
		chain, chainErr := s.approvalProvider.GetApprovalChain(ctx, ownerID, "leave")
		if chainErr != nil {
			return chainErr
		}
		if chain == nil {
			return domain.ErrApprovalChainNotConfigured
		}

		if req.CurrentApprovalLevel+1 < chain.TotalLevels {
			return s.repo.AdvanceLeaveApprovalLevel(ctx, ownerID, requestID)
		}
	}

	// Final approval — move pending to used
	if err := s.repo.ApproveLeaveRequest(ctx, ownerID, requestID); err != nil {
		return err
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	year := int32(startDate.Year())

	addErr := s.repo.AddToUsedDays(ctx, ownerID, req.EmployeeID, req.LeaveTypeID, year, req.TotalDays)
	if addErr != nil {
		logger.Ctx(ctx).Error("failed to update balance after leave approval",
			logger.Any("owner_id", ownerID), logger.Any("request_id", requestID), logger.Error(addErr))
	}

	return nil
}

func (s *Service) RejectLeave(ctx context.Context, ownerID, requestID int64, reason string) error {
	req, err := s.repo.GetLeaveRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}

	if err := s.repo.RejectLeaveRequest(ctx, ownerID, requestID, reason); err != nil {
		return err
	}

	// Remove pending days
	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	year := int32(startDate.Year())

	subErr := s.repo.SubtractPendingDays(ctx, ownerID, req.EmployeeID, req.LeaveTypeID, year, req.TotalDays)
	if subErr != nil {
		logger.Ctx(ctx).Error("failed to subtract pending days after rejection",
			logger.Any("owner_id", ownerID), logger.Any("request_id", requestID), logger.Error(subErr))
	}

	return nil
}

func (s *Service) GetLeaveBalances(ctx context.Context, ownerID, employeeID int64, year int32) ([]domain.LeaveBalanceRecord, error) {
	if year == 0 {
		year = int32(time.Now().Year())
	}
	return s.repo.ListLeaveBalances(ctx, ownerID, employeeID, year)
}

func (s *Service) GetLeaveCalendar(ctx context.Context, ownerID int64, params domain.LeaveCalendarParams) ([]domain.LeaveRequestRecord, error) {
	return s.repo.ListLeaveCalendar(ctx, ownerID, params.DateFrom, params.DateTo)
}

// --- Background Job: Monthly Accrual ---

func (s *Service) RunMonthlyAccrual(ctx context.Context) error {
	now := time.Now()
	currentMonth := int32(now.Month())
	currentYear := int32(now.Year())

	ownerIDs, err := s.repo.ListDistinctOwnerIDs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("leave accrual: failed to list owner IDs", logger.Error(err))
		return err
	}

	for _, ownerID := range ownerIDs {
		if err := s.accrueForOwner(ctx, ownerID, currentYear, currentMonth); err != nil {
			logger.Ctx(ctx).Error("leave accrual: failed for owner",
				logger.Any("owner_id", ownerID), logger.Error(err))
		}
	}

	return nil
}

func (s *Service) accrueForOwner(ctx context.Context, ownerID int64, year, month int32) error {
	leaveTypes, err := s.repo.ListAllActiveLeaveTypesForOwner(ctx, ownerID)
	if err != nil {
		return err
	}

	for _, lt := range leaveTypes {
		if lt.AccrualRule == "" || lt.AccrualRule == "none" || lt.AccrualAmount <= 0 {
			continue
		}

		if lt.AccrualRule != "monthly" {
			continue
		}

		balances, balErr := s.repo.ListBalancesForAccrual(ctx, ownerID, year, month)
		if balErr != nil {
			logger.Ctx(ctx).Error("leave accrual: failed to list balances",
				logger.Any("owner_id", ownerID), logger.Any("leave_type_id", lt.ID), logger.Error(balErr))
			continue
		}

		for _, bal := range balances {
			if bal.LeaveTypeID != lt.ID {
				continue
			}
			addErr := s.repo.AddToEntitledDays(ctx, ownerID, bal.EmployeeID, lt.ID, year, lt.AccrualAmount, month)
			if addErr != nil {
				logger.Ctx(ctx).Error("leave accrual: failed to add entitled days",
					logger.Any("owner_id", ownerID), logger.Any("employee_id", bal.EmployeeID),
					logger.Any("leave_type_id", lt.ID), logger.Error(addErr))
			}
		}
	}

	return nil
}
