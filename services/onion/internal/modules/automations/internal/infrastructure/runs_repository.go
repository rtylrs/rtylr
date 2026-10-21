package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/automations/internal/domain"
	"github.com/voxire/onion/internal/modules/automations/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ErrRunNotClaimed is retained as an adapter alias for existing callers.
var ErrRunNotClaimed = domain.ErrRunNotClaimed

// ListDueRuns returns runs ready to execute, oldest due first.
func (r *Repository) ListDueRuns(ctx context.Context, limit int32) ([]domain.Run, error) {
	rows, err := r.queries.ListDueAutomationRuns(ctx, limit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list due automation runs", logger.Error(err))
		return nil, err
	}
	runs := make([]domain.Run, 0, len(rows))
	for _, row := range rows {
		runs = append(runs, domain.Run{
			ID:           row.ID,
			BusinessID:   row.BusinessID,
			AutomationID: row.AutomationID,
			EventID:      row.EventID,
			Attempts:     int(row.Attempts),
			MaxAttempts:  int(row.MaxAttempts),
		})
	}
	return runs, nil
}

// ClaimRun takes ownership of a run. Returns ErrRunNotClaimed when another
// worker got there first, which the caller treats as a skip.
func (r *Repository) ClaimRun(ctx context.Context, runID uint64) error {
	result, err := r.queries.ClaimAutomationRun(ctx, runID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to claim automation run", logger.Error(err))
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRunNotClaimed
	}
	return nil
}

// MarkRunSucceeded records a completed run.
func (r *Repository) MarkRunSucceeded(ctx context.Context, runID uint64) error {
	return r.queries.MarkAutomationRunSucceeded(ctx, runID)
}

// MarkRunFailed schedules a retry, or dead-letters the run when its attempts
// are spent. The status decision is made in SQL against the row's own
// max_attempts so it cannot drift from what the table says.
func (r *Repository) MarkRunFailed(ctx context.Context, runID uint64, retryAfterSeconds int64, cause string) error {
	return r.queries.MarkAutomationRunFailed(ctx, sqlc.MarkAutomationRunFailedParams{
		RetryAfterSeconds: retryAfterSeconds,
		LastError:         sql.NullString{String: truncateError(cause), Valid: cause != ""},
		ID:                runID,
	})
}

// MarkRunDead ends a run that cannot succeed however many times it is tried.
func (r *Repository) MarkRunDead(ctx context.Context, runID uint64, cause string) error {
	return r.queries.MarkAutomationRunDead(ctx, sqlc.MarkAutomationRunDeadParams{
		LastError: sql.NullString{String: truncateError(cause), Valid: cause != ""},
		ID:        runID,
	})
}

// RecoverAbandoned requeues runs left executing by a worker that died, and
// reports how many. Safe to retry because every effect an action performs
// carries a key naming it, so a repeat is recognised and skipped rather than
// sending a customer a second copy of something they already have.
func (r *Repository) RecoverAbandoned(ctx context.Context, staleAfterSeconds int64) (int64, error) {
	result, err := r.queries.RecoverAbandonedAutomationRuns(ctx, staleAfterSeconds)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to recover abandoned automation runs", logger.Error(err))
		return 0, err
	}
	return result.RowsAffected()
}

// GetAutomationForRun loads the automation a run belongs to, scoped by the
// run's business as well as its id.
func (r *Repository) GetAutomationForRun(ctx context.Context, automationID, businessID int64) (domain.Automation, error) {
	row, err := r.queries.GetAutomationForRun(ctx, sqlc.GetAutomationForRunParams{
		ID:         automationID,
		BusinessID: businessID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Automation{}, domain.ErrAutomationNotFound
		}
		return domain.Automation{}, err
	}
	return domain.Automation{
		ID:          row.ID,
		TriggerType: row.TriggerType,
		Actions:     row.Actions,
		IsActive:    row.IsActive.Bool,
	}, nil
}

// RecordExecution writes the run_count and last_run_at that #834 records as
// never written.
func (r *Repository) RecordExecution(ctx context.Context, automationID, businessID int64) error {
	return r.queries.RecordAutomationExecution(ctx, sqlc.RecordAutomationExecutionParams{
		ID:         automationID,
		BusinessID: businessID,
	})
}

// truncateError keeps a stored cause inside the column and readable. The
// column is TEXT, so this is about keeping a log line useful rather than
// about overflow.
func truncateError(cause string) string {
	const limit = 500
	if len(cause) <= limit {
		return cause
	}
	return cause[:limit] + "..."
}
