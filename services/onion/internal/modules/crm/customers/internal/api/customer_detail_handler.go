package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *GetCustomerHandler) Validate(_ context.Context, req *GetCustomerRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	return nil
}

func (h *GetCustomerHandler) Serve(ctx context.Context, req *GetCustomerRequest) (GetCustomerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	customer, err := h.CRMService.GetCustomer(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return GetCustomerResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return GetCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get customer")
	}

	return GetCustomerResponse{
		Customer: newCustomerPayload(customer),
	}, nil
}
