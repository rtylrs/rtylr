package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// GET /insights/financial/expenses?from=&to=

func GetExpensesByCategoryRoutes() []string { return []string{"/insights/financial/expenses"} }

type GetExpensesByCategoryHandler struct{ Service *application.Service }

type GetExpensesByCategoryRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetExpensesByCategoryResponse struct {
	Categories []application.ExpenseByCategoryPayload `json:"categories"`
}

func (h *GetExpensesByCategoryHandler) Serve(ctx context.Context, req *GetExpensesByCategoryRequest) (GetExpensesByCategoryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetExpensesByCategoryResponse{}, errorModels.PermissionDeniedServerError()
	}
	from, to, err := application.ResolveDateRange(req.From, req.To)
	if err != nil {
		return GetExpensesByCategoryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date range",
		}
	}
	rows, err := h.Service.GetExpensesByCategory(ctx, device.User.UserID, from, to)
	if err != nil {
		return GetExpensesByCategoryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load expenses by category",
		}
	}
	return GetExpensesByCategoryResponse{Categories: rows}, nil
}
