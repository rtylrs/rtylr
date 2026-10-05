package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/benefits/internal/application"
	"github.com/voxire/onion/internal/modules/hr/benefits/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrbenefitv1 "github.com/voxire/voxie/gen/go/hr/benefit/v1"
)

type BenefitAPI struct {
	service *application.Service
}

func NewBenefitAPI(service *application.Service) *BenefitAPI {
	return &BenefitAPI{service: service}
}

func (a *BenefitAPI) ListBenefits(ctx context.Context, _ *hrbenefitv1.ListBenefitsRequest) (*hrbenefitv1.ListBenefitsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	items, err := a.service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list benefit plans")
	}

	plans := make([]*hrbenefitv1.BenefitPlan, 0, len(items))
	for _, item := range items {
		plans = append(plans, newBenefitPlan(item))
	}

	return &hrbenefitv1.ListBenefitsResponse{Plans: plans}, nil
}

func (a *BenefitAPI) CreateBenefit(ctx context.Context, req *hrbenefitv1.CreateBenefitRequest) (*hrbenefitv1.CreateBenefitResponse, error) {
	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name is required")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.service.Create(ctx, businessID, domain.BenefitPlanInput{
		Name:         name,
		Type:         req.GetType(),
		Provider:     req.GetProvider(),
		CostEmployer: req.GetCostEmployer(),
		CostEmployee: req.GetCostEmployee(),
		IsActive:     req.GetIsActive(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create benefit plan")
	}

	return &hrbenefitv1.CreateBenefitResponse{Id: id, Message: "Benefit plan created successfully"}, nil
}

func newBenefitPlan(plan domain.BenefitPlan) *hrbenefitv1.BenefitPlan {
	return &hrbenefitv1.BenefitPlan{
		Id:           plan.ID,
		OwnerId:      plan.OwnerID,
		Name:         plan.Name,
		Type:         plan.Type,
		Provider:     plan.Provider,
		CostEmployer: plan.CostEmployer,
		CostEmployee: plan.CostEmployee,
		IsActive:     plan.IsActive,
		CreatedAt:    plan.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
