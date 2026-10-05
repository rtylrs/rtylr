package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const asyncSeatChangeTimeout = 10 * time.Second

type seatChangeRecorder interface {
	BusinessOwnerID(ctx context.Context, fallbackOwnerID, businessID int64) (int64, error)
	RecordSeatChange(ctx context.Context, ownerID int64) error
}

func recordSeatChangeAsync(ctx context.Context, service seatChangeRecorder, actorID, businessID int64) {
	if service == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("billing seat change goroutine panicked", logger.Any("panic", r))
			}
		}()

		bgCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), asyncSeatChangeTimeout)
		defer cancel()

		ownerID, err := service.BusinessOwnerID(bgCtx, actorID, businessID)
		if err != nil {
			logger.Ctx(bgCtx).Error("failed to resolve owner for billing seat change",
				logger.Any("actor_id", actorID),
				logger.Error(err),
			)
			return
		}

		if err := service.RecordSeatChange(bgCtx, ownerID); err != nil {
			logger.Ctx(bgCtx).Error("failed to record seat change for billing",
				logger.Any("owner_id", ownerID),
				logger.Error(err),
			)
		}
	}()
}
