package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/pos/orders/internal/application"
)

type VoidOrderApprovalResponse struct {
	Message string `json:"message"`
}

func (r *VoidOrderApprovalRequest) orderID(ctx context.Context) (int64, error) {
	return resolveOrderID(ctx, r.OrderID)
}

func serveVoidOrderWithPermission(ctx context.Context, req *VoidOrderApprovalRequest, service *application.Service, permission string) (VoidOrderApprovalResponse, error) {
	user, authErr := requirePOSPermission(ctx, permission)
	if authErr != nil {
		return VoidOrderApprovalResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return VoidOrderApprovalResponse{}, bizErr
	}
	orderID, orderErr := req.orderID(ctx)
	if orderErr != nil {
		return VoidOrderApprovalResponse{}, orderErr
	}
	if err := service.VoidOrder(ctx, businessID, orderID, req.Reason); err != nil {
		return VoidOrderApprovalResponse{}, voidOrderError(err)
	}
	return VoidOrderApprovalResponse{Message: "Order voided successfully"}, nil
}

type ApproveVoidOrderHandler struct{ Service *application.Service }

func ApproveVoidOrderRoutes() []string {
	return []string{"/pos/orders/{id}/void/approve"}
}

func (h *ApproveVoidOrderHandler) Validate(ctx context.Context, req *VoidOrderApprovalRequest) error {
	_, err := req.orderID(ctx)
	return err
}

func (h *ApproveVoidOrderHandler) Serve(ctx context.Context, req *VoidOrderApprovalRequest) (VoidOrderApprovalResponse, error) {
	return serveVoidOrderWithPermission(ctx, req, h.Service, "pos:orders:void:approve")
}

type OverrideVoidOrderHandler struct{ Service *application.Service }

func OverrideVoidOrderRoutes() []string {
	return []string{"/pos/orders/{id}/void/override"}
}

func (h *OverrideVoidOrderHandler) Validate(ctx context.Context, req *VoidOrderApprovalRequest) error {
	_, err := req.orderID(ctx)
	return err
}

func (h *OverrideVoidOrderHandler) Serve(ctx context.Context, req *VoidOrderApprovalRequest) (VoidOrderApprovalResponse, error) {
	return serveVoidOrderWithPermission(ctx, req, h.Service, "pos:orders:void:override")
}

// ---- Hold Order ----

type HoldOrderHandler struct{ Service *application.Service }

type HoldOrderRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

type HoldOrderResponse struct {
	Message string `json:"message"`
}

// ---- Send to Kitchen ----

type SendToKitchenHandler struct{ Service *application.Service }

type SendToKitchenRequest struct {
	OrderID int64 `json:"order_id"`
}

type SendToKitchenResponse struct {
	Success bool `json:"success"`
}

// ---- Void Order Item ----

type VoidOrderItemHandler struct{ Service *application.Service }

type VoidOrderItemRequest struct{}

type VoidOrderItemResponse struct {
	Success bool `json:"success"`
}

// ---- Transfer Table ----
