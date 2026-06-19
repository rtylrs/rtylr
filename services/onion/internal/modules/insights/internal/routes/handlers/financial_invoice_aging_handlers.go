package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/financial/invoice-aging

func GetInvoiceAgingRoutes() []string { return []string{"/insights/financial/invoice-aging"} }

type GetInvoiceAgingHandler struct{ Service *application.Service }

type GetInvoiceAgingRequest struct{}

type GetInvoiceAgingResponse struct {
	Buckets []application.InvoiceAgingBucketPayload `json:"buckets"`
}

func (h *GetInvoiceAgingHandler) Serve(ctx context.Context, _ *GetInvoiceAgingRequest) (GetInvoiceAgingResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetInvoiceAgingResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetInvoiceAgingResponse{}, errorModels.PermissionDeniedServerError()
	}
	rows, err := h.Service.GetInvoiceAging(ctx, businessID)
	if err != nil {
		return GetInvoiceAgingResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load invoice aging",
		}
	}
	return GetInvoiceAgingResponse{Buckets: rows}, nil
}
