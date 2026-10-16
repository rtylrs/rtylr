package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/store/shipping/internal/application"
	"github.com/voxire/onion/internal/modules/store/shipping/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListRoutes() []string   { return []string{"/store/shipping"} }
func CreateRoutes() []string { return []string{"/store/shipping"} }
func UpdateRoutes() []string { return []string{"/store/shipping"} }
func DeleteRoutes() []string { return []string{"/store/shipping"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Methods []shippingMethodPayload `json:"methods"`
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
		return ListResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list shipping methods")
	}
	return ListResponse{Methods: payloadShippingMethodList(items)}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name          string  `json:"name"`
		Price         float64 `json:"price"`
		EstimatedDays int     `json:"estimated_days"`
		IsActive      bool    `json:"is_active"`
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
	if req.Price < 0 {
		return onekiterrors.New(http.StatusBadRequest, "Price cannot be negative")
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.IsBusinessAdmin(authUser) {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, onekiterrors.PermissionDenied()
	}
	id, err := h.Service.Create(ctx, authUser.UserID, businessID, domain.ShippingMethodInput{Name: req.Name, Price: req.Price, EstimatedDays: req.EstimatedDays, IsActive: req.IsActive})
	if err != nil {
		return CreateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create shipping method")
	}
	return CreateResponse{ID: id, Message: "Shipping method created"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID            int64   `json:"id"`
		Name          string  `json:"name"`
		Price         float64 `json:"price"`
		EstimatedDays int     `json:"estimated_days"`
		IsActive      bool    `json:"is_active"`
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
	if !authPkg.IsBusinessAdmin(authUser) {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.ShippingMethodInput{Name: req.Name, Price: req.Price, EstimatedDays: req.EstimatedDays, IsActive: req.IsActive}); err != nil {
		return UpdateResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update shipping method")
	}
	return UpdateResponse{Message: "Shipping method updated"}, nil
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
	if !authPkg.IsBusinessAdmin(authUser) {
		return DeleteResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteResponse{}, onekiterrors.PermissionDenied()
	}
	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		return DeleteResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete shipping method")
	}
	return DeleteResponse{Message: "Shipping method deleted"}, nil
}
