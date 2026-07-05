package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	hrovertimev1 "github.com/voxire/voxie/gen/go/hr/overtime/v1"
)

type OvertimeService struct {
	service *application.Service
}

func NewOvertimeService(service *application.Service) *OvertimeService {
	return &OvertimeService{service: service}
}

func (s *OvertimeService) ListOvertimeRequests(ctx context.Context, req *hrovertimev1.ListOvertimeRequestsRequest) (*hrovertimev1.ListOvertimeRequestsResponse, error) {
	res, err := (&ListOvertimeRequestsHandler{Service: s.service}).Serve(ctx, &ListOvertimeRequestsRequest{})
	if err != nil {
		return nil, err
	}
	return &hrovertimev1.ListOvertimeRequestsResponse{
		Requests: res.Requests,
		Total:    res.Total,
		Limit:    res.Limit,
		Offset:   res.Offset,
	}, nil
}

func (s *OvertimeService) CreateOvertimeRequest(ctx context.Context, req *hrovertimev1.CreateOvertimeRequestRequest) (*hrovertimev1.CreateOvertimeRequestResponse, error) {
	in := &CreateOvertimeRequestRequest{
		EmployeeID:          req.GetEmployeeId(),
		Date:                req.GetDate(),
		PlannedExtraMinutes: req.GetPlannedExtraMinutes(),
		Reason:              req.GetReason(),
	}
	handler := &CreateOvertimeRequestHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrovertimev1.CreateOvertimeRequestResponse{Request: res.Request}, nil
}

func (s *OvertimeService) ApproveOvertime(ctx context.Context, req *hrovertimev1.ApproveOvertimeRequest) (*hrovertimev1.ApproveOvertimeResponse, error) {
	body := req.GetBody()
	in := &ApproveOvertimeRequestReq{
		RequestID:         req.GetOvertimeRequestId(),
		Approved:          body.GetApproved(),
		MultiplierApplied: body.GetMultiplierApplied(),
	}
	handler := &ApproveOvertimeHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrovertimev1.ApproveOvertimeResponse{Success: res.Success}, nil
}
