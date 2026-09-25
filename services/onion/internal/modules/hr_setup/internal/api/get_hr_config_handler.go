package api

import (
	"context"
	"errors"
	"net/http"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func GetHRConfigRoutes() []string { return []string{"/hr/config"} }

type GetHRConfigRequest struct{}

type HRConfigPayload struct {
	ID                    int64   `json:"id"`
	OwnerID               int64   `json:"owner_id"`
	TimeTrackingModel     string  `json:"time_tracking_model"`
	SetupCompletedAt      *int64  `json:"setup_completed_at"`
	CompanyName           string  `json:"company_name,omitempty"`
	FiscalYearStartMonth  int32   `json:"fiscal_year_start_month,omitempty"`
	DefaultCurrency       string  `json:"default_currency,omitempty"`
	DefaultTimezone       string  `json:"default_timezone,omitempty"`
	DefaultWorkScheduleID int64   `json:"default_work_schedule_id,omitempty"`
	OvertimeEnabled       bool    `json:"overtime_enabled,omitempty"`
	OvertimeMultiplier    float64 `json:"overtime_multiplier,omitempty"`
	DirectoryEnabled      bool    `json:"directory_enabled,omitempty"`
	CreatedAt             int64   `json:"created_at"`
	UpdatedAt             int64   `json:"updated_at"`
}

type GetHRConfigResponse struct {
	Config *HRConfigPayload `json:"config"`
}

type GetHRConfigHandler struct {
	Service     *application.Service
	CoreService *hrcore.Module
}

func (h *GetHRConfigHandler) Serve(ctx context.Context, _ *GetHRConfigRequest) (GetHRConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetHRConfigResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return GetHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, setupErr := h.Service.GetHRConfig(ctx, device.User.UserID)
	setupNotFound := errors.Is(setupErr, domain.ErrConfigNotFound)
	if setupErr != nil && !errors.Is(setupErr, domain.ErrConfigNotFound) {
		logger.Ctx(ctx).Error("failed to get hr setup config", logger.Any("owner_id", device.User.UserID), logger.Error(setupErr))
		return GetHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get HR config")
	}

	var coreCfg *hrcore.Config
	if h.CoreService != nil {
		coreCfg, setupErr = h.CoreService.GetConfig(ctx, businessID)
		if setupErr != nil && !hrcore.IsConfigNotFound(setupErr) {
			logger.Ctx(ctx).Error("failed to get hr core config", logger.Any("business_id", businessID), logger.Error(setupErr))
			return GetHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get HR config")
		}
	}
	if setupNotFound {
		cfg = nil
	}
	if cfg == nil && coreCfg == nil {
		return GetHRConfigResponse{Config: nil}, nil
	}

	return GetHRConfigResponse{Config: toConfigPayload(cfg, coreCfg)}, nil
}

func toConfigPayload(cfg *domain.HRConfig, coreCfg *hrcore.Config) *HRConfigPayload {
	p := &HRConfigPayload{TimeTrackingModel: "both"}
	if cfg != nil {
		p.ID = cfg.ID
		p.OwnerID = cfg.OwnerID
		p.CreatedAt = cfg.CreatedAt.Unix()
		p.UpdatedAt = cfg.UpdatedAt.Unix()
		if cfg.SetupCompletedAt != nil {
			unix := cfg.SetupCompletedAt.Unix()
			p.SetupCompletedAt = &unix
		}
	}
	if coreCfg != nil {
		if p.ID == 0 {
			p.ID = coreCfg.ID
		}
		if p.OwnerID == 0 {
			p.OwnerID = coreCfg.OwnerID
		}
		if p.CreatedAt == 0 {
			p.CreatedAt = coreCfg.CreatedAt.Unix()
		}
		if p.UpdatedAt == 0 {
			p.UpdatedAt = coreCfg.UpdatedAt.Unix()
		}
		p.CompanyName = coreCfg.CompanyName
		p.FiscalYearStartMonth = coreCfg.FiscalYearStartMonth
		p.DefaultCurrency = coreCfg.DefaultCurrency
		p.DefaultTimezone = coreCfg.DefaultTimezone
		p.DefaultWorkScheduleID = coreCfg.DefaultWorkScheduleID
		p.OvertimeEnabled = coreCfg.OvertimeEnabled
		p.OvertimeMultiplier = coreCfg.OvertimeMultiplier
		p.DirectoryEnabled = coreCfg.DirectoryEnabled
	}
	return p
}
