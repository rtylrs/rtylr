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
)

func CustomerDeleteRoutes() []string {
	return []string{"/crm/customers"}
}

type DeleteCustomerHandler struct {
	CRMService *application.Service
}

type DeleteCustomerRequest struct {
	ID int64 `json:"id"`
}

type DeleteCustomerResponse struct {
	Message string `json:"message"`
}

func (h *DeleteCustomerHandler) Validate(_ context.Context, req *DeleteCustomerRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	return nil
}

func (h *DeleteCustomerHandler) Serve(ctx context.Context, req *DeleteCustomerRequest) (DeleteCustomerResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.CRMService.DeleteCustomer(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return DeleteCustomerResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Customer not found",
			}
		}

		return DeleteCustomerResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete customer",
		}
	}

	return DeleteCustomerResponse{Message: "Customer deleted successfully"}, nil
}
