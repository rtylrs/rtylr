package api

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/pos_orders/internal/application"
	"github.com/voxire/onion/internal/modules/pos_orders/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

// resolveOrderID returns the order id from either the chi URL param or the request body.
// The URL param takes precedence when present.
func resolveOrderID(ctx context.Context, bodyID int64) (int64, error) {
	if raw := chi.URLParamFromCtx(ctx, "id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return 0, onekiterrors.New(http.StatusBadRequest, "Invalid order id")
		}
		return parsed, nil
	}
	if bodyID == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}
	return bodyID, nil
}

// ---- Split Order ----

func SplitOrderRoutes() []string { return []string{"/pos/orders/split", "/pos/orders/{id}/split"} }

type SplitOrderHandler struct{ Service *application.Service }

type SplitOrderRequest struct {
	OrderID int64     `json:"order_id"`
	Splits  [][]int64 `json:"splits"`
}

type SplitOrderResponse struct {
	OrderIDs []int64 `json:"order_ids"`
	Message  string  `json:"message"`
}

func (h *SplitOrderHandler) Validate(ctx context.Context, req *SplitOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if len(req.Splits) < 2 {
		return onekiterrors.New(http.StatusBadRequest, "At least two splits are required")
	}
	return nil
}

func (h *SplitOrderHandler) Serve(ctx context.Context, req *SplitOrderRequest) (SplitOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return SplitOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return SplitOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return SplitOrderResponse{}, sErr
	}

	orderIDs, err := h.Service.SplitOrder(ctx, businessID, orderID, req.Splits)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return SplitOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeSplit) {
			return SplitOrderResponse{}, onekiterrors.New(http.StatusConflict, "Order cannot be split in its current status")
		}
		return SplitOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to split order")
	}
	return SplitOrderResponse{OrderIDs: orderIDs, Message: "Order split successfully"}, nil
}

// ---- Apply Discount ----

func DiscountOrderRoutes() []string {
	return []string{"/pos/orders/discount", "/pos/orders/{id}/discount"}
}

type DiscountOrderHandler struct{ Service *application.Service }

type DiscountOrderRequest struct {
	OrderID       int64   `json:"order_id"`
	DiscountType  string  `json:"discount_type"`
	DiscountValue float64 `json:"discount_value"`
	Reason        string  `json:"reason"`
}

type DiscountOrderResponse struct {
	Message string `json:"message"`
}

func (h *DiscountOrderHandler) Validate(ctx context.Context, req *DiscountOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if req.DiscountValue <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Discount value must be positive")
	}
	return nil
}

func (h *DiscountOrderHandler) Serve(ctx context.Context, req *DiscountOrderRequest) (DiscountOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:discount")
	if authErr != nil {
		return DiscountOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return DiscountOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return DiscountOrderResponse{}, sErr
	}

	if err := h.Service.ApplyDiscount(ctx, businessID, user.UserID, domain.DiscountInput{
		// Capabilities are resolved server-side from the authenticated POS user.
		// Client-provided discount values never bypass that check.
		OrderID:       orderID,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		Reason:        req.Reason,
	}); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeDiscounted) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusConflict, "Paid or refunded orders cannot be discounted")
		}
		if errors.Is(err, domain.ErrDiscountNotAllowed) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusForbidden, "You cannot apply discounts")
		}
		if errors.Is(err, domain.ErrDiscountExceedsLimit) {
			return DiscountOrderResponse{}, onekiterrors.New(http.StatusBadRequest, "Discount exceeds your allowed limit")
		}
		return DiscountOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to apply discount")
	}
	return DiscountOrderResponse{Message: "Discount applied successfully"}, nil
}

// ---- Void Order ----

func VoidOrderRoutes() []string { return []string{"/pos/orders/void", "/pos/orders/{id}/void"} }

type VoidOrderHandler struct{ Service *application.Service }

type VoidOrderRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

type VoidOrderResponse struct {
	Message string `json:"message"`
}

func (h *VoidOrderHandler) Validate(ctx context.Context, req *VoidOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if req.Reason == "" {
		return onekiterrors.New(http.StatusBadRequest, "Reason is required")
	}
	return nil
}

func (h *VoidOrderHandler) Serve(ctx context.Context, req *VoidOrderRequest) (VoidOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:void")
	if authErr != nil {
		return VoidOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return VoidOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return VoidOrderResponse{}, sErr
	}

	if err := h.Service.VoidOrder(ctx, businessID, orderID, req.Reason); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return VoidOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return VoidOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order")
	}
	return VoidOrderResponse{Message: "Order voided successfully"}, nil
}

type VoidOrderApprovalRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

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
		if errors.Is(err, domain.ErrOrderNotFound) {
			return VoidOrderApprovalResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return VoidOrderApprovalResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order")
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

func HoldOrderRoutes() []string { return []string{"/pos/orders/hold", "/pos/orders/{id}/hold"} }

type HoldOrderHandler struct{ Service *application.Service }

type HoldOrderRequest struct {
	OrderID int64  `json:"order_id"`
	Reason  string `json:"reason"`
}

type HoldOrderResponse struct {
	Message string `json:"message"`
}

func (h *HoldOrderHandler) Validate(ctx context.Context, req *HoldOrderRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	return nil
}

func (h *HoldOrderHandler) Serve(ctx context.Context, req *HoldOrderRequest) (HoldOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return HoldOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return HoldOrderResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return HoldOrderResponse{}, sErr
	}

	if err := h.Service.HoldOrder(ctx, businessID, orderID, req.Reason); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return HoldOrderResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return HoldOrderResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to hold order")
	}
	return HoldOrderResponse{Message: "Order held successfully"}, nil
}

// ---- Send to Kitchen ----

func SendToKitchenRoutes() []string {
	return []string{"/pos/orders/send-to-kitchen", "/pos/orders/{id}/send-to-kitchen"}
}

type SendToKitchenHandler struct{ Service *application.Service }

type SendToKitchenRequest struct {
	OrderID int64 `json:"order_id"`
}

type SendToKitchenResponse struct {
	Success bool `json:"success"`
}

func (h *SendToKitchenHandler) Validate(ctx context.Context, req *SendToKitchenRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	return nil
}

func (h *SendToKitchenHandler) Serve(ctx context.Context, req *SendToKitchenRequest) (SendToKitchenResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return SendToKitchenResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return SendToKitchenResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return SendToKitchenResponse{}, sErr
	}

	if err := h.Service.SendToKitchen(ctx, businessID, orderID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return SendToKitchenResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		return SendToKitchenResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, err.Error())
	}
	return SendToKitchenResponse{Success: true}, nil
}

// ---- Void Order Item ----

func VoidOrderItemRoutes() []string {
	return []string{"/pos/orders/{id}/items/{itemId}"}
}

type VoidOrderItemHandler struct{ Service *application.Service }

type VoidOrderItemRequest struct{}

type VoidOrderItemResponse struct {
	Success bool `json:"success"`
}

func (h *VoidOrderItemHandler) Validate(ctx context.Context, _ *VoidOrderItemRequest) error {
	if chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if chi.URLParamFromCtx(ctx, "itemId") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Item ID is required")
	}
	return nil
}

func (h *VoidOrderItemHandler) Serve(ctx context.Context, _ *VoidOrderItemRequest) (VoidOrderItemResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:void")
	if authErr != nil {
		return VoidOrderItemResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return VoidOrderItemResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, 0)
	if sErr != nil {
		return VoidOrderItemResponse{}, sErr
	}

	rawItemID := chi.URLParamFromCtx(ctx, "itemId")
	itemID, err := strconv.ParseInt(rawItemID, 10, 64)
	if err != nil || itemID <= 0 {
		return VoidOrderItemResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid item id")
	}

	if err := h.Service.VoidOrderItem(ctx, businessID, orderID, itemID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return VoidOrderItemResponse{}, onekiterrors.New(http.StatusNotFound, "Order item not found")
		}
		if errors.Is(err, domain.ErrItemNotVoidable) {
			return VoidOrderItemResponse{}, onekiterrors.New(http.StatusConflict, "Order item cannot be voided in its current status")
		}
		return VoidOrderItemResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to void order item")
	}
	return VoidOrderItemResponse{Success: true}, nil
}

// ---- Transfer Table ----

func TransferTableRoutes() []string {
	return []string{"/pos/orders/transfer-table", "/pos/orders/{id}/transfer-table"}
}

type TransferTableHandler struct{ Service *application.Service }

type TransferTableRequest struct {
	OrderID    int64 `json:"order_id"`
	NewTableID int64 `json:"new_table_id"`
}

type TransferTableResponse struct {
	Success bool `json:"success"`
}

func (h *TransferTableHandler) Validate(ctx context.Context, req *TransferTableRequest) error {
	if req.OrderID == 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order ID is required")
	}
	if req.NewTableID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "New table ID is required")
	}
	return nil
}

func (h *TransferTableHandler) Serve(ctx context.Context, req *TransferTableRequest) (TransferTableResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return TransferTableResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return TransferTableResponse{}, bizErr
	}

	orderID, sErr := resolveOrderID(ctx, req.OrderID)
	if sErr != nil {
		return TransferTableResponse{}, sErr
	}

	if err := h.Service.TransferTable(ctx, businessID, orderID, req.NewTableID); err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			return TransferTableResponse{}, onekiterrors.New(http.StatusNotFound, "Order not found")
		}
		if errors.Is(err, domain.ErrTableNotFound) {
			return TransferTableResponse{}, onekiterrors.New(http.StatusNotFound, "Table not found")
		}
		if errors.Is(err, domain.ErrOrderCannotBeTransferred) {
			return TransferTableResponse{}, onekiterrors.New(http.StatusUnprocessableEntity, err.Error())
		}
		return TransferTableResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to transfer table")
	}
	return TransferTableResponse{Success: true}, nil
}

// ---- Compatibility order create/resume -------------------------------------

// The generated order contract predates combo-aware line inputs and item
// replacement. Keep this narrow compatibility boundary in the POS module until
// the public contract can carry those fields; it still uses the same scoped
// application service and pricing resolver as the generated create path.
type OrderCompatItemInput struct {
	MenuItemID     int64   `json:"menu_item_id"`
	ComboID        int64   `json:"combo_id"`
	ModifierIDs    []int64 `json:"modifier_ids"`
	Name           string  `json:"name"`
	SKU            string  `json:"sku"`
	Quantity       float64 `json:"quantity"`
	UnitPrice      float64 `json:"unit_price"`
	DiscountPct    float64 `json:"discount_pct"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxRate        float64 `json:"tax_rate"`
	Notes          string  `json:"notes"`
}

type CreateOrderCompatRequest struct {
	SessionID     int64                  `json:"session_id"`
	TableID       int64                  `json:"table_id"`
	LocationID    int64                  `json:"location_id"`
	Type          string                 `json:"type"`
	CustomerID    int64                  `json:"customer_id"`
	CustomerName  string                 `json:"customer_name"`
	PaymentMethod string                 `json:"payment_method"`
	Notes         string                 `json:"notes"`
	KitchenNotes  string                 `json:"kitchen_notes"`
	Items         []OrderCompatItemInput `json:"items"`
}

type CreateOrderCompatResponse struct {
	OrderID     int64  `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Message     string `json:"message"`
}

type ResumeOrderRequest struct {
	OrderID      int64                  `json:"order_id"`
	TableID      int64                  `json:"table_id"`
	LocationID   int64                  `json:"location_id"`
	Type         string                 `json:"type"`
	CustomerID   int64                  `json:"customer_id"`
	CustomerName string                 `json:"customer_name"`
	Notes        string                 `json:"notes"`
	KitchenNotes string                 `json:"kitchen_notes"`
	Items        []OrderCompatItemInput `json:"items"`
}

type ResumeOrderResponse struct {
	Success bool  `json:"success"`
	OrderID int64 `json:"order_id"`
}

func validateOrderCompatItem(item OrderCompatItemInput) error {
	if item.Quantity <= 0 || math.IsNaN(item.Quantity) || math.IsInf(item.Quantity, 0) {
		return onekiterrors.New(http.StatusBadRequest, "Item quantity must be positive")
	}
	if item.MenuItemID <= 0 && item.ComboID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "Each item must reference a menu item or combo")
	}
	if item.MenuItemID > 0 && item.ComboID > 0 {
		return onekiterrors.New(http.StatusBadRequest, "Each item cannot reference both a menu item and combo")
	}
	if item.UnitPrice < 0 || item.DiscountPct < 0 || item.DiscountPct > 100 || item.DiscountAmount < 0 || item.TaxRate < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Invalid order item amounts")
	}
	for _, modifierID := range item.ModifierIDs {
		if modifierID <= 0 {
			return onekiterrors.New(http.StatusBadRequest, "Modifier ids must be positive")
		}
	}
	return nil
}

func validateOrderCompatItems(items []OrderCompatItemInput) error {
	if len(items) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Order must have at least one item")
	}
	for _, item := range items {
		if err := validateOrderCompatItem(item); err != nil {
			return err
		}
	}
	return nil
}

func compatItems(items []OrderCompatItemInput) []domain.OrderItemInput {
	result := make([]domain.OrderItemInput, 0, len(items))
	for _, item := range items {
		result = append(result, domain.OrderItemInput{
			MenuItemID:     item.MenuItemID,
			ComboID:        item.ComboID,
			ModifierIDs:    append([]int64(nil), item.ModifierIDs...),
			Name:           item.Name,
			SKU:            item.SKU,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			DiscountPct:    item.DiscountPct,
			DiscountAmount: item.DiscountAmount,
			TaxRate:        item.TaxRate,
			Notes:          item.Notes,
		})
	}
	return result
}

type CreateOrderCompatHandler struct{ Service *application.Service }

func (h *CreateOrderCompatHandler) Validate(_ context.Context, req *CreateOrderCompatRequest) error {
	if req.SessionID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "An open register session is required")
	}
	if req.Type != "" && !domain.IsOrderType(domain.OrderType(req.Type)) {
		return onekiterrors.New(http.StatusBadRequest, "Unsupported order type")
	}
	return validateOrderCompatItems(req.Items)
}

func (h *CreateOrderCompatHandler) Serve(ctx context.Context, req *CreateOrderCompatRequest) (CreateOrderCompatResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return CreateOrderCompatResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return CreateOrderCompatResponse{}, bizErr
	}
	session, sessionErr := h.Service.GetCurrentOpenSession(ctx, businessID, user.UserID)
	if sessionErr != nil {
		return CreateOrderCompatResponse{}, onekiterrors.New(http.StatusBadRequest, "Could not verify the register session")
	}
	if session == nil || session.ID != req.SessionID {
		return CreateOrderCompatResponse{}, onekiterrors.New(http.StatusBadRequest, "The register session is not open for this cashier")
	}

	orderID, err := h.Service.CreateOrder(ctx, businessID, businessID, user.UserID, domain.OrderInput{
		SessionID:     req.SessionID,
		TableID:       req.TableID,
		LocationID:    req.LocationID,
		Type:          domain.OrderType(req.Type),
		CustomerID:    req.CustomerID,
		CustomerName:  req.CustomerName,
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		Notes:         req.Notes,
		KitchenNotes:  req.KitchenNotes,
		Items:         compatItems(req.Items),
	})
	if err != nil {
		return CreateOrderCompatResponse{}, mapOrderCompatError(err)
	}
	order, getErr := h.Service.GetOrder(ctx, businessID, orderID)
	if getErr != nil {
		return CreateOrderCompatResponse{OrderID: orderID, Message: "Order created"}, nil
	}
	return CreateOrderCompatResponse{OrderID: orderID, OrderNumber: order.OrderNumber, Message: "Order created"}, nil
}

type ResumeOrderHandler struct{ Service *application.Service }

func (h *ResumeOrderHandler) Validate(ctx context.Context, req *ResumeOrderRequest) error {
	if req.OrderID <= 0 && chi.URLParamFromCtx(ctx, "id") == "" {
		return onekiterrors.New(http.StatusBadRequest, "Order id is required")
	}
	if req.Type != "" && !domain.IsOrderType(domain.OrderType(req.Type)) {
		return onekiterrors.New(http.StatusBadRequest, "Unsupported order type")
	}
	return validateOrderCompatItems(req.Items)
}

func (h *ResumeOrderHandler) Serve(ctx context.Context, req *ResumeOrderRequest) (ResumeOrderResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:create")
	if authErr != nil {
		return ResumeOrderResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ResumeOrderResponse{}, bizErr
	}
	orderID, orderErr := resolveOrderID(ctx, req.OrderID)
	if orderErr != nil {
		return ResumeOrderResponse{}, orderErr
	}

	if err := h.Service.ResumeOrder(ctx, businessID, user.UserID, orderID, domain.OrderInput{
		TableID:      req.TableID,
		LocationID:   req.LocationID,
		Type:         domain.OrderType(req.Type),
		CustomerID:   req.CustomerID,
		CustomerName: req.CustomerName,
		Notes:        req.Notes,
		KitchenNotes: req.KitchenNotes,
		Items:        compatItems(req.Items),
	}); err != nil {
		return ResumeOrderResponse{}, mapOrderCompatError(err)
	}
	return ResumeOrderResponse{Success: true, OrderID: orderID}, nil
}

func mapOrderCompatError(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return onekiterrors.New(http.StatusNotFound, "Order not found")
	case errors.Is(err, domain.ErrOrderNotEditable):
		return onekiterrors.New(http.StatusConflict, "Order cannot be edited in its current status")
	case errors.Is(err, domain.ErrPricedItemNotFound):
		return onekiterrors.New(http.StatusBadRequest, "Item price could not be resolved")
	case errors.Is(err, domain.ErrDiscountNotAllowed):
		return onekiterrors.New(http.StatusForbidden, "You cannot apply discounts")
	case errors.Is(err, domain.ErrDiscountExceedsLimit):
		return onekiterrors.New(http.StatusBadRequest, "Discount exceeds your allowed limit")
	default:
		return onekiterrors.New(http.StatusInternalServerError, "Failed to update order")
	}
}

// ---- List Held Orders ----

func ListHeldOrdersRoutes() []string { return []string{"/pos/orders/held"} }

type ListHeldOrdersHandler struct{ Service *application.Service }

type ListHeldOrdersRequest struct{}

type ListHeldOrdersResponse struct {
	Orders []OrderPayload `json:"orders"`
}

func (h *ListHeldOrdersHandler) Serve(ctx context.Context, _ *ListHeldOrdersRequest) (ListHeldOrdersResponse, error) {
	user, authErr := requirePOSPermission(ctx, "pos:orders:view")
	if authErr != nil {
		return ListHeldOrdersResponse{}, authErr
	}
	businessID, bizErr := resolveBusinessID(ctx, user)
	if bizErr != nil {
		return ListHeldOrdersResponse{}, bizErr
	}

	orders, err := h.Service.ListHeldOrders(ctx, businessID)
	if err != nil {
		return ListHeldOrdersResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list held orders")
	}

	payloads := make([]OrderPayload, 0, len(orders))
	for _, o := range orders {
		payloads = append(payloads, newOrderPayload(o))
	}
	return ListHeldOrdersResponse{Orders: payloads}, nil
}
