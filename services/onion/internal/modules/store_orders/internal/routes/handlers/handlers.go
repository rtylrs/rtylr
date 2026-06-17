package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/store_orders/internal/application"
	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ListRoutes() []string   { return []string{"/store/orders"} }
func DetailRoutes() []string { return []string{"/store/orders/detail"} }
func CreateRoutes() []string { return []string{"/store/orders"} }
func UpdateRoutes() []string { return []string{"/store/orders"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Orders []domain.StoreOrder `json:"orders"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list orders"}
	}
	return ListResponse{Orders: items}, nil
}

type (
	DetailHandler struct{ Service *application.Service }
	DetailRequest struct {
		ID int64 `json:"id"`
	}
)

type DetailResponse struct {
	Order domain.StoreOrder       `json:"order"`
	Items []domain.StoreOrderItem `json:"items"`
}

func (h *DetailHandler) Validate(_ context.Context, req *DetailRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *DetailHandler) Serve(ctx context.Context, req *DetailRequest) (DetailResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DetailResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	detail, err := h.Service.GetDetail(ctx, authUser.UserID, req.ID)
	if err != nil {
		return DetailResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to get order detail"}
	}
	return DetailResponse{Order: detail.Order, Items: detail.Items}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		CustomerEmail   string                  `json:"customer_email"`
		CustomerName    string                  `json:"customer_name"`
		CustomerPhone   string                  `json:"customer_phone"`
		Status          string                  `json:"status"`
		Subtotal        float64                 `json:"subtotal"`
		Tax             float64                 `json:"tax"`
		ShippingCost    float64                 `json:"shipping_cost"`
		Total           float64                 `json:"total"`
		ShippingAddress string                  `json:"shipping_address"`
		Notes           string                  `json:"notes"`
		Items           []domain.OrderItemInput `json:"items"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.CustomerEmail == "" && req.CustomerName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Customer info is required"}
	}
	if req.Subtotal < 0 || req.Tax < 0 || req.ShippingCost < 0 || req.Total < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Monetary amounts cannot be negative"}
	}
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Item quantity must be positive"}
		}
		if item.Price < 0 || item.Total < 0 {
			return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Item amounts cannot be negative"}
		}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	id, err := h.Service.Create(ctx, authUser.UserID, domain.StoreOrderInput{
		CustomerEmail: req.CustomerEmail, CustomerName: req.CustomerName, CustomerPhone: req.CustomerPhone,
		Status: req.Status, Subtotal: req.Subtotal, Tax: req.Tax, ShippingCost: req.ShippingCost,
		Total: req.Total, ShippingAddress: req.ShippingAddress, Notes: req.Notes, Items: req.Items,
	})
	if err != nil {
		return CreateResponse{}, errorModels.FromError(err, "Failed to create order")
	}
	return CreateResponse{ID: id, Message: "Order created successfully"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID              int64   `json:"id"`
		CustomerEmail   string  `json:"customer_email"`
		CustomerName    string  `json:"customer_name"`
		CustomerPhone   string  `json:"customer_phone"`
		Status          string  `json:"status"`
		Subtotal        float64 `json:"subtotal"`
		Tax             float64 `json:"tax"`
		ShippingCost    float64 `json:"shipping_cost"`
		Total           float64 `json:"total"`
		ShippingAddress string  `json:"shipping_address"`
		Notes           string  `json:"notes"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	if req.Subtotal < 0 || req.Tax < 0 || req.ShippingCost < 0 || req.Total < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Monetary amounts cannot be negative"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.StoreOrderInput{
		CustomerEmail: req.CustomerEmail, CustomerName: req.CustomerName, CustomerPhone: req.CustomerPhone,
		Status: req.Status, Subtotal: req.Subtotal, Tax: req.Tax, ShippingCost: req.ShippingCost,
		Total: req.Total, ShippingAddress: req.ShippingAddress, Notes: req.Notes,
	}); err != nil {
		return UpdateResponse{}, errorModels.FromError(err, "Failed to update order")
	}
	return UpdateResponse{Message: "Order updated successfully"}, nil
}
