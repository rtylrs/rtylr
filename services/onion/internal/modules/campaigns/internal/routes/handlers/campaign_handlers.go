package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
)

type CampaignPayload struct {
	ID                int64   `json:"id"`
	OwnerID           int64   `json:"ownerId"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Status            string  `json:"status"`
	AudienceSegmentID *int64  `json:"audienceSegmentId"`
	Subject           string  `json:"subject"`
	Body              string  `json:"body"`
	SentCount         int     `json:"sentCount"`
	OpenCount         int     `json:"openCount"`
	ScheduledAt       *string `json:"scheduledAt"`
	SentAt            *string `json:"sentAt"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
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

type ListCampaignsHandler struct{ Service *application.Service }
type ListCampaignsRequest struct{}
type ListCampaignsResponse struct {
	Campaigns []CampaignPayload `json:"campaigns"`
}

func (h *ListCampaignsHandler) Serve(ctx context.Context, _ *ListCampaignsRequest) (ListCampaignsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListCampaignsResponse{}, errorModels.PermissionDeniedServerError()
	}
	campaigns, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListCampaignsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list campaigns"}
	}
	payloads := make([]CampaignPayload, len(campaigns))
	for i, c := range campaigns {
		payloads[i] = newCampaignPayload(c)
	}
	return ListCampaignsResponse{Campaigns: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func CampaignCreateRoutes() []string { return []string{"/crm/campaigns"} }

type CreateCampaignHandler struct{ Service *application.Service }

type CreateCampaignRequest struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	AudienceSegmentID *int64  `json:"audienceSegmentId"`
	Subject           string  `json:"subject"`
	Body              string  `json:"body"`
	ScheduledAt       *string `json:"scheduledAt"`
}

type CreateCampaignResponse struct {
	CampaignID int64  `json:"campaignId"`
	Message    string `json:"message"`
}

func (h *CreateCampaignHandler) Validate(_ context.Context, req *CreateCampaignRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateCampaignHandler) Serve(ctx context.Context, req *CreateCampaignRequest) (CreateCampaignResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateCampaignResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.CampaignInput{
		OwnerID:           device.User.UserID,
		Name:              req.Name,
		Type:              req.Type,
		AudienceSegmentID: req.AudienceSegmentID,
		Subject:           req.Subject,
		Body:              req.Body,
	}
	if req.ScheduledAt != nil {
		t, err := time.Parse("2006-01-02T15:04:05Z", *req.ScheduledAt)
		if err == nil {
			input.ScheduledAt = &t
		}
	}
	id, err := h.Service.Create(ctx, input)
	if err != nil {
		return CreateCampaignResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create campaign"}
	}
	return CreateCampaignResponse{CampaignID: id, Message: "Campaign created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func CampaignUpdateRoutes() []string { return []string{"/crm/campaigns/{id}"} }

type UpdateCampaignHandler struct{ Service *application.Service }

type UpdateCampaignRequest struct {
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	AudienceSegmentID *int64  `json:"audienceSegmentId"`
	Subject           string  `json:"subject"`
	Body              string  `json:"body"`
	ScheduledAt       *string `json:"scheduledAt"`
}

type UpdateCampaignResponse struct {
	Message string `json:"message"`
}

func (h *UpdateCampaignHandler) Validate(_ context.Context, req *UpdateCampaignRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateCampaignHandler) Serve(ctx context.Context, req *UpdateCampaignRequest) (UpdateCampaignResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateCampaignResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid campaign ID"}
	}
	input := domain.CampaignUpdateInput{
		ID:                id,
		OwnerID:           device.User.UserID,
		Name:              req.Name,
		Type:              req.Type,
		AudienceSegmentID: req.AudienceSegmentID,
		Subject:           req.Subject,
		Body:              req.Body,
	}
	if req.ScheduledAt != nil {
		t, parseErr := time.Parse("2006-01-02T15:04:05Z", *req.ScheduledAt)
		if parseErr == nil {
			input.ScheduledAt = &t
		}
	}
	if updateErr := h.Service.Update(ctx, input); updateErr != nil {
		return UpdateCampaignResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update campaign"}
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

func (h *UpdateCampaignStatusHandler) Validate(_ context.Context, req *UpdateCampaignStatusRequest) *errorModels.ServerError {
	if req.Status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Status is required"}
	}
	return nil
}

func (h *UpdateCampaignStatusHandler) Serve(ctx context.Context, req *UpdateCampaignStatusRequest) (UpdateCampaignStatusResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateCampaignStatusResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateCampaignStatusResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid campaign ID"}
	}
	if updateErr := h.Service.UpdateStatus(ctx, device.User.UserID, id, req.Status); updateErr != nil {
		return UpdateCampaignStatusResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update campaign status"}
	}
	return UpdateCampaignStatusResponse{Message: "Campaign status updated successfully"}, nil
}
