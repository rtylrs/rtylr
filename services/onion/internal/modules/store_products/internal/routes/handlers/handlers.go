package handlers

import (
	"context"
	"net/http"

	storev1 "github.com/voxire/voxie/gen/go/store/v1"

	"github.com/voxire/onion/internal/modules/store_products/internal/application"
	"github.com/voxire/onion/internal/modules/store_products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func toProtoProduct(p domain.StoreProduct) *storev1.StoreProduct {
	return &storev1.StoreProduct{
		Id: p.ID, OwnerId: p.OwnerID, ProductId: p.ProductID, Title: p.Title,
		Slug: p.Slug, Description: p.Description, Images: p.Images,
		Price: p.Price, CompareAtPrice: p.CompareAtPrice, IsPublished: p.IsPublished,
		SortOrder: int32(p.SortOrder),
		CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ListRoutes() []string   { return []string{"/store/products"} }
func CreateRoutes() []string { return []string{"/store/products"} }
func UpdateRoutes() []string { return []string{"/store/products"} }
func DeleteRoutes() []string { return []string{"/store/products"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Products []*storev1.StoreProduct `json:"products"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list store products"}
	}
	out := make([]*storev1.StoreProduct, 0, len(items))
	for _, i := range items {
		out = append(out, toProtoProduct(i))
	}
	return ListResponse{Products: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
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
type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Title is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	id, err := h.Service.Create(ctx, authUser.UserID, domain.StoreProductInput{
		ProductID: req.ProductID, Title: req.Title, Slug: req.Slug, Description: req.Description,
		Images: req.Images, Price: req.Price, CompareAtPrice: req.CompareAtPrice,
		IsPublished: req.IsPublished, SortOrder: req.SortOrder,
	})
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create store product"}
	}
	return CreateResponse{ID: id, Message: "Store product created successfully"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
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
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.StoreProductInput{
		ProductID: req.ProductID, Title: req.Title, Slug: req.Slug, Description: req.Description,
		Images: req.Images, Price: req.Price, CompareAtPrice: req.CompareAtPrice,
		IsPublished: req.IsPublished, SortOrder: req.SortOrder,
	}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update store product"}
	}
	return UpdateResponse{Message: "Store product updated successfully"}, nil
}

type DeleteHandler struct{ Service *application.Service }
type DeleteRequest struct{ ID int64 `json:"id"` }
type DeleteResponse struct{ Message string `json:"message"` }

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	if err := h.Service.Delete(ctx, authUser.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete store product"}
	}
	return DeleteResponse{Message: "Store product deleted successfully"}, nil
}
