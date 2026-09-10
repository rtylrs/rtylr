package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"time"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/emailqueue"
	"github.com/voxire/onion/internal/platform/logger"
)

// defaultRunBatch bounds one pass. Each run is claimed and executed
// individually, so this governs how much work a single tick takes rather than
// how long anything is locked.
const defaultRunBatch = 50

// baseRetrySeconds is the first backoff step. Doubling from here gives
// 1, 2, 4 minutes across the default three attempts.
const baseRetrySeconds = 60

// maxRetrySeconds caps the backoff so a long-lived run does not drift to a
// retry hours away.
const maxRetrySeconds = 3600

// abandonedAfter is how long a claimed run may sit before it is treated as
// abandoned. Comfortably longer than any legitimate execution, which is a
// handful of queries and an enqueue, so a slow run is never mistaken for a
// dead worker.
const abandonedAfter = 15 * time.Minute

// errPermanent marks a failure that retrying cannot fix. Wrapping rather than
// a sentinel value so the reason survives to last_error, where an operator
// reads it.
type errPermanent struct{ cause error }

func (e errPermanent) Error() string { return e.cause.Error() }
func (e errPermanent) Unwrap() error { return e.cause }

func permanent(format string, args ...any) error {
	return errPermanent{cause: fmt.Errorf(format, args...)}
}

// Executor drains the automation run queue.
//
// It lives in this module rather than beside the dispatcher because the
// automation definition and its run_count belong to crm_automations, which
// this module owns. The queue itself is a platform table and is read here as
// an annotated cross-schema read.
type Executor struct {
	repo      *infrastructure.Repository
	db        *sql.DB
	batchSize int32
}

// NewExecutor builds an executor over the shared pool.
func NewExecutor(db *sql.DB) *Executor {
	return &Executor{
		repo:      infrastructure.NewRepository(db),
		db:        db,
		batchSize: defaultRunBatch,
	}
}

// ProcessDue executes one batch of due runs.
//
// A run that fails is recorded against itself and the batch continues, so one
// broken automation cannot stop every other tenant's from running.
func (e *Executor) ProcessDue(ctx context.Context) error {
	e.recoverAbandoned(ctx)

	runs, err := e.repo.ListDueRuns(ctx, e.batchSize)
	if err != nil {
		return fmt.Errorf("listing due automation runs: %w", err)
	}

	for _, run := range runs {
		if claimErr := e.repo.ClaimRun(ctx, run.ID); claimErr != nil {
			if errors.Is(claimErr, infrastructure.ErrRunNotClaimed) {
				continue
			}
			logger.Ctx(ctx).Error("automations: failed to claim run",
				logger.Any("run_id", run.ID), logger.Error(claimErr))
			continue
		}

		if runErr := e.execute(ctx, run); runErr != nil {
			e.recordFailure(ctx, run, runErr)
			continue
		}

		if markErr := e.repo.MarkRunSucceeded(ctx, run.ID); markErr != nil {
			logger.Ctx(ctx).Error("automations: run succeeded but could not be marked",
				logger.Any("run_id", run.ID), logger.Error(markErr))
		}
	}
	return nil
}

// recoverAbandoned requeues runs left in 'running' by a worker that died
// between claiming a row and recording its outcome. Without it those rows are
// never looked at again, because the due query only reads 'pending' and
// 'failed'.
//
// These used to be dead-lettered, because re-running could send a customer a
// second copy of a message they already had. Every effect an action performs
// now carries a key naming it, so a repeat is recognised and skipped, and the
// run can simply be tried again.
//
// Runs here at the start of every pass rather than as a separate job: it is a
// single indexed UPDATE that matches nothing in normal operation, and tying it
// to the same cadence means there is no second schedule to get wrong.
//
// A failure to recover does not stop the pass. The due runs are still worth
// processing, and the abandoned ones will still be there next time.
func (e *Executor) recoverAbandoned(ctx context.Context) {
	recovered, err := e.repo.RecoverAbandoned(ctx, int64(abandonedAfter.Seconds()))
	if err != nil {
		logger.Ctx(ctx).Error("automations: could not recover abandoned runs", logger.Error(err))
		return
	}
	// Warn rather than info: a non-zero count means a worker died mid-run, and
	// each one is an automation that did not finish.
	if recovered > 0 {
		logger.Ctx(ctx).Warn("automations: requeued abandoned runs",
			logger.Any("runs", recovered))
	}
}

// recordFailure routes a failure to a retry or to the dead letter, and never
// lets the bookkeeping failure hide the original cause.
func (e *Executor) recordFailure(ctx context.Context, run domain.Run, cause error) {
	logger.Ctx(ctx).Error("automations: run failed",
		logger.Any("run_id", run.ID),
		logger.Any("automation_id", run.AutomationID),
		logger.Error(cause))

	var permanentErr errPermanent
	if errors.As(cause, &permanentErr) {
		if err := e.repo.MarkRunDead(ctx, run.ID, cause.Error()); err != nil {
			logger.Ctx(ctx).Error("automations: could not dead-letter run",
				logger.Any("run_id", run.ID), logger.Error(err))
		}
		return
	}

	if err := e.repo.MarkRunFailed(ctx, run.ID, retryDelaySeconds(run.Attempts), cause.Error()); err != nil {
		logger.Ctx(ctx).Error("automations: could not reschedule run",
			logger.Any("run_id", run.ID), logger.Error(err))
	}
}

// retryDelaySeconds doubles from the base with each attempt, capped.
func retryDelaySeconds(attempts int) int64 {
	delay := int64(baseRetrySeconds)
	for i := 1; i < attempts && delay < maxRetrySeconds; i++ {
		delay *= 2
	}
	if delay > maxRetrySeconds {
		return maxRetrySeconds
	}
	return delay
}

// execute runs one automation against the event that triggered it.
func (e *Executor) execute(ctx context.Context, run domain.Run) error {
	automation, err := e.repo.GetAutomationForRun(ctx, run.AutomationID, run.BusinessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return permanent("automation %d no longer exists in business %d", run.AutomationID, run.BusinessID)
		}
		return fmt.Errorf("loading automation: %w", err)
	}
	// Switched off between the event and now. Terminal rather than retryable:
	// nothing about waiting makes a disabled automation run.
	if !automation.IsActive {
		return permanent("automation %d is not active", run.AutomationID)
	}

	event, err := e.repo.GetEventForRun(ctx, run.EventID, run.BusinessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return permanent("event %d no longer exists in business %d", run.EventID, run.BusinessID)
		}
		return fmt.Errorf("loading event: %w", err)
	}

	actions, err := planActions(automation.Actions)
	if err != nil {
		return err
	}

	// Every action is checked before any is performed. A partly executed
	// automation is worse than one that did not run: the operator cannot tell
	// which half happened, and a retry would repeat it.
	for _, action := range actions {
		if !isExecutable(action.Type) {
			return permanent("action %q is declared but not implemented by this build", action.Type)
		}
	}

	for index, action := range actions {
		if err := e.perform(ctx, run, event, action, index); err != nil {
			return err
		}
	}

	// Only after every action succeeded, so the counters mean what they say.
	if err := e.repo.RecordExecution(ctx, run.AutomationID, run.BusinessID); err != nil {
		logger.Ctx(ctx).Error("automations: executed but could not record the run",
			logger.Any("automation_id", run.AutomationID), logger.Error(err))
	}
	return nil
}

// plannedAction is one action from a stored definition.
type plannedAction struct {
	Type    string
	Subject string
	Body    string
}

// isExecutable reports whether this build can perform an action. The domain
// declares four; only send_email is implemented, and the rest fail loudly
// rather than being skipped, because silently doing nothing is the complaint
// in #834.
func isExecutable(actionType string) bool {
	return actionType == domain.ActionSendEmail
}

// planActions canonicalises a stored definition and reads the fields the
// executor understands. Reuses the module's own normalizer so the legacy
// bare-string form is handled in one place.
func planActions(raw []byte) ([]plannedAction, error) {
	normalized, err := domain.NormalizeAutomationActions(raw)
	if err != nil {
		return nil, permanent("stored actions are invalid: %v", err)
	}

	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(normalized, &entries); err != nil {
		return nil, permanent("stored actions could not be read: %v", err)
	}
	if len(entries) == 0 {
		return nil, permanent("automation has no actions")
	}

	planned := make([]plannedAction, 0, len(entries))
	for _, entry := range entries {
		action := plannedAction{
			Type:    stringField(entry, "type"),
			Subject: stringField(entry, "subject"),
			Body:    stringField(entry, "body"),
		}
		planned = append(planned, action)
	}
	return planned, nil
}

func stringField(entry map[string]json.RawMessage, key string) string {
	raw, ok := entry[key]
	if !ok {
		return ""
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return value
}

// perform carries out a single action.
//
// index identifies the action within its automation, and together with the run
// id it names the effect so a repeat can be recognised. That is what makes an
// abandoned run safe to retry.
func (e *Executor) perform(ctx context.Context, run domain.Run, event domain.Event, action plannedAction, index int) error {
	switch action.Type {
	case domain.ActionSendEmail:
		return e.sendEmail(ctx, run, event, action, index)
	default:
		// Unreachable: isExecutable gates this above. Kept so a new action
		// added to the switch without a matching isExecutable entry fails
		// loudly rather than silently doing nothing.
		return permanent("action %q has no implementation", action.Type)
	}
}

// sendEmail queues a message to the customer the event is about.
//
// The email is enqueued rather than sent, so this hands off to the existing
// email worker and does not block the run on an SMTP round trip.
func (e *Executor) sendEmail(ctx context.Context, run domain.Run, event domain.Event, action plannedAction, index int) error {
	customerID := payloadInt64(event.Payload, "customer_id")
	if customerID <= 0 {
		return permanent("send_email needs a customer and event %d has none", event.ID)
	}

	contact, err := e.repo.GetCustomerContact(ctx, customerID, run.BusinessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return permanent("customer %d is not in business %d", customerID, run.BusinessID)
		}
		return fmt.Errorf("loading customer: %w", err)
	}
	if contact.Email == "" {
		return permanent("customer %d has no email address", customerID)
	}

	subject := action.Subject
	if subject == "" {
		subject = "Thank you for your order"
	}
	body := action.Body
	if body == "" {
		body = "Thank you for your order."
	}

	// Escaped because both fields are operator-authored text stored as JSON,
	// not markup, and they end up in an HTML body.
	//
	// The dedupe key names this exact effect: one action of one run. It is
	// stable across retries because a recovered run keeps its id, so an
	// abandoned run that had already queued this message enqueues nothing the
	// second time. Without it a retry would send the customer a second copy,
	// which is why abandoned runs used to be dead-lettered instead.
	return emailqueue.Enqueue(ctx, e.db, emailqueue.Job{
		To:        contact.Email,
		Subject:   subject,
		HTML:      fmt.Sprintf("<p>%s</p>", html.EscapeString(body)),
		Text:      body,
		DedupeKey: fmt.Sprintf("automation_run:%d:action:%d", run.ID, index),
	})
}

// payloadInt64 reads a numeric field from a producer's JSON payload. Accepts
// the string form too, because an int64 crosses the wire as a string in this
// product and a producer may well have written one.
func payloadInt64(payload []byte, key string) int64 {
	if len(payload) == 0 {
		return 0
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(payload, &fields) != nil {
		return 0
	}
	raw, ok := fields[key]
	if !ok {
		return 0
	}
	var number int64
	if json.Unmarshal(raw, &number) == nil {
		return number
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		var parsed int64
		if _, err := fmt.Sscanf(text, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return 0
}
