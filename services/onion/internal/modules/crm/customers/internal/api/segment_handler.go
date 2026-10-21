package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/crm/customers/internal/application"
	"github.com/voxire/onion/internal/modules/crm/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/segmentrules"
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
		Filters:       string(seg.Rules),
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

func (h *ListSegmentsHandler) Serve(ctx context.Context, _ *ListSegmentsRequest) (ListSegmentsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListSegmentsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListSegmentsResponse{}, onekiterrors.PermissionDenied()
	}

	segments, err := h.CRMService.ListSegments(ctx, businessID)
	if err != nil {
		return ListSegmentsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list segments")
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

func (h *CreateSegmentHandler) Validate(_ context.Context, req *CreateSegmentRequest) error {
	if req == nil || req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Segment name is required")
	}

	return validateSegmentRules(req.Rules)
}

func (h *CreateSegmentHandler) Serve(ctx context.Context, req *CreateSegmentRequest) (CreateSegmentResponse, error) {
	if req == nil {
		return CreateSegmentResponse{}, onekiterrors.New(http.StatusBadRequest, "Segment name is required")
	}
	if err := validateSegmentRules(req.Rules); err != nil {
		return CreateSegmentResponse{}, err
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateSegmentResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateSegmentResponse{}, onekiterrors.PermissionDenied()
	}

	rules := []byte(req.Rules)

	id, err := h.CRMService.CreateSegment(ctx, domain.SegmentInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       rules,
	})
	if err != nil {
		return CreateSegmentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create segment")
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

func (h *UpdateSegmentHandler) Validate(_ context.Context, req *UpdateSegmentRequest) error {
	if req == nil || req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Segment ID is required")
	}

	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Segment name is required")
	}

	return validateSegmentRules(req.Rules)
}

func (h *UpdateSegmentHandler) Serve(ctx context.Context, req *UpdateSegmentRequest) (UpdateSegmentResponse, error) {
	if req == nil {
		return UpdateSegmentResponse{}, onekiterrors.New(http.StatusBadRequest, "Segment ID is required")
	}
	if err := validateSegmentRules(req.Rules); err != nil {
		return UpdateSegmentResponse{}, err
	}

	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateSegmentResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateSegmentResponse{}, onekiterrors.PermissionDenied()
	}

	rules := []byte(req.Rules)

	if err := h.CRMService.UpdateSegment(ctx, domain.SegmentUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Rules:       rules,
	}); err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return UpdateSegmentResponse{}, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}

		return UpdateSegmentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update segment")
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

func (h *DeleteSegmentHandler) Validate(_ context.Context, req *DeleteSegmentRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Segment ID is required")
	}

	return nil
}

func (h *DeleteSegmentHandler) Serve(ctx context.Context, req *DeleteSegmentRequest) (DeleteSegmentResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteSegmentResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCRMFeaturePermission(ctx, "crm:customers:edit"); err != nil {
		return DeleteSegmentResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteSegmentResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.CRMService.DeleteSegment(ctx, businessID, req.ID); err != nil {
		if errors.Is(err, domain.ErrSegmentNotFound) {
			return DeleteSegmentResponse{}, onekiterrors.New(http.StatusNotFound, "Segment not found")
		}

		return DeleteSegmentResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete segment")
	}

	return DeleteSegmentResponse{Message: "Segment deleted successfully"}, nil
}

func validateSegmentRules(raw json.RawMessage) error {
	if len(raw) == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Segment rules are required")
	}

	var document struct {
		Rules *[]string `json:"rules"`
	}
	if err := json.Unmarshal(raw, &document); err != nil || document.Rules == nil {
		return onekiterrors.New(http.StatusBadRequest, "Segment rules must be a JSON array")
	}

	// An explicit empty rules array is the supported all-customers segment.
	// Omission is still rejected above, so an omitted field cannot erase rules.
	for _, rule := range *document.Rules {
		if strings.TrimSpace(rule) == "" {
			return onekiterrors.New(http.StatusBadRequest, "Segment rules contain an empty rule")
		}
		wrapped, err := json.Marshal(map[string][]string{"rules": []string{rule}})
		if err != nil {
			return onekiterrors.New(http.StatusBadRequest, "Segment rules are invalid")
		}
		fragment, _ := segmentrules.Filter(wrapped)
		if fragment == "" {
			return onekiterrors.New(http.StatusBadRequest, "Segment rules contain an unsupported rule")
		}
	}
	return nil
}
