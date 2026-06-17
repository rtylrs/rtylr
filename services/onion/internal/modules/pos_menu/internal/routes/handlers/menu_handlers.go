package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/application"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type MenuItemPayload struct {
	ID           int64   `json:"id"`
	OwnerID      int64   `json:"owner_id"`
	CategoryID   int64   `json:"category_id"`
	Category     string  `json:"category"`
	CategoryName string  `json:"category_name"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Price        float64 `json:"price"`
	SalePrice    float64 `json:"sale_price"`
	Image        string  `json:"image,omitempty"`
	ImageURL     string  `json:"image_url,omitempty"`
	Available    bool    `json:"available"`
	IsActive     bool    `json:"is_active"`
	Stock        float64 `json:"stock"`
	Description  string  `json:"description,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func toMenuItemPayload(item domain.MenuItem) MenuItemPayload {
	createdAt := ""
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	updatedAt := ""
	if !item.UpdatedAt.IsZero() {
		updatedAt = item.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	return MenuItemPayload{
		ID:           item.ID,
		OwnerID:      item.OwnerID,
		CategoryID:   item.CategoryID,
		Category:     item.Category,
		CategoryName: item.Category,
		Name:         item.Name,
		SKU:          item.SKU,
		Price:        item.SalePrice,
		SalePrice:    item.SalePrice,
		Image:        item.ImageURL,
		ImageURL:     item.ImageURL,
		Available:    item.IsActive,
		IsActive:     item.IsActive,
		Stock:        item.Stock,
		Description:  item.Description,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

func resolveAuthenticatedMenuOwnerID(ctx context.Context) (int64, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil {
		return 0, errorModels.PermissionDeniedServerError()
	}

	return tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID), nil
}

// --- List ---

func ListMenuItemsRoutes() []string { return []string{"/pos/menu/items", "/pos/menu-items"} }

type (
	ListMenuItemsHandler  struct{ Service *application.Service }
	ListMenuItemsRequest  struct{}
	ListMenuItemsResponse struct {
		Items []MenuItemPayload `json:"items"`
	}
)

func (h *ListMenuItemsHandler) Serve(ctx context.Context, _ *ListMenuItemsRequest) (ListMenuItemsResponse, *errorModels.ServerError) {
	ownerID, ownerErr := resolveAuthenticatedMenuOwnerID(ctx)
	if ownerErr != nil {
		return ListMenuItemsResponse{}, ownerErr
	}
	items, err := h.Service.List(ctx, ownerID)
	if err != nil {
		return ListMenuItemsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list menu items"}
	}
	payloads := make([]MenuItemPayload, len(items))
	for i, item := range items {
		payloads[i] = toMenuItemPayload(item)
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
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.SalePrice < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Sale price must be zero or greater"}
	}
	return nil
}

func (h *CreateMenuItemHandler) Serve(ctx context.Context, req *CreateMenuItemRequest) (CreateMenuItemResponse, *errorModels.ServerError) {
	ownerID, ownerErr := resolveAuthenticatedMenuOwnerID(ctx)
	if ownerErr != nil {
		return CreateMenuItemResponse{}, ownerErr
	}
	id, err := h.Service.QuickCreate(ctx, ownerID, domain.MenuItemInput{
		Name:      req.Name,
		SKU:       req.SKU,
		SalePrice: req.SalePrice,
	})
	if err != nil {
		return CreateMenuItemResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create menu item"}
	}
	return CreateMenuItemResponse{ItemID: id, Message: "Menu item created successfully"}, nil
}
