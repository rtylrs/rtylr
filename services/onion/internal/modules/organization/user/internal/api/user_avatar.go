package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/organization/user/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hubuserv1 "github.com/voxire/voxie/gen/go/hub/user/v1"
)

// UpdateAvatar sets the caller's avatar, or resets it to the default when the
// request carries none.
func (a *UserAPI) UpdateAvatar(ctx context.Context, req *hubuserv1.UpdateAvatarRequest) (*hubuserv1.UpdateAvatarResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var next *domain.Avatar
	if in := req.GetAvatar(); in != nil {
		converted := avatarFromProto(in)
		next = &converted
	}
	if _, err := a.UserService.UpdateAvatar(ctx, device.User.UserID, next); err != nil {
		if errors.Is(err, domain.ErrAvatarInvalid) {
			return nil, onekiterrors.New(http.StatusBadRequest, "That avatar can't be used")
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update avatar")
	}

	me, err := a.GetMe(ctx, &hubuserv1.GetMeRequest{})
	if err != nil {
		return nil, err
	}
	return &hubuserv1.UpdateAvatarResponse{User: me}, nil
}

// ListBusinessAvatars returns the chosen avatars of the caller's colleagues in
// their active business. An avatar is shown to everyone the person works with,
// so any member may read the business's set; nothing else about them is sent.
func (a *UserAPI) ListBusinessAvatars(ctx context.Context, _ *hubuserv1.ListBusinessAvatarsRequest) (*hubuserv1.ListBusinessAvatarsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if device.User.BusinessID <= 0 {
		return &hubuserv1.ListBusinessAvatarsResponse{Avatars: []*hubuserv1.BusinessUserAvatar{}}, nil
	}

	entries, err := a.UserService.ListBusinessAvatars(ctx, device.User.BusinessID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list avatars")
	}
	out := make([]*hubuserv1.BusinessUserAvatar, 0, len(entries))
	for _, entry := range entries {
		out = append(out, &hubuserv1.BusinessUserAvatar{
			UserId: entry.UserID,
			Avatar: avatarToProto(entry.Avatar),
		})
	}
	return &hubuserv1.ListBusinessAvatarsResponse{Avatars: out}, nil
}

// attachAvatar adds the user's chosen avatar to a profile response. A failed
// read leaves the default rather than failing the whole profile.
func (a *UserAPI) attachAvatar(ctx context.Context, response *hubuserv1.GetMeResponse) {
	if response == nil || a.UserService == nil {
		return
	}
	avatar, err := a.UserService.GetAvatar(ctx, response.Id)
	if err != nil {
		logger.Ctx(ctx).Warn("user: avatar lookup failed", logger.Error(err))
		return
	}
	if avatar != nil {
		response.Avatar = avatarToProto(*avatar)
	}
}

func avatarFromProto(in *hubuserv1.UserAvatar) domain.Avatar {
	out := domain.Avatar{Kind: in.GetKind(), Color: in.Color}
	if in.Url != nil {
		out.URL = *in.Url
	}
	if in.Style != nil {
		out.Style = *in.Style
	}
	if in.Seed != nil {
		out.Seed = *in.Seed
	}
	if in.Emoji != nil {
		out.Emoji = *in.Emoji
	}
	return out
}

func avatarToProto(in domain.Avatar) *hubuserv1.UserAvatar {
	out := &hubuserv1.UserAvatar{Kind: in.Kind, Color: in.Color}
	if in.URL != "" {
		out.Url = &in.URL
	}
	if in.Style != "" {
		out.Style = &in.Style
	}
	if in.Seed != "" {
		out.Seed = &in.Seed
	}
	if in.Emoji != "" {
		out.Emoji = &in.Emoji
	}
	return out
}
