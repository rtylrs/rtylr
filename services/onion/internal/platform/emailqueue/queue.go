package emailqueue

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/platform/email"
	"github.com/voxire/onion/internal/platform/logger"
)

// Priority levels — higher number = sent first.
const (
	PriorityCritical = 10 // password reset, security alerts
	PriorityHigh     = 5  // invitations, billing emails
	PriorityNormal   = 1  // welcome emails, marketing
)

// Job holds everything needed to send one email.
type Job struct {
	Priority  int
	To        string
	Subject   string
	HTML      string
	Text      string
	From      string
	ReplyTo   string
}

// Enqueue inserts a job into email_jobs to be sent by the worker.
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

// Worker polls email_jobs and sends pending emails in priority order.
// Call Run in a goroutine; it blocks until ctx is cancelled.
type Worker struct {
	db       *sql.DB
	interval time.Duration
}

func NewWorker(db *sql.DB, interval time.Duration) *Worker {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &Worker{db: db, interval: interval}
}

func (w *Worker) Run(ctx context.Context) {
	logger.Ctx(ctx).Info("email queue worker started")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Ctx(ctx).Info("email queue worker stopped")
			return
		case <-ticker.C:
			w.flush(ctx)
		}
	}
}

func (w *Worker) flush(ctx context.Context) {
	for {
		sent := w.processOne(ctx)
		if !sent {
			return
		}
	}
}

func (w *Worker) processOne(ctx context.Context) bool {
	tx, err := w.db.BeginTx(ctx, nil)
	if err != nil {
		return false
	}
	defer func() { _ = tx.Rollback() }()

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

	err = tx.QueryRowContext(ctx, `
		SELECT id, to_email, subject, html, text_body, from_email, reply_to, attempts, max_attempts
		FROM email_jobs
		WHERE status = 'pending' AND scheduled_at <= NOW()
		ORDER BY priority DESC, scheduled_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED`,
	).Scan(&id, &toEmail, &subject, &htmlBody, &textBody, &fromEmail, &replyTo, &attempts, &maxAtt)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		return false
	}

	_, err = tx.ExecContext(ctx, `UPDATE email_jobs SET status='processing' WHERE id=?`, id)
	if err != nil {
		return false
	}
	if err = tx.Commit(); err != nil {
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
		if newAttempts >= maxAtt {
			newStatus = "failed"
		}
		_, _ = w.db.ExecContext(ctx, `
			UPDATE email_jobs
			SET status=?, attempts=?, last_error=?, scheduled_at=DATE_ADD(NOW(), INTERVAL ? SECOND)
			WHERE id=?`,
			newStatus, newAttempts, sendErr.Error(), retryDelay(newAttempts), id,
		)
	} else {
		_, _ = w.db.ExecContext(ctx, `
			UPDATE email_jobs SET status='sent', sent_at=NOW(), attempts=? WHERE id=?`,
			attempts+1, id,
		)
	}

	return true
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
