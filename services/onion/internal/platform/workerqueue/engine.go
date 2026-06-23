package workerqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

type Message struct {
	ID         string
	Body       []byte
	ReceivedAt time.Time
	Ack        func(context.Context) error
	Nack       func(context.Context, error) error
}

type Queue interface {
	Name() string
	Receive(context.Context, int) ([]Message, error)
}

type HandlerFunc[T any] func(context.Context, T) error

type Engine struct {
	queue       Queue
	workers     int
	batchSize   int
	pollDelay   time.Duration
	handlerName string
	handler     func(context.Context, Message) error
}

type Option func(*Engine)

func WithWorkers(workers int) Option {
	return func(e *Engine) {
		if workers > 0 {
			e.workers = workers
		}
	}
}

func WithBatchSize(batchSize int) Option {
	return func(e *Engine) {
		if batchSize > 0 {
			e.batchSize = batchSize
		}
	}
}

func WithPollDelay(delay time.Duration) Option {
	return func(e *Engine) {
		if delay > 0 {
			e.pollDelay = delay
		}
	}
}

func New[T any](queue Queue, handler HandlerFunc[T], opts ...Option) *Engine {
	e := &Engine{
		queue:     queue,
		workers:   4,
		batchSize: 10,
		pollDelay: time.Second,
	}
	if handler != nil {
		e.handlerName = fmt.Sprintf("%T", *new(T))
		e.handler = func(ctx context.Context, msg Message) error {
			var payload T
			if err := json.Unmarshal(msg.Body, &payload); err != nil {
				return fmt.Errorf("decode %s: %w", e.handlerName, err)
			}
			return handler(ctx, payload)
		}
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

func (e *Engine) Run(ctx context.Context) error {
	if e == nil || e.queue == nil {
		return fmt.Errorf("queue is required")
	}
	if e.handler == nil {
		return fmt.Errorf("handler is required")
	}

	logger.Ctx(ctx).Info("worker queue starting",
		logger.String("queue", e.queue.Name()),
		logger.Int("workers", e.workers),
		logger.Int("batch_size", e.batchSize),
	)

	jobs := make(chan Message, e.batchSize*e.workers)
	var wg sync.WaitGroup
	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			e.worker(ctx, workerID, jobs)
		}(i + 1)
	}

	defer func() {
		close(jobs)
		wg.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		messages, err := e.queue.Receive(ctx, e.batchSize)
		if err != nil {
			logger.Ctx(ctx).Warn("queue receive failed",
				logger.String("queue", e.queue.Name()),
				logger.Error(err),
			)
			if !sleep(ctx, e.pollDelay) {
				return nil
			}
			continue
		}
		if len(messages) == 0 {
			if !sleep(ctx, e.pollDelay) {
				return nil
			}
			continue
		}
		for _, msg := range messages {
			select {
			case <-ctx.Done():
				return nil
			case jobs <- msg:
			}
		}
	}
}

func (e *Engine) worker(ctx context.Context, workerID int, jobs <-chan Message) {
	ctx = logger.WithLogger(ctx, logger.Ctx(ctx).With(
		"queue", e.queue.Name(),
		"worker_id", workerID,
	))

	for msg := range jobs {
		start := time.Now()
		msgCtx := logger.WithLogger(ctx, logger.Ctx(ctx).With(
			"message_id", msg.ID,
			"message_type", e.handlerName,
		))

		err := e.safeHandle(msgCtx, msg)
		if err != nil {
			logger.Ctx(msgCtx).Error("queue message failed",
				logger.Error(err),
				logger.Any("duration", time.Since(start).String()),
			)
			if msg.Nack != nil {
				_ = msg.Nack(msgCtx, err)
			}
			continue
		}
		if msg.Ack != nil {
			if ackErr := msg.Ack(msgCtx); ackErr != nil {
				logger.Ctx(msgCtx).Warn("queue message ack failed", logger.Error(ackErr))
			}
		}
		logger.Ctx(msgCtx).Info("queue message processed",
			logger.Any("duration", time.Since(start).String()),
		)
	}
}

func (e *Engine) safeHandle(ctx context.Context, msg Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handler panic: %v", r)
			logger.Ctx(ctx).Error("queue handler panicked",
				logger.Any("panic", r),
				logger.String("stack", string(debug.Stack())),
			)
		}
	}()
	return e.handler(ctx, msg)
}

func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
