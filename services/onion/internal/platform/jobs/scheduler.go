package jobs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/voxire/onion/internal/platform/logger"
)

const (
	EnvSchedulerEnabled  = "SCHEDULER_ENABLED"
	EnvSchedulerDisabled = "SCHEDULER_DISABLED"
	EnvRunJob            = "RUN_JOB"
	EnvJobTimeout        = "JOB_TIMEOUT"

	DefaultJobTimeout = 30 * time.Minute
)

var (
	ErrJobAlreadyRunning  = errors.New("job already running")
	ErrJobLockNotAcquired = errors.New("job lock not acquired")
)

type Locker interface {
	TryLock(ctx context.Context, name string) (release func(context.Context) error, acquired bool, err error)
}

type Scheduler struct {
	registry       *Registry
	cron           *cron.Cron
	locker         Locker
	defaultTimeout time.Duration

	mu      sync.Mutex
	running map[string]struct{}
}

type SchedulerOption func(*Scheduler)

func NewScheduler(registry *Registry, opts ...SchedulerOption) *Scheduler {
	if registry == nil {
		registry = defaultRegistry
	}
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	s := &Scheduler{
		registry:       registry,
		cron:           cron.New(cron.WithParser(parser), cron.WithLocation(time.UTC)),
		defaultTimeout: DefaultJobTimeout,
		running:        make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithLocker(locker Locker) SchedulerOption {
	return func(s *Scheduler) {
		s.locker = locker
	}
}

func WithDefaultTimeout(timeout time.Duration) SchedulerOption {
	return func(s *Scheduler) {
		if timeout > 0 {
			s.defaultTimeout = timeout
		}
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("scheduler is nil")
	}

	registered := 0
	for _, job := range s.registry.Jobs() {
		if strings.TrimSpace(job.Schedule) == "" {
			continue
		}

		if _, err := s.cron.AddFunc(job.Schedule, func() {
			if _, err := s.Run(ctx, job.Name); err != nil {
				if errors.Is(err, ErrJobAlreadyRunning) || errors.Is(err, ErrJobLockNotAcquired) {
					logger.Ctx(ctx).Info("scheduled job skipped",
						logger.String("job_name", job.Name),
						logger.Error(err),
					)
					return
				}
				logger.Ctx(ctx).Error("scheduled job failed",
					logger.String("job_name", job.Name),
					logger.Error(err),
				)
			}
		}); err != nil {
			return fmt.Errorf("register scheduled job %q: %w", job.Name, err)
		}
		registered++
		logger.Ctx(ctx).Info("scheduled job registered",
			logger.String("job_name", job.Name),
			logger.String("schedule", job.Schedule),
		)
	}

	if registered == 0 {
		logger.Ctx(ctx).Warn("scheduler started with no scheduled jobs")
	}

	s.cron.Start()
	go func() {
		<-ctx.Done()
		stopCtx := s.cron.Stop()
		select {
		case <-stopCtx.Done():
			logger.Ctx(ctx).Info("scheduler stopped")
		case <-time.After(10 * time.Second):
			logger.Ctx(ctx).Warn("scheduler stop timed out")
		}
	}()

	logger.Ctx(ctx).Info("scheduler started", logger.Int("jobs", registered))
	return nil
}

func (s *Scheduler) Run(ctx context.Context, name string) (Result, error) {
	if s == nil {
		return Result{}, fmt.Errorf("scheduler is nil")
	}

	job, ok := s.registry.Get(name)
	if !ok {
		return Result{}, fmt.Errorf("unknown job %q", name)
	}

	if !s.tryStartLocal(name) {
		return Result{}, fmt.Errorf("%w: %s", ErrJobAlreadyRunning, name)
	}
	defer s.finishLocal(name)

	var release func(context.Context) error
	if s.locker != nil {
		var acquired bool
		var err error
		release, acquired, err = s.locker.TryLock(ctx, name)
		if err != nil {
			return Result{}, fmt.Errorf("lock job %q: %w", name, err)
		}
		if !acquired {
			return Result{}, fmt.Errorf("%w: %s", ErrJobLockNotAcquired, name)
		}
	}
	if release != nil {
		defer func() {
			if err := release(context.Background()); err != nil {
				logger.Ctx(ctx).Warn("job lock release failed",
					logger.String("job_name", name),
					logger.Error(err),
				)
			}
		}()
	}

	timeout := job.Timeout
	if timeout <= 0 {
		timeout = s.defaultTimeout
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return s.registry.Run(ctx, name)
}

func (s *Scheduler) tryStartLocal(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.running[name]; ok {
		return false
	}
	s.running[name] = struct{}{}
	return true
}

func (s *Scheduler) finishLocal(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, name)
}

func SchedulerEnabledFromEnv() bool {
	if disabled, ok := boolEnv(EnvSchedulerDisabled); ok && disabled {
		return false
	}
	if enabled, ok := boolEnv(EnvSchedulerEnabled); ok {
		return enabled
	}
	return true
}

func RunJobFromEnv() string {
	return strings.TrimSpace(os.Getenv(EnvRunJob))
}

func DurationFromEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err == nil {
		return parsed
	}
	if seconds, convErr := strconv.Atoi(value); convErr == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func ScheduleFromEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func boolEnv(key string) (bool, bool) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
}
