package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/application"
	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

type DocumentPayload struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"ownerId"`
	StaffID    int64     `json:"staffId"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	FileURL    string    `json:"fileUrl"`
	ExpiryDate time.Time `json:"expiryDate"`
	CreatedAt  time.Time `json:"createdAt"`
}

func newPayload(d domain.Document) DocumentPayload {
	return DocumentPayload{
		ID: d.ID, OwnerID: d.OwnerID, StaffID: d.StaffID, Name: d.Name,
		Type: d.Type, FileURL: d.FileURL, ExpiryDate: d.ExpiryDate, CreatedAt: d.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/documents"} }

type ListHandler struct{ Service *application.Service }
type ListRequest struct{}
type ListResponse struct {
	Documents []DocumentPayload `json:"documents"`
}

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list documents"}
	}
	payloads := make([]DocumentPayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Documents: payloads}, nil
}

func CreateRoutes() []string { return []string{"/hr/documents"} }

type CreateHandler struct{ Service *application.Service }
type CreateRequest struct {
	StaffID    int64  `json:"staffId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	FileURL    string `json:"fileUrl"`
	ExpiryDate string `json:"expiryDate"`
}
type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.DocumentInput{
		StaffID: req.StaffID, Name: req.Name, Type: req.Type,
		FileURL: req.FileURL, ExpiryDate: req.ExpiryDate,
	}
	id, err := h.Service.Create(ctx, device.User.UserID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create document"}
	}
	return CreateResponse{ID: id, Message: "Document created successfully"}, nil
}

func DeleteRoutes() []string { return []string{"/hr/documents"} }

type DeleteHandler struct{ Service *application.Service }
type DeleteRequest struct {
	ID int64 `json:"id"`
}
type DeleteResponse struct {
	Message string `json:"message"`
}

func (h *DeleteHandler) Validate(_ context.Context, req *DeleteRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Delete(ctx, device.User.UserID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete document"}
	}
	return DeleteResponse{Message: "Document deleted successfully"}, nil
}
