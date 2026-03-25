package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	creditsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/credits/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/credits/internal/application"
)

func PurchaseCreditsRoutes() []string {
	return []string{
		"/zontriq/credits/purchase",
	}
}

type PurchaseCreditsHandler struct {
	CreditsService *application.Service
}

func (h *PurchaseCreditsHandler) Validate(_ context.Context, req *creditsv1.PurchaseCreditsRequest) *errorModels.ServerError {
	if req.GetAmount() <= 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Amount must be greater than 0",
		}
	}

	if req.GetPaymentMethod() == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Payment method is required",
		}
	}

	return nil
}

func (h *PurchaseCreditsHandler) Serve(ctx context.Context, req *creditsv1.PurchaseCreditsRequest) (creditsv1.PurchaseCreditsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return creditsv1.PurchaseCreditsResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	result, err := h.CreditsService.PurchaseCredits(ctx, zontriqUserID, req.GetAmount(), req.GetPaymentMethod())
	if err != nil {
		return creditsv1.PurchaseCreditsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to purchase credits",
		}
	}

	return creditsv1.PurchaseCreditsResponse{
		Credits: &creditsv1.Credits{
			Id:             int64(result.Credits.ID),
			Balance:        result.Credits.Balance,
			TotalPurchased: result.Credits.TotalPurchased,
			TotalSpent:     result.Credits.TotalSpent,
			Status:         result.Credits.Status,
			CreatedAt:      result.Credits.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:      result.Credits.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		Transaction: &creditsv1.CreditTransaction{
			Id:            int64(result.Transaction.ID),
			Amount:        result.Transaction.Amount,
			Type:          result.Transaction.Type,
			Description:   result.Transaction.Description,
			ReferenceType: result.Transaction.ReferenceType,
			ReferenceId:   int64(result.Transaction.ReferenceID),
			Status:        result.Transaction.Status,
			CreatedAt:     result.Transaction.CreatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
