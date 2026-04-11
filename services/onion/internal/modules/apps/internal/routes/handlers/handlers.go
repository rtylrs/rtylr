package handlers

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func AppRoutes() []string { return []string{"/apps"} }

type ListAppsHandler struct{ Service *application.Service }
type ListAppsRequest struct{}

type AppPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IsEnabled   bool   `json:"isEnabled"`
}

type ListAppsResponse struct {
	Apps []AppPayload `json:"apps"`
}

func (h *ListAppsHandler) Serve(ctx context.Context, _ *ListAppsRequest) (ListAppsResponse, *errorModels.ServerError) {
	allApps := h.Service.ListApps()

	// Check if user is owner or has specific app permissions
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		// No auth context - return core apps only
		payloads := make([]AppPayload, 0, len(allApps))
		for _, a := range allApps {
			if a.AddonGated == "" {
				payloads = append(payloads, AppPayload{ID: a.ID, Name: a.Name, Description: a.Description, Icon: a.Icon, IsEnabled: a.IsEnabled})
			}
		}
		return ListAppsResponse{Apps: payloads}, nil
	}

	// Get addon enablement status for this business
	enabledAddons := getEnabledAddons(ctx, authUser.BusinessID)

	// Build app list with addon gating applied
	buildPayload := func(a domain.App) AppPayload {
		enabled := a.IsEnabled
		if a.AddonGated != "" {
			enabled = enabledAddons[a.AddonGated]
		}
		return AppPayload{ID: a.ID, Name: a.Name, Description: a.Description, Icon: a.Icon, IsEnabled: enabled}
	}

	if authUser.Role == "owner" || authUser.Role == "admin" {
		// Owners see all apps (gated apps show as disabled if addon not enabled)
		payloads := make([]AppPayload, 0, len(allApps))
		for _, a := range allApps {
			payloads = append(payloads, buildPayload(a))
		}
		return ListAppsResponse{Apps: payloads}, nil
	}

	// Staff user: only return apps they have access to
	allowedApps := getUserAllowedApps(ctx, authUser.BusinessID, authUser.UserID)
	allowed := make(map[string]bool)
	for _, a := range allowedApps {
		allowed[a] = true
	}

	payloads := make([]AppPayload, 0)
	for _, a := range allApps {
		if allowed[a.ID] {
			payloads = append(payloads, buildPayload(a))
		}
	}

	return ListAppsResponse{Apps: payloads}, nil
}

// getEnabledAddons returns a map of addon IDs that are enabled for the given business.
func getEnabledAddons(ctx context.Context, businessID int64) map[string]bool {
	pool := dbclient.Pool()
	rows, err := pool.QueryContext(ctx,
		"SELECT addon_id FROM business_addons WHERE business_id = ? AND is_enabled = 1",
		businessID,
	)
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()

	enabled := make(map[string]bool)
	for rows.Next() {
		var addonID string
		if scanErr := rows.Scan(&addonID); scanErr == nil {
			enabled[addonID] = true
		}
	}
	return enabled
}

func getUserAllowedApps(ctx context.Context, businessID, userID int64) []string {
	pool := dbclient.Pool()
	var appsJSON []byte
	err := pool.QueryRowContext(ctx,
		"SELECT apps FROM user_app_permissions WHERE (business_id = ? OR owner_id = ?) AND user_id = ?",
		businessID, businessID, userID,
	).Scan(&appsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}
		}
		return []string{}
	}

	var apps []string
	if jsonErr := json.Unmarshal(appsJSON, &apps); jsonErr != nil {
		return []string{}
	}
	return apps
}
