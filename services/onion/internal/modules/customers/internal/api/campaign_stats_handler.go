package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/customers/internal/application"
	"github.com/voxire/onion/internal/modules/customers/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcampaignv1 "github.com/voxire/voxie/gen/go/crm/campaign/v1"
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
	CampaignID int64 `json:"campaign_id"`
}

func newCampaignStatsPayload(s domain.CampaignStats) *crmcampaignv1.CampaignStats {
	return &crmcampaignv1.CampaignStats{
		CampaignId:     s.CampaignID,
		AudienceCount:  int64(s.AudienceCount),
		SentCount:      int64(s.SentCount),
		DeliveredCount: int64(s.DeliveredCount),
		OpenedCount:    int64(s.OpenedCount),
		ClickedCount:   int64(s.ClickedCount),
		BouncedCount:   int64(s.BouncedCount),
		FailedCount:    int64(s.FailedCount),
	}
}

type GetCampaignStatsResponse struct {
	Stats *crmcampaignv1.CampaignStats `json:"stats"`
}

func (h *GetCampaignStatsHandler) Validate(_ context.Context, req *GetCampaignStatsRequest) error {
	if req.CampaignID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Campaign ID is required")
	}

	return nil
}

func (h *GetCampaignStatsHandler) Serve(ctx context.Context, req *GetCampaignStatsRequest) (GetCampaignStatsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCampaignStatsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCampaignStatsResponse{}, onekiterrors.PermissionDenied()
	}

	stats, err := h.CRMService.GetCampaignStats(ctx, businessID, req.CampaignID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return GetCampaignStatsResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}

		return GetCampaignStatsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get campaign stats")
	}

	return GetCampaignStatsResponse{
		Stats: newCampaignStatsPayload(stats),
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
	Campaign *crmcampaignv1.Campaign      `json:"campaign"`
	Stats    *crmcampaignv1.CampaignStats `json:"stats"`
}

func (h *GetCampaignDetailHandler) Validate(_ context.Context, req *GetCampaignDetailRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "Campaign ID is required")
	}

	return nil
}

func (h *GetCampaignDetailHandler) Serve(ctx context.Context, req *GetCampaignDetailRequest) (GetCampaignDetailResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCampaignDetailResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCampaignDetailResponse{}, onekiterrors.PermissionDenied()
	}

	detail, err := h.CRMService.GetCampaignDetail(ctx, businessID, req.ID)
	if err != nil {
		if errors.Is(err, domain.ErrCampaignNotFound) {
			return GetCampaignDetailResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}

		return GetCampaignDetailResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get campaign detail")
	}

	return GetCampaignDetailResponse{
		Campaign: newCampaignPayload(detail.Campaign),
		Stats:    newCampaignStatsPayload(detail.Stats),
	}, nil
}
