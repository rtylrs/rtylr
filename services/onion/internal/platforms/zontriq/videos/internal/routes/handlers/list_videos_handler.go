package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	commonv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/common/v1"
	videosv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/videos/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/application"
)

func ListVideosRoutes() []string {
	return []string{
		"/zontriq/videos",
	}
}

type ListVideosHandler struct {
	VideoService *application.Service
}

func (h *ListVideosHandler) Serve(ctx context.Context, req *videosv1.ListVideosRequest) (videosv1.ListVideosResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return videosv1.ListVideosResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	zontriqUserID := int32(device.User.PlatformUserID)

	var (
		page     int32 = 1
		pageSize int32 = 20
	)

	if req.GetPagination() != nil {
		if req.GetPagination().GetPage() > 0 {
			page = req.GetPagination().GetPage()
		}

		if req.GetPagination().GetPageSize() > 0 {
			pageSize = req.GetPagination().GetPageSize()
		}
	}

	params := application.ListVideosParams{
		UserID:   zontriqUserID,
		Page:     page,
		PageSize: pageSize,
	}

	if req.GetSeriesId() > 0 {
		seriesID := int32(req.GetSeriesId())
		params.SeriesID = &seriesID
	}

	if req.GetState() != "" {
		state := req.GetState()
		params.State = &state
	}

	result, err := h.VideoService.ListVideos(ctx, params)
	if err != nil {
		return videosv1.ListVideosResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list videos",
		}
	}

	protoVideos := make([]*videosv1.Video, 0, len(result.Videos))
	for i := range result.Videos {
		v := &result.Videos[i]
		protoVideos = append(protoVideos, &videosv1.Video{
			Id:            int64(v.ID),
			SeriesId:      int64(v.SeriesID),
			Title:         v.Title,
			Script:        v.Script,
			VideoUrl:      v.VideoUrl,
			ThumbnailUrl:  v.ThumbnailUrl,
			State:         v.State,
			Duration:      v.Duration,
			EpisodeNumber: v.EpisodeNumber,
			ErrorMessage:  v.ErrorMessage,
			CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	totalPages := int32(result.TotalCount) / pageSize
	if int32(result.TotalCount)%pageSize > 0 {
		totalPages++
	}

	return videosv1.ListVideosResponse{
		Videos: protoVideos,
		Pagination: &commonv1.PaginationResponse{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
			TotalCount:  result.TotalCount,
			HasNext:     page < totalPages,
			HasPrev:     page > 1,
		},
	}, nil
}
