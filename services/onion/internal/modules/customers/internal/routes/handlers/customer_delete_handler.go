package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *DeleteCustomerHandler) Validate(_ context.Context, req *DeleteCustomerRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	return nil
}

func (h *DeleteCustomerHandler) Serve(ctx context.Context, req *DeleteCustomerRequest) (DeleteCustomerResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteCustomerResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:delete"); err != nil {
		return DeleteCustomerResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteCustomerResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.CRMService.DeleteCustomer(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return DeleteCustomerResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return DeleteCustomerResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete customer")
	}

	return DeleteCustomerResponse{Message: "Customer deleted successfully"}, nil
}
