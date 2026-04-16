package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_performance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_performance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func UpsertConfigRoutes() []string { return []string{"/hr/performance/config"} }

type UpsertConfigHandler struct{ Service *application.Service }

type UpsertConfigRequest struct {
	RatingScaleType     string `json:"rating_scale_type"`
	RatingMin           int32  `json:"rating_min"`
	RatingMax           int32  `json:"rating_max"`
	RatingLabels        string `json:"rating_labels"`
	PeerFeedbackEnabled bool   `json:"peer_feedback_enabled"`
}

type UpsertConfigResponse struct {
	Config *domain.PerformanceConfigRecord `json:"config"`
}

func (h *UpsertConfigHandler) Validate(_ context.Context, req *UpsertConfigRequest) *errorModels.ServerError {
	if req.RatingScaleType != "" && req.RatingScaleType != "numeric" && req.RatingScaleType != "descriptive" {
		return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: "rating_scale_type must be 'numeric' or 'descriptive'"}
	}
	return nil
}

func (h *UpsertConfigHandler) Serve(ctx context.Context, req *UpsertConfigRequest) (UpsertConfigResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertConfigResponse{}, errorModels.PermissionDeniedServerError()
	}

	cfg, err := h.Service.UpsertPerformanceConfig(ctx, device.User.UserID, domain.UpsertConfigParams{
		RatingScaleType:     req.RatingScaleType,
		RatingMin:           req.RatingMin,
		RatingMax:           req.RatingMax,
		RatingLabels:        req.RatingLabels,
		PeerFeedbackEnabled: req.PeerFeedbackEnabled,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert performance config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UpsertConfigResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to update performance config"}
	}

	return UpsertConfigResponse{Config: cfg}, nil
}
