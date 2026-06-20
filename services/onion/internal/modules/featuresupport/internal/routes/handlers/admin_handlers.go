package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/featuresupport/internal/application"
	"github.com/voxire/onion/internal/modules/featuresupport/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// requireAdmin gates flag/target management to elevated roles. Feature flags
// are global platform config, so only owner/admin may change them.
func requireAdmin(ctx context.Context) *errorModels.ServerError {
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil {
		return errorModels.PermissionDeniedServerError()
	}
	role := strings.ToLower(user.Role)
	if role != "owner" && role != "admin" && role != "console_admin" {
		return errorModels.PermissionDeniedServerError()
	}
	return nil
}

type FlagPayload struct {
	Key               string `json:"key"`
	Description       string `json:"description"`
	Enabled           bool   `json:"enabled"`
	RolloutPercentage int32  `json:"rollout_percentage"`
	VariantOn         string `json:"variant_on"`
	VariantOff        string `json:"variant_off"`
}

type TargetPayload struct {
	FeatureKey string `json:"feature_key"`
	OwnerID    int64  `json:"owner_id"`
	Enabled    bool   `json:"enabled"`
	Variant    string `json:"variant"`
}

func badRequest(msg string) *errorModels.ServerError {
	return &errorModels.ServerError{HTTPCode: http.StatusBadRequest, Code: http.StatusBadRequest, Message: msg}
}

func internalError(msg string) *errorModels.ServerError {
	return &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: msg}
}

// -- GET /features/admin/flags ----------------------------------------------

func ListFlagsRoutes() []string { return []string{"/features/admin/flags"} }

type ListFlagsHandler struct{ Service *application.Service }

type ListFlagsRequest struct{}

type ListFlagsResponse struct {
	Flags []FlagPayload `json:"flags"`
}

func (h *ListFlagsHandler) Serve(ctx context.Context, _ *ListFlagsRequest) (ListFlagsResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return ListFlagsResponse{}, err
	}
	flags, err := h.Service.ListFlags(ctx)
	if err != nil {
		return ListFlagsResponse{}, internalError("Failed to list feature flags")
	}
	payloads := make([]FlagPayload, 0, len(flags))
	for _, f := range flags {
		payloads = append(payloads, FlagPayload{
			Key:               f.Key,
			Description:       f.Description,
			Enabled:           f.Enabled,
			RolloutPercentage: f.RolloutPercentage,
			VariantOn:         f.VariantOn,
			VariantOff:        f.VariantOff,
		})
	}
	return ListFlagsResponse{Flags: payloads}, nil
}

// -- PUT /features/admin/flags (create or update) ---------------------------

func UpsertFlagRoutes() []string { return []string{"/features/admin/flags"} }

type UpsertFlagHandler struct{ Service *application.Service }

type UpsertFlagRequest struct {
	Key               string `json:"key"`
	Description       string `json:"description"`
	Enabled           bool   `json:"enabled"`
	RolloutPercentage int32  `json:"rollout_percentage"`
	VariantOn         string `json:"variant_on"`
	VariantOff        string `json:"variant_off"`
}

type UpsertFlagResponse struct {
	Message string `json:"message"`
}

func (h *UpsertFlagHandler) Validate(_ context.Context, req *UpsertFlagRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Key) == "" {
		return badRequest("Feature key is required")
	}
	return nil
}

func (h *UpsertFlagHandler) Serve(ctx context.Context, req *UpsertFlagRequest) (UpsertFlagResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return UpsertFlagResponse{}, err
	}
	if err := h.Service.UpsertFlag(ctx, domain.Flag{
		Key:               strings.TrimSpace(req.Key),
		Description:       req.Description,
		Enabled:           req.Enabled,
		RolloutPercentage: req.RolloutPercentage,
		VariantOn:         req.VariantOn,
		VariantOff:        req.VariantOff,
	}); err != nil {
		return UpsertFlagResponse{}, internalError("Failed to save feature flag")
	}
	return UpsertFlagResponse{Message: "Feature flag saved"}, nil
}

// -- DELETE /features/admin/flags -------------------------------------------

func DeleteFlagRoutes() []string { return []string{"/features/admin/flags"} }

type DeleteFlagHandler struct{ Service *application.Service }

type DeleteFlagRequest struct {
	Key string `json:"key"`
}

type DeleteFlagResponse struct {
	Message string `json:"message"`
}

func (h *DeleteFlagHandler) Validate(_ context.Context, req *DeleteFlagRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.Key) == "" {
		return badRequest("Feature key is required")
	}
	return nil
}

func (h *DeleteFlagHandler) Serve(ctx context.Context, req *DeleteFlagRequest) (DeleteFlagResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return DeleteFlagResponse{}, err
	}
	if err := h.Service.DeleteFlag(ctx, strings.TrimSpace(req.Key)); err != nil {
		return DeleteFlagResponse{}, internalError("Failed to delete feature flag")
	}
	return DeleteFlagResponse{Message: "Feature flag deleted"}, nil
}

// -- GET /features/admin/targets?feature_key= -------------------------------

func ListTargetsRoutes() []string { return []string{"/features/admin/targets"} }

type ListTargetsHandler struct{ Service *application.Service }

type ListTargetsRequest struct {
	FeatureKey string `json:"feature_key"`
}

type ListTargetsResponse struct {
	Targets []TargetPayload `json:"targets"`
}

func (h *ListTargetsHandler) Serve(ctx context.Context, req *ListTargetsRequest) (ListTargetsResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return ListTargetsResponse{}, err
	}
	if strings.TrimSpace(req.FeatureKey) == "" {
		return ListTargetsResponse{}, badRequest("feature_key is required")
	}
	targets, err := h.Service.ListTargets(ctx, strings.TrimSpace(req.FeatureKey))
	if err != nil {
		return ListTargetsResponse{}, internalError("Failed to list feature targets")
	}
	payloads := make([]TargetPayload, 0, len(targets))
	for _, t := range targets {
		payloads = append(payloads, TargetPayload{
			FeatureKey: t.FeatureKey,
			OwnerID:    t.OwnerID,
			Enabled:    t.Enabled,
			Variant:    t.Variant,
		})
	}
	return ListTargetsResponse{Targets: payloads}, nil
}

// -- PUT /features/admin/targets (target a specific user) -------------------

func UpsertTargetRoutes() []string { return []string{"/features/admin/targets"} }

type UpsertTargetHandler struct{ Service *application.Service }

type UpsertTargetRequest struct {
	FeatureKey string `json:"feature_key"`
	OwnerID    int64  `json:"owner_id"`
	Enabled    bool   `json:"enabled"`
	Variant    string `json:"variant"`
}

type UpsertTargetResponse struct {
	Message string `json:"message"`
}

func (h *UpsertTargetHandler) Validate(_ context.Context, req *UpsertTargetRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.FeatureKey) == "" {
		return badRequest("feature_key is required")
	}
	if req.OwnerID == 0 {
		return badRequest("owner_id is required")
	}
	return nil
}

func (h *UpsertTargetHandler) Serve(ctx context.Context, req *UpsertTargetRequest) (UpsertTargetResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return UpsertTargetResponse{}, err
	}
	if err := h.Service.UpsertTarget(ctx, domain.Target{
		FeatureKey: strings.TrimSpace(req.FeatureKey),
		OwnerID:    req.OwnerID,
		Enabled:    req.Enabled,
		Variant:    strings.TrimSpace(req.Variant),
	}); err != nil {
		return UpsertTargetResponse{}, internalError("Failed to save feature target")
	}
	return UpsertTargetResponse{Message: "Feature target saved"}, nil
}

// -- DELETE /features/admin/targets -----------------------------------------

func DeleteTargetRoutes() []string { return []string{"/features/admin/targets"} }

type DeleteTargetHandler struct{ Service *application.Service }

type DeleteTargetRequest struct {
	FeatureKey string `json:"feature_key"`
	OwnerID    int64  `json:"owner_id"`
}

type DeleteTargetResponse struct {
	Message string `json:"message"`
}

func (h *DeleteTargetHandler) Validate(_ context.Context, req *DeleteTargetRequest) *errorModels.ServerError {
	if strings.TrimSpace(req.FeatureKey) == "" {
		return badRequest("feature_key is required")
	}
	if req.OwnerID == 0 {
		return badRequest("owner_id is required")
	}
	return nil
}

func (h *DeleteTargetHandler) Serve(ctx context.Context, req *DeleteTargetRequest) (DeleteTargetResponse, *errorModels.ServerError) {
	if err := requireAdmin(ctx); err != nil {
		return DeleteTargetResponse{}, err
	}
	if err := h.Service.DeleteTarget(ctx, strings.TrimSpace(req.FeatureKey), req.OwnerID); err != nil {
		return DeleteTargetResponse{}, internalError("Failed to delete feature target")
	}
	return DeleteTargetResponse{Message: "Feature target deleted"}, nil
}
