package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/tax/internal/application"
	"github.com/voxire/onion/internal/modules/tax/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	financetaxv1 "github.com/voxire/voxie/gen/go/finance/tax/v1"
)

type TaxAPI struct {
	service *application.Service
}

func NewTaxAPI(service *application.Service) *TaxAPI {
	return &TaxAPI{service: service}
}

func (a *TaxAPI) GetTaxConfig(ctx context.Context, _ *financetaxv1.GetTaxConfigRequest) (*financetaxv1.GetTaxConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	config, err := a.service.GetConfig(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get tax config")
	}

	return &financetaxv1.GetTaxConfigResponse{Config: newTaxConfig(config)}, nil
}

func (a *TaxAPI) UpdateTaxConfig(ctx context.Context, req *financetaxv1.UpdateTaxConfigRequest) (*financetaxv1.UpdateTaxConfigResponse, error) {
	if req.GetTaxName() == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Tax name is required")
	}
	if req.GetTaxRate() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Tax rate must not be negative")
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	err := a.service.UpdateConfig(ctx, domain.TaxConfigInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		TaxName:     req.GetTaxName(),
		TaxRate:     req.GetTaxRate(),
		IsInclusive: req.GetIsInclusive(),
		IsActive:    req.GetIsActive(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update tax config")
	}

	return &financetaxv1.UpdateTaxConfigResponse{Message: "Tax config updated successfully"}, nil
}

func (a *TaxAPI) GetPLReport(ctx context.Context, _ *financetaxv1.GetPLReportRequest) (*financetaxv1.GetPLReportResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	report, err := a.service.GetPLReport(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get P&L report")
	}

	return &financetaxv1.GetPLReportResponse{
		TotalRevenue:  report.TotalRevenue,
		TotalExpenses: report.TotalExpenses,
		GrossProfit:   report.GrossProfit,
		NetProfit:     report.NetProfit,
	}, nil
}

func newTaxConfig(config domain.TaxConfig) *financetaxv1.TaxConfig {
	return &financetaxv1.TaxConfig{
		Id:          config.ID,
		OwnerId:     config.OwnerID,
		TaxName:     config.TaxName,
		TaxRate:     config.TaxRate,
		IsInclusive: config.IsInclusive,
		IsActive:    config.IsActive,
		CreatedAt:   config.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   config.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
