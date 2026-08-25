package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	posordersv1 "github.com/voxire/voxie/gen/go/pos/orders/v1"
)

type OrderAPI struct {
	Service *application.Service
}

func NewOrderAPI(service *application.Service) *OrderAPI {
	return &OrderAPI{Service: service}
}

func (a *OrderAPI) ListOrders(ctx context.Context, req *posordersv1.ListOrdersRequest) (*posordersv1.ListOrdersResponse, error) {
	if err := validateOrderListDates(req); err != nil {
		return nil, err
	}

	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	page := req.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.GetPageSize()
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}

	orders, total, hasMore, err := a.Service.ListOrdersPage(ctx, businessID, req.GetSessionId(), req.GetStatus(), req.GetDateFrom(), req.GetDateTo(), req.GetSearch(), page, pageSize)
	if err != nil {
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to list orders")
	}

	payloads := make([]*posordersv1.Order, 0, len(orders))
	for _, order := range orders {
		payloads = append(payloads, orderToProto(order))
	}

	return &posordersv1.ListOrdersResponse{
		Orders:   payloads,
		Total:    total,
		HasMore:  hasMore,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (a *OrderAPI) CreateOrder(ctx context.Context, req *posordersv1.CreateOrderRequest) (*posordersv1.CreateOrderResponse, error) {
	if err := validateCreateOrderRequest(req); err != nil {
		return nil, err
	}

	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	items := make([]domain.OrderItemInput, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		if item.GetMenuItemId() <= 0 {
			return nil, orderAPIStatusError(http.StatusBadRequest, "Each item must reference a menu item")
		}
		items = append(items, domain.OrderItemInput{
			MenuItemID:  item.GetMenuItemId(),
			Name:        item.GetName(),
			SKU:         item.GetSku(),
			Quantity:    item.GetQuantity(),
			UnitPrice:   item.GetUnitPrice(),
			DiscountPct: item.GetDiscountPct(),
			TaxRate:     item.GetTaxRate(),
			Notes:       item.GetNotes(),
		})
	}

	orderID, err := a.Service.CreateOrder(ctx, businessID, businessID, user.UserID, domain.OrderInput{
		SessionID:     req.GetSessionId(),
		TableID:       req.GetTableId(),
		LocationID:    req.GetLocationId(),
		Type:          domain.OrderType(req.GetType()),
		CustomerID:    req.GetCustomerId(),
		CustomerName:  req.GetCustomerName(),
		PaymentMethod: domain.PaymentMethod(req.GetPaymentMethod()),
		Notes:         req.GetNotes(),
		KitchenNotes:  req.GetKitchenNotes(),
		Items:         items,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPricedItemNotFound) {
			return nil, orderAPIStatusError(http.StatusBadRequest, "Item price could not be resolved")
		}
		if errors.Is(err, domain.ErrDiscountNotAllowed) {
			return nil, orderAPIStatusError(http.StatusForbidden, "You cannot apply discounts")
		}
		if errors.Is(err, domain.ErrDiscountExceedsLimit) {
			return nil, orderAPIStatusError(http.StatusBadRequest, "Discount exceeds your allowed limit")
		}
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to create order")
	}

	order, getErr := a.Service.GetOrder(ctx, businessID, orderID)
	if getErr != nil {
		return &posordersv1.CreateOrderResponse{OrderId: orderID, Message: "Order created"}, nil
	}

	return &posordersv1.CreateOrderResponse{OrderId: orderID, OrderNumber: order.OrderNumber, Message: "Order created"}, nil
}

func (a *OrderAPI) CreatePublicOrder(ctx context.Context, req *posordersv1.CreatePublicOrderRequest) (*posordersv1.CreatePublicOrderResponse, error) {
	if req.GetOwnerId() <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "owner_id is required")
	}
	if len(req.GetItems()) == 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "order must have at least one item")
	}
	for _, item := range req.GetItems() {
		if item.GetMenuItemId() <= 0 {
			return nil, orderAPIStatusError(http.StatusBadRequest, "each item must have a menu_item_id")
		}
		if item.GetQuantity() <= 0 {
			return nil, orderAPIStatusError(http.StatusBadRequest, "item quantity must be positive")
		}
	}

	items := make([]application.PublicOrderItemInput, 0, len(req.GetItems()))
	for _, it := range req.GetItems() {
		items = append(items, application.PublicOrderItemInput{
			MenuItemID: it.GetMenuItemId(),
			Name:       it.GetName(),
			Quantity:   it.GetQuantity(),
			UnitPrice:  it.GetUnitPrice(),
			Notes:      it.GetNotes(),
		})
	}

	orderID, orderNumber, publicToken, err := a.Service.CreatePublicOrder(ctx, req.GetOwnerId(), application.PublicOrderInput{
		CustomerName:  req.GetCustomerName(),
		CustomerPhone: req.GetCustomerPhone(),
		CustomerEmail: req.GetCustomerEmail(),
		Notes:         req.GetNotes(),
		Items:         items,
	})
	if err != nil {
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to create order")
	}

	return &posordersv1.CreatePublicOrderResponse{
		OrderId:     orderID,
		OrderNumber: orderNumber,
		Status:      "pending",
		PublicToken: publicToken,
	}, nil
}

func (a *OrderAPI) GetPublicOrderStatus(ctx context.Context, req *posordersv1.GetPublicOrderStatusRequest) (*posordersv1.GetPublicOrderStatusResponse, error) {
	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Invalid order id")
	}
	businessID := req.GetBusinessId()
	if businessID <= 0 {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Invalid business id")
	}
	if req.GetPublicToken() == "" {
		return nil, orderAPIStatusError(http.StatusBadRequest, "public_token is required")
	}

	orderNumber, status, err := a.Service.GetPublicOrderStatus(ctx, businessID, orderID, req.GetPublicToken())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to fetch order status")
	}

	return &posordersv1.GetPublicOrderStatusResponse{
		OrderId:     orderID,
		OrderNumber: orderNumber,
		Status:      status,
	}, nil
}

func (a *OrderAPI) GetOrderDetail(ctx context.Context, req *posordersv1.GetOrderDetailRequest) (*posordersv1.GetOrderDetailResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	order, err := a.Service.GetOrder(ctx, businessID, req.GetOrderId())
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to get order")
	}

	return &posordersv1.GetOrderDetailResponse{Order: orderToProto(*order)}, nil
}

func (a *OrderAPI) UpdateOrderStatus(ctx context.Context, req *posordersv1.UpdateOrderStatusRequest) (*posordersv1.UpdateOrderStatusResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	body := req.GetBody()
	if body == nil || body.GetStatus() == "" {
		return nil, orderAPIStatusError(http.StatusBadRequest, "Status is required")
	}

	if err := a.Service.UpdateOrderStatus(ctx, businessID, req.GetOrderId(), domain.OrderStatus(body.GetStatus())); err != nil {
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to update order status")
	}

	return &posordersv1.UpdateOrderStatusResponse{Message: "Order status updated"}, nil
}

func (a *OrderAPI) VoidOrder(ctx context.Context, req *posordersv1.VoidOrderRequest) (*posordersv1.VoidOrderResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}

	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	if err := a.Service.VoidOrder(ctx, businessID, req.GetOrderId(), req.GetReason()); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, orderAPIStatusError(http.StatusNotFound, "Order not found")
		}
		return nil, orderAPIStatusError(http.StatusInternalServerError, "Failed to void order")
	}

	return &posordersv1.VoidOrderResponse{Message: "Order voided successfully"}, nil
}

func orderAPIError(err error) error {
	return err
}

func orderAPIStatusError(status int, message string) error {
	return onekiterrors.New(status, message)
}

func validateOrderListDates(req *posordersv1.ListOrdersRequest) error {
	const layout = "2006-01-02"
	if req.GetDateFrom() != "" {
		if _, err := time.Parse(layout, req.GetDateFrom()); err != nil {
			return orderAPIStatusError(http.StatusBadRequest, "date_from must be in YYYY-MM-DD format")
		}
	}
	if req.GetDateTo() != "" {
		if _, err := time.Parse(layout, req.GetDateTo()); err != nil {
			return orderAPIStatusError(http.StatusBadRequest, "date_to must be in YYYY-MM-DD format")
		}
	}
	return nil
}

func validateCreateOrderRequest(req *posordersv1.CreateOrderRequest) error {
	if len(req.GetItems()) == 0 {
		return orderAPIStatusError(http.StatusBadRequest, "Order must have at least one item")
	}

	if req.GetType() != "" && !domain.IsOrderType(domain.OrderType(req.GetType())) {
		return orderAPIStatusError(http.StatusBadRequest, "Unsupported order type")
	}

	for _, item := range req.GetItems() {
		if item.GetQuantity() <= 0 {
			return orderAPIStatusError(http.StatusBadRequest, "Item quantity must be positive")
		}
		if item.GetMenuItemId() <= 0 {
			return orderAPIStatusError(http.StatusBadRequest, "Each item must reference a menu item")
		}
		// Unit price and tax rate are accepted for wire compatibility but are
		// resolved from the server-side product record during order creation.
	}

	return nil
}

func orderToProto(order domain.Order) *posordersv1.Order {
	payload := &posordersv1.Order{
		Id:             order.ID,
		OwnerId:        order.OwnerID,
		SessionId:      order.SessionID,
		TableId:        order.TableID,
		LocationId:     order.LocationID,
		OrderNumber:    order.OrderNumber,
		Type:           string(order.Type),
		Status:         string(order.Status),
		CustomerId:     order.CustomerID,
		CustomerName:   order.CustomerName,
		StaffId:        order.StaffID,
		Subtotal:       order.Subtotal,
		DiscountAmount: order.DiscountAmount,
		TaxAmount:      order.TaxAmount,
		Total:          order.Total,
		PaymentMethod:  string(order.PaymentMethod),
		PaymentStatus:  string(order.PaymentStatus),
		AmountPaid:     order.AmountPaid,
		ChangeDue:      order.ChangeDue,
		Notes:          order.Notes,
		KitchenNotes:   order.KitchenNotes,
		CreatedAt:      formatOrderTime(order.CreatedAt),
		UpdatedAt:      formatOrderTime(order.UpdatedAt),
		PaidAt:         formatOrderTimePtr(order.PaidAt),
	}

	if len(order.Items) > 0 {
		payload.Items = make([]*posordersv1.OrderItem, 0, len(order.Items))
		for _, item := range order.Items {
			payload.Items = append(payload.Items, &posordersv1.OrderItem{
				Id:             item.ID,
				OrderId:        item.OrderID,
				MenuItemId:     item.MenuItemID,
				Name:           item.Name,
				Sku:            item.SKU,
				Quantity:       item.Quantity,
				UnitPrice:      item.UnitPrice,
				DiscountPct:    item.DiscountPct,
				DiscountAmount: item.DiscountAmount,
				TaxRate:        item.TaxRate,
				Total:          item.Total,
				Notes:          item.Notes,
				Status:         string(item.Status),
				CreatedAt:      formatOrderTime(item.CreatedAt),
			})
		}
	}

	return payload
}

func formatOrderTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339Nano)
}

func formatOrderTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return formatOrderTime(*t)
}

func (a *OrderAPI) SendToKitchen(ctx context.Context, req *posordersv1.SendToKitchenRequest) (*posordersv1.SendToKitchenResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	if orderID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}

	if err := a.Service.SendToKitchen(ctx, businessID, orderID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return nil, onekiterrors.New(http.StatusUnprocessableEntity, err.Error())
	}
	return &posordersv1.SendToKitchenResponse{Success: true}, nil
}

func (a *OrderAPI) TransferTable(ctx context.Context, req *posordersv1.TransferTableRequest) (*posordersv1.TransferTableResponse, error) {
	user, authErr := authenticatedPOSUser(ctx)
	if authErr != nil {
		return nil, orderAPIError(authErr)
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return nil, orderAPIError(bizErr)
	}

	orderID := req.GetOrderId()
	newTableID := req.GetNewTableId()
	if orderID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if newTableID <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "New table ID is required")
	}

	if err := a.Service.TransferTable(ctx, businessID, orderID, newTableID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeTransferred) {
			return nil, onekiterrors.New(http.StatusUnprocessableEntity, err.Error())
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to transfer table")
	}
	return &posordersv1.TransferTableResponse{Success: true}, nil
}
