package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type CampaignPayload struct {
	ID                int64   `json:"id"`
	OwnerID           int64   `json:"owner_id"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Status            string  `json:"status"`
	AudienceSegmentID *int64  `json:"audience_segment_id"`
	Subject           string  `json:"subject"`
	Body              string  `json:"body"`
	SentCount         int     `json:"sent_count"`
	OpenCount         int     `json:"open_count"`
	ScheduledAt       *string `json:"scheduled_at"`
	SentAt            *string `json:"sent_at"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

func newCampaignPayload(c domain.Campaign) CampaignPayload {
	p := CampaignPayload{
		ID:                c.ID,
		OwnerID:           c.OwnerID,
		Name:              c.Name,
		Type:              c.Type,
		Status:            c.Status,
		AudienceSegmentID: c.AudienceSegmentID,
		Subject:           c.Subject,
		Body:              c.Body,
		SentCount:         c.SentCount,
		OpenCount:         c.OpenCount,
		CreatedAt:         c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.ScheduledAt != nil {
		s := c.ScheduledAt.Format("2006-01-02T15:04:05Z")
		p.ScheduledAt = &s
	}
	if c.SentAt != nil {
		s := c.SentAt.Format("2006-01-02T15:04:05Z")
		p.SentAt = &s
	}
	return p
}

// ─── List ─────────────────────────────────────────────────────────────────────

func CampaignListRoutes() []string { return []string{"/crm/campaigns"} }

type (
	ListCampaignsHandler struct{ Service *application.Service }
	ListCampaignsRequest struct {
		Limit  int32 `json:"limit"`
		Offset int32 `json:"offset"`
	}
	ListCampaignsResponse struct {
		Campaigns []CampaignPayload `json:"campaigns"`
		Total     int64             `json:"total"`
		Limit     int32             `json:"limit"`
		Offset    int32             `json:"offset"`
	}
)

func (h *ListCampaignsHandler) Serve(ctx context.Context, req *ListCampaignsRequest) (ListCampaignsResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:view"); err != nil {
		return ListCampaignsResponse{}, err
	}
	device := authPkg.DeviceCtx(ctx)
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListCampaignsResponse{}, onekiterrors.PermissionDenied()
	}
	const maxCampaignPageSize int32 = 100
	if req.Limit <= 0 || req.Limit > maxCampaignPageSize {
		req.Limit = maxCampaignPageSize
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	page, err := h.Service.ListPaginated(ctx, businessID, req.Limit, req.Offset)
	if err != nil {
		return ListCampaignsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list campaigns")
	}
	payloads := make([]CampaignPayload, len(page.Campaigns))
	for i, c := range page.Campaigns {
		payloads[i] = newCampaignPayload(c)
	}
	return ListCampaignsResponse{
		Campaigns: payloads,
		Total:     page.Total,
		Limit:     page.Limit,
		Offset:    page.Offset,
	}, nil
}

// ─── Get By ID ────────────────────────────────────────────────────────────────

func CampaignGetRoutes() []string { return []string{"/crm/campaigns/{id}"} }

type (
	GetCampaignHandler  struct{ Service *application.Service }
	GetCampaignRequest  struct{}
	GetCampaignResponse struct {
		Campaign CampaignPayload `json:"campaign"`
	}
)

func (h *GetCampaignHandler) Serve(ctx context.Context, _ *GetCampaignRequest) (GetCampaignResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:view"); err != nil {
		return GetCampaignResponse{}, err
	}
	device := authPkg.DeviceCtx(ctx)
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return GetCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign ID")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	campaign, getErr := h.Service.GetByID(ctx, businessID, id)
	if getErr != nil {
		if errors.Is(getErr, domain.ErrCampaignNotFound) {
			return GetCampaignResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}
		return GetCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get campaign")
	}
	return GetCampaignResponse{Campaign: newCampaignPayload(campaign)}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func CampaignCreateRoutes() []string { return []string{"/crm/campaigns"} }

type CreateCampaignHandler struct{ Service *application.Service }

type CreateCampaignRequest struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	AudienceSegmentID *int64  `json:"audience_segment_id"`
	Subject           string  `json:"subject"`
	Body              string  `json:"body"`
	ScheduledAt       *string `json:"scheduled_at"`
}

type CreateCampaignResponse struct {
	CampaignID int64  `json:"campaign_id"`
	Message    string `json:"message"`
}

func (h *CreateCampaignHandler) Validate(_ context.Context, req *CreateCampaignRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.Type == "" {
		return onekiterrors.New(http.StatusBadRequest, "Type is required")
	}
	return nil
}

func (h *CreateCampaignHandler) Serve(ctx context.Context, req *CreateCampaignRequest) (CreateCampaignResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:create"); err != nil {
		return CreateCampaignResponse{}, err
	}
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.CampaignInput{
		OwnerID:           device.User.UserID,
		BusinessID:        businessID,
		Name:              req.Name,
		Type:              req.Type,
		AudienceSegmentID: req.AudienceSegmentID,
		Subject:           req.Subject,
		Body:              req.Body,
	}
	if req.ScheduledAt != nil {
		t, parseErr := time.Parse("2006-01-02T15:04:05Z", *req.ScheduledAt)
		if parseErr != nil {
			return CreateCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid scheduled_at format")
		}
		input.ScheduledAt = &t
	}
	id, err := h.Service.Create(ctx, input)
	if err != nil {
		return CreateCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create campaign")
	}
	return CreateCampaignResponse{CampaignID: id, Message: "Campaign created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func CampaignUpdateRoutes() []string { return []string{"/crm/campaigns/{id}"} }

type UpdateCampaignHandler struct{ Service *application.Service }

type UpdateCampaignRequest struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	AudienceSegmentID *int64  `json:"audience_segment_id"`
	Subject           string  `json:"subject"`
	Body              string  `json:"body"`
	ScheduledAt       *string `json:"scheduled_at"`
}

type UpdateCampaignResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCampaignHandler) Validate(_ context.Context, req *UpdateCampaignRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Name is required")
	}
	if req.Type == "" {
		return onekiterrors.New(http.StatusBadRequest, "Type is required")
	}
	return nil
}

func (h *UpdateCampaignHandler) Serve(ctx context.Context, req *UpdateCampaignRequest) (UpdateCampaignResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:create"); err != nil {
		return UpdateCampaignResponse{}, err
	}
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign ID")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.CampaignUpdateInput{
		ID:                id,
		OwnerID:           device.User.UserID,
		BusinessID:        businessID,
		Name:              req.Name,
		Type:              req.Type,
		AudienceSegmentID: req.AudienceSegmentID,
		Subject:           req.Subject,
		Body:              req.Body,
	}
	if req.ScheduledAt != nil {
		t, parseErr := time.Parse("2006-01-02T15:04:05Z", *req.ScheduledAt)
		if parseErr != nil {
			return UpdateCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid scheduled_at format")
		}
		input.ScheduledAt = &t
	}
	if updateErr := h.Service.Update(ctx, input); updateErr != nil {
		if errors.Is(updateErr, domain.ErrCampaignNotFound) {
			return UpdateCampaignResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}
		return UpdateCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update campaign")
	}
	return UpdateCampaignResponse{Message: "Campaign updated successfully"}, nil
}

// ─── Update Status ────────────────────────────────────────────────────────────

func CampaignStatusRoutes() []string { return []string{"/crm/campaigns/{id}/status"} }

type UpdateCampaignStatusHandler struct{ Service *application.Service }

type UpdateCampaignStatusRequest struct {
	Status string `json:"status"`
}

type UpdateCampaignStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCampaignStatusHandler) Validate(_ context.Context, req *UpdateCampaignStatusRequest) error {
	if req.Status == "" {
		return onekiterrors.New(http.StatusBadRequest, "Status is required")
	}
	return nil
}

func (h *UpdateCampaignStatusHandler) Serve(ctx context.Context, req *UpdateCampaignStatusRequest) (UpdateCampaignStatusResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpdateCampaignStatusResponse{}, onekiterrors.PermissionDenied()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign ID")
	}
	status := strings.TrimSpace(req.Status)
	switch status {
	case "draft", "scheduled", "active", "paused", "completed", "sent", "cancelled":
	default:
		return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign status")
	}
	permission := "crm:campaigns:create"
	if status == "active" || status == "sent" {
		permission = "crm:campaigns:send"
	}
	if err := requireCampaignPermission(ctx, permission); err != nil {
		return UpdateCampaignStatusResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateCampaignStatusResponse{}, onekiterrors.PermissionDenied()
	}
	if updateErr := h.Service.UpdateStatus(ctx, businessID, id, status); updateErr != nil {
		if errors.Is(updateErr, domain.ErrCampaignNotFound) {
			return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}
		if errors.Is(updateErr, domain.ErrInvalidCampaignStatus) {
			return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign status")
		}
		return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update campaign status")
	}
	return UpdateCampaignStatusResponse{Message: "Campaign status updated successfully"}, nil
}
