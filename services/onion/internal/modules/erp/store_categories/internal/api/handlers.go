package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/erp/store_categories/internal/application"
	"github.com/voxire/onion/internal/modules/erp/store_categories/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListRoutes() []string   { return []string{"/store/categories"} }
func CreateRoutes() []string { return []string{"/store/categories"} }
func UpdateRoutes() []string { return []string{"/store/categories"} }
func DeleteRoutes() []string { return []string{"/store/categories"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Categories []domain.StoreCategory `json:"categories"`
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
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list store categories")
	}
	return ListResponse{Categories: items}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
		ParentID    int64  `json:"parent_id"`
		SortOrder   int    `json:"sort_order"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return onekiterrors.New(http.StatusBadRequest, "Name is required")
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
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.StoreCategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description, ImageURL: req.ImageURL, ParentID: req.ParentID, SortOrder: req.SortOrder})
	if err != nil {
		if errors.Is(err, domain.ErrParentNotFound) {
			return CreateResponse{}, onekiterrors.New(http.StatusBadRequest, "Parent category not found")
		}
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create store category")
	}
	return CreateResponse{ID: id, Message: "Store category created successfully"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
		ImageURL    string `json:"image_url"`
		ParentID    int64  `json:"parent_id"`
		SortOrder   int    `json:"sort_order"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "ID is required")
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
	if err := h.Service.Update(ctx, businessID, req.ID, domain.StoreCategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description, ImageURL: req.ImageURL, ParentID: req.ParentID, SortOrder: req.SortOrder}); err != nil {
		if errors.Is(err, domain.ErrParentNotFound) {
			return UpdateResponse{}, onekiterrors.New(http.StatusBadRequest, "Parent category not found")
		}
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update store category")
	}
	return UpdateResponse{Message: "Store category updated successfully"}, nil
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
		return DeleteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete store category")
	}
	return DeleteResponse{Message: "Store category deleted successfully"}, nil
}
