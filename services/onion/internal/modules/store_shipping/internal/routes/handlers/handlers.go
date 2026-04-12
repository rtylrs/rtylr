package handlers

import (
	"context"
	"net/http"

	storev1 "github.com/voxire/voxie/gen/go/store/v1"

	"github.com/voxire/onion/internal/modules/store_shipping/internal/application"
	"github.com/voxire/onion/internal/modules/store_shipping/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func toProtoShipping(s domain.ShippingMethod) *storev1.ShippingMethod {
	return &storev1.ShippingMethod{Id: s.ID, OwnerId: s.OwnerID, Name: s.Name, Price: s.Price, EstimatedDays: int32(s.EstimatedDays), IsActive: s.IsActive, CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}
}

func ListRoutes() []string   { return []string{"/store/shipping"} }
func CreateRoutes() []string { return []string{"/store/shipping"} }
func UpdateRoutes() []string { return []string{"/store/shipping"} }
func DeleteRoutes() []string { return []string{"/store/shipping"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct{ Methods []*storev1.ShippingMethod `json:"methods"` }

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	items, err := h.Service.List(ctx, authUser.UserID)
	if err != nil { return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list shipping methods"} }
	out := make([]*storev1.ShippingMethod, 0, len(items))
	for _, i := range items { out = append(out, toProtoShipping(i)) }
	return ListResponse{Methods: out}, nil
}

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct { Name string `json:"name"`; Price float64 `json:"price"`; EstimatedDays int `json:"estimated_days"`; IsActive bool `json:"is_active"` }
type CreateResponse struct { ID int64 `json:"id"`; Message string `json:"message"` }

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"} }
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	id, err := h.Service.Create(ctx, authUser.UserID, domain.ShippingMethodInput{Name: req.Name, Price: req.Price, EstimatedDays: req.EstimatedDays, IsActive: req.IsActive})
	if err != nil { return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create shipping method"} }
	return CreateResponse{ID: id, Message: "Shipping method created"}, nil
}

type UpdateHandler struct{ Service *application.Service }
type UpdateRequest struct { ID int64 `json:"id"`; Name string `json:"name"`; Price float64 `json:"price"`; EstimatedDays int `json:"estimated_days"`; IsActive bool `json:"is_active"` }
type UpdateResponse struct{ Message string `json:"message"` }

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 { return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"} }
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok { return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"} }
	if err := h.Service.Update(ctx, authUser.UserID, req.ID, domain.ShippingMethodInput{Name: req.Name, Price: req.Price, EstimatedDays: req.EstimatedDays, IsActive: req.IsActive}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update shipping method"}
	}
	return UpdateResponse{Message: "Shipping method updated"}, nil
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
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete shipping method"}
	}
	return DeleteResponse{Message: "Shipping method deleted"}, nil
}
