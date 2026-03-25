package async

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"codnect.io/chrono"
	"github.com/voxire/onion/internal/platform/logger"
)

type ScheduledFunc = func()

func ScheduleFunc(ctx context.Context, task ScheduledFunc, duration time.Duration, taskName string) {
	taskScheduler := chrono.NewDefaultTaskScheduler()

	_, err := taskScheduler.ScheduleAtFixedRate(func(context.Context) {
		task()
	}, duration)
	if err != nil {
		logger.Ctx(ctx).Warn(fmt.Sprintf("Error while scheduling '%s'", taskName),
			logger.Error(err),
		)
	}
}

type Provider[T any] interface {
	Get() T
	Refresh()
}

type ScheduledProvider[T any] struct {
	rwMu  sync.RWMutex
	fetch func(context.Context) (T, bool)
	data  T
}

func NewScheduledProvider[T any](ctx context.Context, fetch func(context.Context) (T, bool), duration time.Duration) *ScheduledProvider[T] {
	initialValue, success := fetch(ctx)
	if !success {
		logger.Ctx(ctx).Warn(fmt.Sprintf("Error while fetching initial value for '%s'", identifier[T]()))
	}

	provider := &ScheduledProvider[T]{
		fetch: fetch,
		data:  initialValue,
	}

	taskScheduler := chrono.NewDefaultTaskScheduler()

	_, err := taskScheduler.ScheduleAtFixedRate(func(context.Context) {
		provider.Refresh()
	},
		duration,
		chrono.WithTime(time.Now().Add(duration)),
	)
	if err != nil {
		logger.Ctx(ctx).Warn(fmt.Sprintf("Error while scheduling '%s'", identifier[T]()),
			logger.Error(err),
		)
	}

	return provider
}

func (provider *ScheduledProvider[T]) Get() T {
	provider.rwMu.RLock()
	defer provider.rwMu.RUnlock()

	return provider.data
}

func (provider *ScheduledProvider[T]) Refresh() {
	provider.rwMu.Lock()

	data, success := provider.fetch(context.Background())
	if success {
		provider.data = data
	}

	provider.rwMu.Unlock()
}

func identifier[T any]() string {
	return reflect.TypeOf(*new(T)).String()
}
