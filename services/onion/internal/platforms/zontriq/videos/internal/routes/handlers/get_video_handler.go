package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	videosv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/videos/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/application"
	"github.com/voxire/onion/internal/platforms/zontriq/videos/internal/domain"
)

func GetVideoRoutes() []string {
	return []string{
		"/zontriq/videos/{id}",
	}
}

type GetVideoHandler struct {
	VideoService *application.Service
}

func (h *GetVideoHandler) Serve(ctx context.Context, _ *videosv1.GetVideoRequest) (videosv1.GetVideoResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return videosv1.GetVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
	}

	idStr := chi.RouteContext(ctx).URLParam("id")

	videoID, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return videosv1.GetVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid video ID",
		}
	}

	result, err := h.VideoService.GetVideoByID(ctx, int32(videoID))
	if err != nil {
		return videosv1.GetVideoResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get video",
		}
	}

	protoVideo := &videosv1.Video{
		Id:            int64(result.Video.ID),
		SeriesId:      int64(result.Video.SeriesID),
		SeriesName:    result.Video.SeriesName,
		Title:         result.Video.Title,
		Script:        result.Video.Script,
		VideoUrl:      result.Video.VideoUrl,
		ThumbnailUrl:  result.Video.ThumbnailUrl,
		State:         result.Video.State,
		Duration:      result.Video.Duration,
		EpisodeNumber: result.Video.EpisodeNumber,
		ErrorMessage:  result.Video.ErrorMessage,
		CreatedAt:     result.Video.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     result.Video.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if result.ActiveJob != nil {
		protoVideo.Job = &videosv1.JobProgress{
			Id:           int64(result.ActiveJob.ID),
			Type:         result.ActiveJob.Type,
			State:        result.ActiveJob.State,
			Progress:     result.ActiveJob.Progress,
			CurrentStep:  result.ActiveJob.CurrentStep,
			ErrorMessage: result.ActiveJob.ErrorMessage,
		}

		if !result.ActiveJob.StartedAt.IsZero() {
			protoVideo.Job.StartedAt = result.ActiveJob.StartedAt.Format("2006-01-02T15:04:05Z")
		}

		if !result.ActiveJob.CompletedAt.IsZero() {
			protoVideo.Job.CompletedAt = result.ActiveJob.CompletedAt.Format("2006-01-02T15:04:05Z")
		}
	}

	return videosv1.GetVideoResponse{
		Video:      protoVideo,
		Targets:    mapTargets(result.Targets),
		PublishLog: mapPublishLog(result.PublishLog),
	}, nil
}

func mapTargets(targets []domain.VideoTarget) []*videosv1.VideoTarget {
	proto := make([]*videosv1.VideoTarget, 0, len(targets))

	for i := range targets {
		t := &targets[i]
		pt := &videosv1.VideoTarget{
			Id:             int64(t.ID),
			VideoId:        int64(t.VideoID),
			ChannelId:      int64(t.ChannelID),
			ChannelName:    t.ChannelName,
			Platform:       t.Platform,
			State:          t.State,
			PlatformPostId: t.PlatformPostID,
			ErrorMessage:   t.ErrorMessage,
		}

		if !t.PublishedAt.IsZero() {
			pt.PublishedAt = t.PublishedAt.Format("2006-01-02T15:04:05Z")
		}

		proto = append(proto, pt)
	}

	return proto
}

func mapPublishLog(log []domain.PublishLogEntry) []*videosv1.PublishLogEntry {
	proto := make([]*videosv1.PublishLogEntry, 0, len(log))

	for i := range log {
		pl := &log[i]
		entry := &videosv1.PublishLogEntry{
			Id:              int64(pl.ID),
			VideoId:         int64(pl.VideoID),
			ChannelId:       int64(pl.ChannelID),
			Platform:        pl.Platform,
			PlatformPostId:  pl.PlatformPostID,
			PlatformPostUrl: pl.PlatformPostUrl,
			Result:          pl.Result,
			ErrorMessage:    pl.ErrorMessage,
		}

		if !pl.PublishedAt.IsZero() {
			entry.PublishedAt = pl.PublishedAt.Format("2006-01-02T15:04:05Z")
		}

		proto = append(proto, entry)
	}

	return proto
}
