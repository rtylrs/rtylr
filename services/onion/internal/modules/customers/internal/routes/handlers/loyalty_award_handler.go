package handlers

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *AwardPointsHandler) Validate(_ context.Context, req *AwardPointsRequest) error {
	if req.CustomerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	if req.Points <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Points must be greater than zero")
	}
	// The ledger write casts to int32; anything above MaxInt32 would wrap
	// negative before the merge path's clamp ever sees it (#844).
	if req.Points > math.MaxInt32 {
		return onekiterrors.New(http.StatusBadRequest, "Points exceed the maximum allowed value")
	}

	return nil
}

func (h *AwardPointsHandler) Serve(ctx context.Context, req *AwardPointsRequest) (AwardPointsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return AwardPointsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMManager(ctx); err != nil {
		return AwardPointsResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return AwardPointsResponse{}, onekiterrors.PermissionDenied()
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
			return AwardPointsResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}

		return AwardPointsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to award points")
	}

	return AwardPointsResponse{Message: "Points awarded successfully"}, nil
}
