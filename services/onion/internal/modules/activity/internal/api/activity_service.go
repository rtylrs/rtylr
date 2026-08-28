package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/activity/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubactivityv1 "github.com/voxire/voxie/gen/go/hub/activity/v1"
)

type ActivityAPI struct {
	Service *application.Service
}

func NewActivityAPI(service *application.Service) *ActivityAPI {
	return &ActivityAPI{Service: service}
}

func (a *ActivityAPI) ListActivity(ctx context.Context, req *hubactivityv1.ListActivityRequest) (*hubactivityv1.ListActivityResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil {
		return nil, onekiterrors.New(http.StatusUnauthorized, "unauthorized")
	}
	if !authPkg.HasDashboardPermission(authUser, "hub:audit:view") {
		return nil, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	page, err := a.Service.ListActivity(ctx, businessID, "", limit, 0)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list activity")
	}

	payloads := make([]*hubactivityv1.ActivityEntry, 0, len(page.Entries))
	for _, entry := range page.Entries {
		payloads = append(payloads, &hubactivityv1.ActivityEntry{
			Id:          entry.ID,
			OwnerId:     entry.OwnerID,
			App:         entry.App,
			EntityType:  entry.EntityType,
			EntityId:    entry.EntityID,
			Action:      entry.Action,
			Description: entry.Description,
			UserId:      entry.UserID,
			UserName:    entry.UserName,
			CreatedAt:   entry.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &hubactivityv1.ListActivityResponse{Activities: payloads}, nil
}
