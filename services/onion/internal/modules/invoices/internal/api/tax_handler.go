package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
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

// ListTaxConfigResponse is a top-level array (TaxConfig[]) rather than an
// object envelope. The frontend hits this endpoint to render a select/list
// of configured tax rates; returning a bare array matches the shape the
// other config-list endpoints in this repo already expose and keeps the
// frontend adapter minimal.
type ListTaxConfigResponse []TaxConfigPayload

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

func (h *ListTaxConfigHandler) Serve(ctx context.Context, _ *ListTaxConfigRequest) (ListTaxConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return ListTaxConfigResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:tax:configure"); err != nil {
		return ListTaxConfigResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListTaxConfigResponse{}, onekiterrors.PermissionDenied()
	}

	configs, err := h.FinanceService.ListTaxConfigs(ctx, businessID)
	if err != nil {
		return ListTaxConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list tax configs")
	}

	payloads := make(ListTaxConfigResponse, 0, len(configs))
	for _, c := range configs {
		payloads = append(payloads, newTaxConfigPayload(c))
	}

	return payloads, nil
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

func (h *CreateTaxConfigHandler) Validate(_ context.Context, req *CreateTaxConfigRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Tax config name is required")
	}

	if req.Rate < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Tax rate must be non-negative")
	}

	return nil
}

func (h *CreateTaxConfigHandler) Serve(ctx context.Context, req *CreateTaxConfigRequest) (CreateTaxConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return CreateTaxConfigResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:tax:configure"); err != nil {
		return CreateTaxConfigResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateTaxConfigResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.TaxConfigInput{
		Name:      req.Name,
		Rate:      req.Rate,
		Type:      req.Type,
		IsDefault: req.IsDefault,
	}

	t, err := h.FinanceService.CreateTaxConfig(ctx, businessID, input)
	if err != nil {
		return CreateTaxConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create tax config")
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

func (h *TaxReportHandler) Serve(ctx context.Context, req *TaxReportRequest) (TaxReportResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return TaxReportResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:reports:view"); err != nil {
		return TaxReportResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return TaxReportResponse{}, onekiterrors.PermissionDenied()
	}

	period := req.Period
	if period == "" {
		period = "month"
	}

	report, err := h.FinanceService.GetTaxReport(ctx, businessID, period, req.Date)
	if err != nil {
		return TaxReportResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to generate tax report")
	}

	return TaxReportResponse{
		Period:       report.Period,
		TaxCollected: report.TaxCollected,
		TaxOwed:      report.TaxOwed,
	}, nil
}
