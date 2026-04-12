package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/tax/internal/application"
	"github.com/voxire/onion/internal/modules/tax/internal/domain"
)

// ─── Get Tax Config ───────────────────────────────────────────────────────────

func TaxConfigGetRoutes() []string { return []string{"/finance/tax"} }

type GetTaxConfigHandler struct{ Service *application.Service }
type GetTaxConfigRequest struct{}

type TaxConfigPayload struct {
	ID          int64   `json:"id"`
	OwnerID     int64   `json:"owner_id"`
	TaxName     string  `json:"tax_name"`
	TaxRate     float64 `json:"tax_rate"`
	IsInclusive bool    `json:"is_inclusive"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type GetTaxConfigResponse struct {
	Config TaxConfigPayload `json:"config"`
}

func (h *GetTaxConfigHandler) Serve(ctx context.Context, _ *GetTaxConfigRequest) (GetTaxConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTaxConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	cfg, err := h.Service.GetConfig(ctx, device.User.UserID)
	if err != nil {
		return GetTaxConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get tax config"}
	}
	return GetTaxConfigResponse{Config: TaxConfigPayload{
		ID:          cfg.ID,
		OwnerID:     cfg.OwnerID,
		TaxName:     cfg.TaxName,
		TaxRate:     cfg.TaxRate,
		IsInclusive: cfg.IsInclusive,
		IsActive:    cfg.IsActive,
		CreatedAt:   cfg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}}, nil
}

// ─── Update Tax Config ────────────────────────────────────────────────────────

func TaxConfigUpdateRoutes() []string { return []string{"/finance/tax"} }

type UpdateTaxConfigHandler struct{ Service *application.Service }

type UpdateTaxConfigRequest struct {
	TaxName     string  `json:"tax_name"`
	TaxRate     float64 `json:"tax_rate"`
	IsInclusive bool    `json:"is_inclusive"`
	IsActive    bool    `json:"is_active"`
}

type UpdateTaxConfigResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTaxConfigHandler) Serve(ctx context.Context, req *UpdateTaxConfigRequest) (UpdateTaxConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTaxConfigResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.UpdateConfig(ctx, domain.TaxConfigInput{
		OwnerID:     device.User.UserID,
		TaxName:     req.TaxName,
		TaxRate:     req.TaxRate,
		IsInclusive: req.IsInclusive,
		IsActive:    req.IsActive,
	})
	if err != nil {
		return UpdateTaxConfigResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update tax config"}
	}
	return UpdateTaxConfigResponse{Message: "Tax config updated successfully"}, nil
}

// ─── P&L Report ───────────────────────────────────────────────────────────────

func PLReportRoutes() []string { return []string{"/finance/pl"} }

type GetPLReportHandler struct{ Service *application.Service }
type GetPLReportRequest struct{}

type PLReportResponse struct {
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	GrossProfit   float64 `json:"gross_profit"`
	NetProfit     float64 `json:"net_profit"`
}

func (h *GetPLReportHandler) Serve(ctx context.Context, _ *GetPLReportRequest) (PLReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return PLReportResponse{}, errorModels.PermissionDeniedServerError()
	}
	report, err := h.Service.GetPLReport(ctx, device.User.UserID)
	if err != nil {
		return PLReportResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get P&L report"}
	}
	return PLReportResponse{
		TotalRevenue:  report.TotalRevenue,
		TotalExpenses: report.TotalExpenses,
		GrossProfit:   report.GrossProfit,
		NetProfit:     report.NetProfit,
	}, nil
}
