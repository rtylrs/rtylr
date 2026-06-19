package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/booking_services/internal/application"
	"github.com/voxire/onion/internal/modules/booking_services/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ListRoutes() []string   { return []string{"/book/services"} }
func CreateRoutes() []string { return []string{"/book/services"} }
func UpdateRoutes() []string { return []string{"/book/services"} }
func DeleteRoutes() []string { return []string{"/book/services"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Services []domain.BookingService `json:"services"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to list booking services"}
	}
	return ListResponse{Services: items}, nil
}

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name            string  `json:"name"`
		Description     string  `json:"description"`
		DurationMinutes int     `json:"duration_minutes"`
		Price           float64 `json:"price"`
		Color           string  `json:"color"`
		IsActive        bool    `json:"is_active"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.DurationMinutes <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "duration_minutes must be greater than 0"}
	}
	if req.Price < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "price must not be negative"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, businessID, domain.BookingServiceInput{Name: req.Name, Description: req.Description, DurationMinutes: req.DurationMinutes, Price: req.Price, Color: req.Color, IsActive: req.IsActive})
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to create booking service"}
	}
	return CreateResponse{ID: id, Message: "Booking service created"}, nil
}

type (
	UpdateHandler struct{ Service *application.Service }
	UpdateRequest struct {
		ID              int64   `json:"id"`
		Name            string  `json:"name"`
		Description     string  `json:"description"`
		DurationMinutes int     `json:"duration_minutes"`
		Price           float64 `json:"price"`
		Color           string  `json:"color"`
		IsActive        bool    `json:"is_active"`
	}
)

type UpdateResponse struct {
	Message string `json:"message"`
}

func (h *UpdateHandler) Validate(_ context.Context, req *UpdateRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "ID is required"}
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.DurationMinutes <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "duration_minutes must be greater than 0"}
	}
	if req.Price < 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, HTTPCode: http.StatusBadRequest, Message: "price must not be negative"}
	}
	return nil
}

func (h *UpdateHandler) Serve(ctx context.Context, req *UpdateRequest) (UpdateResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusUnauthorized, HTTPCode: http.StatusUnauthorized, Message: "unauthorized"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Update(ctx, businessID, req.ID, domain.BookingServiceInput{Name: req.Name, Description: req.Description, DurationMinutes: req.DurationMinutes, Price: req.Price, Color: req.Color, IsActive: req.IsActive}); err != nil {
		return UpdateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to update booking service"}
	}
	return UpdateResponse{Message: "Booking service updated"}, nil
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, HTTPCode: http.StatusInternalServerError, Message: "Failed to delete booking service"}
	}
	return DeleteResponse{Message: "Booking service deleted"}, nil
}
