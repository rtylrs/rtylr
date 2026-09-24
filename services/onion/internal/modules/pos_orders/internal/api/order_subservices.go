package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	posordersdiscountsv1 "github.com/voxire/voxie/gen/go/pos/orders/discounts/v1"
	posordersholdsv1 "github.com/voxire/voxie/gen/go/pos/orders/holds/v1"
	posordersitemsv1 "github.com/voxire/voxie/gen/go/pos/orders/items/v1"
	posorderspaymentsv1 "github.com/voxire/voxie/gen/go/pos/orders/payments/v1"
	posordersrefundsv1 "github.com/voxire/voxie/gen/go/pos/orders/refunds/v1"
	posorderssplitv1 "github.com/voxire/voxie/gen/go/pos/orders/splits/v1"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
)

type PaymentAPI struct {
	Service *application.Service
}

func NewPaymentAPI(service *application.Service) *PaymentAPI {
	return &PaymentAPI{Service: service}
}

func (a *PaymentAPI) ProcessPayment(ctx context.Context, req *posorderspaymentsv1.ProcessPaymentRequest) (*posorderspaymentsv1.ProcessPaymentResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id is required")
	}
	if req.GetAmountPaid() < 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Amount paid cannot be negative")
	}

	if err := a.Service.ProcessPayment(ctx, businessID, domain.PaymentInput{
		OrderID:       orderID,
		PaymentMethod: domain.PaymentMethod(req.GetPaymentMethod()),
		AmountPaid:    req.GetAmountPaid(),
	}); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderAlreadyPaid) {
			return nil, orderAPIStatusError(http.StatusConflict, "Order is already paid")
		}
		if errors.Is(err, domain.ErrPaymentInsufficient) {
			return nil, orderAPIStatusError(http.StatusBadRequest, "Amount paid is less than order total")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to process payment")
	}

	return &posorderspaymentsv1.ProcessPaymentResponse{Message: "Payment processed"}, nil
}

type RefundAPI struct {
	Service *application.Service
}

func NewRefundAPI(service *application.Service) *RefundAPI {
	return &RefundAPI{Service: service}
}

func (a *RefundAPI) ProcessRefund(ctx context.Context, req *posordersrefundsv1.ProcessRefundRequest) (*posordersrefundsv1.ProcessRefundResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:refund")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id is required")
	}
	if req.GetAmount() <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Refund amount must be positive")
	}

	refundID, err := a.Service.ProcessRefund(ctx, businessID, user.UserID, domain.RefundInput{
		OrderID: orderID,
		Amount:  req.GetAmount(),
		Reason:  req.GetReason(),
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		case errors.Is(err, domain.ErrOrderNotPaid):
			return nil, orderAPIStatusError(http.StatusConflict, "Order has not been paid")
		case errors.Is(err, domain.ErrRefundExceedsTotal):
			return nil, orderAPIStatusError(http.StatusBadRequest, "Refund exceeds order total")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to process refund")
		}
	}

	return &posordersrefundsv1.ProcessRefundResponse{RefundId: refundID, Message: "Refund processed"}, nil
}

type SplitAPI struct {
	Service *application.Service
}

func NewSplitAPI(service *application.Service) *SplitAPI {
	return &SplitAPI{Service: service}
}

func isValidOrderItemPartition(order *domain.Order, splits [][]int64) bool {
	if order == nil || len(splits) < 2 {
		return false
	}

	expected := make(map[int64]struct{}, len(order.Items))
	for _, item := range order.Items {
		if item.Status != domain.ItemStatusCancelled {
			expected[item.ID] = struct{}{}
		}
	}
	if len(expected) == 0 {
		return false
	}

	seen := make(map[int64]struct{}, len(expected))
	for _, group := range splits {
		if len(group) == 0 {
			return false
		}
		for _, itemID := range group {
			if itemID <= 0 {
				return false
			}
			if _, ok := expected[itemID]; !ok {
				return false
			}
			if _, duplicate := seen[itemID]; duplicate {
				return false
			}
			seen[itemID] = struct{}{}
		}
	}
	return len(seen) == len(expected)
}

func (a *SplitAPI) SplitOrder(ctx context.Context, req *posorderssplitv1.SplitOrderRequest) (*posorderssplitv1.SplitOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id is required")
	}

	splits := req.GetSplits()
	if len(splits) < 2 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "At least two splits are required")
	}

	splitGroups := make([][]int64, 0, len(splits))
	for _, group := range splits {
		splitGroups = append(splitGroups, group.GetItemIds())
	}

	order, err := a.Service.GetOrder(ctx, businessID, orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to split order")
	}
	if !isValidOrderItemPartition(order, splitGroups) {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Split groups must contain every active order item exactly once")
	}

	orderIDs, err := a.Service.SplitOrder(ctx, businessID, orderID, splitGroups)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeSplit) {
			return nil, orderAPIStatusError(http.StatusConflict, "Order cannot be split in its current status")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to split order")
	}

	return &posorderssplitv1.SplitOrderResponse{OrderIds: orderIDs, Message: "Order split successfully"}, nil
}

type DiscountAPI struct {
	Service *application.Service
}

func NewDiscountAPI(service *application.Service) *DiscountAPI {
	return &DiscountAPI{Service: service}
}

func (a *DiscountAPI) DiscountOrder(ctx context.Context, req *posordersdiscountsv1.DiscountOrderRequest) (*posordersdiscountsv1.DiscountOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:discount")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id is required")
	}
	if req.GetDiscountValue() <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Discount value must be positive")
	}

	if err := a.Service.ApplyDiscount(ctx, businessID, user.UserID, domain.DiscountInput{
		OrderID:       orderID,
		DiscountType:  req.GetDiscountType(),
		DiscountValue: req.GetDiscountValue(),
		Reason:        req.GetReason(),
	}); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeDiscounted) {
			return nil, orderAPIStatusError(http.StatusConflict, "Paid or refunded orders cannot be discounted")
		}
		if errors.Is(err, domain.ErrDiscountNotAllowed) {
			return nil, orderAPIStatusError(http.StatusForbidden, "You cannot apply discounts")
		}
		if errors.Is(err, domain.ErrDiscountExceedsLimit) {
			return nil, orderAPIStatusError(http.StatusBadRequest, "Discount exceeds your allowed limit")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to apply discount")
	}

	return &posordersdiscountsv1.DiscountOrderResponse{Message: "Discount applied successfully"}, nil
}

type HoldAPI struct {
	Service *application.Service
}

func NewHoldAPI(service *application.Service) *HoldAPI {
	return &HoldAPI{Service: service}
}

func (a *HoldAPI) HoldOrder(ctx context.Context, req *posordersholdsv1.HoldOrderRequest) (*posordersholdsv1.HoldOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id is required")
	}

	if err := a.Service.HoldOrder(ctx, businessID, orderID, req.GetReason()); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to hold order")
	}

	return &posordersholdsv1.HoldOrderResponse{Message: "Order held successfully"}, nil
}

func (a *HoldAPI) ListHeldOrders(ctx context.Context, _ *posordersholdsv1.ListHeldOrdersRequest) (*posordersholdsv1.ListHeldOrdersResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:view")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orders, err := a.Service.ListHeldOrders(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list held orders")
	}

	payloads := make([]*posordersv1.Order, 0, len(orders))
	for _, o := range orders {
		payloads = append(payloads, orderToProto(o))
	}

	return &posordersholdsv1.ListHeldOrdersResponse{Orders: payloads}, nil
}

type ItemAPI struct {
	Service *application.Service
}

func NewItemAPI(service *application.Service) *ItemAPI {
	return &ItemAPI{Service: service}
}

func (a *ItemAPI) UpdateOrderItemStatus(ctx context.Context, req *posordersitemsv1.UpdateOrderItemStatusRequest) (*posordersitemsv1.UpdateOrderItemStatusResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	itemID := req.GetItemId()
	if orderID <= 0 || itemID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id and item id are required")
	}

	status := ""
	if req.GetBody() != nil {
		status = req.GetBody().GetStatus()
	}
	if status == "" {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Status is required")
	}

	if err := a.Service.UpdateOrderItemStatus(ctx, businessID, orderID, itemID, domain.ItemStatus(status)); err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderNotFound):
			return nil, orderAPIStatusError(http.StatusNotFound, "Order item not found")
		case errors.Is(err, domain.ErrInvalidItemStatus):
			return nil, orderAPIStatusError(http.StatusBadRequest, "Invalid item status")
		default:
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update item status")
		}
	}

	return &posordersitemsv1.UpdateOrderItemStatusResponse{Message: "Item status updated"}, nil
}

func (a *ItemAPI) VoidOrderItem(ctx context.Context, req *posordersitemsv1.VoidOrderItemRequest) (*posordersitemsv1.VoidOrderItemResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:void")
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	itemID := req.GetItemId()
	if orderID <= 0 || itemID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Order id and item id are required")
	}

	if err := a.Service.VoidOrderItem(ctx, businessID, orderID, itemID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order item not found")
		}
		if errors.Is(err, domain.ErrItemNotVoidable) {
			return nil, orderAPIStatusError(http.StatusConflict, "Order item cannot be voided in its current status")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to void order item")
	}

	return &posordersitemsv1.VoidOrderItemResponse{Success: true}, nil
}
