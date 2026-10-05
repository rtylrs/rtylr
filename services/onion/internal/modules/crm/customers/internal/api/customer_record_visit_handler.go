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
)

func CustomerRecordVisitRoutes() []string {
	return []string{"/crm/customers/record-visit"}
}

type RecordVisitHandler struct {
	CRMService *application.Service
}

type RecordVisitRequest struct {
	CustomerID  int64   `json:"customer_id"`
	Amount      float64 `json:"amount"`
	AwardPoints bool    `json:"award_points"`
}

type RecordVisitResponse struct {
	Message string `json:"message"`
}

func (h *RecordVisitHandler) Validate(_ context.Context, req *RecordVisitRequest) error {
	if req.CustomerID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Customer ID is required")
	}

	if req.Amount < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Amount must not be negative")
	}

	return nil
}

func (h *RecordVisitHandler) Serve(ctx context.Context, req *RecordVisitRequest) (RecordVisitResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return RecordVisitResponse{}, onekiterrors.PermissionDenied()
	}
	if req.AwardPoints {
		if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
			return RecordVisitResponse{}, err
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return RecordVisitResponse{}, onekiterrors.PermissionDenied()
	}

	if _, err := h.CRMService.GetCustomer(ctx, businessID, req.CustomerID); err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return RecordVisitResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}
		return RecordVisitResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to record visit")
	}

	err := h.CRMService.RecordVisit(ctx, domain.RecordVisitInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		CustomerID:  req.CustomerID,
		Amount:      req.Amount,
		AwardPoints: req.AwardPoints,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCustomerNotFound) {
			return RecordVisitResponse{}, onekiterrors.New(http.StatusNotFound, "Customer not found")
		}
		return RecordVisitResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to record visit")
	}

	return RecordVisitResponse{Message: "Visit recorded successfully"}, nil
}
