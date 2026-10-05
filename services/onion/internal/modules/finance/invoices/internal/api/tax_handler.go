package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/finance/invoices/internal/application"
	"github.com/voxire/onion/internal/modules/finance/invoices/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ---------------------------------------------------------------------------
// List Tax Configs
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Tax Report
// ---------------------------------------------------------------------------

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
