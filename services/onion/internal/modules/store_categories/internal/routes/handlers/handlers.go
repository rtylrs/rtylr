package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/store_categories/internal/application"
	"github.com/voxire/onion/internal/modules/store_categories/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type Payload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"owner_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	ParentID    int64  `json:"parent_id"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
}

func toPayload(c domain.StoreCategory) Payload {
	return Payload{
		ID: c.ID, OwnerID: c.OwnerID, Name: c.Name, Slug: c.Slug,
		Description: c.Description, ImageURL: c.ImageURL, ParentID: c.ParentID,
		SortOrder: c.SortOrder, CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ListRoutes() []string   { return []string{"/store/categories"} }
func CreateRoutes() []string { return []string{"/store/categories"} }
func UpdateRoutes() []string { return []string{"/store/categories"} }
func DeleteRoutes() []string { return []string{"/store/categories"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Categories []Payload `json:"categories"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list store categories"}
	}
	out := make([]Payload, 0, len(items))
	for _, i := range items { out = append(out, toPayload(i)) }
	return ListResponse{Categories: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	Name string `json:"name"`; Slug string `json:"slug"`; Description string `json:"description"`
	ImageURL string `json:"image_url"`; ParentID int64 `json:"parent_id"`; SortOrder int `json:"sort_order"`
}
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"} }
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.StoreCategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description, ImageURL: req.ImageURL, ParentID: req.ParentID, SortOrder: req.SortOrder})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create store category"} }
	return CreateResponse{ID: id, Message: "Store category created successfully"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct {
	ID int64 `json:"id"`; Name string `json:"name"`; Slug string `json:"slug"`; Description string `json:"description"`
	ImageURL string `json:"image_url"`; ParentID int64 `json:"parent_id"`; SortOrder int `json:"sort_order"`
}
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.StoreCategoryInput{Name: req.Name, Slug: req.Slug, Description: req.Description, ImageURL: req.ImageURL, ParentID: req.ParentID, SortOrder: req.SortOrder}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update store category"}
	}
	return UpdateResponse{Message: "Store category updated successfully"}, nil
}

type DeleteHandler struct{ Service *application.Service }
type DeleteRequest struct{ ID int64 `json:"id"` }
type DeleteResponse struct{ Message string `json:"message"` }

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Delete(ctx, authUser.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete store category"}
	}
	return DeleteResponse{Message: "Store category deleted successfully"}, nil
}
