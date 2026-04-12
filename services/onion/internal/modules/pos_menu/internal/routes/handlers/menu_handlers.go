package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
)

type MenuItemPayload struct {
	ID         int64   `json:"id"`
	OwnerID    int64   `json:"owner_id"`
	CategoryID int64   `json:"category_id"`
	Name       string  `json:"name"`
	SKU        string  `json:"sku"`
	SalePrice  float64 `json:"sale_price"`
	IsActive   bool    `json:"is_active"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func newMenuItemPayload(item domain.MenuItem) MenuItemPayload {
	return MenuItemPayload{
		ID:         item.ID,
		OwnerID:    item.OwnerID,
		CategoryID: item.CategoryID,
		Name:       item.Name,
		SKU:        item.SKU,
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
	Items []MenuItemPayload `json:"items"`
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
	payloads := make([]MenuItemPayload, len(items))
	for i, item := range items {
		payloads[i] = newMenuItemPayload(item)
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
