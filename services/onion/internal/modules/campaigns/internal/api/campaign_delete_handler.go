package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/campaigns/internal/application"
	"github.com/voxire/onion/internal/modules/campaigns/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func CampaignDeleteRoutes() []string { return []string{"/crm/campaigns/{id}"} }

type (
	DeleteCampaignHandler  struct{ Service *application.Service }
	DeleteCampaignRequest  struct{}
	DeleteCampaignResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteCampaignHandler) Serve(ctx context.Context, _ *DeleteCampaignRequest) (DeleteCampaignResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireCampaignPermission(ctx, "crm:campaigns:create"); err != nil {
		return DeleteCampaignResponse{}, err
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return DeleteCampaignResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid campaign ID")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return DeleteCampaignResponse{}, onekiterrors.PermissionDenied()
	}
	if deleteErr := h.Service.Delete(ctx, businessID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrCampaignNotFound) {
			return DeleteCampaignResponse{}, onekiterrors.New(http.StatusNotFound, "Campaign not found")
		}
		return DeleteCampaignResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to delete campaign")
	}
	return DeleteCampaignResponse{Message: "Campaign deleted successfully"}, nil
}
