package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/segments/internal/application"
	"github.com/voxire/onion/internal/modules/segments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type SegmentPayload struct {
	ID            int64           `json:"id"`
	OwnerID       int64           `json:"owner_id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Filters       json.RawMessage `json:"filters"`
	CustomerCount int             `json:"customer_count"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

func newSegmentPayload(seg domain.Segment) SegmentPayload {
	filters := json.RawMessage(seg.Filters)
	if filters == nil {
		filters = json.RawMessage("{}")
	}
	return SegmentPayload{
		ID:            seg.ID,
		OwnerID:       seg.OwnerID,
		Name:          seg.Name,
		Description:   seg.Description,
		Filters:       filters,
		CustomerCount: seg.CustomerCount,
		CreatedAt:     seg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     seg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func SegmentListRoutes() []string { return []string{"/crm/segments"} }

type (
	ListSegmentsHandler  struct{ Service *application.Service }
	ListSegmentsRequest  struct{}
	ListSegmentsResponse struct {
		Segments []SegmentPayload `json:"segments"`
	}
)

func (h *ListSegmentsHandler) Serve(ctx context.Context, _ *ListSegmentsRequest) (ListSegmentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListSegmentsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListSegmentsResponse{}, errorModels.PermissionDeniedServerError()
	}
	segs, err := h.Service.List(ctx, businessID)
	if err != nil {
		return ListSegmentsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list segments"}
	}
	payloads := make([]SegmentPayload, len(segs))
	for i, s := range segs {
		payloads[i] = newSegmentPayload(s)
	}
	return ListSegmentsResponse{Segments: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func SegmentCreateRoutes() []string { return []string{"/crm/segments"} }

type CreateSegmentHandler struct{ Service *application.Service }

type CreateSegmentRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Filters     json.RawMessage `json:"filters"`
}

type CreateSegmentResponse struct {
	SegmentID int64  `json:"segment_id"`
	Message   string `json:"message"`
}

func (h *CreateSegmentHandler) Validate(_ context.Context, req *CreateSegmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateSegmentHandler) Serve(ctx context.Context, req *CreateSegmentRequest) (CreateSegmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	filters := []byte(req.Filters)
	if filters == nil {
		filters = []byte("{}")
	}
	id, err := h.Service.Create(ctx, domain.SegmentInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Filters:     filters,
	})
	if err != nil {
		return CreateSegmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create segment"}
	}
	return CreateSegmentResponse{SegmentID: id, Message: "Segment created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func SegmentUpdateRoutes() []string { return []string{"/crm/segments/{id}"} }

type UpdateSegmentHandler struct{ Service *application.Service }

type UpdateSegmentRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Filters     json.RawMessage `json:"filters"`
}

type UpdateSegmentResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSegmentHandler) Validate(_ context.Context, req *UpdateSegmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateSegmentHandler) Serve(ctx context.Context, req *UpdateSegmentRequest) (UpdateSegmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateSegmentResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid segment ID"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	filters := []byte(req.Filters)
	if filters == nil {
		filters = []byte("{}")
	}
	if updateErr := h.Service.Update(ctx, domain.SegmentUpdateInput{
		ID:          id,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Filters:     filters,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrSegmentNotFound) {
			return UpdateSegmentResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Segment not found"}
		}
		return UpdateSegmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update segment"}
	}
	return UpdateSegmentResponse{Message: "Segment updated successfully"}, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func SegmentDeleteRoutes() []string { return []string{"/crm/segments/{id}"} }

type (
	DeleteSegmentHandler  struct{ Service *application.Service }
	DeleteSegmentRequest  struct{}
	DeleteSegmentResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteSegmentHandler) Serve(ctx context.Context, _ *DeleteSegmentRequest) (DeleteSegmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteSegmentResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid segment ID"}
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	if deleteErr := h.Service.Delete(ctx, businessID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrSegmentNotFound) {
			return DeleteSegmentResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Segment not found"}
		}
		return DeleteSegmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete segment"}
	}
	return DeleteSegmentResponse{Message: "Segment deleted successfully"}, nil
}
