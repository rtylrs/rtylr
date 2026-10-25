package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/application"
	"github.com/voxire/onion/internal/modules/erp/products/internal/catalog/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	posmenuitemv1 "github.com/voxire/voxie/gen/go/pos/menu_item/v1"
)

type MenuItemAPI struct {
	Service *application.Service
}

func NewMenuItemAPI(service *application.Service) *MenuItemAPI {
	return &MenuItemAPI{Service: service}
}

func (a *MenuItemAPI) ListPOSCatalog(ctx context.Context, _ *posmenuitemv1.ListMenuItemsRequest) (*posmenuitemv1.ListMenuItemsResponse, error) {
	user := authPkg.DeviceCtx(ctx).User
	if !authPkg.HasAppPermission(user, authorization.AppPOS, "pos:orders:create") || !authPkg.HasAnyAppScope(user, authorization.AppPOS) {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, user.BusinessID, user.UserID)
	if err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	items, err := a.Service.List(ctx, businessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load POS catalog")
	}
	payloads := make([]*posmenuitemv1.MenuItem, len(items))
	for i, item := range items {
		payloads[i] = toProtoMenuItem(item)
		// Selling a product does not grant access to its purchase cost.
		payloads[i].UnitCost = nil
	}
	return &posmenuitemv1.ListMenuItemsResponse{Items: payloads}, nil
}

func (a *MenuItemAPI) ListMenuItems(ctx context.Context, _ *posmenuitemv1.ListMenuItemsRequest) (*posmenuitemv1.ListMenuItemsResponse, error) {
	_, businessID, err := requireMenuPermission(ctx, "menu:menus:view")
	if err != nil {
		return nil, err
	}

	items, listErr := a.Service.List(ctx, businessID)
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list menu items")
	}

	payloads := make([]*posmenuitemv1.MenuItem, len(items))
	for i, item := range items {
		payloads[i] = toProtoMenuItem(item)
	}
	return &posmenuitemv1.ListMenuItemsResponse{Items: payloads}, nil
}

func (a *MenuItemAPI) ListPublicMenuItems(ctx context.Context, req *posmenuitemv1.ListPublicMenuItemsRequest) (*posmenuitemv1.ListPublicMenuItemsResponse, error) {
	if req.GetBusinessId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "business_id is required")
	}

	items, listErr := a.Service.ListPublic(ctx, req.GetBusinessId())
	if listErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load public menu")
	}

	payloads := make([]*posmenuitemv1.MenuItem, len(items))
	for i, item := range items {
		payloads[i] = toPublicProtoMenuItem(item)
	}
	return &posmenuitemv1.ListPublicMenuItemsResponse{Items: payloads}, nil
}

func toPublicProtoMenuItem(item domain.MenuItem) *posmenuitemv1.MenuItem {
	return &posmenuitemv1.MenuItem{
		Id:          item.ID,
		Name:        item.Name,
		SalePrice:   item.SalePrice,
		IsActive:    item.IsActive,
		ImageUrl:    item.ImageURL,
		Category:    item.Category,
		Description: item.Description,
	}
}

func (a *MenuItemAPI) CreateMenuItem(ctx context.Context, req *posmenuitemv1.CreateMenuItemRequest) (*posmenuitemv1.CreateMenuItemResponse, error) {
	_, businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.GetName())
	if name == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.GetSalePrice() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Sale price must be zero or greater")
	}

	id, createErr := a.Service.QuickCreate(ctx, businessID, domain.MenuItemInput{
		Name:      name,
		SKU:       req.GetSku(),
		SalePrice: req.GetSalePrice(),
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create menu item")
	}
	return &posmenuitemv1.CreateMenuItemResponse{ItemId: id, Message: "Menu item created successfully"}, nil
}

func (a *MenuItemAPI) UpdateMenuItem(ctx context.Context, req *posmenuitemv1.UpdateMenuItemRequest) (*posmenuitemv1.UpdateMenuItemResponse, error) {
	_, businessID, err := requireMenuPermission(ctx, "menu:menus:edit")
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Menu item body is required")
	}
	if body.GetSalePrice() < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Sale price must be zero or greater")
	}

	if updateErr := a.Service.Update(ctx, businessID, req.GetId(), domain.MenuItemUpdateInput{
		Name:      body.GetName(),
		SalePrice: body.GetSalePrice(),
		Stock:     body.GetStock(),
		IsActive:  body.GetIsActive(),
	}); updateErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update menu item")
	}
	return &posmenuitemv1.UpdateMenuItemResponse{Message: "Menu item updated successfully"}, nil
}

func requireMenuPermission(ctx context.Context, permission string) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || !authPkg.HasAppPermission(authUser, authorization.AppMenu, permission) {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	return authUser, businessID, nil
}

func toProtoMenuItem(item domain.MenuItem) *posmenuitemv1.MenuItem {
	createdAt := ""
	if !item.CreatedAt.IsZero() {
		createdAt = item.CreatedAt.Format("2006-01-02T15:04:05Z")
	}
	updatedAt := ""
	if !item.UpdatedAt.IsZero() {
		updatedAt = item.UpdatedAt.Format("2006-01-02T15:04:05Z")
	}

	productId := item.ID
	unitCost := item.CostPrice
	return &posmenuitemv1.MenuItem{
		Id:          item.ID,
		ProductId:   &productId,
		UnitCost:    &unitCost,
		OwnerId:     item.OwnerID,
		CategoryId:  item.CategoryID,
		Name:        item.Name,
		Sku:         item.SKU,
		SalePrice:   item.SalePrice,
		IsActive:    item.IsActive,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		ImageUrl:    item.ImageURL,
		Category:    item.Category,
		Barcode:     item.Barcode,
		Description: item.Description,
		TaxRate:     item.TaxRate,
		Stock:       item.Stock,
	}
}
