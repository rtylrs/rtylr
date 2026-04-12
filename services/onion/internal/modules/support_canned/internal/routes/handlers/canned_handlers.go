package handlers

import (
	"context"
	"net/http"

	suppv1 "github.com/voxire/voxie/gen/go/support/v1"

	"github.com/voxire/onion/internal/modules/support_canned/internal/application"
	"github.com/voxire/onion/internal/modules/support_canned/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ListCannedRoutes() []string   { return []string{"/support/canned"} }
func CreateCannedRoutes() []string { return []string{"/support/canned"} }
func UpdateCannedRoutes() []string { return []string{"/support/canned"} }
func DeleteCannedRoutes() []string { return []string{"/support/canned"} }

type ListCannedHandler struct{ Service *application.Service }
type ListCannedRequest struct{}
type ListCannedResponse struct {
	CannedResponses []*suppv1.CannedResponse `json:"canned_responses"`
}

func (h *ListCannedHandler) Serve(ctx context.Context, _ *ListCannedRequest) (ListCannedResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCannedResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListCannedResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list canned responses"}
	}
	payloads := make([]*suppv1.CannedResponse, len(items))
	for i, c := range items {
		payloads[i] = &suppv1.CannedResponse{
			Id: c.ID, BusinessId: c.BusinessID, Name: c.Name,
			Body: c.Body, Category: c.Category, Shortcut: c.Shortcut,
			CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return ListCannedResponse{CannedResponses: payloads}, nil
}

type CreateCannedHandler struct{ Service *application.Service }
type CreateCannedRequest struct {
	Name     string `json:"name"`
	Body     string `json:"body"`
	Category string `json:"category"`
	Shortcut string `json:"shortcut"`
}
type CreateCannedResponse struct {
	CannedResponseID int64  `json:"canned_response_id"`
	Message          string `json:"message"`
}

func (h *CreateCannedHandler) Validate(_ context.Context, req *CreateCannedRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.Body == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Body is required"}
	}
	return nil
}

func (h *CreateCannedHandler) Serve(ctx context.Context, req *CreateCannedRequest) (CreateCannedResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCannedResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.CannedResponseInput{
		BusinessID: device.User.UserID, Name: req.Name, Body: req.Body,
		Category: req.Category, Shortcut: req.Shortcut,
	})
	if err != nil {
		return CreateCannedResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create canned response"}
	}
	return CreateCannedResponse{CannedResponseID: id, Message: "Canned response created"}, nil
}

type UpdateCannedHandler struct{ Service *application.Service }
type UpdateCannedRequest struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Body     string `json:"body"`
	Category string `json:"category"`
	Shortcut string `json:"shortcut"`
}
type UpdateCannedResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCannedHandler) Validate(_ context.Context, req *UpdateCannedRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *UpdateCannedHandler) Serve(ctx context.Context, req *UpdateCannedRequest) (UpdateCannedResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCannedResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Update(ctx, device.User.UserID, req.ID, domain.CannedResponseInput{
		Name: req.Name, Body: req.Body, Category: req.Category, Shortcut: req.Shortcut,
	})
	if err != nil {
		return UpdateCannedResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update canned response"}
	}
	return UpdateCannedResponse{Message: "Canned response updated"}, nil
}

type DeleteCannedHandler struct{ Service *application.Service }
type DeleteCannedRequest struct {
	ID int64 `json:"id"`
}
type DeleteCannedResponse struct {
	Message string `json:"message"`
}

func (h *DeleteCannedHandler) Validate(_ context.Context, req *DeleteCannedRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "ID is required"}
	}
	return nil
}

func (h *DeleteCannedHandler) Serve(ctx context.Context, req *DeleteCannedRequest) (DeleteCannedResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteCannedResponse{}, errorModels.PermissionDeniedServerError()
	}
	err := h.Service.Delete(ctx, device.User.UserID, req.ID)
	if err != nil {
		return DeleteCannedResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete canned response"}
	}
	return DeleteCannedResponse{Message: "Canned response deleted"}, nil
}
