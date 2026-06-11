package handlers

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const asyncUserHookTimeout = 10 * time.Second

func runAsyncUserHook(ctx context.Context, userID int64, hook func(context.Context, int64)) {
	if hook == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("user async hook panicked", logger.Any("panic", r))
			}
		}()

		hookCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), asyncUserHookTimeout)
		defer cancel()
		hook(hookCtx, userID)
	}()
}
