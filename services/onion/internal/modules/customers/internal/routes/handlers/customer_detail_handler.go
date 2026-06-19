package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcustomerv1 "github.com/voxire/voxie/gen/go/crm/customer/v1"
)

func CustomerDetailRoutes() []string {
	return []string{"/crm/customers/detail"}
}

type GetCustomerHandler struct {
	CRMService *application.Service
}

type GetCustomerRequest struct {
	ID int64 `json:"id"`
}

type GetCustomerResponse struct {
	Customer *crmcustomerv1.Customer `json:"customer"`
}

func (h *GetCustomerHandler) Validate(_ context.Context, req *GetCustomerRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	return nil
}

func (h *GetCustomerHandler) Serve(ctx context.Context, req *GetCustomerRequest) (GetCustomerResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}

	customer, err := h.CRMService.GetCustomer(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return GetCustomerResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Customer not found",
			}
		}

		return GetCustomerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get customer",
		}
	}

	return GetCustomerResponse{
		Customer: newCustomerPayload(customer),
	}, nil
}
