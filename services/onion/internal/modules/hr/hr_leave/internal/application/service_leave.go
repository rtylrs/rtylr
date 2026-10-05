package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

func (s *Service) ListLeaveRequests(ctx context.Context, ownerID, callerUserID int64, params domain.ListLeaveRequestsParams) ([]domain.LeaveRequestRecord, int64, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	effectiveEmployeeID, err := s.scopeLeaveRequests(ctx, ownerID, callerUserID, params.EmployeeID)
	if err != nil {
		return nil, 0, err
	}
	params.EmployeeID = effectiveEmployeeID
	params.Status = strings.ToLower(strings.TrimSpace(params.Status))
	type lr struct {
		Items []domain.LeaveRequestRecord
		Total int64
	}
	key := cacheKeys.Scoped(hrLeaveScope, ownerID, fmt.Sprintf("reqs|e=%d|s=%s|f=%s|t=%s|l=%d|o=%d", params.EmployeeID, params.Status, derefStr(params.FromDate), derefStr(params.ToDate), params.Limit, params.Offset))
	res, err := cacher.Of[lr]().GetOrComputeSingle(ctx, key+"|domain=v2",
		func(ctx context.Context) (lr, error) {
			items, total, err := s.repo.ListLeaveRequests(ctx, ownerID, params)
			return lr{Items: items, Total: total}, err
		},
		cacher.WithExpiry(ctx, hrLeaveTTL),
	)
	return res.Items, res.Total, err
}

func (s *Service) ApproveLeave(ctx context.Context, ownerID, requestID, approverUserID int64) error {
	req, lt, err := s.authorizedLeaveDecision(ctx, ownerID, requestID, approverUserID)
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
			if err := s.repo.AdvanceLeaveApprovalLevel(ctx, ownerID, requestID, req.CurrentApprovalLevel); err != nil {
				return err
			}
			invalidateHRLeaveScope(ctx, ownerID)
			return nil
		}
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return parseErr
	}

	// Final approval atomically moves the request from pending to approved and
	// moves every year's allocation from pending to used.
	if err := s.repo.ApproveLeaveRequestWithAllocations(ctx, ownerID, requestID, req.EmployeeID, req.LeaveTypeID, leaveDayAllocations(startDate, endDate)); err != nil {
		return err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return nil
}

func (s *Service) RejectLeave(ctx context.Context, ownerID, requestID, approverUserID int64, reason string) error {
	req, _, err := s.authorizedLeaveDecision(ctx, ownerID, requestID, approverUserID)
	if err != nil {
		return err
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return parseErr
	}
	if err := s.repo.RejectLeaveRequestWithBalance(ctx, ownerID, requestID, req.EmployeeID, req.LeaveTypeID, leaveDayAllocations(startDate, endDate), reason); err != nil {
		return err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return nil
}

// CancelLeaveRequest withdraws the caller's own pending leave request.
// Unlike ApproveLeave/RejectLeave, which are an approver's decision over
// someone else's request, this is self-service: only the request's own
// subject may cancel it - the leave approval permission grants no override
// here - and only while it is still pending. Once a request has been
// decided, undoing it can have payroll or scheduling consequences a bare
// status flip should not trigger silently; that path is a request to a
// manager, not a self-service button.
func (s *Service) CancelLeaveRequest(ctx context.Context, ownerID, callerUserID, requestID int64) error {
	if s.permissions == nil || callerUserID <= 0 {
		return domain.ErrLeaveAccessDenied
	}
	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil || employeeID <= 0 {
		return domain.ErrLeaveAccessDenied
	}

	req, err := s.repo.GetLeaveRequest(ctx, ownerID, requestID)
	if err != nil {
		return err
	}
	if req.EmployeeID != employeeID {
		return domain.ErrLeaveAccessDenied
	}
	if req.Status != domain.LeaveStatusPending {
		return domain.ErrLeaveRequestNotPending
	}

	startDate, parseErr := time.Parse("2006-01-02", req.StartDate)
	if parseErr != nil {
		return parseErr
	}
	endDate, parseErr := time.Parse("2006-01-02", req.EndDate)
	if parseErr != nil {
		return parseErr
	}
	if err := s.repo.CancelLeaveRequestWithBalance(ctx, ownerID, requestID, req.EmployeeID, req.LeaveTypeID, leaveDayAllocations(startDate, endDate)); err != nil {
		return err
	}

	invalidateHRLeaveScope(ctx, ownerID)
	return nil
}

// scopeLeaveRequests resolves the authenticated user's employee row before
// applying any client-supplied filter. A caller with the leave approval
// permission may request the team feed; every other caller is forced to their
// own employee row, even when employee_id is supplied by the client.
func (s *Service) scopeLeaveRequests(ctx context.Context, ownerID, callerUserID, requestedEmployeeID int64) (int64, error) {
	if s.permissions == nil || callerUserID <= 0 {
		return 0, domain.ErrLeaveAccessDenied
	}

	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil || employeeID <= 0 {
		return 0, domain.ErrLeaveAccessDenied
	}

	canViewTeam, permissionErr := s.permissions.HasPermission(ctx, ownerID, employeeID, permissionApproveTimeOff)
	if permissionErr == nil && canViewTeam {
		return requestedEmployeeID, nil
	}
	return employeeID, nil
}

// authorizeLeaveSubject verifies the caller may act on employeeID's leave -
// either it is their own, or they hold the leave approval permission.
// Unlike scopeLeaveRequests, which narrows an unauthorized list to the
// caller's own view rather than fail the read, a write silently substituting
// a different employee than the one requested would file the wrong person's
// leave without anyone noticing - so this denies outright instead.
func (s *Service) authorizeLeaveSubject(ctx context.Context, ownerID, callerUserID, requestedEmployeeID int64) error {
	if s.permissions == nil || callerUserID <= 0 {
		return domain.ErrLeaveAccessDenied
	}

	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, callerUserID)
	if err != nil || employeeID <= 0 {
		return domain.ErrLeaveAccessDenied
	}
	if employeeID == requestedEmployeeID {
		return nil
	}

	canActForOthers, permissionErr := s.permissions.HasPermission(ctx, ownerID, employeeID, permissionApproveTimeOff)
	if permissionErr == nil && canActForOthers {
		return nil
	}
	return domain.ErrLeaveAccessDenied
}

// authorizedLeaveDecision performs the permission and approval-chain checks
// before loading the requested row. This keeps an unauthorised caller from
// using a request id to probe another employee's leave data or mutate it.
func (s *Service) authorizedLeaveDecision(ctx context.Context, ownerID, requestID, approverUserID int64) (*domain.LeaveRequestRecord, *domain.LeaveTypeRecord, error) {
	if s.permissions == nil || approverUserID <= 0 {
		return nil, nil, domain.ErrLeaveAccessDenied
	}

	employeeID, err := s.permissions.GetEmployeeIDByUserID(ctx, ownerID, approverUserID)
	if err != nil || employeeID <= 0 {
		return nil, nil, domain.ErrLeaveAccessDenied
	}
	allowed, err := s.permissions.HasPermission(ctx, ownerID, employeeID, permissionApproveTimeOff)
	if err != nil {
		return nil, nil, err
	}
	if !allowed {
		return nil, nil, domain.ErrLeaveApproverNotAllowed
	}

	req, err := s.repo.GetLeaveRequest(ctx, ownerID, requestID)
	if err != nil {
		return nil, nil, err
	}
	if req.EmployeeID == employeeID {
		return nil, nil, domain.ErrLeaveSelfApproval
	}
	lt, err := s.repo.GetLeaveType(ctx, ownerID, req.LeaveTypeID)
	if err != nil {
		return nil, nil, err
	}

	if !lt.RequiresApproval {
		return req, lt, nil
	}
	if s.approvalProvider == nil {
		return nil, nil, domain.ErrApprovalChainNotConfigured
	}
	chain, err := s.approvalProvider.GetApprovalChain(ctx, ownerID, "leave")
	if err != nil {
		return nil, nil, err
	}
	if chain == nil {
		return nil, nil, domain.ErrApprovalChainNotConfigured
	}
	if len(chain.Levels) == 0 {
		return req, lt, nil
	}

	matched, err := s.approverMatchesChain(ctx, ownerID, employeeID, req.CurrentApprovalLevel, chain.Levels)
	if err != nil {
		return nil, nil, err
	}
	if !matched {
		return nil, nil, domain.ErrLeaveApproverNotAllowed
	}
	return req, lt, nil
}
