package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/store_products/internal/application"
	"github.com/voxire/onion/internal/modules/erp/store_products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListRoutes() []string   { return []string{"/store/products"} }
func CreateRoutes() []string { return []string{"/store/products"} }
func UpdateRoutes() []string { return []string{"/store/products"} }
func DeleteRoutes() []string { return []string{"/store/products"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Products []domain.StoreProduct `json:"products"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, onekiterrors.PermissionDenied()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list store products")
	}
	return ListResponse{Products: items}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		ProductID      int64   `json:"product_id"`
		Title          string  `json:"title"`
		Slug           string  `json:"slug"`
		Description    string  `json:"description"`
		Images         string  `json:"images"`
		Price          float64 `json:"price"`
		CompareAtPrice float64 `json:"compare_at_price"`
		IsPublished    bool    `json:"is_published"`
		SortOrder      int     `json:"sort_order"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Title is required")
	}
	if req.Price < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Price cannot be negative")
	}
	if req.CompareAtPrice < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Compare-at price cannot be negative")
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.StoreProductInput{
		ProductID: req.ProductID, Title: req.Title, Slug: req.Slug, Description: req.Description,
		Images: req.Images, Price: req.Price, CompareAtPrice: req.CompareAtPrice,
		IsPublished: req.IsPublished, SortOrder: req.SortOrder,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPrice) {
			return CreateResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create store product")
	}
	return CreateResponse{ID: id, Message: "Store product created successfully"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID             int64   `json:"id"`
		ProductID      int64   `json:"product_id"`
		Title          string  `json:"title"`
		Slug           string  `json:"slug"`
		Description    string  `json:"description"`
		Images         string  `json:"images"`
		Price          float64 `json:"price"`
		CompareAtPrice float64 `json:"compare_at_price"`
		IsPublished    bool    `json:"is_published"`
		SortOrder      int     `json:"sort_order"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	if req.Price < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Price cannot be negative")
	}
	if req.CompareAtPrice < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Compare-at price cannot be negative")
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.StoreProductInput{
		ProductID: req.ProductID, Title: req.Title, Slug: req.Slug, Description: req.Description,
		Images: req.Images, Price: req.Price, CompareAtPrice: req.CompareAtPrice,
		IsPublished: req.IsPublished, SortOrder: req.SortOrder,
	}); err != nil {
		if errors.Is(err, domain.ErrInvalidPrice) {
			return UpdateResponse{}, onekiterrors.New(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, domain.ErrStoreProductNotFound) {
			return UpdateResponse{}, onekiterrors.New(http.StatusNotFound, "Store product not found")
		}
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update store product")
	}
	return UpdateResponse{Message: "Store product updated successfully"}, nil
}

type (
	DeleteHandler struct{ Service *application.Service }
	DeleteRequest struct {
		ID int64 `json:"id"`
	}
)

type DeleteResponse struct {
	Message string `json:"message"`
}

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		return DeleteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete store product")
	}
	return DeleteResponse{Message: "Store product deleted successfully"}, nil
}
