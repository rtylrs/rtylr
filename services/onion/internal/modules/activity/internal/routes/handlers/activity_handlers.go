package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/activity/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	hubactivityv1 "github.com/voxire/voxie/gen/go/hub/activity/v1"
)

// --- List Activity -----------------------------------------------------------

func ListActivityRoutes() []string {
	return []string{"/erp/activity", "/activity"}
}

type ListActivityHandler struct {
	Service *application.Service
}

type ListActivityRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	App    string `json:"app"`
}

type ListActivityResponse struct {
	Activities []*hubactivityv1.ActivityEntry `json:"activities"`
	Total      int64                          `json:"total"`
	Limit      int                            `json:"limit"`
	Offset     int                            `json:"offset"`
}

func (h *ListActivityHandler) Serve(ctx context.Context, req *ListActivityRequest) (ListActivityResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListActivityResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListActivityResponse{}, errorModels.PermissionDeniedServerError()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	page, err := h.Service.ListActivity(ctx, businessID, req.App, limit, offset)
	if err != nil {
		return ListActivityResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list activity",
		}
	}

	payloads := make([]*hubactivityv1.ActivityEntry, 0, len(page.Entries))
	for _, a := range page.Entries {
		payloads = append(payloads, &hubactivityv1.ActivityEntry{
			Id:          a.ID,
			OwnerId:     a.OwnerID,
			App:         a.App,
			EntityType:  a.EntityType,
			EntityId:    a.EntityID,
			Action:      a.Action,
			Description: a.Description,
			UserId:      a.UserID,
			UserName:    a.UserName,
			CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return ListActivityResponse{Activities: payloads, Total: page.Total, Limit: limit, Offset: offset}, nil
}
