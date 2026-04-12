package handlers

import (
	"context"
	"net/http"

	storev1 "github.com/voxire/voxie/gen/go/store/v1"

	"github.com/voxire/onion/internal/modules/store_orders/internal/application"
	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func toProtoOrder(o domain.StoreOrder) *storev1.StoreOrder {
	return &storev1.StoreOrder{
		Id: o.ID, OwnerId: o.OwnerID, CustomerEmail: o.CustomerEmail,
		CustomerName: o.CustomerName, CustomerPhone: o.CustomerPhone,
		Status: o.Status, Subtotal: o.Subtotal, Tax: o.Tax,
		ShippingCost: o.ShippingCost, Total: o.Total,
		ShippingAddress: o.ShippingAddress, Notes: o.Notes,
		CreatedAt: o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func toProtoOrderItem(i domain.StoreOrderItem) *storev1.StoreOrderItem {
	return &storev1.StoreOrderItem{
		Id: i.ID, OrderId: i.OrderID, ProductId: i.ProductID, Name: i.Name,
		Quantity: int32(i.Quantity), Price: i.Price, Total: i.Total,
		CreatedAt: i.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ListRoutes() []string   { return []string{"/store/orders"} }
func DetailRoutes() []string { return []string{"/store/orders/detail"} }
func CreateRoutes() []string { return []string{"/store/orders"} }
func UpdateRoutes() []string { return []string{"/store/orders"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Orders []*storev1.StoreOrder `json:"orders"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil { return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list orders"} }
	out := make([]*storev1.StoreOrder, 0, len(items))
	for _, i := range items { out = append(out, toProtoOrder(i)) }
	return ListResponse{Orders: out}, nil
}

type DetailHandler struct{ Service *application.Service }
type DetailRequest struct{ ID int64 `json:"id"` }
type DetailResponse struct {
	Order *storev1.StoreOrder      `json:"order"`
	Items []*storev1.StoreOrderItem `json:"items"`
}

func (h *DetailHandler) Serve(ctx context.Context, req *DetailRequest) (DetailResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DetailResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	detail, err := h.Service.GetDetail(ctx, authUser.UserID, req.ID)
	if err != nil { return DetailResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to get order detail"} }
	items := make([]*storev1.StoreOrderItem, 0, len(detail.Items))
	for _, i := range detail.Items { items = append(items, toProtoOrderItem(i)) }
	return DetailResponse{Order: toProtoOrder(detail.Order), Items: items}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	CustomerEmail   string                 `json:"customer_email"`
	CustomerName    string                 `json:"customer_name"`
	CustomerPhone   string                 `json:"customer_phone"`
	Status          string                 `json:"status"`
	Subtotal        float64                `json:"subtotal"`
	Tax             float64                `json:"tax"`
	ShippingCost    float64                `json:"shipping_cost"`
	Total           float64                `json:"total"`
	ShippingAddress string                 `json:"shipping_address"`
	Notes           string                 `json:"notes"`
	Items           []domain.OrderItemInput `json:"items"`
}
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.CustomerEmail == "" && req.CustomerName == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Customer info is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.StoreOrderInput{
		CustomerEmail: req.CustomerEmail, CustomerName: req.CustomerName, CustomerPhone: req.CustomerPhone,
		Status: req.Status, Subtotal: req.Subtotal, Tax: req.Tax, ShippingCost: req.ShippingCost,
		Total: req.Total, ShippingAddress: req.ShippingAddress, Notes: req.Notes, Items: req.Items,
	})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create order"} }
	return CreateResponse{ID: id, Message: "Order created successfully"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
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
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.StoreOrderInput{
		CustomerEmail: req.CustomerEmail, CustomerName: req.CustomerName, CustomerPhone: req.CustomerPhone,
		Status: req.Status, Subtotal: req.Subtotal, Tax: req.Tax, ShippingCost: req.ShippingCost,
		Total: req.Total, ShippingAddress: req.ShippingAddress, Notes: req.Notes,
	}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update order"}
	}
	return UpdateResponse{Message: "Order updated successfully"}, nil
}
