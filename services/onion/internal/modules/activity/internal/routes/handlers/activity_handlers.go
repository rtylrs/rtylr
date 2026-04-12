package handlers

import (
	"context"
	"net/http"

	hubv1 "github.com/voxire/voxie/gen/go/hub/v1"

	"github.com/voxire/onion/internal/modules/activity/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// --- List Activity -----------------------------------------------------------

func ListActivityRoutes() []string {
	return []string{"/erp/activity", "/activity"}
}

type ListActivityHandler struct {
	Service *application.Service
}

type ListActivityRequest struct {
	Limit int `json:"limit"`
}

type ListActivityResponse struct {
	Activities []*hubv1.ActivityEntry `json:"activities"`
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

	activities, err := h.Service.ListActivity(ctx, authUser.UserID, req.Limit)
	if err != nil {
		return ListActivityResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list activity",
		}
	}

	payloads := make([]*hubv1.ActivityEntry, 0, len(activities))
	for _, a := range activities {
		payloads = append(payloads, &hubv1.ActivityEntry{
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

	return ListActivityResponse{Activities: payloads}, nil
}
