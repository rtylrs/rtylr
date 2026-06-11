package handlers

import (
	"context"

	"github.com/voxire/onion/internal/modules/apps/internal/application"
	"github.com/voxire/onion/internal/modules/apps/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	hubappv1 "github.com/voxire/voxie/gen/go/hub/app/v1"
)

func AppRoutes() []string { return []string{"/apps"} }

type (
	ListAppsHandler struct{ Service *application.Service }
	ListAppsRequest struct{}
)

type ListAppsResponse struct {
	Apps []*hubappv1.App `json:"apps"`
}

func (h *ListAppsHandler) Serve(ctx context.Context, _ *ListAppsRequest) (ListAppsResponse, *errorModels.ServerError) {
	var businessID, userID int64
	var role string
	if u, ok := authPkg.GetAuthUser(ctx); ok {
		businessID, userID, role = u.BusinessID, u.UserID, u.Role
	}

	apps, err := h.Service.GetVisibleAppsForUser(ctx, businessID, userID, role)
	if err != nil {
		return ListAppsResponse{}, errorModels.FromError(err, "Failed to load apps")
	}

	payloads := make([]*hubappv1.App, 0, len(apps))
	for _, a := range apps {
		payloads = append(payloads, toProtoApp(a))
	}
	return ListAppsResponse{Apps: payloads}, nil
}

func toProtoApp(a domain.App) *hubappv1.App {
	return &hubappv1.App{
		Id:          a.ID,
		Name:        a.Name,
		Description: a.Description,
		Icon:        a.Icon,
		IsEnabled:   a.IsEnabled,
	}
}
