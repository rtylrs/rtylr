package emailqueue

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/voxire/onion/internal/platform/email"
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
}

// Enqueue inserts a job into email_jobs to be sent by the pool.
func Enqueue(ctx context.Context, db *sql.DB, job Job) error {
	priority := job.Priority
	if priority <= 0 {
		priority = PriorityNormal
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO email_jobs (priority, to_email, subject, html, text_body, from_email, reply_to)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		priority, job.To, job.Subject, job.HTML, job.Text, job.From, job.ReplyTo,
	)
	return err
}

// Pool watches the email_jobs table and spawns up to maxWorkers goroutines
// to send emails concurrently. Each goroutine claims jobs via a unique job_id
// so no two goroutines process the same row.
type Pool struct {
	db     *sql.DB
	sem    chan struct{}
	active atomic.Int32
}

func NewPool(db *sql.DB) *Pool {
	return &Pool{
		db:  db,
		sem: make(chan struct{}, maxWorkers),
	}
}

func (p *Pool) Run(ctx context.Context) {
	logger.Ctx(ctx).Info("email pool started")

	watchTick := time.NewTicker(500 * time.Millisecond)
	defer watchTick.Stop()

	recoveryTick := time.NewTicker(60 * time.Second)
	defer recoveryTick.Stop()

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
		case <-watchTick.C:
			p.spawnWorkers(ctx)
		}
	}
}

func (p *Pool) spawnWorkers(ctx context.Context) {
	var pending int
	row := p.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM email_jobs WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW()`)
	if err := row.Scan(&pending); err != nil || pending == 0 {
		return
	}

	current := int(p.active.Load())
	want := min(pending, maxWorkers)
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
	claimID := uuid.New().String()

	res, err := p.db.ExecContext(ctx, `
		UPDATE email_jobs
		SET job_id = ?, status = 'processing'
		WHERE status = 'pending' AND job_id IS NULL AND scheduled_at <= NOW()
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT 1`,
		claimID,
	)
	if err != nil {
		return false
	}
	n, err := res.RowsAffected()
	if err != nil || n == 0 {
		return false
	}

	var (
		id        int64
		toEmail   string
		subject   string
		htmlBody  string
		textBody  sql.NullString
		fromEmail sql.NullString
		replyTo   sql.NullString
		attempts  int
		maxAtt    int
	)

	err = p.db.QueryRowContext(ctx, `
		SELECT id, to_email, subject, html, text_body, from_email, reply_to, attempts, max_attempts
		FROM email_jobs WHERE job_id = ?`, claimID,
	).Scan(&id, &toEmail, &subject, &htmlBody, &textBody, &fromEmail, &replyTo, &attempts, &maxAtt)
	if err != nil {
		return false
	}

	params := email.SendParams{
		To:      []string{toEmail},
		Subject: subject,
		HTML:    htmlBody,
	}
	if textBody.Valid {
		params.Text = textBody.String
	}
	if fromEmail.Valid && fromEmail.String != "" {
		params.From = fromEmail.String
	}
	if replyTo.Valid && replyTo.String != "" {
		params.ReplyTo = replyTo.String
	}

	_, sendErr := email.Send(ctx, params)

	if sendErr != nil {
		newAttempts := attempts + 1
		newStatus := "pending"
		var newJobID *string // nil resets job_id to NULL so it can be retried
		if newAttempts >= maxAtt {
			newStatus = "failed"
		}
		_, _ = p.db.ExecContext(ctx, `
			UPDATE email_jobs
			SET status = ?, attempts = ?, last_error = ?, job_id = ?,
			    scheduled_at = DATE_ADD(NOW(), INTERVAL ? SECOND)
			WHERE id = ?`,
			newStatus, newAttempts, sendErr.Error(), newJobID, retryDelay(newAttempts), id,
		)
	} else {
		_, _ = p.db.ExecContext(ctx, `
			UPDATE email_jobs SET status = 'sent', sent_at = NOW(), attempts = ?, job_id = NULL WHERE id = ?`,
			attempts+1, id,
		)
	}

	return true
}

func (p *Pool) recoverStuck(ctx context.Context) {
	res, err := p.db.ExecContext(ctx, `
		UPDATE email_jobs
		SET status = 'pending', job_id = NULL
		WHERE status = 'processing' AND updated_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE)`)
	if err != nil {
		logger.Ctx(ctx).Error("stuck job recovery failed", logger.Error(err))
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
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
