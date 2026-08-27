package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/invoices/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
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

func (h *DeleteExpenseHandler) Validate(_ context.Context, req *DeleteExpenseRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Expense ID is required")
	}

	return nil
}

func (h *DeleteExpenseHandler) Serve(ctx context.Context, req *DeleteExpenseRequest) (DeleteExpenseResponse, error) {
	device := authPkg.DeviceCtx(ctx)

	if device.User == nil {
		return DeleteExpenseResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireFinancePermission(ctx, "finance:expenses:create"); err != nil {
		return DeleteExpenseResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteExpenseResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.FinanceService.DeleteExpense(ctx, businessID, req.ID); err != nil {
		return DeleteExpenseResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete expense")
	}

	return DeleteExpenseResponse{Message: "Expense deleted successfully"}, nil
}
