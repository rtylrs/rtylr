package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	crmcampaignv1 "github.com/voxire/voxie/gen/go/crm/campaign/v1"
)

type campaignService interface {
	ListByStatus(context.Context, int64, string) ([]domain.Campaign, error)
	ListPaginated(context.Context, int64, int32, int32) (domain.CampaignPage, error)
	GetByID(context.Context, int64, int64) (domain.Campaign, error)
	Create(context.Context, domain.CampaignInput) (int64, error)
	Update(context.Context, domain.CampaignUpdateInput) error
	UpdateStatus(context.Context, int64, int64, string) error
}

// CampaignAPI adapts the campaign application service to the generated CRM
// campaign/v1 contract. The legacy /crm/campaigns routes remain registered for
// older clients, while this adapter serves the contract used by the frontend.
type CampaignAPI struct {
	service campaignService
}

func NewCampaignAPI(service *application.Service) *CampaignAPI {
	return &CampaignAPI{service: service}
}

func (a *CampaignAPI) ListCampaigns(ctx context.Context, req *crmcampaignv1.ListCampaignsRequest) (*crmcampaignv1.ListCampaignsResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:view"); err != nil {
		return nil, err
	}
	businessID, _, err := resolveCampaignIdentity(ctx)
	if err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	if status := strings.TrimSpace(req.GetStatus()); status != "" {
		items, listErr := a.service.ListByStatus(ctx, businessID, status)
		if listErr != nil {
			return nil, campaignAPIError(listErr, "Failed to list campaigns")
		}

		start := int(offset)
		if start > len(items) {
			start = len(items)
		}
		end := start + int(limit)
		if end > len(items) {
			end = len(items)
		}

		payloads := make([]*crmcampaignv1.Campaign, 0, end-start)
		for _, item := range items[start:end] {
			payloads = append(payloads, newCampaignProto(item))
		}
		return &crmcampaignv1.ListCampaignsResponse{
			Campaigns: payloads,
			Total:     int64(len(items)),
			Limit:     limit,
			Offset:    offset,
		}, nil
	}

	page, err := a.service.ListPaginated(ctx, businessID, int32(limit), int32(offset))
	if err != nil {
		return nil, campaignAPIError(err, "Failed to list campaigns")
	}
	payloads := make([]*crmcampaignv1.Campaign, 0, len(page.Campaigns))
	for _, item := range page.Campaigns {
		payloads = append(payloads, newCampaignProto(item))
	}
	return &crmcampaignv1.ListCampaignsResponse{
		Campaigns: payloads,
		Total:     page.Total,
		Limit:     int64(page.Limit),
		Offset:    int64(page.Offset),
	}, nil
}

func (a *CampaignAPI) CreateCampaign(ctx context.Context, req *crmcampaignv1.CreateCampaignRequest) (*crmcampaignv1.CreateCampaignResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:create"); err != nil {
		return nil, err
	}
	if req == nil || strings.TrimSpace(req.GetName()) == "" || strings.TrimSpace(req.GetChannel()) == "" || strings.TrimSpace(req.GetContent()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name, channel, and content are required")
	}
	businessID, ownerID, err := resolveCampaignIdentity(ctx)
	if err != nil {
		return nil, err
	}
	scheduledAt, err := parseCampaignTime(req.GetScheduledAt())
	if err != nil {
		return nil, err
	}

	id, err := a.service.Create(ctx, domain.CampaignInput{
		OwnerID:           ownerID,
		BusinessID:        businessID,
		Name:              req.GetName(),
		Description:       req.GetDescription(),
		Type:              req.GetChannel(),
		AudienceSegmentID: req.SegmentId,
		Subject:           req.GetSubject(),
		Body:              req.GetContent(),
		ScheduledAt:       scheduledAt,
	})
	if err != nil {
		return nil, campaignAPIError(err, "Failed to create campaign")
	}
	return &crmcampaignv1.CreateCampaignResponse{CampaignId: id, Message: "Campaign created successfully"}, nil
}

func (a *CampaignAPI) GetCampaignDetail(ctx context.Context, req *crmcampaignv1.GetCampaignDetailRequest) (*crmcampaignv1.GetCampaignDetailResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:view"); err != nil {
		return nil, err
	}
	if req == nil || req.GetId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "id must be greater than 0")
	}
	businessID, _, err := resolveCampaignIdentity(ctx)
	if err != nil {
		return nil, err
	}
	campaign, err := a.service.GetByID(ctx, businessID, req.GetId())
	if err != nil {
		return nil, campaignAPIError(err, "Failed to get campaign")
	}
	return &crmcampaignv1.GetCampaignDetailResponse{
		Campaign: newCampaignProto(campaign),
		Stats:    newCampaignStatsProto(campaign),
	}, nil
}

func (a *CampaignAPI) GetCampaignStats(ctx context.Context, req *crmcampaignv1.GetCampaignStatsRequest) (*crmcampaignv1.GetCampaignStatsResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:view"); err != nil {
		return nil, err
	}
	if req == nil || req.GetCampaignId() <= 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "campaign_id must be greater than 0")
	}
	businessID, _, err := resolveCampaignIdentity(ctx)
	if err != nil {
		return nil, err
	}
	campaign, err := a.service.GetByID(ctx, businessID, req.GetCampaignId())
	if err != nil {
		return nil, campaignAPIError(err, "Failed to get campaign stats")
	}
	return &crmcampaignv1.GetCampaignStatsResponse{Stats: newCampaignStatsProto(campaign)}, nil
}

func (a *CampaignAPI) UpdateCampaign(ctx context.Context, req *crmcampaignv1.UpdateCampaignRequest) (*crmcampaignv1.UpdateCampaignResponse, error) {
	if err := requireCampaignPermission(ctx, "crm:campaigns:create"); err != nil {
		return nil, err
	}
	if req == nil || req.GetId() <= 0 || req.GetBody() == nil || strings.TrimSpace(req.GetBody().GetName()) == "" || strings.TrimSpace(req.GetBody().GetChannel()) == "" || strings.TrimSpace(req.GetBody().GetContent()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "name, channel, and content are required")
	}
	businessID, ownerID, err := resolveCampaignIdentity(ctx)
	if err != nil {
		return nil, err
	}
	body := req.GetBody()
	scheduledAt, err := parseCampaignTime(body.GetScheduledAt())
	if err != nil {
		return nil, err
	}
	if err := a.service.Update(ctx, domain.CampaignUpdateInput{
		ID:                req.GetId(),
		OwnerID:           ownerID,
		BusinessID:        businessID,
		Name:              body.GetName(),
		Description:       body.GetDescription(),
		Type:              body.GetChannel(),
		AudienceSegmentID: body.SegmentId,
		Subject:           body.GetSubject(),
		Body:              body.GetContent(),
		ScheduledAt:       scheduledAt,
	}); err != nil {
		return nil, campaignAPIError(err, "Failed to update campaign")
	}
	return &crmcampaignv1.UpdateCampaignResponse{Message: "Campaign updated successfully"}, nil
}

func (a *CampaignAPI) UpdateCampaignStatus(ctx context.Context, req *crmcampaignv1.UpdateCampaignStatusRequest) (*crmcampaignv1.UpdateCampaignStatusResponse, error) {
	if req == nil || req.GetId() <= 0 || req.GetBody() == nil || strings.TrimSpace(req.GetBody().GetStatus()) == "" {
		return nil, onekiterrors.New(http.StatusBadRequest, "status is required")
	}
	status := strings.TrimSpace(req.GetBody().GetStatus())
	permission := "crm:campaigns:create"
	if status == domain.StatusActive || status == domain.StatusSent {
		permission = "crm:campaigns:send"
	}
	if err := requireCampaignPermission(ctx, permission); err != nil {
		return nil, err
	}
	businessID, _, err := resolveCampaignIdentity(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.service.UpdateStatus(ctx, businessID, req.GetId(), status); err != nil {
		return nil, campaignAPIError(err, "Failed to update campaign status")
	}
	return &crmcampaignv1.UpdateCampaignStatusResponse{Message: "Campaign status updated successfully"}, nil
}

func resolveCampaignIdentity(ctx context.Context) (businessID, ownerID int64, err error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	return businessID, device.User.UserID, nil
}

func parseCampaignTime(value string) (*time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid scheduled_at format")
	}
	return &parsed, nil
}

func campaignAPIError(err error, fallback string) error {
	switch {
	case errors.Is(err, domain.ErrCampaignNotFound):
		return onekiterrors.New(http.StatusNotFound, "Campaign not found")
	case errors.Is(err, domain.ErrInvalidCampaignStatus),
		errors.Is(err, domain.ErrEmptyCampaignName),
		errors.Is(err, domain.ErrEmptyCampaignType),
		errors.Is(err, domain.ErrInvalidCampaignID),
		errors.Is(err, domain.ErrInvalidCampaignOwner):
		return onekiterrors.New(http.StatusBadRequest, "Invalid campaign")
	default:
		return onekiterrors.New(http.StatusInternalServerError, fallback)
	}
}

func newCampaignProto(campaign domain.Campaign) *crmcampaignv1.Campaign {
	payload := &crmcampaignv1.Campaign{
		Id:             campaign.ID,
		OwnerId:        campaign.OwnerID,
		Name:           campaign.Name,
		Description:    campaign.Description,
		Channel:        campaign.Type,
		Status:         campaign.Status,
		Subject:        campaign.Subject,
		Content:        campaign.Body,
		AudienceCount:  int64(campaign.AudienceCount),
		SentCount:      int64(campaign.SentCount),
		DeliveredCount: int64(campaign.DeliveredCount),
		OpenedCount:    int64(campaign.OpenCount),
		ClickedCount:   int64(campaign.ClickedCount),
		CreatedAt:      formatCampaignTime(campaign.CreatedAt),
		UpdatedAt:      formatCampaignTime(campaign.UpdatedAt),
	}
	if campaign.AudienceSegmentID != nil {
		segmentID := *campaign.AudienceSegmentID
		payload.SegmentId = &segmentID
	}
	if campaign.ScheduledAt != nil {
		scheduledAt := formatCampaignTime(*campaign.ScheduledAt)
		payload.ScheduledAt = &scheduledAt
	}
	if campaign.SentAt != nil {
		sentAt := formatCampaignTime(*campaign.SentAt)
		payload.SentAt = &sentAt
	}
	return payload
}

func newCampaignStatsProto(campaign domain.Campaign) *crmcampaignv1.CampaignStats {
	return &crmcampaignv1.CampaignStats{
		CampaignId:     campaign.ID,
		AudienceCount:  int64(campaign.AudienceCount),
		SentCount:      int64(campaign.SentCount),
		DeliveredCount: int64(campaign.DeliveredCount),
		OpenedCount:    int64(campaign.OpenCount),
		ClickedCount:   int64(campaign.ClickedCount),
	}
}

func formatCampaignTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}
