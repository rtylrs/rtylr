package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	hrcore "github.com/voxire/onion/internal/modules/hr_core"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func UpsertHRConfigRoutes() []string { return []string{"/hr/config"} }

type UpsertHRConfigHandler struct {
	Service     *application.Service
	CoreService *hrcore.Module
}

type UpsertHRConfigRequest struct {
	SetupCompletedAt      *int64  `json:"setup_completed_at"`
	CompanyName           string  `json:"company_name"`
	FiscalYearStartMonth  int32   `json:"fiscal_year_start_month"`
	DefaultCurrency       string  `json:"default_currency"`
	DefaultTimezone       string  `json:"default_timezone"`
	DefaultWorkScheduleID int64   `json:"default_work_schedule_id"`
	OvertimeEnabled       bool    `json:"overtime_enabled"`
	OvertimeMultiplier    float64 `json:"overtime_multiplier"`
	DirectoryEnabled      bool    `json:"directory_enabled"`
	hasCoreConfigFields   bool
}

type UpsertHRConfigResponse struct {
	Config *HRConfigPayload `json:"config"`
}

func (h *UpsertHRConfigHandler) Validate(_ context.Context, req *UpsertHRConfigRequest) error {
	if !req.hasCoreConfigFields {
		return nil
	}
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
	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return UpsertHRConfigResponse{}, onekiterrors.PermissionDenied()
	}

	cfg, err := h.Service.UpsertHRConfig(ctx, device.User.UserID, req.SetupCompletedAt)
	if err != nil {
		logger.Ctx(ctx).Error("failed to upsert hr config", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return UpsertHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update HR config")
	}

	var coreCfg *hrcore.Config
	if req.hasCoreConfigFields {
		if h.CoreService == nil {
			return UpsertHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "HR core configuration is unavailable")
		}
		coreCfg, err = h.CoreService.UpsertConfig(ctx, businessID, hrcore.ConfigInput{
			CompanyName:           req.CompanyName,
			FiscalYearStartMonth:  req.FiscalYearStartMonth,
			DefaultCurrency:       req.DefaultCurrency,
			DefaultTimezone:       req.DefaultTimezone,
			DefaultWorkScheduleID: req.DefaultWorkScheduleID,
			OvertimeEnabled:       req.OvertimeEnabled,
			OvertimeMultiplier:    req.OvertimeMultiplier,
			DirectoryEnabled:      req.DirectoryEnabled,
		})
		if err != nil {
			logger.Ctx(ctx).Error("failed to upsert hr core config", logger.Any("business_id", businessID), logger.Error(err))
			return UpsertHRConfigResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update HR config")
		}
	}

	return UpsertHRConfigResponse{Config: toConfigPayload(cfg, coreCfg)}, nil
}

func (r *UpsertHRConfigRequest) UnmarshalJSON(data []byte) error {
	type requestPayload struct {
		SetupCompletedAtSnake *json.RawMessage `json:"setup_completed_at"`
		CompanyNameSnake      string           `json:"company_name"`
		FiscalMonthSnake      int32            `json:"fiscal_year_start_month"`
		CurrencySnake         string           `json:"default_currency"`
		TimezoneSnake         string           `json:"default_timezone"`
		ScheduleSnake         int64            `json:"default_work_schedule_id"`
		OvertimeEnabledSnake  bool             `json:"overtime_enabled"`
		MultiplierSnake       float64          `json:"overtime_multiplier"`
		DirectorySnake        bool             `json:"directory_enabled"`
	}
	var payload requestPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	decode := func(key string, dst any) error {
		raw, ok := fields[key]
		if !ok {
			return nil
		}
		return json.Unmarshal(raw, dst)
	}
	var setupCompletedAtCamel *json.RawMessage
	var companyNameCamel string
	var fiscalMonthCamel int32
	var currencyCamel string
	var timezoneCamel string
	var scheduleCamel int64
	var overtimeEnabledCamel bool
	var multiplierCamel float64
	var directoryCamel bool
	if raw, ok := fields["setupCompletedAt"]; ok {
		setupCompletedAtCamel = &raw
	}
	if err := decode("companyName", &companyNameCamel); err != nil {
		return err
	}
	if err := decode("fiscalYearStartMonth", &fiscalMonthCamel); err != nil {
		return err
	}
	if err := decode("defaultCurrency", &currencyCamel); err != nil {
		return err
	}
	if err := decode("defaultTimezone", &timezoneCamel); err != nil {
		return err
	}
	if err := decode("defaultWorkScheduleId", &scheduleCamel); err != nil {
		return err
	}
	if err := decode("overtimeEnabled", &overtimeEnabledCamel); err != nil {
		return err
	}
	if err := decode("overtimeMultiplier", &multiplierCamel); err != nil {
		return err
	}
	if err := decode("directoryEnabled", &directoryCamel); err != nil {
		return err
	}

	setupRaw := payload.SetupCompletedAtSnake
	if setupRaw == nil {
		setupRaw = setupCompletedAtCamel
	}
	setupCompletedAt, err := parseOptionalInt64(setupRaw)
	if err != nil {
		return fmt.Errorf("setup_completed_at: %w", err)
	}

	*r = UpsertHRConfigRequest{
		SetupCompletedAt:      setupCompletedAt,
		CompanyName:           firstString(payload.CompanyNameSnake, companyNameCamel),
		FiscalYearStartMonth:  firstInt32(payload.FiscalMonthSnake, fiscalMonthCamel),
		DefaultCurrency:       firstString(payload.CurrencySnake, currencyCamel),
		DefaultTimezone:       firstString(payload.TimezoneSnake, timezoneCamel),
		DefaultWorkScheduleID: firstInt64(payload.ScheduleSnake, scheduleCamel),
		OvertimeEnabled:       payload.OvertimeEnabledSnake || overtimeEnabledCamel,
		OvertimeMultiplier:    firstFloat64(payload.MultiplierSnake, multiplierCamel),
		DirectoryEnabled:      payload.DirectorySnake || directoryCamel,
	}
	for _, key := range []string{
		"company_name", "companyName", "fiscal_year_start_month", "fiscalYearStartMonth",
		"default_currency", "defaultCurrency", "default_timezone", "defaultTimezone",
		"default_work_schedule_id", "defaultWorkScheduleId", "overtime_enabled", "overtimeEnabled",
		"overtime_multiplier", "overtimeMultiplier", "directory_enabled", "directoryEnabled",
	} {
		if _, ok := fields[key]; ok {
			r.hasCoreConfigFields = true
			break
		}
	}
	return nil
}

func parseOptionalInt64(raw *json.RawMessage) (*int64, error) {
	if raw == nil {
		return nil, nil
	}
	var stringValue string
	if err := json.Unmarshal(*raw, &stringValue); err == nil {
		value, err := strconv.ParseInt(strings.TrimSpace(stringValue), 10, 64)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	var value int64
	if err := json.Unmarshal(*raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

func firstString(snake, camel string) string {
	if snake != "" {
		return snake
	}
	return camel
}

func firstInt32(snake, camel int32) int32 {
	if snake != 0 {
		return snake
	}
	return camel
}

func firstInt64(snake, camel int64) int64 {
	if snake != 0 {
		return snake
	}
	return camel
}

func firstFloat64(snake, camel float64) float64 {
	if snake != 0 {
		return snake
	}
	return camel
}
