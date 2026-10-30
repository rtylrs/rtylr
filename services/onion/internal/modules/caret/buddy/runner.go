package buddy

import (
	"context"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	pollInterval    = 10 * time.Second
	maintainEvery   = time.Minute
	maxConcurrent   = 2
	runTimeout      = 5 * time.Minute
	runLease        = 10 * time.Minute
	finishTimeout   = 10 * time.Second
	reportPanicText = "the run crashed"
)

// Task is one unattended check handed to the agent.
type Task struct {
	RunID        int64
	BusinessID   int64
	UserID       int64
	Instructions string
}

type Outcome struct {
	Report    string
	ToolCalls int
	Proposals int
	Tokens    int64
}

// Agent runs a task with the identity already attached to the context.
type Agent interface {
	Run(ctx context.Context, task Task) (Outcome, error)
}

// StartRunner polls for queued runs until ctx ends. Runs are leased, so a
// process that dies mid-run hands its work to the next one after the lease.
func (m *Module) StartRunner(ctx context.Context, agent Agent, access authPkg.EffectiveAccessResolver) {
	if !Enabled() || agent == nil || access == nil {
		return
	}
	go m.runLoop(ctx, agent, access)
}

func (m *Module) runLoop(ctx context.Context, agent Agent, access authPkg.EffectiveAccessResolver) {
	slots := make(chan struct{}, maxConcurrent)
	poll := time.NewTicker(pollInterval)
	defer poll.Stop()
	maintain := time.NewTicker(maintainEvery)
	defer maintain.Stop()
	log := logger.Ctx(ctx)
	log.Info("caret buddy runner started")

	for {
		select {
		case <-ctx.Done():
			return
		case <-maintain.C:
			if err := m.repo.RecoverStaleRuns(ctx); err != nil {
				log.Warn("caret buddy: recover stale runs failed", logger.Error(err))
			}
			if err := m.repo.ExpireActions(ctx); err != nil {
				log.Warn("caret buddy: expire actions failed", logger.Error(err))
			}
		case <-poll.C:
			for m.claimAndRun(ctx, agent, access, slots) {
			}
		}
	}
}

// claimAndRun starts one queued run if a slot is free and reports whether it did.
func (m *Module) claimAndRun(ctx context.Context, agent Agent, access authPkg.EffectiveAccessResolver, slots chan struct{}) bool {
	select {
	case slots <- struct{}{}:
	default:
		return false
	}
	run, buddy, err := m.repo.ClaimRun(ctx, runLease)
	if err != nil || run == nil {
		<-slots
		if err != nil && ctx.Err() == nil {
			logger.Ctx(ctx).Warn("caret buddy: claim failed", logger.Error(err))
		}
		return false
	}
	go func() {
		defer func() { <-slots }()
		m.execute(ctx, agent, access, *run, *buddy)
	}()
	return true
}

func (m *Module) execute(parent context.Context, agent Agent, access authPkg.EffectiveAccessResolver, run domain.Run, buddy domain.Buddy) {
	result := m.perform(parent, agent, access, run, buddy)
	// The run may have been cut short by shutdown, so the result is saved on a
	// fresh context.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(parent), finishTimeout)
	defer cancel()
	if err := m.repo.FinishRun(ctx, run.ID, result); err != nil {
		logger.Ctx(ctx).Error("caret buddy: saving run failed", logger.Any("run_id", run.ID), logger.Error(err))
	}
}

func (m *Module) perform(parent context.Context, agent Agent, access authPkg.EffectiveAccessResolver, run domain.Run, buddy domain.Buddy) (result domain.RunResult) {
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Ctx(parent).Error("caret buddy: run panicked", logger.Any("run_id", run.ID), logger.Any("panic", recovered))
			result = domain.RunResult{Error: reportPanicText}
		}
	}()
	ctx, cancel := context.WithTimeout(parent, runTimeout)
	defer cancel()

	effective, err := access.GetEffectiveAccess(ctx, buddy.BusinessID, buddy.RunAsUserID)
	if err != nil || effective == nil {
		return domain.RunResult{Error: "the buddy's account no longer has access to this business"}
	}
	ctx = authPkg.WithAuthUser(ctx, &authPkg.AuthUser{
		UserID: buddy.RunAsUserID, BusinessID: buddy.BusinessID, EffectiveAccess: effective,
	})
	outcome, err := agent.Run(ctx, Task{RunID: run.ID, BusinessID: buddy.BusinessID, UserID: buddy.RunAsUserID, Instructions: buddy.Instructions})
	result = domain.RunResult{Report: outcome.Report, ToolCalls: outcome.ToolCalls, Tokens: outcome.Tokens}
	if err != nil {
		result.Error = trimError(err)
	}
	return result
}

func trimError(err error) string {
	message := err.Error()
	if errors.Is(err, context.DeadlineExceeded) {
		message = "the run took too long"
	}
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
