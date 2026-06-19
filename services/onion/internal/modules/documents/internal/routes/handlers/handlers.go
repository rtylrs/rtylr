package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/documents/internal/application"
	"github.com/voxire/onion/internal/modules/documents/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

var expiryDateLayouts = []string{
	"2006-01-02",
	time.RFC3339,
	"2006-01-02 15:04:05",
}

func isValidExpiryDate(s string) bool {
	for _, layout := range expiryDateLayouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

type DocumentPayload struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	StaffID    int64     `json:"staff_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	FileURL    string    `json:"file_url"`
	ExpiryDate time.Time `json:"expiry_date"`
	CreatedAt  time.Time `json:"created_at"`
}

func newPayload(d domain.Document) DocumentPayload {
	return DocumentPayload{
		ID: d.ID, OwnerID: d.OwnerID, StaffID: d.StaffID, Name: d.Name,
		Type: d.Type, FileURL: d.FileURL, ExpiryDate: d.ExpiryDate, CreatedAt: d.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/documents"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Documents []DocumentPayload `json:"documents"`
	}
)

func (h *ListHandler) Serve(ctx context.Context, _ *ListRequest) (ListResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListResponse{}, errorModels.PermissionDeniedServerError()
	}
	items, err := h.Service.List(ctx, businessID)
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

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		StaffID    int64  `json:"staff_id"`
		Name       string `json:"name"`
		Type       string `json:"type"`
		FileURL    string `json:"file_url"`
		ExpiryDate string `json:"expiry_date"`
	}
)

type CreateResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h *CreateHandler) Validate(_ context.Context, req *CreateRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	if req.ExpiryDate != "" && !isValidExpiryDate(req.ExpiryDate) {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "expiry_date is not a valid date"}
	}
	return nil
}

func (h *CreateHandler) Serve(ctx context.Context, req *CreateRequest) (CreateResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.DocumentInput{
		StaffID: req.StaffID, Name: req.Name, Type: req.Type,
		FileURL: req.FileURL, ExpiryDate: req.ExpiryDate,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create document"}
	}
	return CreateResponse{ID: id, Message: "Document created successfully"}, nil
}

func DeleteRoutes() []string { return []string{"/hr/documents"} }

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
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteHandler) Serve(ctx context.Context, req *DeleteRequest) (DeleteResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.Delete(ctx, businessID, req.ID); err != nil {
		return DeleteResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete document"}
	}
	return DeleteResponse{Message: "Document deleted successfully"}, nil
}
