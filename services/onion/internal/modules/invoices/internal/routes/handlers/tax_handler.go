package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// List Tax Configs
// ---------------------------------------------------------------------------

func ListTaxConfigRoutes() []string {
	return []string{"/finance/tax/config"}
}

type ListTaxConfigHandler struct {
	FinanceService *application.Service
}

type ListTaxConfigRequest struct{}

type TaxConfigPayload struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Name      string    `json:"name"`
	Rate      float64   `json:"rate"`
	Type      string    `json:"type"`
	IsDefault bool      `json:"is_default"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type ListTaxConfigResponse struct {
	Configs []TaxConfigPayload `json:"configs"`
}

func newTaxConfigPayload(t domain.TaxConfig) TaxConfigPayload {
	return TaxConfigPayload{
		ID:        t.ID,
		OwnerID:   t.OwnerID,
		Name:      t.Name,
		Rate:      t.Rate,
		Type:      t.Type,
		IsDefault: t.IsDefault,
		IsActive:  t.IsActive,
		CreatedAt: t.CreatedAt,
	}
}

func (h *ListTaxConfigHandler) Serve(ctx context.Context, _ *ListTaxConfigRequest) (ListTaxConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	configs, err := h.FinanceService.ListTaxConfigs(ctx, device.User.UserID)
	if err != nil {
		return ListTaxConfigResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list tax configs",
		}
	}

	payloads := make([]TaxConfigPayload, 0, len(configs))
	for _, c := range configs {
		payloads = append(payloads, newTaxConfigPayload(c))
	}

	return ListTaxConfigResponse{Configs: payloads}, nil
}

// ---------------------------------------------------------------------------
// Create Tax Config
// ---------------------------------------------------------------------------

func CreateTaxConfigRoutes() []string {
	return []string{"/finance/tax/config"}
}

type CreateTaxConfigHandler struct {
	FinanceService *application.Service
}

type CreateTaxConfigRequest struct {
	Name      string  `json:"name"`
	Rate      float64 `json:"rate"`
	Type      string  `json:"type"`
	IsDefault bool    `json:"is_default"`
}

type CreateTaxConfigResponse struct {
	Config TaxConfigPayload `json:"config"`
}

func (h *CreateTaxConfigHandler) Validate(_ context.Context, req *CreateTaxConfigRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Tax config name is required",
		}
	}

	if req.Rate < 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Tax rate must be non-negative",
		}
	}

	return nil
}

func (h *CreateTaxConfigHandler) Serve(ctx context.Context, req *CreateTaxConfigRequest) (CreateTaxConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	input := domain.TaxConfigInput{
		Name:      req.Name,
		Rate:      req.Rate,
		Type:      req.Type,
		IsDefault: req.IsDefault,
	}

	t, err := h.FinanceService.CreateTaxConfig(ctx, device.User.UserID, input)
	if err != nil {
		return CreateTaxConfigResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create tax config",
		}
	}

	return CreateTaxConfigResponse{Config: newTaxConfigPayload(*t)}, nil
}

// ---------------------------------------------------------------------------
// Tax Report
// ---------------------------------------------------------------------------

func TaxReportRoutes() []string {
	return []string{"/finance/tax/report"}
}

type TaxReportHandler struct {
	FinanceService *application.Service
}

type TaxReportRequest struct {
	Period string `json:"period"`
	Date   string `json:"date"`
}

type TaxReportResponse struct {
	Period       string  `json:"period"`
	TaxCollected float64 `json:"tax_collected"`
	TaxOwed      float64 `json:"tax_owed"`
}

func (h *TaxReportHandler) Serve(ctx context.Context, req *TaxReportRequest) (TaxReportResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	period := req.Period
	if period == "" {
		period = "month"
	}

	report, err := h.FinanceService.GetTaxReport(ctx, device.User.UserID, period, req.Date)
	if err != nil {
		return TaxReportResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate tax report",
		}
	}

	return TaxReportResponse{
		Period:       report.Period,
		TaxCollected: report.TaxCollected,
		TaxOwed:      report.TaxOwed,
	}, nil
}
