package handlers

import (
	"context"
	"errors"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
)

func CampaignStatsRoutes() []string {
	return []string{"/crm/campaigns/stats"}
}

func CampaignDetailFullRoutes() []string {
	return []string{"/crm/campaigns/detail"}
}

// -- Stats -------------------------------------------------------------------

type GetCampaignStatsHandler struct {
	CRMService *application.Service
}

type GetCampaignStatsRequest struct {
	CampaignID int64 `json:"campaignId"`
}

type CampaignStatsPayload struct {
	CampaignID     int64 `json:"campaignId"`
	AudienceCount  int   `json:"audienceCount"`
	SentCount      int   `json:"sentCount"`
	DeliveredCount int   `json:"deliveredCount"`
	OpenedCount    int   `json:"openedCount"`
	ClickedCount   int   `json:"clickedCount"`
	BouncedCount   int   `json:"bouncedCount"`
	FailedCount    int   `json:"failedCount"`
}

type GetCampaignStatsResponse struct {
	Stats CampaignStatsPayload `json:"stats"`
}

func (h *GetCampaignStatsHandler) Validate(_ context.Context, req *GetCampaignStatsRequest) *errorModels.ServerError {
	if req.CampaignID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign ID is required",
		}
	}

	return nil
}

func (h *GetCampaignStatsHandler) Serve(ctx context.Context, req *GetCampaignStatsRequest) (GetCampaignStatsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCampaignStatsResponse{}, errorModels.PermissionDeniedServerError()
	}

	stats, err := h.CRMService.GetCampaignStats(ctx, device.User.UserID, req.CampaignID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return GetCampaignStatsResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Campaign not found",
			}
		}

		return GetCampaignStatsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaign stats",
		}
	}

	return GetCampaignStatsResponse{
		Stats: CampaignStatsPayload{
			CampaignID:     stats.CampaignID,
			AudienceCount:  stats.AudienceCount,
			SentCount:      stats.SentCount,
			DeliveredCount: stats.DeliveredCount,
			OpenedCount:    stats.OpenedCount,
			ClickedCount:   stats.ClickedCount,
			BouncedCount:   stats.BouncedCount,
			FailedCount:    stats.FailedCount,
		},
	}, nil
}

// -- Detail ------------------------------------------------------------------

type GetCampaignDetailHandler struct {
	CRMService *application.Service
}

type GetCampaignDetailRequest struct {
	ID int64 `json:"id"`
}

type GetCampaignDetailResponse struct {
	Campaign CampaignPayload      `json:"campaign"`
	Stats    CampaignStatsPayload `json:"stats"`
}

func (h *GetCampaignDetailHandler) Validate(_ context.Context, req *GetCampaignDetailRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Campaign ID is required",
		}
	}

	return nil
}

func (h *GetCampaignDetailHandler) Serve(ctx context.Context, req *GetCampaignDetailRequest) (GetCampaignDetailResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCampaignDetailResponse{}, errorModels.PermissionDeniedServerError()
	}

	detail, err := h.CRMService.GetCampaignDetail(ctx, device.User.UserID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return GetCampaignDetailResponse{}, &errorModels.ServerError{
				Code:    http.StatusNotFound,
				Message: "Campaign not found",
			}
		}

		return GetCampaignDetailResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get campaign detail",
		}
	}

	return GetCampaignDetailResponse{
		Campaign: newCampaignPayload(detail.Campaign),
		Stats: CampaignStatsPayload{
			CampaignID:     detail.Stats.CampaignID,
			AudienceCount:  detail.Stats.AudienceCount,
			SentCount:      detail.Stats.SentCount,
			DeliveredCount: detail.Stats.DeliveredCount,
			OpenedCount:    detail.Stats.OpenedCount,
			ClickedCount:   detail.Stats.ClickedCount,
			BouncedCount:   detail.Stats.BouncedCount,
			FailedCount:    detail.Stats.FailedCount,
		},
	}, nil
}
