package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/activity/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// --- List Activity -----------------------------------------------------------

func ListActivityRoutes() []string {
	return []string{"/erp/activity"}
}

type ListActivityHandler struct {
	Service *application.Service
}

type ListActivityRequest struct {
	Limit int `json:"limit"`
}

type ActivityPayload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"ownerId"`
	EntityType  string `json:"entityType"`
	EntityID    int64  `json:"entityId"`
	Action      string `json:"action"`
	Description string `json:"description"`
	UserID      int64  `json:"userId"`
	CreatedAt   string `json:"createdAt"`
}

type ListActivityResponse struct {
	Activities []ActivityPayload `json:"activities"`
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

	payloads := make([]ActivityPayload, 0, len(activities))
	for _, a := range activities {
		payloads = append(payloads, ActivityPayload{
			ID:          a.ID,
			OwnerID:     a.OwnerID,
			EntityType:  a.EntityType,
			EntityID:    a.EntityID,
			Action:      a.Action,
			Description: a.Description,
			UserID:      a.UserID,
			CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return ListActivityResponse{Activities: payloads}, nil
}
