package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos/kds/internal/application"
	"github.com/voxire/onion/internal/modules/pos/kds/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	poskdsv1 "github.com/voxire/voxie/gen/go/pos/kds/v1"
)

type KDSAPI struct {
	Service *application.Service
}

func NewKDSAPI(service *application.Service) *KDSAPI {
	return &KDSAPI{Service: service}
}

func (a *KDSAPI) ListKDSOrders(ctx context.Context, _ *poskdsv1.ListKDSOrdersRequest) (*poskdsv1.ListKDSOrdersResponse, error) {
	businessID, err := authenticatedKDSBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	orders, listErr := a.Service.ListActiveOrdersByStation(ctx, businessID, "")
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list KDS orders")
	}

	payloads := make([]*poskdsv1.KDSOrder, len(orders))
	for i, order := range orders {
		payloads[i] = toProtoKDSOrder(order)
	}
	return &poskdsv1.ListKDSOrdersResponse{Orders: payloads}, nil
}

func (a *KDSAPI) UpdateKDSItemByID(ctx context.Context, req *poskdsv1.UpdateKDSItemByIDRequest) (*poskdsv1.UpdateKDSItemByIDResponse, error) {
	businessID, err := authenticatedKDSBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetStatus()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Status is required")
	}
	if req.GetItemId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid item ID")
	}

	normStatus, normErr := domain.NormalizeItemStatus(body.GetStatus())
	if normErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, normErr.Error())
	}

	if updateErr := a.Service.UpdateItemStatusByItemID(ctx, businessID, req.GetItemId(), normStatus); updateErr != nil {
		if errors.Is(updateErr, domain.ErrItemNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Item not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update item status")
	}
	return &poskdsv1.UpdateKDSItemByIDResponse{Message: "Item status updated"}, nil
}

func (a *KDSAPI) UpdateKDSItemStatus(ctx context.Context, req *poskdsv1.UpdateKDSItemStatusRequest) (*poskdsv1.UpdateKDSItemStatusResponse, error) {
	device, businessID, err := authenticatedKDSDevice(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil || strings.TrimSpace(body.GetStatus()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Status is required")
	}
	if req.GetOrderId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid order ID")
	}
	if req.GetItemId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid item ID")
	}

	normStatus, normErr := domain.NormalizeItemStatus(body.GetStatus())
	if normErr != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, normErr.Error())
	}

	if updateErr := a.Service.UpdateItemStatus(ctx, domain.UpdateItemStatusInput{
		OwnerID:    device.User.UserID,
		BusinessID: businessID,
		OrderID:    req.GetOrderId(),
		ItemID:     req.GetItemId(),
		Status:     normStatus,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrItemNotFound) {
			return nil, onekiterrors.New(http.StatusNotFound, "Item not found")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update item status")
	}
	return &poskdsv1.UpdateKDSItemStatusResponse{Message: "Item status updated"}, nil
}

func authenticatedKDSBusinessID(ctx context.Context) (int64, error) {
	_, businessID, err := authenticatedKDSDevice(ctx)
	return businessID, err
}

func authenticatedKDSDevice(ctx context.Context) (authPkg.Device, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return device, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return device, 0, onekiterrors.PermissionDenied()
	}
	return device, businessID, nil
}

func toProtoKDSOrder(order domain.KDSOrder) *poskdsv1.KDSOrder {
	items := make([]*poskdsv1.KDSItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &poskdsv1.KDSItem{
			Id:       item.ID,
			OrderId:  item.OrderID,
			Name:     item.Name,
			Quantity: item.Quantity,
			Notes:    item.Notes,
			Status:   item.Status,
		}
	}
	return &poskdsv1.KDSOrder{
		Id:           order.ID,
		OrderNumber:  order.OrderNumber,
		Type:         order.Type,
		Status:       order.Status,
		TableId:      order.TableID,
		CustomerName: order.CustomerName,
		KitchenNotes: order.KitchenNotes,
		CreatedAt:    order.CreatedAt.Format("2006-01-02T15:04:05Z"),
		Items:        items,
	}
}
