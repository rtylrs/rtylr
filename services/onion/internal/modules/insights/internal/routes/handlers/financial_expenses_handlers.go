package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
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

func (h *GetExpensesByCategoryHandler) Validate(_ context.Context, req *GetExpensesByCategoryRequest) *errorModels.ServerError {
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	return nil
}

func (h *GetExpensesByCategoryHandler) Serve(ctx context.Context, req *GetExpensesByCategoryRequest) (GetExpensesByCategoryResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetExpensesByCategoryResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
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
	rows, err := h.Service.GetExpensesByCategory(ctx, businessID, from, to)
	if err != nil {
		return GetExpensesByCategoryResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load expenses by category",
		}
	}
	return GetExpensesByCategoryResponse{Categories: rows}, nil
}
