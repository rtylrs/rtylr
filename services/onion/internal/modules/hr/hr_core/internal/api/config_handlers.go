package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_core/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

type HRConfigPayload struct {
	ID                    int64   `json:"id"`
	CompanyName           string  `json:"company_name"`
	FiscalYearStartMonth  int32   `json:"fiscal_year_start_month"`
	DefaultCurrency       string  `json:"default_currency"`
	DefaultTimezone       string  `json:"default_timezone"`
	DefaultWorkScheduleID int64   `json:"default_work_schedule_id"`
	OvertimeEnabled       bool    `json:"overtime_enabled"`
	OvertimeMultiplier    float64 `json:"overtime_multiplier"`
	DirectoryEnabled      bool    `json:"directory_enabled"`
	CreatedAt             string  `json:"created_at"`
	UpdatedAt             string  `json:"updated_at"`
}

func configToPayload(c domain.HRCompanyConfigRecord) *HRConfigPayload {
	return &HRConfigPayload{
		ID: c.ID, CompanyName: c.CompanyName,
		FiscalYearStartMonth: c.FiscalYearStartMonth, DefaultCurrency: c.DefaultCurrency,
		DefaultTimezone: c.DefaultTimezone, DefaultWorkScheduleID: c.DefaultWorkScheduleID,
		OvertimeEnabled: c.OvertimeEnabled, OvertimeMultiplier: c.OvertimeMultiplier,
		DirectoryEnabled: c.DirectoryEnabled,
		CreatedAt:        c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- Get HR Config ---

// The onboarding/setup config owns /hr/config. The full company config
// API is exposed under /hr/company-config to avoid two handlers competing for
// the same method and path.
func GetHRConfigRoutes() []string { return []string{"/hr/company-config"} }

type (
	GetHRConfigHandler  struct{ Service *application.Service }
	GetHRConfigRequest  struct{}
	GetHRConfigResponse struct {
		Config *HRConfigPayload `json:"config"`
	}
)

func (h *GetHRConfigHandler) Serve(ctx context.Context, _ *GetHRConfigRequest) (GetHRConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetHRConfigResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.GetHRConfig(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrSetupNotComplete) {
			return GetHRConfigResponse{}, onekiterrors.New(http.StatusNotFound, "HR configuration not found. Please complete setup first.")
		}
		logger.Ctx(ctx).Error("failed to get hr config", logger.Any("business_id", businessID), logger.Error(err))
		return GetHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get HR config")
	}

	return GetHRConfigResponse{Config: configToPayload(*cfg)}, nil
}

// --- Upsert HR Config ---

func UpsertHRConfigRoutes() []string { return []string{"/hr/company-config"} }

type (
	UpsertHRConfigHandler struct{ Service *application.Service }
	UpsertHRConfigRequest struct {
		CompanyName           string  `json:"company_name"`
		FiscalYearStartMonth  int32   `json:"fiscal_year_start_month"`
		DefaultCurrency       string  `json:"default_currency"`
		DefaultTimezone       string  `json:"default_timezone"`
		DefaultWorkScheduleID int64   `json:"default_work_schedule_id"`
		OvertimeEnabled       bool    `json:"overtime_enabled"`
		OvertimeMultiplier    float64 `json:"overtime_multiplier"`
		DirectoryEnabled      bool    `json:"directory_enabled"`
	}
)

type UpsertHRConfigResponse struct {
	Config *HRConfigPayload `json:"config"`
}

func (h *UpsertHRConfigHandler) Validate(_ context.Context, req *UpsertHRConfigRequest) error {
	req.CompanyName = strings.TrimSpace(req.CompanyName)
	if req.CompanyName == "" {
		return onekiterrors.New(http.StatusBadRequest, "company_name is required")
	}
	return nil
}

func (h *UpsertHRConfigHandler) Serve(ctx context.Context, req *UpsertHRConfigRequest) (UpsertHRConfigResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return UpsertHRConfigResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireHRPermission(ctx, "hr:employees:manage"); err != nil {
		return UpsertHRConfigResponse{}, err
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return UpsertHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg := domain.HRCompanyConfigRecord{
		CompanyName:           req.CompanyName,
		FiscalYearStartMonth:  req.FiscalYearStartMonth,
		DefaultCurrency:       req.DefaultCurrency,
		DefaultTimezone:       req.DefaultTimezone,
		DefaultWorkScheduleID: req.DefaultWorkScheduleID,
		OvertimeEnabled:       req.OvertimeEnabled,
		OvertimeMultiplier:    req.OvertimeMultiplier,
		DirectoryEnabled:      req.DirectoryEnabled,
	}

	updated, err := h.Service.UpsertHRConfig(ctx, businessID, cfg)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert hr config", logger.Any("business_id", businessID), logger.Error(err))
		return UpsertHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update HR config")
	}

	return UpsertHRConfigResponse{Config: configToPayload(*updated)}, nil
}
