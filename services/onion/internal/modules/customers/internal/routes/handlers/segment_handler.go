package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmsegmentv1 "github.com/voxire/voxie/gen/go/crm/segment/v1"
)

func SegmentListRoutes() []string {
	return []string{"/crm/segments"}
}

func SegmentCreateRoutes() []string {
	return []string{"/crm/segments"}
}

func SegmentUpdateRoutes() []string {
	return []string{"/crm/segments"}
}

func SegmentDeleteRoutes() []string {
	return []string{"/crm/segments"}
}

func newSegmentPayload(seg domain.Segment) *crmsegmentv1.Segment {
	p := &crmsegmentv1.Segment{
		Id:            seg.ID,
		OwnerId:       seg.OwnerID,
		Name:          seg.Name,
		Description:   seg.Description,
		Rules:         string(seg.Rules),
		CustomerCount: int64(seg.CustomerCount),
		CreatedAt:     seg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     seg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if seg.LastCalculatedAt != nil {
		s := seg.LastCalculatedAt.Format("2006-01-02T15:04:05Z")
		p.LastCalculatedAt = &s
	}

	return p
}

// ─── LIST ─────────────────────────────────────────────────────────────────────

type ListSegmentsHandler struct {
	CRMService *application.Service
}

type ListSegmentsRequest struct{}

type ListSegmentsResponse struct {
	Segments []*crmsegmentv1.Segment `json:"segments"`
}

func (h *ListSegmentsHandler) Serve(ctx context.Context, _ *ListSegmentsRequest) (ListSegmentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListSegmentsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListSegmentsResponse{}, errorModels.PermissionDeniedServerError()
	}

	segments, err := h.CRMService.ListSegments(ctx, businessID)
	if err != nil {
		return ListSegmentsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list segments",
		}
	}

	payloads := make([]*crmsegmentv1.Segment, len(segments))
	for i, seg := range segments {
		payloads[i] = newSegmentPayload(seg)
	}

	return ListSegmentsResponse{Segments: payloads}, nil
}

// ─── CREATE ───────────────────────────────────────────────────────────────────

type CreateSegmentHandler struct {
	CRMService *application.Service
}

type CreateSegmentRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rules       json.RawMessage `json:"rules"`
}

type CreateSegmentResponse struct {
	SegmentID int64  `json:"segment_id"`
	Message   string `json:"message"`
}

func (h *CreateSegmentHandler) Validate(_ context.Context, req *CreateSegmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Segment name is required",
		}
	}

	if len(req.Rules) == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Segment rules are required",
		}
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

	rules := []byte(req.Rules)
	if rules == nil {
		rules = []byte("[]")
	}

	id, err := h.CRMService.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       rules,
	})
	if err != nil {
		return CreateSegmentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create segment",
		}
	}

	return CreateSegmentResponse{SegmentID: id, Message: "Segment created successfully"}, nil
}

// ─── UPDATE ───────────────────────────────────────────────────────────────────

type UpdateSegmentHandler struct {
	CRMService *application.Service
}

type UpdateSegmentRequest struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rules       json.RawMessage `json:"rules"`
}

type UpdateSegmentResponse struct {
	Message string `json:"message"`
}

func (h *UpdateSegmentHandler) Validate(_ context.Context, req *UpdateSegmentRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Segment ID is required",
		}
	}

	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Segment name is required",
		}
	}

	return nil
}

func (h *UpdateSegmentHandler) Serve(ctx context.Context, req *UpdateSegmentRequest) (UpdateSegmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	rules := []byte(req.Rules)
	if rules == nil {
		rules = []byte("[]")
	}

	if err := h.CRMService.UpdateSegment(ctx, domain.SegmentUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       rules,
	}); err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return UpdateSegmentResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Segment not found",
			}
		}

		return UpdateSegmentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update segment",
		}
	}

	return UpdateSegmentResponse{Message: "Segment updated successfully"}, nil
}

// ─── DELETE ───────────────────────────────────────────────────────────────────

type DeleteSegmentHandler struct {
	CRMService *application.Service
}

type DeleteSegmentRequest struct {
	ID int64 `json:"id"`
}

type DeleteSegmentResponse struct {
	Message string `json:"message"`
}

func (h *DeleteSegmentHandler) Validate(_ context.Context, req *DeleteSegmentRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Segment ID is required",
		}
	}

	return nil
}

func (h *DeleteSegmentHandler) Serve(ctx context.Context, req *DeleteSegmentRequest) (DeleteSegmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteSegmentResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.CRMService.DeleteSegment(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return DeleteSegmentResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Segment not found",
			}
		}

		return DeleteSegmentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete segment",
		}
	}

	return DeleteSegmentResponse{Message: "Segment deleted successfully"}, nil
}
