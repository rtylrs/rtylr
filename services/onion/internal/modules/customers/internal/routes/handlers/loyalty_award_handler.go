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

func LoyaltyAwardRoutes() []string {
	return []string{"/crm/loyalty/award"}
}

type AwardPointsHandler struct {
	CRMService *application.Service
}

type AwardPointsRequest struct {
	CustomerID    int64  `json:"customer_id"`
	Points        int    `json:"points"`
	Description   string `json:"description"`
	ReferenceID   *int64 `json:"reference_id"`
	ReferenceType string `json:"reference_type"`
}

type AwardPointsResponse struct {
	Message string `json:"message"`
}

func (h *AwardPointsHandler) Validate(_ context.Context, req *AwardPointsRequest) *errorModels.ServerError {
	if req.CustomerID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Customer ID is required",
		}
	}

	if req.Points <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Points must be greater than zero",
		}
	}

	return nil
}

func (h *AwardPointsHandler) Serve(ctx context.Context, req *AwardPointsRequest) (AwardPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return AwardPointsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return AwardPointsResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.CRMService.AwardPoints(ctx, domain.AwardPointsInput{
		OwnerID:       device.User.UserID,
		BusinessID:    businessID,
		CustomerID:    req.CustomerID,
		Points:        req.Points,
		Description:   req.Description,
		ReferenceID:   req.ReferenceID,
		ReferenceType: req.ReferenceType,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return AwardPointsResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Customer not found",
			}
		}

		return AwardPointsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to award points",
		}
	}

	return AwardPointsResponse{Message: "Points awarded successfully"}, nil
}
