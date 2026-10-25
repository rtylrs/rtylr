package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/aiquota/internal/application"
	"github.com/voxire/onion/internal/modules/aiquota/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubaiv1 "github.com/voxire/voxie/gen/go/hub/ai/v1"
)

type AiAPI struct {
	Service *application.Service
}

func NewAiAPI(service *application.Service) *AiAPI {
	return &AiAPI{Service: service}
}

func quotaStatusToProto(st domain.QuotaStatus) *hubaiv1.AiQuotaStatus {
	return &hubaiv1.AiQuotaStatus{
		PlanSlug:        st.PlanSlug,
		Mode:            string(st.Mode),
		AllowanceTokens: st.AllowanceTokens,
		UsedTokens:      st.UsedTokens,
		RemainingTokens: st.RemainingTokens,
		OverageTokens:   st.OverageTokens,
		AllowOverage:    st.AllowOverage,
		Blocked:         st.Blocked,
		PeriodEnd:       st.PeriodEnd.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (a *AiAPI) GetAiUsage(ctx context.Context, _ *hubaiv1.GetAiUsageRequest) (*hubaiv1.GetAiUsageResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	st, err := a.Service.Status(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load AI usage")
	}

	return &hubaiv1.GetAiUsageResponse{Usage: quotaStatusToProto(st)}, nil
}

func (a *AiAPI) GetAiQuotaSettings(ctx context.Context, _ *hubaiv1.GetAiQuotaSettingsRequest) (*hubaiv1.GetAiQuotaSettingsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	mode, err := a.Service.GetMode(ctx, device.User.BusinessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to load AI quota settings")
	}

	return &hubaiv1.GetAiQuotaSettingsResponse{
		Settings: &hubaiv1.AiQuotaSettings{Mode: string(mode)},
	}, nil
}

func (a *AiAPI) UpdateAiQuotaSettings(ctx context.Context, req *hubaiv1.UpdateAiQuotaSettingsRequest) (*hubaiv1.UpdateAiQuotaSettingsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if !canUpdateAiQuotaSettings(device.User) {
		return nil, onekiterrors.PermissionDenied()
	}

	mode := domain.Mode(req.GetMode())
	if mode != domain.ModePool && mode != domain.ModePerPerson {
		return nil, onekiterrors.New(http.StatusBadRequest, "mode must be 'pool' or 'per_person'")
	}

	if err := a.Service.SetMode(ctx, device.User.BusinessID, mode); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update AI quota settings")
	}

	normalized := string(mode.Normalize())
	return &hubaiv1.UpdateAiQuotaSettingsResponse{
		Settings: &hubaiv1.AiQuotaSettings{Mode: normalized},
	}, nil
}

func canUpdateAiQuotaSettings(user *authPkg.AuthUser) bool {
	return authPkg.IsBusinessAdmin(user) || authPkg.HasDashboardPermission(user, "hub:ai:quota:manage")
}
