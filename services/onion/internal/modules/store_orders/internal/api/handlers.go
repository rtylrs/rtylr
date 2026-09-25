package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/store_orders/internal/application"
	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListRoutes() []string   { return []string{"/store/orders"} }
func DetailRoutes() []string { return []string{"/store/orders/detail"} }
func CreateRoutes() []string { return []string{"/store/orders"} }
func UpdateRoutes() []string { return []string{"/store/orders"} }

type (
	ListHandler struct{ Service *application.Service }
	ListRequest struct {
		Limit  int32 `json:"limit"`
		Offset int32 `json:"offset"`
	}
	ListResponse struct {
		Orders []domain.StoreOrder `json:"orders"`
		Total  int64               `json:"total"`
		Limit  int32               `json:"limit"`
		Offset int32               `json:"offset"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, req *ListRequest) (ListResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}
	page, err := h.Service.ListPaginated(ctx, businessID, req.Limit, req.Offset)
	if err != nil {
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list orders")
	}
	return ListResponse{
		Orders: page.Orders,
		Total:  page.Total,
		Limit:  page.Limit,
		Offset: page.Offset,
	}, nil
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

func (h *DetailHandler) Validate(_ context.Context, req *DetailRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *DetailHandler) Serve(ctx context.Context, req *DetailRequest) (DetailResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DetailResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DetailResponse{}, onekiterrors.PermissionDenied()
	}
	detail, err := h.Service.GetDetail(ctx, businessID, req.ID)
	if err != nil {
		return DetailResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get order detail")
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

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	if req.CustomerEmail == "" && req.CustomerName == "" {
		return onekiterrors.New(http.StatusBadRequest, "Customer info is required")
	}
	if status := strings.TrimSpace(req.Status); status != "" && !strings.EqualFold(status, "pending") {
		return onekiterrors.New(http.StatusBadRequest, "New orders must start with status pending")
	}
	if req.Subtotal < 0 || req.Tax < 0 || req.ShippingCost < 0 || req.Total < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Monetary amounts cannot be negative")
	}
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return onekiterrors.New(http.StatusBadRequest, "Item quantity must be positive")
		}
		if item.Price < 0 || item.Total < 0 {
			return onekiterrors.New(http.StatusBadRequest, "Item amounts cannot be negative")
		}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.StoreOrderInput{
		CustomerEmail:   req.CustomerEmail,
		CustomerName:    req.CustomerName,
		CustomerPhone:   req.CustomerPhone,
		Status:          req.Status,
		Subtotal:        req.Subtotal,
		Tax:             req.Tax,
		ShippingCost:    req.ShippingCost,
		Total:           req.Total,
		ShippingAddress: req.ShippingAddress,
		Notes:           req.Notes,
		Items:           req.Items,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			return CreateResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create order")
	}
	return CreateResponse{ID: id, Message: "Order created successfully"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID              int64   `json:"id"`
		CustomerEmail   *string `json:"customer_email,omitempty"`
		CustomerName    *string `json:"customer_name,omitempty"`
		CustomerPhone   *string `json:"customer_phone,omitempty"`
		Status          *string `json:"status,omitempty"`
		ShippingAddress *string `json:"shipping_address,omitempty"`
		Notes           *string `json:"notes,omitempty"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.IsBusinessAdmin(authUser) {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	detail, err := h.Service.GetDetail(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UpdateResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get order for update")
	}
	current := detail.Order
	input := domain.StoreOrderInput{
		CustomerEmail:   current.CustomerEmail,
		CustomerName:    current.CustomerName,
		CustomerPhone:   current.CustomerPhone,
		Status:          current.Status,
		ShippingAddress: current.ShippingAddress,
		Notes:           current.Notes,
	}
	if req.CustomerEmail != nil {
		input.CustomerEmail = *req.CustomerEmail
	}
	if req.CustomerName != nil {
		input.CustomerName = *req.CustomerName
	}
	if req.CustomerPhone != nil {
		input.CustomerPhone = *req.CustomerPhone
	}
	if req.Status != nil {
		input.Status = *req.Status
	}
	if req.ShippingAddress != nil {
		input.ShippingAddress = *req.ShippingAddress
	}
	if req.Notes != nil {
		input.Notes = *req.Notes
	}

	if err := h.Service.Update(ctx, businessID, req.ID, input); err != nil {
		if errors.Is(err, domain.ErrInvalidStatus) {
			return UpdateResponse{}, onekiterrors.New(http.StatusConflict, "Order status transition is not allowed")
		}
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update order")
	}
	return UpdateResponse{Message: "Order updated successfully"}, nil
}
