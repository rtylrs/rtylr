package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowsettingv1 "github.com/voxire/voxie/gen/go/flow/setting/v1"
)

type SettingAPI struct {
	Service *application.Service
}

func NewSettingAPI(service *application.Service) *SettingAPI {
	return &SettingAPI{Service: service}
}

func (a *SettingAPI) GetUserSettings(ctx context.Context, _ *flowsettingv1.GetUserSettingsRequest) (*flowsettingv1.GetUserSettingsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	s, err := a.Service.GetUserSettings(ctx, authUser.UserID)
	if err != nil {
		return nil, workspaceInternalError(ctx, "GetUserSettings failed", err)
	}

	response := &flowsettingv1.GetUserSettingsResponse{
		TaskAssigned:       s.TaskAssigned,
		TaskComments:       s.TaskComments,
		Mentions:           s.Mentions,
		StatusChanges:      s.StatusChanges,
		WeeklyDigest:       s.WeeklyDigest,
		EmailNotifications: s.EmailNotifications,
	}
	if s.SidebarPreferencesJSON != "" {
		response.SidebarPreferences = &s.SidebarPreferencesJSON
	}
	return response, nil
}

func (a *SettingAPI) UpdateUserSettings(ctx context.Context, req *flowsettingv1.UpdateUserSettingsRequest) (*flowsettingv1.UpdateUserSettingsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	// No permission check: the row written below is keyed on authUser.UserID,
	// so a caller can only ever change their own preferences. The business is
	// still resolved, which keeps a session with no business out.
	if _, businessErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID); businessErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	sidebarPreferencesJSON := req.GetSidebarPreferences()
	if req.SidebarPreferences == nil {
		current, err := a.Service.GetUserSettings(ctx, authUser.UserID)
		if err != nil {
			return nil, workspaceInternalError(ctx, "GetUserSettings before update failed", err)
		}
		sidebarPreferencesJSON = current.SidebarPreferencesJSON
	}

	if err := a.Service.UpdateUserSettings(ctx, domain.UserSettings{
		UserID:                 authUser.UserID,
		TaskAssigned:           req.GetTaskAssigned(),
		TaskComments:           req.GetTaskComments(),
		Mentions:               req.GetMentions(),
		StatusChanges:          req.GetStatusChanges(),
		WeeklyDigest:           req.GetWeeklyDigest(),
		EmailNotifications:     req.GetEmailNotifications(),
		SidebarPreferencesJSON: sidebarPreferencesJSON,
	}); err != nil {
		return nil, workspaceInternalError(ctx, "UpdateUserSettings failed", err)
	}
	return &flowsettingv1.UpdateUserSettingsResponse{}, nil
}
