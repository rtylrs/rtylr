package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/training/internal/application"
	"github.com/voxire/onion/internal/modules/training/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type TrainingCoursePayload struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DurationHours float64   `json:"duration_hours"`
	IsRequired    bool      `json:"is_required"`
	CreatedAt     time.Time `json:"created_at"`
}

func newPayload(t domain.TrainingCourse) TrainingCoursePayload {
	return TrainingCoursePayload{
		ID: t.ID, OwnerID: t.OwnerID, Name: t.Name, Description: t.Description,
		DurationHours: t.DurationHours, IsRequired: t.IsRequired, CreatedAt: t.CreatedAt,
	}
}

func ListRoutes() []string { return []string{"/hr/training"} }

type (
	ListHandler  struct{ Service *application.Service }
	ListRequest  struct{}
	ListResponse struct {
		Courses []TrainingCoursePayload `json:"courses"`
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
		return ListResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list training courses"}
	}
	payloads := make([]TrainingCoursePayload, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newPayload(item))
	}
	return ListResponse{Courses: payloads}, nil
}

func CreateRoutes() []string { return []string{"/hr/training"} }

type (
	CreateHandler struct{ Service *application.Service }
	CreateRequest struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		DurationHours float64 `json:"duration_hours"`
		IsRequired    bool    `json:"is_required"`
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
	input := domain.TrainingCourseInput{
		Name: req.Name, Description: req.Description,
		DurationHours: req.DurationHours, IsRequired: req.IsRequired,
	}
	id, err := h.Service.Create(ctx, businessID, input)
	if err != nil {
		return CreateResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create training course"}
	}
	return CreateResponse{ID: id, Message: "Training course created successfully"}, nil
}
