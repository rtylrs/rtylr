package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/purchase_orders/internal/application"
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	erppurchaseorderv1 "github.com/voxire/voxie/gen/go/erp/purchase_order/v1"
)

// ─── List Purchase Orders ─────────────────────────────────────────────────────

func ListPurchaseOrdersRoutes() []string {
	return []string{"/erp/purchase-orders"}
}

type ListPurchaseOrdersHandler struct {
	Service *application.Service
}

type ListPurchaseOrdersRequest struct {
	Status   string `json:"status"`
	Page     int64  `json:"page"`
	PageSize int64  `json:"page_size"`
}

type ListPurchaseOrdersResponse struct {
	PurchaseOrders []*erppurchaseorderv1.PurchaseOrder `json:"purchase_orders"`
	Page           int64                               `json:"page"`
	PageSize       int64                               `json:"page_size"`
	HasMore        bool                                `json:"has_more"`
}

func newPurchaseOrderPayload(po domain.PurchaseOrder) *erppurchaseorderv1.PurchaseOrder {
	payload := &erppurchaseorderv1.PurchaseOrder{
		Id:         po.ID,
		OwnerId:    po.OwnerID,
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
		Items:      make([]*erppurchaseorderv1.PurchaseOrderItem, 0, len(po.Items)),
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
		payload.Items = append(payload.Items, &erppurchaseorderv1.PurchaseOrderItem{
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListPurchaseOrdersResponse{}, errorModels.PermissionDeniedServerError()
	}

	orders, hasMore, err := h.Service.ListPurchaseOrders(ctx, businessID, req.Status, req.Page, req.PageSize)
	if err != nil {
		return ListPurchaseOrdersResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list purchase orders",
		}
	}

	payloads := make([]*erppurchaseorderv1.PurchaseOrder, 0, len(orders))
	for _, po := range orders {
		payloads = append(payloads, newPurchaseOrderPayload(po))
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListPurchaseOrdersResponse{PurchaseOrders: payloads, Page: page, PageSize: pageSize, HasMore: hasMore}, nil
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
	if strings.TrimSpace(req.PONumber) == "" {
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

	for _, item := range req.Items {
		if item.QuantityOrdered <= 0 {
			return &errorModels.ServerError{
				Code:     http.StatusBadRequest,
				HTTPCode: http.StatusBadRequest,
				Message:  "Item quantity_ordered must be greater than zero",
			}
		}
		if item.UnitCost < 0 {
			return &errorModels.ServerError{
				Code:     http.StatusBadRequest,
				HTTPCode: http.StatusBadRequest,
				Message:  "Item unit_cost cannot be negative",
			}
		}
	}

	if req.ExpectedDelivery != "" {
		if _, err := time.Parse("2006-01-02", req.ExpectedDelivery); err != nil {
			return &errorModels.ServerError{
				Code:     http.StatusBadRequest,
				HTTPCode: http.StatusBadRequest,
				Message:  "expected_delivery must be in YYYY-MM-DD format",
			}
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreatePurchaseOrderResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := h.Service.CreatePurchaseOrder(ctx, businessID, input, authUser.UserID)
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

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdatePurchaseOrderStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.Service.UpdatePurchaseOrderStatus(ctx, businessID, req.ID, req.Status); err != nil {
		return UpdatePurchaseOrderStatusResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update purchase order status",
		}
	}

	return UpdatePurchaseOrderStatusResponse{Message: "Purchase order status updated successfully"}, nil
}

// ─── Get Purchase Order by ID ─────────────────────────────────────────────────

func GetPurchaseOrderRoutes() []string {
	return []string{"/erp/purchase-orders/{id}"}
}

type GetPurchaseOrderHandler struct {
	Service *application.Service
}

type GetPurchaseOrderRequest struct{}

type GetPurchaseOrderResponse struct {
	PurchaseOrder *erppurchaseorderv1.PurchaseOrder `json:"purchase_order"`
}

func (h *GetPurchaseOrderHandler) Serve(ctx context.Context, _ *GetPurchaseOrderRequest) (GetPurchaseOrderResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetPurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GetPurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid purchase order ID",
		}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return GetPurchaseOrderResponse{}, errorModels.PermissionDeniedServerError()
	}

	po, getErr := h.Service.GetByID(ctx, businessID, id)
	if getErr != nil {
		if errors.Is(getErr, domain.ErrPurchaseOrderNotFound) {
			return GetPurchaseOrderResponse{}, &errorModels.ServerError{
				Code:     http.StatusNotFound,
				HTTPCode: http.StatusNotFound,
				Message:  "Purchase order not found",
			}
		}
		return GetPurchaseOrderResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to get purchase order",
		}
	}
	return GetPurchaseOrderResponse{PurchaseOrder: newPurchaseOrderPayload(po)}, nil
}
