package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/application"
	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	erppurchaseorderv1 "github.com/voxire/voxie/gen/go/erp/purchase_order/v1"
)

type PurchaseOrderAPI struct {
	Service *application.Service
}

func NewPurchaseOrderAPI(service *application.Service) *PurchaseOrderAPI {
	return &PurchaseOrderAPI{Service: service}
}

func (a *PurchaseOrderAPI) ListPurchaseOrders(ctx context.Context, req *erppurchaseorderv1.ListPurchaseOrdersRequest) (*erppurchaseorderv1.ListPurchaseOrdersResponse, error) {
	ownerID, err := purchaseOrderOwnerID(ctx, "erp:purchasing:view")
	if err != nil {
		return nil, err
	}
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}

	orders, hasMore, listErr := a.Service.ListPurchaseOrders(ctx, ownerID, req.GetStatus(), int64(page), int64(pageSize))
	if listErr != nil {
		return nil, purchaseOrderStatusError(http.StatusInternalServerError, "Failed to list purchase orders")
	}
	out := make([]*erppurchaseorderv1.PurchaseOrder, 0, len(orders))
	for _, po := range orders {
		out = append(out, newPurchaseOrderPayload(po))
	}
	return &erppurchaseorderv1.ListPurchaseOrdersResponse{
		PurchaseOrders: out,
		HasMore:        hasMore,
		Page:           page,
		PageSize:       pageSize,
	}, nil
}

func (a *PurchaseOrderAPI) CreatePurchaseOrder(ctx context.Context, req *erppurchaseorderv1.CreatePurchaseOrderRequest) (*erppurchaseorderv1.CreatePurchaseOrderResponse, error) {
	ownerID, err := purchaseOrderOwnerID(ctx, "erp:purchasing:create")
	if err != nil {
		return nil, err
	}
	if req.GetExpectedDelivery() != "" {
		if _, parseErr := time.Parse("2006-01-02", req.GetExpectedDelivery()); parseErr != nil {
			return nil, purchaseOrderStatusError(http.StatusBadRequest, "expected_delivery must be in YYYY-MM-DD format")
		}
	}
	items := make([]domain.PurchaseOrderItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, domain.PurchaseOrderItemInput{
			ProductID:       item.GetProductId(),
			QuantityOrdered: item.GetQuantityOrdered(),
			UnitCost:        item.GetUnitCost(),
		})
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, purchaseOrderStatusError(http.StatusUnauthorized, "unauthorized")
	}
	id, createErr := a.Service.CreatePurchaseOrder(ctx, ownerID, domain.PurchaseOrderInput{
		SupplierID:       req.GetSupplierId(),
		PONumber:         strings.TrimSpace(req.GetPoNumber()),
		TaxAmount:        req.GetTaxAmount(),
		ExpectedDelivery: req.GetExpectedDelivery(),
		Notes:            req.GetNotes(),
		Items:            items,
	}, authUser.UserID)
	if createErr != nil {
		if errors.Is(createErr, domain.ErrInvalidPurchaseOrder) {
			return nil, purchaseOrderStatusError(http.StatusBadRequest, createErr.Error())
		}
		if errors.Is(createErr, domain.ErrPurchaseOrderReferenceNotFound) {
			return nil, purchaseOrderStatusError(http.StatusNotFound, createErr.Error())
		}
		return nil, purchaseOrderStatusError(http.StatusInternalServerError, "Failed to create purchase order")
	}
	return &erppurchaseorderv1.CreatePurchaseOrderResponse{PurchaseOrderId: id, Message: "Purchase order created successfully"}, nil
}

func (a *PurchaseOrderAPI) GetPurchaseOrder(ctx context.Context, req *erppurchaseorderv1.GetPurchaseOrderRequest) (*erppurchaseorderv1.GetPurchaseOrderResponse, error) {
	ownerID, err := purchaseOrderOwnerID(ctx, "erp:purchasing:view")
	if err != nil {
		return nil, err
	}
	po, getErr := a.Service.GetByID(ctx, ownerID, req.GetId())
	if getErr != nil {
		if errors.Is(getErr, domain.ErrPurchaseOrderNotFound) {
			return nil, purchaseOrderStatusError(http.StatusNotFound, "Purchase order not found")
		}
		return nil, purchaseOrderStatusError(http.StatusInternalServerError, "Failed to get purchase order")
	}
	return &erppurchaseorderv1.GetPurchaseOrderResponse{PurchaseOrder: newPurchaseOrderPayload(po)}, nil
}

func (a *PurchaseOrderAPI) UpdatePurchaseOrderStatus(ctx context.Context, req *erppurchaseorderv1.UpdatePurchaseOrderStatusRequest) (*erppurchaseorderv1.UpdatePurchaseOrderStatusResponse, error) {
	ownerID, err := purchaseOrderOwnerID(ctx, "erp:purchasing:create")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil || body.GetStatus() == "" {
		return nil, purchaseOrderStatusError(http.StatusBadRequest, "Status is required")
	}
	validStatuses := map[string]struct{}{
		string(domain.POStatusDraft):     {},
		string(domain.POStatusSent):      {},
		string(domain.POStatusApproved):  {},
		string(domain.POStatusPartial):   {},
		string(domain.POStatusReceived):  {},
		string(domain.POStatusCancelled): {},
	}
	if _, ok := validStatuses[body.GetStatus()]; !ok {
		return nil, purchaseOrderStatusError(http.StatusBadRequest, "Invalid status. Must be one of: draft, sent, approved, partial, received, cancelled")
	}
	var submitted []domain.PurchaseOrderItem
	if body.GetStatus() == string(domain.POStatusReceived) && len(body.GetItems()) > 0 {
		submitted = make([]domain.PurchaseOrderItem, 0, len(body.GetItems()))
		for _, item := range body.GetItems() {
			if item == nil {
				continue
			}
			submitted = append(submitted, domain.PurchaseOrderItem{
				ID:               item.GetId(),
				PurchaseOrderID:  item.GetPurchaseOrderId(),
				ProductID:        item.GetProductId(),
				QuantityOrdered:  item.GetQuantityOrdered(),
				QuantityReceived: item.GetQuantityReceived(),
				UnitCost:         item.GetUnitCost(),
				TotalCost:        item.GetTotalCost(),
			})
		}
	}
	if body.GetStatus() == string(domain.POStatusReceived) {
		if receiveErr := a.Service.ReceivePurchaseOrder(ctx, ownerID, req.GetId(), ownerID, submitted); receiveErr != nil {
			return nil, purchaseOrderMutationError(receiveErr, "Failed to receive purchase order")
		}
		return &erppurchaseorderv1.UpdatePurchaseOrderStatusResponse{Message: "Purchase order status updated successfully"}, nil
	}
	if updateErr := a.Service.UpdatePurchaseOrderStatus(ctx, ownerID, req.GetId(), body.GetStatus()); updateErr != nil {
		return nil, purchaseOrderMutationError(updateErr, "Failed to update purchase order status")
	}
	return &erppurchaseorderv1.UpdatePurchaseOrderStatusResponse{Message: "Purchase order status updated successfully"}, nil
}

func (a *PurchaseOrderAPI) ApprovePurchaseOrder(ctx context.Context, req *erppurchaseorderv1.ApprovePurchaseOrderRequest) (*erppurchaseorderv1.ApprovePurchaseOrderResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, purchaseOrderStatusError(http.StatusUnauthorized, "unauthorized")
	}
	ownerID, err := purchaseOrderOwnerID(ctx, "erp:purchasing:approve")
	if err != nil {
		return nil, err
	}
	if approveErr := a.Service.ApprovePurchaseOrder(ctx, ownerID, req.GetId(), authUser.UserID); approveErr != nil {
		if errors.Is(approveErr, domain.ErrPurchaseOrderNotFound) {
			return nil, purchaseOrderStatusError(http.StatusNotFound, "Purchase order not found")
		}
		if errors.Is(approveErr, domain.ErrPurchaseOrderAlreadyApproved) {
			return nil, purchaseOrderStatusError(http.StatusConflict, "Purchase order is already approved")
		}
		if errors.Is(approveErr, domain.ErrPurchaseOrderAlreadyReceived) {
			return nil, purchaseOrderStatusError(http.StatusConflict, "Received purchase orders cannot be reopened")
		}
		if errors.Is(approveErr, domain.ErrPurchaseOrderSelfApproval) {
			return nil, purchaseOrderStatusError(http.StatusForbidden, "Purchase order creators cannot approve their own orders")
		}
		if errors.Is(approveErr, domain.ErrPurchaseOrderInvalidTransition) {
			return nil, purchaseOrderStatusError(http.StatusConflict, "Purchase order cannot be approved from its current status")
		}
		return nil, purchaseOrderStatusError(http.StatusInternalServerError, "Failed to approve purchase order")
	}
	return &erppurchaseorderv1.ApprovePurchaseOrderResponse{Message: "Purchase order approved successfully"}, nil
}

func purchaseOrderOwnerID(ctx context.Context, permission string) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || !authPkg.HasAppPermission(authUser, authorization.AppERP, permission) {
		return 0, onekiterrors.PermissionDenied()
	}
	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return ownerID, nil
}

func purchaseOrderStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}
