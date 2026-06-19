package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func DeleteExpenseRoutes() []string {
	return []string{"/finance/expenses"}
}

type DeleteExpenseHandler struct {
	FinanceService *application.Service
}

type DeleteExpenseRequest struct {
	ID int64 `json:"id"`
}

type DeleteExpenseResponse struct {
	Message string `json:"message"`
}

func (h *DeleteExpenseHandler) Validate(_ context.Context, req *DeleteExpenseRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Expense ID is required",
		}
	}

	return nil
}

func (h *DeleteExpenseHandler) Serve(ctx context.Context, req *DeleteExpenseRequest) (DeleteExpenseResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return DeleteExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteExpenseResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FinanceService.DeleteExpense(ctx, businessID, req.ID); err != nil {
		return DeleteExpenseResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete expense",
		}
	}

	return DeleteExpenseResponse{Message: "Expense deleted successfully"}, nil
}
