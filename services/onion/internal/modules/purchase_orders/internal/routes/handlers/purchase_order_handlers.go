package handlers

import (
	"context"
	"net/http"

	erpv1 "github.com/voxire/voxie/gen/go/erp/v1"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/application"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Purchase Orders ─────────────────────────────────────────────────────

func ListPurchaseOrdersRoutes() []string {
	return []string{"/erp/purchase-orders"}
}

type ListPurchaseOrdersHandler struct {
	Service *application.Service
}

type ListPurchaseOrdersRequest struct {
	Status string `json:"status"`
}

type ListPurchaseOrdersResponse struct {
	PurchaseOrders []*erpv1.PurchaseOrder `json:"purchase_orders"`
}

func newPurchaseOrderPayload(po domain.PurchaseOrder) *erpv1.PurchaseOrder {
	payload := &erpv1.PurchaseOrder{
		Id:         po.ID,
		OwnerId:   po.OwnerID,
		SupplierId: po.SupplierID,
		PoNumber:   po.PONumber,
		Status:     string(po.Status),
		Subtotal:   po.Subtotal,
		TaxAmount:  po.TaxAmount,
		Total:      po.Total,
		Notes:      po.Notes,
		CreatedBy:  po.CreatedBy,
		CreatedAt:  po.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  po.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Items:      make([]*erpv1.PurchaseOrderItem, 0, len(po.Items)),
	}

	if po.ExpectedDelivery != nil {
		formatted := po.ExpectedDelivery.Format("2006-01-02")
		payload.ExpectedDelivery = &formatted
	}

	if po.ReceivedAt != nil {
		formatted := po.ReceivedAt.Format("2006-01-02T15:04:05Z07:00")
		payload.ReceivedAt = &formatted
	}

	for _, item := range po.Items {
		payload.Items = append(payload.Items, &erpv1.PurchaseOrderItem{
			Id:               item.ID,
			PurchaseOrderId:  item.PurchaseOrderID,
			ProductId:        item.ProductID,
			QuantityOrdered:  item.QuantityOrdered,
			QuantityReceived: item.QuantityReceived,
			UnitCost:         item.UnitCost,
			TotalCost:        item.TotalCost,
		})
	}

	return payload
}

func (h *ListPurchaseOrdersHandler) Serve(ctx context.Context, req *ListPurchaseOrdersRequest) (ListPurchaseOrdersResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListPurchaseOrdersResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	orders, err := h.Service.ListPurchaseOrders(ctx, authUser.UserID, req.Status)
	if err != nil {
		return ListPurchaseOrdersResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list purchase orders",
		}
	}

	payloads := make([]*erpv1.PurchaseOrder, 0, len(orders))
	for _, po := range orders {
		payloads = append(payloads, newPurchaseOrderPayload(po))
	}

	return ListPurchaseOrdersResponse{PurchaseOrders: payloads}, nil
}

// ─── Create Purchase Order ────────────────────────────────────────────────────

func CreatePurchaseOrderRoutes() []string {
	return []string{"/erp/purchase-orders"}
}

type CreatePurchaseOrderHandler struct {
	Service *application.Service
}

type CreatePurchaseOrderItemRequest struct {
	ProductID       int64   `json:"product_id"`
	QuantityOrdered float64 `json:"quantity_ordered"`
	UnitCost        float64 `json:"unit_cost"`
}

type CreatePurchaseOrderRequest struct {
	SupplierID       int64                            `json:"supplier_id"`
	PONumber         string                           `json:"po_number"`
	TaxAmount        float64                          `json:"tax_amount"`
	ExpectedDelivery string                           `json:"expected_delivery"`
	Notes            string                           `json:"notes"`
	Items            []CreatePurchaseOrderItemRequest `json:"items"`
}

type CreatePurchaseOrderResponse struct {
	PurchaseOrderID int64  `json:"purchase_order_id"`
	Message         string `json:"message"`
}

func (h *CreatePurchaseOrderHandler) Validate(_ context.Context, req *CreatePurchaseOrderRequest) *errorModels.ServerError {
	if req.SupplierID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Supplier ID is required",
		}
	}
	if req.PONumber == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "PO number is required",
		}
	}
	if len(req.Items) == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "At least one item is required",
		}
	}
	return nil
}

func (h *CreatePurchaseOrderHandler) Serve(ctx context.Context, req *CreatePurchaseOrderRequest) (CreatePurchaseOrderResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreatePurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	items := make([]domain.PurchaseOrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.PurchaseOrderItemInput{
			ProductID:       item.ProductID,
			QuantityOrdered: item.QuantityOrdered,
			UnitCost:        item.UnitCost,
		})
	}

	input := domain.PurchaseOrderInput{
		SupplierID:       req.SupplierID,
		PONumber:         req.PONumber,
		TaxAmount:        req.TaxAmount,
		ExpectedDelivery: req.ExpectedDelivery,
		Notes:            req.Notes,
		Items:            items,
	}

	id, err := h.Service.CreatePurchaseOrder(ctx, authUser.UserID, input, authUser.UserID)
	if err != nil {
		return CreatePurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create purchase order",
		}
	}

	return CreatePurchaseOrderResponse{PurchaseOrderID: id, Message: "Purchase order created successfully"}, nil
}

// ─── Update Purchase Order Status ─────────────────────────────────────────────

func UpdatePurchaseOrderStatusRoutes() []string {
	return []string{"/erp/purchase-orders/status"}
}

type UpdatePurchaseOrderStatusHandler struct {
	Service *application.Service
}

type UpdatePurchaseOrderStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdatePurchaseOrderStatusResponse struct {
	Message string `json:"message"`
}

var validPOStatuses = map[string]struct{}{
	string(domain.POStatusDraft):     {},
	string(domain.POStatusSent):      {},
	string(domain.POStatusApproved):  {},
	string(domain.POStatusPartial):   {},
	string(domain.POStatusReceived):  {},
	string(domain.POStatusCancelled): {},
}

func (h *UpdatePurchaseOrderStatusHandler) Validate(_ context.Context, req *UpdatePurchaseOrderStatusRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Purchase order ID is required",
		}
	}
	if req.Status == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Status is required",
		}
	}
	if _, valid := validPOStatuses[req.Status]; !valid {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid status. Must be one of: draft, sent, approved, partial, received, cancelled",
		}
	}
	return nil
}

func (h *UpdatePurchaseOrderStatusHandler) Serve(ctx context.Context, req *UpdatePurchaseOrderStatusRequest) (UpdatePurchaseOrderStatusResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdatePurchaseOrderStatusResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	if err := h.Service.UpdatePurchaseOrderStatus(ctx, authUser.UserID, req.ID, req.Status); err != nil {
		return UpdatePurchaseOrderStatusResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update purchase order status",
		}
	}

	return UpdatePurchaseOrderStatusResponse{Message: "Purchase order status updated successfully"}, nil
}
