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

func LoyaltyRedeemRoutes() []string {
	return []string{"/crm/loyalty/redeem"}
}

type RedeemPointsHandler struct {
	CRMService *application.Service
}

type RedeemPointsRequest struct {
	CustomerID    int64  `json:"customer_id"`
	Points        int    `json:"points"`
	Description   string `json:"description"`
	ReferenceID   *int64 `json:"reference_id"`
	ReferenceType string `json:"reference_type"`
}

type RedeemPointsResponse struct {
	Message string `json:"message"`
}

func (h *RedeemPointsHandler) Validate(_ context.Context, req *RedeemPointsRequest) *errorModels.ServerError {
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

func (h *RedeemPointsHandler) Serve(ctx context.Context, req *RedeemPointsRequest) (RedeemPointsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RedeemPointsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RedeemPointsResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.CRMService.RedeemPoints(ctx, domain.RedeemPointsInput{
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
			return RedeemPointsResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Customer not found",
			}
		}

		if errors.Is(err, domain.ErrInsufficientPoints) {
			return RedeemPointsResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Insufficient loyalty points",
			}
		}

		return RedeemPointsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to redeem points",
		}
	}

	return RedeemPointsResponse{Message: "Points redeemed successfully"}, nil
}
