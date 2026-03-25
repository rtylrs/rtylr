package handlers

import (
	"context"
	"net/http"

	auth "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	commonv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/common/v1"
	automationsv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/automations/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/automations/internal/application"
)

func ListAutomationsRoutes() []string {
	return []string{
		"/zontriq/automations",
	}
}

type ListAutomationsHandler struct {
	AutomationsService *application.Service
}

func (h *ListAutomationsHandler) Serve(ctx context.Context, req *automationsv1.ListAutomationsRequest) (automationsv1.ListAutomationsResponse, *errorModels.ServerError) {
	device := auth.DeviceCtx(ctx)

	if device.User == nil {
		return automationsv1.ListAutomationsResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	var stateFilter *string

	if req.GetState() != "" {
		s := req.GetState()
		stateFilter = &s
	}

	page := int32(1)
	pageSize := int32(20)

	if req.GetPagination() != nil {
		if req.GetPagination().GetPage() > 0 {
			page = req.GetPagination().GetPage()
		}

		if req.GetPagination().GetPageSize() > 0 {
			pageSize = req.GetPagination().GetPageSize()
		}
	}

	pageOffset := (page - 1) * pageSize

	automations, totalCount, err := h.AutomationsService.ListAutomations(ctx, int(zontriqUserID), stateFilter, int(pageSize), int(pageOffset))
	if err != nil {
		return automationsv1.ListAutomationsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list automations",
		}
	}

	totalPages := int32(0)
	if totalCount > 0 {
		totalPages = int32((totalCount + int64(pageSize) - 1) / int64(pageSize))
	}

	protoAutomations := make([]*automationsv1.Automation, 0, len(automations))
	for i := range automations {
		a := &automations[i]

		protoChannels := make([]*automationsv1.AutomationChannel, 0, len(a.Channels))
		for _, ch := range a.Channels {
			protoChannels = append(protoChannels, &automationsv1.AutomationChannel{
				Id:          ch.ID,
				ChannelId:   ch.ChannelID,
				ChannelName: ch.ChannelName,
				Platform:    ch.Platform,
			})
		}

		protoAutomations = append(protoAutomations, &automationsv1.Automation{
			Id:              a.ID,
			SeriesId:        a.SeriesID,
			SeriesName:      a.SeriesName,
			Name:            a.Name,
			Frequency:       a.Frequency,
			State:           a.State,
			NextRun:         a.NextRun,
			LastRun:         a.LastRun,
			VideosGenerated: a.VideosGenerated,
			Channels:        protoChannels,
			Status:          a.Status,
			CreatedAt:       a.CreatedAt,
			UpdatedAt:       a.UpdatedAt,
		})
	}

	return automationsv1.ListAutomationsResponse{
		Automations: protoAutomations,
		Pagination: &commonv1.PaginationResponse{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
			TotalCount:  totalCount,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}
