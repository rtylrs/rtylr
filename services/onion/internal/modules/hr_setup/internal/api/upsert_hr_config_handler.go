package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
)

func UpsertHRConfigRoutes() []string { return []string{"/hr/config"} }

type UpsertHRConfigHandler struct{ Service *application.Service }

type UpsertHRConfigRequest struct {
	SetupCompletedAt *int64 `json:"setup_completed_at"`
}

// UnmarshalJSON accepts both the legacy snake_case field and the camelCase
// field emitted by the setup wizard. The frontend currently sends the Unix
// timestamp as a JSON string, while older clients send a JSON number.
func (r *UpsertHRConfigRequest) UnmarshalJSON(data []byte) error {
	var payload struct {
		SetupCompletedAt      json.RawMessage `json:"setup_completed_at"`
		SetupCompletedAtCamel json.RawMessage `json:"setupCompletedAt"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	raw := bytes.TrimSpace(payload.SetupCompletedAt)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		raw = bytes.TrimSpace(payload.SetupCompletedAtCamel)
	}
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		r.SetupCompletedAt = nil
		return nil
	}

	var value int64
	if err := json.Unmarshal(raw, &value); err == nil {
		r.SetupCompletedAt = &value
		return nil
	}

	var textValue string
	if err := json.Unmarshal(raw, &textValue); err != nil {
		return fmt.Errorf("setup_completed_at must be a Unix timestamp: %w", err)
	}
	value, err := strconv.ParseInt(strings.TrimSpace(textValue), 10, 64)
	if err != nil {
		return fmt.Errorf("setup_completed_at must be a Unix timestamp: %w", err)
	}
	r.SetupCompletedAt = &value
	return nil
}

type UpsertHRConfigResponse struct {
	Config *HRConfigPayload `json:"config"`
}

func (h *UpsertHRConfigHandler) Serve(ctx context.Context, req *UpsertHRConfigRequest) (UpsertHRConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.UpsertHRConfig(ctx, device.User.UserID, req.SetupCompletedAt)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert hr config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UpsertHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update HR config")
	}

	return UpsertHRConfigResponse{Config: toConfigPayload(cfg)}, nil
}
