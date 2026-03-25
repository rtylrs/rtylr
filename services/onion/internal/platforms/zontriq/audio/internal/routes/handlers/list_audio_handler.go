package handlers

import (
	"context"
	"net/http"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	commonv1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/common/v1"
	audiov1 "github.com/voxire/onion/internal/platform/models/proto/generated/proto/zontriq/audio/v1"
	"github.com/voxire/onion/internal/platforms/zontriq/audio/internal/application"
)

func ListAudioRoutes() []string {
	return []string{
		"/zontriq/audio",
	}
}

type ListAudioHandler struct {
	AudioService *application.Service
}

func (h *ListAudioHandler) Serve(ctx context.Context, req *audiov1.ListAudioRequest) (audiov1.ListAudioResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return audiov1.ListAudioResponse{}, &errorModels.ServerError{
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

	result, err := h.AudioService.ListAudio(ctx, application.ListAudioParams{
		UserID:   zontriqUserID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return audiov1.ListAudioResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list audio",
		}
	}

	protoClips := make([]*audiov1.AudioClip, 0, len(result.Clips))
	for i := range result.Clips {
		c := &result.Clips[i]
		protoClips = append(protoClips, &audiov1.AudioClip{
			Id:              int64(c.ID),
			Title:           c.Title,
			Script:          c.Script,
			VoiceId:         c.VoiceID,
			VoiceName:       c.VoiceName,
			Provider:        c.Provider,
			AudioUrl:        c.AudioUrl,
			DurationSeconds: c.DurationSeconds,
			State:           c.State,
			ErrorMessage:    c.ErrorMessage,
			CreatedAt:       c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	totalPages := int32(result.TotalCount) / pageSize
	if int32(result.TotalCount)%pageSize > 0 {
		totalPages++
	}

	return audiov1.ListAudioResponse{
		Clips: protoClips,
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
