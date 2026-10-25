package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const asyncSeatChangeTimeout = 10 * time.Second

type seatChangeRecorder interface {
	RecordSeatChange(ctx context.Context, businessID int64) error
}

func recordSeatChangeAsync(ctx context.Context, service seatChangeRecorder, businessID int64) {
	if service == nil || businessID <= 0 {
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

		if err := service.RecordSeatChange(bgCtx, businessID); err != nil {
			logger.Ctx(bgCtx).Error("failed to record seat change for billing",
				logger.Any("business_id", businessID),
				logger.Error(err),
			)
		}
	}()
}
