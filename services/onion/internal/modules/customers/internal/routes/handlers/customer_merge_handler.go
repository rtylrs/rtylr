package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *MergeCustomersHandler) Validate(_ context.Context, req *MergeCustomersRequest) error {
	if req.PrimaryID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Primary customer ID is required")
	}

	if req.SecondaryID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Secondary customer ID is required")
	}

	if req.PrimaryID == req.SecondaryID {
		return onekiterrors.New(http.StatusBadRequest, "Cannot merge a customer with itself")
	}

	return nil
}

func (h *MergeCustomersHandler) Serve(ctx context.Context, req *MergeCustomersRequest) (MergeCustomersResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return MergeCustomersResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:delete"); err != nil {
		return MergeCustomersResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return MergeCustomersResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.CRMService.MergeCustomers(ctx, domain.MergeCustomersInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		PrimaryID:   req.PrimaryID,
		SecondaryID: req.SecondaryID,
	})
	if err != nil {
		return MergeCustomersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to merge customers")
	}

	return MergeCustomersResponse{Message: "Customers merged successfully"}, nil
}
