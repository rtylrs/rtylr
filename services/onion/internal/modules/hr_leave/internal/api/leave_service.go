package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

type LeaveService struct {
	service *application.Service
}

func NewLeaveService(service *application.Service) *LeaveService {
	return &LeaveService{service: service}
}

func (s *LeaveService) ListLeaveRequests(ctx context.Context, req *hrleavev1.ListLeaveRequestsRequest) (*hrleavev1.ListLeaveRequestsResponse, error) {
	if req == nil {
		req = &hrleavev1.ListLeaveRequestsRequest{}
	}
	res, err := (&ListLeaveRequestsHandler{Service: s.service}).Serve(ctx, &ListLeaveRequestsRequest{
		EmployeeID: req.GetEmployeeId(),
		Status:     req.GetStatus(),
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		Limit:      req.GetLimit(),
		Offset:     req.GetOffset(),
	})
	if err != nil {
		return nil, err
	}
	return &hrleavev1.ListLeaveRequestsResponse{
		Requests: res.Requests,
		Total:    res.Total,
		Limit:    res.Limit,
		Offset:   res.Offset,
	}, nil
}

func (s *LeaveService) CreateLeaveRequest(ctx context.Context, req *hrleavev1.CreateLeaveRequestRequest) (*hrleavev1.CreateLeaveRequestResponse, error) {
	in := &CreateLeaveRequestRequest{
		EmployeeID:  req.GetEmployeeId(),
		LeaveTypeID: req.GetLeaveTypeId(),
		StartDate:   req.GetStartDate(),
		EndDate:     req.GetEndDate(),
		TotalDays:   req.GetTotalDays(),
		Notes:       req.GetNotes(),
	}
	handler := &CreateLeaveRequestHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrleavev1.CreateLeaveRequestResponse{Request: res.Request}, nil
}

func (s *LeaveService) ApproveLeave(ctx context.Context, req *hrleavev1.ApproveLeaveRequest) (*hrleavev1.ApproveLeaveResponse, error) {
	body := req.GetBody()
	in := &ApproveLeaveRequest{
		RequestID:       req.GetRequestId(),
		Approved:        body.GetApproved(),
		RejectionReason: body.GetRejectionReason(),
	}
	handler := &ApproveLeaveHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrleavev1.ApproveLeaveResponse{Success: res.Success}, nil
}

func (s *LeaveService) CancelLeaveRequest(ctx context.Context, req *hrleavev1.CancelLeaveRequestRequest) (*hrleavev1.CancelLeaveResponse, error) {
	in := &CancelLeaveRequestRequest{RequestID: req.GetRequestId()}
	handler := &CancelLeaveRequestHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrleavev1.CancelLeaveResponse{Success: res.Success}, nil
}

func (s *LeaveService) ListLeaveTypes(ctx context.Context, _ *hrleavev1.ListLeaveTypesRequest) (*hrleavev1.ListLeaveTypesResponse, error) {
	res, err := (&ListLeaveTypesHandler{Service: s.service}).Serve(ctx, &ListLeaveTypesRequest{})
	if err != nil {
		return nil, err
	}
	return &hrleavev1.ListLeaveTypesResponse{
		LeaveTypes: res.Types,
		Total:      res.Total,
		Limit:      res.Limit,
		Offset:     res.Offset,
	}, nil
}

func (s *LeaveService) CreateLeaveType(ctx context.Context, req *hrleavev1.CreateLeaveTypeRequest) (*hrleavev1.CreateLeaveTypeResponse, error) {
	in := &CreateLeaveTypeRequest{
		Name:                           req.GetName(),
		IsPaid:                         req.GetIsPaid(),
		RequiresApproval:               req.GetRequiresApproval(),
		AccrualRule:                    req.GetAccrualRule(),
		MaxCarryoverDays:               req.GetMaxCarryoverDays(),
		MinNoticeDays:                  req.GetMinNoticeDays(),
		RequiresDocumentationAfterDays: req.GetRequiresDocumentationAfterDays(),
	}
	handler := &CreateLeaveTypeHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrleavev1.CreateLeaveTypeResponse{LeaveType: res.LeaveType}, nil
}

func (s *LeaveService) UpdateLeaveType(ctx context.Context, req *hrleavev1.UpdateLeaveTypeRequest) (*hrleavev1.UpdateLeaveTypeResponse, error) {
	if req == nil {
		return nil, (&UpdateLeaveTypeHandler{Service: s.service}).Validate(ctx, nil)
	}
	body := req.GetBody()
	in := updateLeaveTypeRequestFromProto(req.GetId(), body)
	handler := &UpdateLeaveTypeHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrleavev1.UpdateLeaveTypeResponse{LeaveType: res.LeaveType}, nil
}

func updateLeaveTypeRequestFromProto(id int64, body *hrleavev1.UpdateLeaveTypeBody) *UpdateLeaveTypeRequest {
	in := &UpdateLeaveTypeRequest{ID: id}
	if body == nil {
		return in
	}
	in.Name = body.Name
	in.IsPaid = body.IsPaid
	in.RequiresApproval = body.RequiresApproval
	in.AccrualRule = body.AccrualRule
	in.MaxCarryoverDays = body.MaxCarryoverDays
	in.MinNoticeDays = body.MinNoticeDays
	in.RequiresDocumentationAfterDays = body.RequiresDocumentationAfterDays
	in.IsActive = body.IsActive
	return in
}

func (s *LeaveService) ListLeaveBalances(ctx context.Context, req *hrleavev1.ListLeaveBalancesRequest) (*hrleavev1.ListLeaveBalancesResponse, error) {
	if req == nil {
		req = &hrleavev1.ListLeaveBalancesRequest{}
	}
	res, err := (&GetLeaveBalancesHandler{Service: s.service}).Serve(ctx, &GetLeaveBalancesRequest{
		EmployeeID: req.GetEmployeeId(),
		Year:       req.GetYear(),
	})
	if err != nil {
		return nil, err
	}
	return &hrleavev1.ListLeaveBalancesResponse{
		Balances: res.Balances,
		Total:    res.Total,
		Limit:    res.Limit,
		Offset:   res.Offset,
	}, nil
}
