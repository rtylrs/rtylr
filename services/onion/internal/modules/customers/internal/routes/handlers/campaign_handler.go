package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func CampaignListRoutes() []string {
	return []string{"/crm/campaigns"}
}

func CampaignCreateRoutes() []string {
	return []string{"/crm/campaigns"}
}

func CampaignUpdateRoutes() []string {
	return []string{"/crm/campaigns"}
}

func CampaignStatusRoutes() []string {
	return []string{"/crm/campaigns/status"}
}

type CampaignPayload struct {
	ID             int64   `json:"id"`
	OwnerID        int64   `json:"ownerId"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Channel        string  `json:"channel"`
	Status         string  `json:"status"`
	SegmentID      *int64  `json:"segmentId"`
	Subject        string  `json:"subject"`
	Content        string  `json:"content"`
	ScheduledAt    *string `json:"scheduledAt"`
	SentAt         *string `json:"sentAt"`
	AudienceCount  int     `json:"audienceCount"`
	SentCount      int     `json:"sentCount"`
	DeliveredCount int     `json:"deliveredCount"`
	OpenedCount    int     `json:"openedCount"`
	ClickedCount   int     `json:"clickedCount"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

func newCampaignPayload(c domain.Campaign) CampaignPayload {
	p := CampaignPayload{
		ID:             c.ID,
		OwnerID:        c.OwnerID,
		Name:           c.Name,
		Description:    c.Description,
		Channel:        c.Channel,
		Status:         c.Status,
		SegmentID:      c.SegmentID,
		Subject:        c.Subject,
		Content:        c.Content,
		AudienceCount:  c.AudienceCount,
		SentCount:      c.SentCount,
		DeliveredCount: c.DeliveredCount,
		OpenedCount:    c.OpenedCount,
		ClickedCount:   c.ClickedCount,
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
	Campaigns []CampaignPayload `json:"campaigns"`
}

func (h *ListCampaignsHandler) Serve(ctx context.Context, req *ListCampaignsRequest) (ListCampaignsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCampaignsResponse{}, errorModels.PermissionDeniedServerError()
	}

	campaigns, err := h.CRMService.ListCampaigns(ctx, device.User.UserID, req.Status)
	if err != nil {
		return ListCampaignsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list campaigns",
		}
	}

	payloads := make([]CampaignPayload, len(campaigns))
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
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Channel     string  `json:"channel"`
	SegmentID   *int64  `json:"segmentId"`
	Subject     string  `json:"subject"`
	Content     string  `json:"content"`
	ScheduledAt string  `json:"scheduledAt"`
}

type CreateCampaignResponse struct {
	CampaignID int64  `json:"campaignId"`
	Message    string `json:"message"`
}

func (h *CreateCampaignHandler) Validate(_ context.Context, req *CreateCampaignRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign name is required",
		}
	}

	if req.Channel == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign channel is required",
		}
	}

	if req.Content == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign content is required",
		}
	}

	return nil
}

func (h *CreateCampaignHandler) Serve(ctx context.Context, req *CreateCampaignRequest) (CreateCampaignResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCampaignResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.CampaignInput{
		OwnerID:     device.User.UserID,
		Name:        req.Name,
		Description: req.Description,
		Channel:     req.Channel,
		SegmentID:   req.SegmentID,
		Subject:     req.Subject,
		Content:     req.Content,
	}

	if req.ScheduledAt != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", req.ScheduledAt)
		if err == nil {
			input.ScheduledAt = &t
		}
	}

	id, err := h.CRMService.CreateCampaign(ctx, input)
	if err != nil {
		return CreateCampaignResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create campaign",
		}
	}

	return CreateCampaignResponse{CampaignID: id, Message: "Campaign created successfully"}, nil
}

// ─── UPDATE ───────────────────────────────────────────────────────────────────

type UpdateCampaignHandler struct {
	CRMService *application.Service
}

type UpdateCampaignRequest struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Channel     string  `json:"channel"`
	SegmentID   *int64  `json:"segmentId"`
	Subject     string  `json:"subject"`
	Content     string  `json:"content"`
	ScheduledAt string  `json:"scheduledAt"`
}

type UpdateCampaignResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCampaignHandler) Validate(_ context.Context, req *UpdateCampaignRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign ID is required",
		}
	}

	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign name is required",
		}
	}

	if req.Content == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign content is required",
		}
	}

	return nil
}

func (h *UpdateCampaignHandler) Serve(ctx context.Context, req *UpdateCampaignRequest) (UpdateCampaignResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.CampaignUpdateInput{
		ID:          req.ID,
		OwnerID:     device.User.UserID,
		Name:        req.Name,
		Description: req.Description,
		Channel:     req.Channel,
		SegmentID:   req.SegmentID,
		Subject:     req.Subject,
		Content:     req.Content,
	}

	if req.ScheduledAt != "" {
		t, err := time.Parse("2006-01-02T15:04:05Z", req.ScheduledAt)
		if err == nil {
			input.ScheduledAt = &t
		}
	}

	if err := h.CRMService.UpdateCampaign(ctx, input); err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return UpdateCampaignResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Campaign not found",
			}
		}

		return UpdateCampaignResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update campaign",
		}
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

func (h *UpdateCampaignStatusHandler) Validate(_ context.Context, req *UpdateCampaignStatusRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign ID is required",
		}
	}

	validStatuses := map[string]bool{
		"draft": true, "scheduled": true, "active": true,
		"paused": true, "completed": true, "cancelled": true,
	}

	if !validStatuses[req.Status] {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid campaign status",
		}
	}

	return nil
}

func (h *UpdateCampaignStatusHandler) Serve(ctx context.Context, req *UpdateCampaignStatusRequest) (UpdateCampaignStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.CRMService.UpdateCampaignStatus(ctx, device.User.UserID, req.ID, req.Status); err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return UpdateCampaignStatusResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Campaign not found",
			}
		}

		return UpdateCampaignStatusResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update campaign status",
		}
	}

	return UpdateCampaignStatusResponse{Message: "Campaign status updated successfully"}, nil
}
