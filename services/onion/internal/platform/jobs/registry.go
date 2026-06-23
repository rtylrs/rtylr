package jobs

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"
	"sync"
	"time"

	"github.com/voxire/onion/internal/platform/logger"
)

type Handler func(context.Context) error

type Job struct {
	Name        string
	Description string
	Schedule    string
	Timeout     time.Duration
	Handler     Handler
}

type Result struct {
	Name      string        `json:"name"`
	Duration  time.Duration `json:"duration"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   time.Time     `json:"ended_at"`
}

type Registry struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func NewRegistry() *Registry {
	return &Registry{jobs: make(map[string]Job)}
}

var defaultRegistry = NewRegistry()

func Register(job Job) error {
	return defaultRegistry.Register(job)
}

func Names() []string {
	return defaultRegistry.Names()
}

func Jobs() []Job {
	return defaultRegistry.Jobs()
}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func Run(ctx context.Context, name string) (Result, error) {
	return defaultRegistry.Run(ctx, name)
}

func (r *Registry) Register(job Job) error {
	if r == nil {
		return fmt.Errorf("jobs registry is nil")
	}
	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}
	if job.Handler == nil {
		return fmt.Errorf("job %q handler is required", job.Name)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.Name] = job
	return nil
}

func (r *Registry) Names() []string {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.jobs))
	for name := range r.jobs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) Jobs() []Job {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Job, 0, len(r.jobs))
	for _, job := range r.jobs {
		items = append(items, job)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

func (r *Registry) Get(name string) (Job, bool) {
	if r == nil {
		return Job{}, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[name]
	return job, ok
}

func (r *Registry) Run(ctx context.Context, name string) (result Result, err error) {
	if r == nil {
		return Result{}, fmt.Errorf("jobs registry is nil")
	}

	r.mu.RLock()
	job, ok := r.jobs[name]
	r.mu.RUnlock()
	if !ok {
		return Result{}, fmt.Errorf("unknown job %q", name)
	}

	started := time.Now().UTC()
	ctx = logger.WithLogger(ctx, logger.Ctx(ctx).With("job_name", name))
	logger.Ctx(ctx).Info("job started")

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job %q panicked: %v", name, r)
			logger.Ctx(ctx).Error("job panicked",
				logger.Any("panic", r),
				logger.String("stack", string(debug.Stack())),
			)
		}
		ended := time.Now().UTC()
		result = Result{
			Name:      name,
			StartedAt: started,
			EndedAt:   ended,
			Duration:  ended.Sub(started),
		}
		if err != nil {
			logger.Ctx(ctx).Error("job failed", logger.Error(err), logger.Any("duration", result.Duration.String()))
		} else {
			logger.Ctx(ctx).Info("job completed", logger.Any("duration", result.Duration.String()))
		}
	}()

	err = job.Handler(ctx)
	return result, err
}
