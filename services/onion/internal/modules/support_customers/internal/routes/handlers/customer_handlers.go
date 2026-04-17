package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/support_customers/internal/application"
	"github.com/voxire/onion/internal/modules/support_customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// -- LIST --------------------------------------------------------------------

func ListCustomersRoutes() []string { return []string{"/support/customers"} }

type ListCustomersHandler struct{ Service *application.Service }

type ListCustomersRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ListCustomersResponse struct {
	Customers []domain.Customer `json:"customers"`
	Total     int               `json:"total"`
	Page      int               `json:"page"`
	PageSize  int               `json:"page_size"`
}

func (h *ListCustomersHandler) Serve(ctx context.Context, req *ListCustomersRequest) (ListCustomersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	items, total, err := h.Service.List(ctx, device.User.BusinessID, req.Page, req.PageSize)
	if err != nil {
		return ListCustomersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list support customers"}
	}
	return ListCustomersResponse{Customers: items, Total: total, Page: req.Page, PageSize: req.PageSize}, nil
}

// -- GET DETAIL --------------------------------------------------------------

func GetCustomerRoutes() []string { return []string{"/support/customers/detail"} }

type GetCustomerHandler struct{ Service *application.Service }

type GetCustomerRequest struct {
	ID int64 `json:"id"`
}

type GetCustomerResponse struct {
	Customer *domain.Customer `json:"customer"`
}

func (h *GetCustomerHandler) Validate(_ context.Context, req *GetCustomerRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *GetCustomerHandler) Serve(ctx context.Context, req *GetCustomerRequest) (GetCustomerResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCustomerResponse{}, errorModels.PermissionDeniedServerError()
	}
	c, err := h.Service.GetByID(ctx, device.User.BusinessID, req.ID)
	if err != nil {
		return GetCustomerResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get support customer"}
	}
	return GetCustomerResponse{Customer: c}, nil
}
