package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_customers/internal/application"
	"github.com/voxire/onion/internal/modules/pos_customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func SearchCustomersRoutes() []string { return []string{"/pos/customers"} }

type SearchCustomersHandler struct{ Service *application.Service }

type SearchCustomersRequest struct {
	Query string `json:"query"`
}

type SearchCustomersResponse struct {
	Customers []domain.Customer `json:"customers"`
}

func (h *SearchCustomersHandler) Serve(ctx context.Context, req *SearchCustomersRequest) (SearchCustomersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return SearchCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}

	ownerID := tenant.ResolveOwnerID(ctx, device.User.BusinessID, device.User.UserID)
	customers, err := h.Service.Search(ctx, ownerID, req.Query)
	if err != nil {
		return SearchCustomersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to search customers"}
	}
	return SearchCustomersResponse{Customers: customers}, nil
}
