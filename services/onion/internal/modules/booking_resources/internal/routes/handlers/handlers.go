package handlers

import (
	bookv1 "github.com/voxire/voxie/gen/go/book/v1"

	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/booking_resources/internal/application"
	"github.com/voxire/onion/internal/modules/booking_resources/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func toProto(r domain.Resource) *bookv1.Resource {
	return &bookv1.Resource{
		Id:        r.ID,
		Name:      r.Name,
		Title:     r.Type,
		IsActive:  r.IsActive,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ListRoutes() []string   { return []string{"/book/resources"} }
func CreateRoutes() []string { return []string{"/book/resources"} }
func UpdateRoutes() []string { return []string{"/book/resources"} }
func DeleteRoutes() []string { return []string{"/book/resources"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Resources []*bookv1.Resource `json:"resources"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil { return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list resources"} }
	out := make([]*bookv1.Resource, 0, len(items)); for _, i := range items { out = append(out, toProto(i)) }
	return ListResponse{Resources: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct { Name string `json:"name"`; Type string `json:"type"`; Capacity int `json:"capacity"`; IsActive bool `json:"is_active"` }
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"} }; return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.ResourceInput{Name: req.Name, Type: req.Type, Capacity: req.Capacity, IsActive: req.IsActive})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create resource"} }
	return CreateResponse{ID: id, Message: "Resource created"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct { ID int64 `json:"id"`; Name string `json:"name"`; Type string `json:"type"`; Capacity int `json:"capacity"`; IsActive bool `json:"is_active"` }
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.ResourceInput{Name: req.Name, Type: req.Type, Capacity: req.Capacity, IsActive: req.IsActive}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update resource"}
	}
	return UpdateResponse{Message: "Resource updated"}, nil
}

type DeleteHandler struct{ Service *application.Service }
type DeleteRequest struct{ ID int64 `json:"id"` }
type DeleteResponse struct{ Message string `json:"message"` }

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }; return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Delete(ctx, authUser.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete resource"}
	}
	return DeleteResponse{Message: "Resource deleted"}, nil
}
