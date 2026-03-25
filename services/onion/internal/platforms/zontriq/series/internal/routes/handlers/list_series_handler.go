package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	commonv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/common/v1"
	seriesv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/series/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/series/internal/application"
)

func ListSeriesRoutes() []string {
	return []string{
		"/zontriq/series",
	}
}

type ListSeriesHandler struct {
	SeriesService *application.Service
}

func (h *ListSeriesHandler) Serve(ctx context.Context, req *seriesv1.ListSeriesRequest) (seriesv1.ListSeriesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return seriesv1.ListSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	userID := int32(device.User.PlatformUserID)

	state := ""
	if req != nil {
		state = req.GetState()
	}

	var (
		page     int32 = 1
		pageSize int32 = 20
	)

	if req != nil && req.GetPagination() != nil {
		if req.GetPagination().GetPage() > 0 {
			page = req.GetPagination().GetPage()
		}

		if req.GetPagination().GetPageSize() > 0 {
			pageSize = req.GetPagination().GetPageSize()
		}
	}

	seriesList, totalCount, err := h.SeriesService.ListSeries(ctx, userID, state, page, pageSize)
	if err != nil {
		return seriesv1.ListSeriesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list series",
		}
	}

	totalPages := int32(totalCount) / pageSize
	if int32(totalCount)%pageSize > 0 {
		totalPages++
	}

	protoSeries := make([]*seriesv1.Series, 0, len(seriesList))
	for i := range seriesList {
		s := &seriesList[i]
		protoSeries = append(protoSeries, &seriesv1.Series{
			Id:          s.Id,
			Name:        s.Name,
			Description: s.Description,
			Niche:       s.Niche,
			Voice:       s.Voice,
			Style:       s.Style,
			Frequency:   s.Frequency,
			Config:      s.Config,
			PublishTime: s.PublishTime,
			State:       s.State,
			VideoCount:  s.VideoCount,
			Status:      s.Status,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}

	return seriesv1.ListSeriesResponse{
		Series: protoSeries,
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
