package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcampaignv1 "github.com/voxire/voxie/gen/go/crm/campaign/v1"
)





func newCampaignPayload(c domain.Campaign) *crmcampaignv1.Campaign {
	p := &crmcampaignv1.Campaign{
		Id:             c.ID,
		OwnerId:        c.OwnerID,
		Name:           c.Name,
		Description:    c.Description,
		Channel:        c.Channel,
		Status:         c.Status,
		SegmentId:      c.SegmentID,
		Subject:        c.Subject,
		Content:        c.Content,
		AudienceCount:  int64(c.AudienceCount),
		SentCount:      int64(c.SentCount),
		DeliveredCount: int64(c.DeliveredCount),
		OpenedCount:    int64(c.OpenedCount),
		ClickedCount:   int64(c.ClickedCount),
		CreatedAt:      c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

// ─── LIST ─────────────────────────────────────────────────────────────────────

type ListCampaignsHandler struct {
	CRMService *application.Service
}

type ListCampaignsRequest struct {
	Status string `json:"status"`
}

type ListCampaignsResponse struct {
	Campaigns []*crmcampaignv1.Campaign `json:"campaigns"`
}

func (h *ListCampaignsHandler) Serve(ctx context.Context, req *ListCampaignsRequest) (ListCampaignsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCampaignsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListCampaignsResponse{}, onekiterrors.PermissionDenied()
	}

	campaigns, err := h.CRMService.ListCampaigns(ctx, businessID, req.Status)
	if err != nil {
		return ListCampaignsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list campaigns")
	}

	payloads := make([]*crmcampaignv1.Campaign, len(campaigns))
	for i, c := range campaigns {
		payloads[i] = newCampaignPayload(c)
	}

	return ListCampaignsResponse{Campaigns: payloads}, nil
}

// ─── CREATE ───────────────────────────────────────────────────────────────────

type CreateCampaignHandler struct {
	CRMService *application.Service
}

type CreateCampaignRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Channel     string `json:"channel"`
	SegmentID   *int64 `json:"segment_id"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	ScheduledAt string `json:"scheduled_at"`
}

type CreateCampaignResponse struct {
	CampaignID int64  `json:"campaign_id"`
	Message    string `json:"message"`
}

func (h *CreateCampaignHandler) Validate(_ context.Context, req *CreateCampaignRequest) error {
	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Campaign name is required")
	}

	if req.Channel == "" {
		return onekiterrors.New(http.StatusBadRequest, "Campaign channel is required")
	}

	if req.Content == "" {
		return onekiterrors.New(http.StatusBadRequest, "Campaign content is required")
	}

	return nil
}

func (h *CreateCampaignHandler) Serve(ctx context.Context, req *CreateCampaignRequest) (CreateCampaignResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCampaignResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return CreateCampaignResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.CampaignInput{
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Channel:     req.Channel,
		SegmentID:   req.SegmentID,
		Subject:     req.Subject,
		Content:     req.Content,
	}

	if req.ScheduledAt != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", req.ScheduledAt)
		if err != nil {
			return CreateCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid scheduled_at format, expected RFC3339 (YYYY-MM-DDTHH:MM:SSZ)")
		}
		input.ScheduledAt = &t
	}

	id, err := h.CRMService.CreateCampaign(ctx, input)
	if err != nil {
		return CreateCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to create campaign")
	}

	return CreateCampaignResponse{CampaignID: id, Message: "Campaign created successfully"}, nil
}

// ─── UPDATE ───────────────────────────────────────────────────────────────────

type UpdateCampaignHandler struct {
	CRMService *application.Service
}

type UpdateCampaignRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Channel     string `json:"channel"`
	SegmentID   *int64 `json:"segment_id"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	ScheduledAt string `json:"scheduled_at"`
}

type UpdateCampaignResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCampaignHandler) Validate(_ context.Context, req *UpdateCampaignRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Campaign ID is required")
	}

	if req.Name == "" {
		return onekiterrors.New(http.StatusBadRequest, "Campaign name is required")
	}

	if req.Content == "" {
		return onekiterrors.New(http.StatusBadRequest, "Campaign content is required")
	}

	return nil
}

func (h *UpdateCampaignHandler) Serve(ctx context.Context, req *UpdateCampaignRequest) (UpdateCampaignResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateCampaignResponse{}, onekiterrors.PermissionDenied()
	}

	input := domain.CampaignUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		BusinessID:  businessID,
		Name:        req.Name,
		Description: req.Description,
		Channel:     req.Channel,
		SegmentID:   req.SegmentID,
		Subject:     req.Subject,
		Content:     req.Content,
	}

	if req.ScheduledAt != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", req.ScheduledAt)
		if err != nil {
			return UpdateCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid scheduled_at format, expected RFC3339 (YYYY-MM-DDTHH:MM:SSZ)")
		}
		input.ScheduledAt = &t
	}

	if err := h.CRMService.UpdateCampaign(ctx, input); err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return UpdateCampaignResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}

		return UpdateCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update campaign")
	}

	return UpdateCampaignResponse{Message: "Campaign updated successfully"}, nil
}

// ─── UPDATE STATUS ────────────────────────────────────────────────────────────

type UpdateCampaignStatusHandler struct {
	CRMService *application.Service
}

type UpdateCampaignStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdateCampaignStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCampaignStatusHandler) Validate(_ context.Context, req *UpdateCampaignStatusRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Campaign ID is required")
	}

	validStatuses := map[string]bool{
		"draft": true, "scheduled": true, "active": true,
		"paused": true, "completed": true, "cancelled": true,
	}

	if !validStatuses[req.Status] {
		return onekiterrors.New(http.StatusBadRequest, "Invalid campaign status")
	}

	return nil
}

func (h *UpdateCampaignStatusHandler) Serve(ctx context.Context, req *UpdateCampaignStatusRequest) (UpdateCampaignStatusResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignStatusResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpdateCampaignStatusResponse{}, onekiterrors.PermissionDenied()
	}

	if err := h.CRMService.UpdateCampaignStatus(ctx, businessID, req.ID, req.Status); err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}

		return UpdateCampaignStatusResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update campaign status")
	}

	return UpdateCampaignStatusResponse{Message: "Campaign status updated successfully"}, nil
}
