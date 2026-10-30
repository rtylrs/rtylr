package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	"github.com/voxire/onion/internal/modules/caret/buddy/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	db *sql.DB
	q  *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, q: sqlc.New(db)}
}

func (r *Repository) UpsertBuddy(ctx context.Context, b domain.Buddy) error {
	return r.q.UpsertBuddy(ctx, sqlc.UpsertBuddyParams{
		BusinessID:      b.BusinessID,
		RunAsUserID:     b.RunAsUserID,
		Enabled:         b.Enabled,
		Instructions:    b.Instructions,
		IntervalMinutes: int32(b.IntervalMinutes),
		DailyRunCap:     int32(b.DailyRunCap),
		EnabledFlag:     boolToInt(b.Enabled),
	})
}

func (r *Repository) BuddyByBusiness(ctx context.Context, businessID int64) (domain.Buddy, error) {
	row, err := r.q.GetBuddyByBusiness(ctx, businessID)
	if err != nil {
		return domain.Buddy{}, err
	}
	return buddyFromRow(row), nil
}

// EnqueueDue queues one run for every buddy whose time has come and moves its
// next run forward. A buddy over its daily cap is advanced without a run.
func (r *Repository) EnqueueDue(ctx context.Context, limit int) (queued int, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.q.WithTx(tx)

	due, err := q.ListDueBuddiesForUpdate(ctx, int32(limit))
	if err != nil {
		return 0, fmt.Errorf("list due buddies: %w", err)
	}
	for _, b := range due {
		runs, err := q.CountRunsSince(ctx, b.ID)
		if err != nil {
			return 0, fmt.Errorf("count runs: %w", err)
		}
		if runs < int64(b.DailyRunCap) {
			if _, err := q.InsertRun(ctx, sqlc.InsertRunParams{BuddyID: b.ID, BusinessID: b.BusinessID, TriggerKind: domain.TriggerSchedule}); err != nil {
				return 0, fmt.Errorf("insert run: %w", err)
			}
			queued++
		}
		if err := q.AdvanceBuddy(ctx, b.ID); err != nil {
			return 0, fmt.Errorf("advance buddy: %w", err)
		}
	}
	return queued, tx.Commit()
}

// ClaimRun leases the oldest queued run to the caller, or returns nil when
// nothing is waiting.
func (r *Repository) ClaimRun(ctx context.Context, lease time.Duration) (*domain.Run, *domain.Buddy, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.q.WithTx(tx)

	row, err := q.ClaimNextRunForUpdate(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("claim run: %w", err)
	}
	if err := q.MarkRunRunning(ctx, sqlc.MarkRunRunningParams{LeaseSeconds: int64(lease.Seconds()), ID: row.ID}); err != nil {
		return nil, nil, fmt.Errorf("mark run running: %w", err)
	}
	buddyRow, err := q.GetBuddy(ctx, row.BuddyID)
	if err != nil {
		return nil, nil, fmt.Errorf("load buddy: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	buddy := buddyFromRow(buddyRow)
	return &domain.Run{
		ID: row.ID, BuddyID: row.BuddyID, BusinessID: row.BusinessID,
		TriggerKind: row.TriggerKind, Attempts: int(row.Attempts) + 1,
	}, &buddy, nil
}

func (r *Repository) FinishRun(ctx context.Context, runID int64, result domain.RunResult) error {
	status := sqlc.CaretBuddyRunsStatusDone
	if result.Error != "" {
		status = sqlc.CaretBuddyRunsStatusFailed
	}
	return r.q.FinishRun(ctx, sqlc.FinishRunParams{
		Status:    status,
		Report:    sql.NullString{String: result.Report, Valid: result.Report != ""},
		Error:     sql.NullString{String: result.Error, Valid: result.Error != ""},
		ToolCalls: int32(result.ToolCalls),
		Tokens:    result.Tokens,
		ID:        runID,
	})
}

// RecoverStaleRuns requeues runs whose lease ran out (the process died) and
// fails the ones that already used all their attempts.
func (r *Repository) RecoverStaleRuns(ctx context.Context) error {
	if _, err := r.q.FailExhaustedRuns(ctx, domain.MaxRunAttempts); err != nil {
		return err
	}
	_, err := r.q.RequeueExpiredRuns(ctx, domain.MaxRunAttempts)
	return err
}

// ProposeAction records a write the buddy wants to make. It is idempotent per
// run and call, so a retried run does not queue the same proposal twice.
func (r *Repository) ProposeAction(ctx context.Context, runID, businessID int64, tool string, input map[string]any, fingerprint string, ttl time.Duration) error {
	encoded, err := json.Marshal(input)
	if err != nil {
		return err
	}
	_, err = r.q.InsertAction(ctx, sqlc.InsertActionParams{
		RunID: runID, BusinessID: businessID, ToolName: tool, Input: encoded,
		Fingerprint: fingerprint, TtlHours: int64(ttl.Hours()),
	})
	return err
}

func (r *Repository) ExpireActions(ctx context.Context) error {
	_, err := r.q.ExpireActions(ctx)
	return err
}

func buddyFromRow(row sqlc.CaretBuddy) domain.Buddy {
	b := domain.Buddy{
		ID: row.ID, BusinessID: row.BusinessID, RunAsUserID: row.RunAsUserID,
		Enabled: row.Enabled, Instructions: row.Instructions,
		IntervalMinutes: int(row.IntervalMinutes), DailyRunCap: int(row.DailyRunCap),
	}
	if row.NextRunAt.Valid {
		b.NextRunAt = &row.NextRunAt.Time
	}
	if row.LastRunAt.Valid {
		b.LastRunAt = &row.LastRunAt.Time
	}
	return b
}

func boolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func (r *Repository) EnqueueEvent(ctx context.Context, eventID uint64, businessID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.q.WithTx(tx)
	b, err := q.GetBuddyByBusinessForUpdate(ctx, businessID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("lock buddy: %w", err)
	}
	if !b.Enabled {
		return nil
	}
	result, err := q.RecordBuddyEvent(ctx, sqlc.RecordBuddyEventParams{EventID: eventID, BuddyID: b.ID})
	if err != nil {
		return fmt.Errorf("record buddy event: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if inserted == 0 {
		return tx.Commit()
	}
	queued, err := q.CountQueuedEventRuns(ctx, b.ID)
	if err != nil {
		return err
	}
	runs, err := q.CountRunsSince(ctx, b.ID)
	if err != nil {
		return err
	}
	if queued == 0 && runs < int64(b.DailyRunCap) {
		if _, err := q.InsertRun(ctx, sqlc.InsertRunParams{BuddyID: b.ID, BusinessID: b.BusinessID, TriggerKind: "event"}); err != nil {
			return fmt.Errorf("insert event run: %w", err)
		}
	}
	return tx.Commit()
}
