package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr/hr_performance/internal/application"
	hrperformancev1 "github.com/voxire/voxie/gen/go/hr/performance/v1"
)

type PerformanceService struct {
	service *application.Service
}

func NewPerformanceService(service *application.Service) *PerformanceService {
	return &PerformanceService{service: service}
}

func (s *PerformanceService) ListReviewCycles(ctx context.Context, _ *hrperformancev1.ListReviewCyclesRequest) (*hrperformancev1.ListReviewCyclesResponse, error) {
	pagination := performancePaginationFromContext(ctx)
	res, err := (&ListCyclesHandler{Service: s.service}).Serve(ctx, &ListCyclesRequest{
		Limit:  pagination.limit,
		Offset: pagination.offset,
	})
	if err != nil {
		return nil, err
	}
	return &hrperformancev1.ListReviewCyclesResponse{
		Cycles: res.Cycles,
		Total:  res.Total,
		Limit:  res.Limit,
		Offset: res.Offset,
	}, nil
}

func (s *PerformanceService) CreateReviewCycle(ctx context.Context, req *hrperformancev1.CreateReviewCycleRequest) (*hrperformancev1.CreateReviewCycleResponse, error) {
	in := &CreateCycleRequest{
		Name:        req.GetName(),
		CycleType:   req.GetType(),
		PeriodStart: req.GetPeriodStart(),
		PeriodEnd:   req.GetPeriodEnd(),
	}
	handler := &CreateCycleHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	res, err := handler.Serve(ctx, in)
	if err != nil {
		return nil, err
	}
	return &hrperformancev1.CreateReviewCycleResponse{Cycle: res.Cycle}, nil
}

func (s *PerformanceService) OpenReviewCycle(ctx context.Context, req *hrperformancev1.OpenReviewCycleRequest) (*hrperformancev1.OpenReviewCycleResponse, error) {
	res, err := (&OpenCycleHandler{Service: s.service}).Serve(ctx, &OpenCycleRequest{ID: req.GetCycleId()})
	if err != nil {
		return nil, err
	}
	if !res.Success {
		return &hrperformancev1.OpenReviewCycleResponse{}, nil
	}
	return &hrperformancev1.OpenReviewCycleResponse{}, nil
}

func (s *PerformanceService) ListGoals(ctx context.Context, req *hrperformancev1.ListGoalsRequest) (*hrperformancev1.ListGoalsResponse, error) {
	if req == nil {
		req = &hrperformancev1.ListGoalsRequest{}
	}
	pagination := performancePaginationFromContext(ctx)
	res, err := (&ListGoalsHandler{Service: s.service}).Serve(ctx, &ListGoalsRequest{
		EmployeeID:    req.GetEmployeeId(),
		ReviewCycleID: req.GetReviewCycleId(),
		Limit:         pagination.limit,
		Offset:        pagination.offset,
	})
	if err != nil {
		return nil, err
	}
	return &hrperformancev1.ListGoalsResponse{
		Goals:  res.Goals,
		Total:  res.Total,
		Limit:  res.Limit,
		Offset: res.Offset,
	}, nil
}

func (s *PerformanceService) SubmitReview(ctx context.Context, req *hrperformancev1.SubmitReviewRequest) (*hrperformancev1.SubmitReviewResponse, error) {
	in := &SubmitReviewRequest{
		ID:        req.GetReviewId(),
		Rating:    req.GetOverallRating(),
		Narrative: req.GetNarrative(),
	}
	handler := &SubmitReviewHandler{Service: s.service}
	if err := handler.Validate(ctx, in); err != nil {
		return nil, err
	}
	if _, err := handler.Serve(ctx, in); err != nil {
		return nil, err
	}
	return &hrperformancev1.SubmitReviewResponse{}, nil
}
