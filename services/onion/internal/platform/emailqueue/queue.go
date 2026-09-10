package emailqueue

import (
	"context"
	"database/sql"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/email"
	sqlc "github.com/voxire/onion/internal/platform/emailqueue/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	PriorityCritical = 10
	PriorityHigh     = 5
	PriorityNormal   = 1
	maxWorkers       = 50
)

// Job holds everything needed to send one email.
type Job struct {
	Priority int
	To       string
	Subject  string
	HTML     string
	Text     string
	From     string
	ReplyTo  string
	// DedupeKey makes the enqueue idempotent. Left empty, the message is
	// queued unconditionally, which is what every existing caller wants. Set,
	// a second enqueue under the same key is a no-op, so a producer that is
	// retried does not send the recipient a second copy.
	DedupeKey string
}

// nullStr wraps a string as a NULL-able column value (empty → NULL).
func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// Enqueue inserts a job into email_jobs to be sent by the pool.
func Enqueue(ctx context.Context, db *sql.DB, job Job) error {
	priority := job.Priority
	if priority <= 0 {
		priority = PriorityNormal
	}
	// Reset and invite mails carry live account-takeover links; the rendered
	// body is encrypted at rest whenever EMAIL_ENCRYPTION_KEY is configured.
	//
	// A caller that can name the message enqueues idempotently instead, so a
	// retry does not send a second copy. Everything else keeps a NULL key and
	// is unaffected, since NULL does not collide in a unique index.
	if job.DedupeKey != "" {
		return sqlc.New(db).EnqueueJobWithDedupeKey(ctx, sqlc.EnqueueJobWithDedupeKeyParams{
			Priority:  int8(priority),
			ToEmail:   job.To,
			Subject:   job.Subject,
			Html:      sealHTML(job.HTML),
			TextBody:  nullStr(job.Text),
			FromEmail: nullStr(job.From),
			ReplyTo:   nullStr(job.ReplyTo),
			DedupeKey: nullStr(job.DedupeKey),
		})
	}

	return sqlc.New(db).EnqueueJob(ctx, sqlc.EnqueueJobParams{
		Priority:  int8(priority),
		ToEmail:   job.To,
		Subject:   job.Subject,
		Html:      sealHTML(job.HTML),
		TextBody:  nullStr(job.Text),
		FromEmail: nullStr(job.From),
		ReplyTo:   nullStr(job.ReplyTo),
	})
}

// Pool watches the email_jobs table and spawns up to maxWorkers goroutines
// to send emails concurrently. Each goroutine claims jobs via a unique job_id
// so no two goroutines process the same row.
type Pool struct {
	db     *sql.DB
	q      *sqlc.Queries
	sem    chan struct{}
	active atomic.Int32
}

func NewPool(db *sql.DB) *Pool {
	return &Pool{
		db:  db,
		q:   sqlc.New(db),
		sem: make(chan struct{}, maxWorkers),
	}
}

func (p *Pool) Run(ctx context.Context) {
	logger.Ctx(ctx).Info("email pool started")

	watchTick := time.NewTicker(500 * time.Millisecond)
	defer watchTick.Stop()

	recoveryTick := time.NewTicker(60 * time.Second)
	defer recoveryTick.Stop()

	retentionTick := time.NewTicker(retentionInterval)
	defer retentionTick.Stop()

	for {
		select {
		case <-ctx.Done():
			// Drain semaphore — wait for all active goroutines to finish.
			for i := 0; i < maxWorkers; i++ {
				p.sem <- struct{}{}
			}
			logger.Ctx(ctx).Info("email pool stopped")
			return
		case <-recoveryTick.C:
			p.recoverStuck(ctx)
		case <-retentionTick.C:
			p.CleanupOld(ctx, defaultRetention)
		case <-watchTick.C:
			p.spawnWorkers(ctx)
		}
	}
}

func (p *Pool) spawnWorkers(ctx context.Context) {
	pending, err := p.q.CountPendingJobs(ctx)
	if err != nil || pending == 0 {
		return
	}

	current := int(p.active.Load())
	want := min(int(pending), maxWorkers)
	toSpawn := want - current
	if toSpawn <= 0 {
		return
	}

	for i := 0; i < toSpawn; i++ {
		select {
		case p.sem <- struct{}{}:
			p.active.Add(1)
			go func() {
				defer func() {
					<-p.sem
					p.active.Add(-1)
				}()
				p.drain(ctx)
			}()
		default:
			return
		}
	}
}

func (p *Pool) drain(ctx context.Context) {
	for ctx.Err() == nil {
		if !p.processOne(ctx) {
			return
		}
	}
}

func (p *Pool) processOne(ctx context.Context) bool {
	claimID := sql.NullString{String: uuid.New().String(), Valid: true}

	res, err := p.q.ClaimJob(ctx, claimID)
	if err != nil {
		return false
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return false
	}

	job, err := p.q.GetClaimedJob(ctx, claimID)
	if err != nil {
		p.releaseClaim(ctx, claimID)
		return false
	}
	// An encrypted body that cannot be decrypted must not loop forever: mark
	// the job failed instead of releasing the claim back to pending.
	html := openHTML(job.Html)
	if job.Html != "" && html == "" {
		if failErr := p.q.RescheduleJob(ctx, sqlc.RescheduleJobParams{
			Status:    sqlc.EmailJobsStatusFailed,
			Attempts:  job.MaxAttempts,
			LastError: sql.NullString{String: "body undecryptable (EMAIL_ENCRYPTION_KEY missing or changed)", Valid: true},
			JobID:     sql.NullString{},
			DATEADD:   0,
			ID:        job.ID,
		}); failErr != nil {
			logger.Ctx(ctx).Error("failed to mark undecryptable email job as failed",
				logger.Any("job_id", job.ID), logger.Error(failErr))
		}
		return true
	}
	params := email.SendParams{
		To:      []string{job.ToEmail},
		Subject: job.Subject,
		HTML:    html,
	}
	if job.TextBody.Valid {
		params.Text = job.TextBody.String
	}
	if job.FromEmail.Valid && job.FromEmail.String != "" {
		params.From = job.FromEmail.String
	}
	if job.ReplyTo.Valid && job.ReplyTo.String != "" {
		params.ReplyTo = job.ReplyTo.String
	}

	_, sendErr := email.Send(ctx, params)

	if sendErr != nil {
		var rateLimitErr *email.RateLimitError
		if errors.As(sendErr, &rateLimitErr) {
			delay := int(rateLimitErr.RetryAfter / time.Second)
			if delay < 1 {
				delay = 1
			}
			// Count the attempt even on rate limits so a permanently
			// throttled job still terminates after MaxAttempts.
			newAttempts := int(job.Attempts) + 1
			newStatus := sqlc.EmailJobsStatusPending
			if newAttempts >= int(job.MaxAttempts) {
				newStatus = sqlc.EmailJobsStatusFailed
			}
			if err := p.q.RescheduleJob(ctx, sqlc.RescheduleJobParams{
				Status:    newStatus,
				Attempts:  uint8(newAttempts), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
				LastError: sql.NullString{String: sendErr.Error(), Valid: true},
				JobID:     sql.NullString{},
				DATEADD:   delay,
				ID:        job.ID,
			}); err != nil {
				logger.Ctx(ctx).Error("failed to reschedule rate-limited email job",
					logger.Any("job_id", job.ID), logger.Error(err))
			}
			return true
		}
		if errors.Is(sendErr, context.Canceled) || errors.Is(sendErr, context.DeadlineExceeded) {
			p.releaseClaim(ctx, claimID)
			return false
		}
		newAttempts := int(job.Attempts) + 1
		newStatus := sqlc.EmailJobsStatusPending // NULL job_id below makes it retryable
		if newAttempts >= int(job.MaxAttempts) {
			newStatus = sqlc.EmailJobsStatusFailed
		}
		if err := p.q.RescheduleJob(ctx, sqlc.RescheduleJobParams{
			Status:    newStatus,
			Attempts:  uint8(newAttempts), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			LastError: sql.NullString{String: sendErr.Error(), Valid: true},
			JobID:     sql.NullString{}, // NULL resets job_id so it can be retried
			DATEADD:   retryDelay(newAttempts),
			ID:        job.ID,
		}); err != nil {
			logger.Ctx(ctx).Error("failed to reschedule email job after send error",
				logger.Any("job_id", job.ID),
				logger.Error(err),
			)
		}
	} else {
		if err := p.q.MarkJobSent(ctx, sqlc.MarkJobSentParams{
			Attempts: uint8(int(job.Attempts) + 1), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			ID:       job.ID,
		}); err != nil {
			logger.Ctx(ctx).Error("failed to mark email job sent",
				logger.Any("job_id", job.ID),
				logger.Error(err),
			)
		}
	}

	return true
}

func (p *Pool) releaseClaim(_ context.Context, claimID sql.NullString) {
	releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := p.q.ReleaseClaim(releaseCtx, claimID); err != nil {
		logger.Ctx(releaseCtx).Error("failed to release email job claim", logger.Error(err))
	}
}

func (p *Pool) recoverStuck(ctx context.Context) {
	res, err := p.q.RecoverStuckJobs(ctx)
	if err != nil {
		logger.Ctx(ctx).Error("stuck job recovery failed", logger.Error(err))
		return
	}
	if n, err := res.RowsAffected(); err != nil {
		logger.Ctx(ctx).Warn("could not read stuck job recovery row count", logger.Error(err))
	} else if n > 0 {
		logger.Ctx(ctx).Info("recovered stuck email jobs", logger.Int("count", int(n)))
	}
}

func retryDelay(attempt int) int {
	switch attempt {
	case 1:
		return 30
	case 2:
		return 120
	default:
		return 300
	}
}
