package handlers

import (
	"context"
	"net/http"

	errorModels "github.com/voxire/onion/internal/platform/errors"
	usersv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/users/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/users/internal/domain"
)

func ListPlansRoutes() []string {
	return []string{
		"/zontriq/plans",
	}
}

type ListPlansHandler struct {
	UsersService *application.Service
}

func (h *ListPlansHandler) Serve(ctx context.Context, _ *usersv1.ListPlansRequest) (usersv1.ListPlansResponse, *errorModels.ServerError) {
	plans, err := h.UsersService.ListPlans(ctx)
	if err != nil {
		return usersv1.ListPlansResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list plans",
		}
	}

	protoPlans := make([]*usersv1.Plan, 0, len(plans))
	for i := range plans {
		protoPlans = append(protoPlans, mapPlanToProto(&plans[i]))
	}

	return usersv1.ListPlansResponse{
		Plans: protoPlans,
	}, nil
}

func mapPlanToProto(p *domain.Plan) *usersv1.Plan {
	return &usersv1.Plan{
		Id:                int64(p.ID),
		Name:              p.Name,
		DisplayName:       p.DisplayName,
		PriceCents:        int32(p.PriceCents),
		MaxVideosPerMonth: int32(p.MaxVideosPerMonth),
		MaxChannels:       int32(p.MaxChannels),
		MaxSeries:         int32(p.MaxSeries),
		MaxAutomations:    int32(p.MaxAutomations),
	}
}
