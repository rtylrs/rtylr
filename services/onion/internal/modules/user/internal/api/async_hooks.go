package api

import (
	"context"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

const asyncUserHookTimeout = 10 * time.Second

type loginRecorder interface {
	RecordLogin(ctx context.Context, userID int64, email, method, ip, userAgent string, success bool) error
}

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

func runAsyncLoginTasks(ctx context.Context, recorder loginRecorder, userID int64, email, ip, userAgent string, hook func(context.Context, int64)) {
	if recorder == nil && hook == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Global().Error("login async task panicked", logger.Any("panic", r))
			}
		}()

		taskCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), asyncUserHookTimeout)
		defer cancel()

		if recorder != nil {
			if err := recorder.RecordLogin(taskCtx, userID, email, "email", ip, userAgent, true); err != nil {
				logger.Ctx(taskCtx).Warn("failed to record login", logger.Int("user_id", int(userID)), logger.Error(err))
			}
		}

		if hook != nil {
			hook(taskCtx, userID)
		}
	}()
}
