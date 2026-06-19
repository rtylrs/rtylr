package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CustomerMergeRoutes() []string {
	return []string{"/crm/customers/merge"}
}

type MergeCustomersHandler struct {
	CRMService *application.Service
}

type MergeCustomersRequest struct {
	PrimaryID   int64 `json:"primary_id"`
	SecondaryID int64 `json:"secondary_id"`
}

type MergeCustomersResponse struct {
	Message string `json:"message"`
}

func (h *MergeCustomersHandler) Validate(_ context.Context, req *MergeCustomersRequest) *errorModels.ServerError {
	if req.PrimaryID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Primary customer ID is required",
		}
	}

	if req.SecondaryID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Secondary customer ID is required",
		}
	}

	if req.PrimaryID == req.SecondaryID {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Cannot merge a customer with itself",
		}
	}

	return nil
}

func (h *MergeCustomersHandler) Serve(ctx context.Context, req *MergeCustomersRequest) (MergeCustomersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MergeCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MergeCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.CRMService.MergeCustomers(ctx, domain.MergeCustomersInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		PrimaryID:   req.PrimaryID,
		SecondaryID: req.SecondaryID,
	})
	if err != nil {
		return MergeCustomersResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to merge customers",
		}
	}

	return MergeCustomersResponse{Message: "Customers merged successfully"}, nil
}
