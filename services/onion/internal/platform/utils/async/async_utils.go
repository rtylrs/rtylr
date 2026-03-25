package async

import (
	"context"
	"sync"

	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/utils/maputils"
)

func CombineMaps[K comparable, V any](steps []map[K]V) map[K]V {
	return maputils.Merge(steps...)
}

type FutureResult[T any] struct {
	wg             sync.WaitGroup
	tasks          []func() T
	reduceFunction func([]T) T
}

func NewFutureResult[T any](reduceFunction func([]T) T) *FutureResult[T] {
	return &FutureResult[T]{
		wg:             sync.WaitGroup{},
		tasks:          []func() T{},
		reduceFunction: reduceFunction,
	}
}

func (f *FutureResult[T]) AddTask(task func() T) {
	f.tasks = append(f.tasks, task)
}

func (f *FutureResult[T]) AddTasks(task ...func() T) {
	f.tasks = append(f.tasks, task...)
}

func (f *FutureResult[T]) Execute() T {
	ctx := context.Background()
	myListOfT := make([]T, len(f.tasks))

	for index, task := range f.tasks {
		i := index

		f.wg.Add(1)

		queuedTask := task

		go func() {
			defer f.panicRecovery(ctx)
			defer f.wg.Done()

			myListOfT[i] = queuedTask()
		}()
	}

	f.wg.Wait()

	return f.reduceFunction(myListOfT)
}

func (f *FutureResult[T]) panicRecovery(ctx context.Context) {
	if r := recover(); r != nil {
		logger.Ctx(ctx).Error("Async func panicked",
			logger.Any("panic", r))
	}
}

type (
	Func   func()
	Future struct {
		wg    sync.WaitGroup
		tasks []Func
	}
)

func NewFuture() *Future {
	return &Future{
		wg:    sync.WaitGroup{},
		tasks: []Func{},
	}
}

func (f *Future) AddTask(task Func) {
	f.tasks = append(f.tasks, task)
}

func (f *Future) AddTasks(task ...Func) {
	f.tasks = append(f.tasks, task...)
}

func (f *Future) Execute() {
	ctx := context.Background()

	for _, task := range f.tasks {
		f.wg.Add(1)

		queuedTask := task

		go func() {
			defer f.panicRecovery(ctx)
			defer f.wg.Done()

			queuedTask()
		}()
	}

	f.wg.Wait()
}

func (f *Future) panicRecovery(ctx context.Context) {
	if r := recover(); r != nil {
		logger.Ctx(ctx).Error("Async func panicked",
			logger.Any("panic", r))
	}
}
