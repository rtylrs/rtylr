package handlers

import (
	"context"
	"net/http"

	posv1 "github.com/voxire/voxie/gen/go/pos/v1"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func toProtoMenuItem(item domain.MenuItem) *posv1.MenuItem {
	return &posv1.MenuItem{
		Id:         item.ID,
		OwnerId:    item.OwnerID,
		CategoryId: item.CategoryID,
		Name:       item.Name,
		Sku:        item.SKU,
		SalePrice:  item.SalePrice,
		IsActive:   item.IsActive,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List ---

func ListMenuItemsRoutes() []string { return []string{"/pos/menu/items"} }

type ListMenuItemsHandler struct{ Service *application.Service }
type ListMenuItemsRequest struct{}
type ListMenuItemsResponse struct {
	Items []*posv1.MenuItem `json:"items"`
}

func (h *ListMenuItemsHandler) Serve(ctx context.Context, _ *ListMenuItemsRequest) (ListMenuItemsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListMenuItemsResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListMenuItemsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list menu items"}
	}
	payloads := make([]*posv1.MenuItem, len(items))
	for i, item := range items {
		payloads[i] = toProtoMenuItem(item)
	}
	return ListMenuItemsResponse{Items: payloads}, nil
}

// --- Quick Create ---

func CreateMenuItemRoutes() []string { return []string{"/pos/menu/items"} }

type CreateMenuItemHandler struct{ Service *application.Service }

type CreateMenuItemRequest struct {
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	SalePrice float64 `json:"sale_price"`
}

type CreateMenuItemResponse struct {
	ItemID  int64  `json:"item_id"`
	Message string `json:"message"`
}

func (h *CreateMenuItemHandler) Validate(_ context.Context, req *CreateMenuItemRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateMenuItemHandler) Serve(ctx context.Context, req *CreateMenuItemRequest) (CreateMenuItemResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateMenuItemResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.QuickCreate(ctx, device.User.UserID, domain.MenuItemInput{
		Name:      req.Name,
		SKU:       req.SKU,
		SalePrice: req.SalePrice,
	})
	if err != nil {
		return CreateMenuItemResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create menu item"}
	}
	return CreateMenuItemResponse{ItemID: id, Message: "Menu item created successfully"}, nil
}
