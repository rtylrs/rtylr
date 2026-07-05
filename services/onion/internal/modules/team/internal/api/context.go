package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/team/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func teamOwnerIDFromCtx(ctx context.Context, service *application.Service) (int64, bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, false
	}
	ownerID, err := service.BusinessOwnerID(ctx, device.User.UserID, device.User.BusinessID)
	return ownerID, err == nil && ownerID != 0
}

func teamManagerFromCtx(ctx context.Context, service *application.Service) (actorID, ownerID int64, ok bool) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return 0, 0, false
	}
	ownerID, err := service.TeamManagerOwnerID(ctx, device.User.UserID, device.User.BusinessID)
	return device.User.UserID, ownerID, err == nil && ownerID != 0
}
