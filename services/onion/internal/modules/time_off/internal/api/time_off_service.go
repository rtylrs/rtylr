package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/time_off/internal/application"
	hrtimeoffv1 "github.com/voxire/voxie/gen/go/hr/time_off/v1"
)

type TimeOffAPI struct {
	Service *application.Service
}

func NewTimeOffAPI(service *application.Service) *TimeOffAPI {
	return &TimeOffAPI{Service: service}
}

func (a *TimeOffAPI) ListTimeOff(ctx context.Context, _ *hrtimeoffv1.ListTimeOffRequest) (*hrtimeoffv1.ListTimeOffResponse, error) {
	result, err := (&ListTimeOffHandler{Service: a.Service}).Serve(ctx, &ListTimeOffRequest{})
	if err != nil {
		return nil, err
	}
	return &hrtimeoffv1.ListTimeOffResponse{Requests: result.Requests}, nil
}

func (a *TimeOffAPI) SubmitTimeOff(ctx context.Context, req *hrtimeoffv1.SubmitTimeOffRequest) (*hrtimeoffv1.SubmitTimeOffResponse, error) {
	result, err := (&SubmitTimeOffHandler{Service: a.Service}).Serve(ctx, &SubmitTimeOffRequest{
		StaffID:   req.GetStaffId(),
		Type:      req.GetType(),
		StartDate: req.GetStartDate(),
		EndDate:   req.GetEndDate(),
		DaysCount: req.GetDaysCount(),
		Reason:    req.GetReason(),
		Notes:     req.GetNotes(),
	})
	if err != nil {
		return nil, err
	}
	return &hrtimeoffv1.SubmitTimeOffResponse{Id: result.ID}, nil
}
