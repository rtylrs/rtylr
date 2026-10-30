package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	"github.com/voxire/onion/internal/modules/caret/buddy/internal/infrastructure/persistence/sqlc"
)

var ErrRoutineNotFound = errors.New("routine not found")

// CreateRoutine adds a routine, creating the business's buddy if it has none.
func (r *Repository) CreateRoutine(ctx context.Context, businessID, userID int64, routine domain.Routine) (domain.Routine, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Routine{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := r.q.WithTx(tx)
	if err := q.EnsureBuddy(ctx, sqlc.EnsureBuddyParams{BusinessID: businessID, RunAsUserID: userID}); err != nil {
		return domain.Routine{}, fmt.Errorf("ensure buddy: %w", err)
	}
	b, err := q.GetBuddyByBusiness(ctx, businessID)
	if err != nil {
		return domain.Routine{}, fmt.Errorf("load buddy: %w", err)
	}
	result, err := q.InsertRoutine(ctx, sqlc.InsertRoutineParams{
		BuddyID: b.ID, BusinessID: businessID, UserID: userID, Name: routine.Name, Instructions: routine.Instructions,
		EveryMinutes: int32(routine.EveryMinutes), NextRunAt: nullTime(routine.NextRunAt),
	})
	if err != nil {
		return domain.Routine{}, fmt.Errorf("insert routine: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Routine{}, err
	}
	row, err := q.GetRoutineByID(ctx, id)
	if err != nil {
		return domain.Routine{}, err
	}
	return routineFromRow(row), tx.Commit()
}

func (r *Repository) Routines(ctx context.Context, businessID, userID int64) ([]domain.Routine, error) {
	rows, err := r.q.ListRoutinesByOwner(ctx, sqlc.ListRoutinesByOwnerParams{BusinessID: businessID, UserID: userID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.Routine, 0, len(rows))
	for _, row := range rows {
		out = append(out, routineFromRow(row))
	}
	return out, nil
}

func (r *Repository) Routine(ctx context.Context, businessID, userID, id int64) (domain.Routine, error) {
	row, err := r.q.GetRoutine(ctx, sqlc.GetRoutineParams{ID: id, BusinessID: businessID, UserID: userID})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Routine{}, ErrRoutineNotFound
	}
	if err != nil {
		return domain.Routine{}, err
	}
	return routineFromRow(row), nil
}

func (r *Repository) UpdateRoutine(ctx context.Context, routine domain.Routine) error {
	n, err := r.q.UpdateRoutine(ctx, sqlc.UpdateRoutineParams{
		Name: routine.Name, Instructions: routine.Instructions, EveryMinutes: int32(routine.EveryMinutes),
		Enabled: routine.Enabled, NextRunAt: nullTime(routine.NextRunAt), ID: routine.ID, BusinessID: routine.BusinessID, UserID: routine.UserID,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		if _, err := r.Routine(ctx, routine.BusinessID, routine.UserID, routine.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) DeleteRoutine(ctx context.Context, businessID, userID, id int64) error {
	n, err := r.q.DeleteRoutine(ctx, sqlc.DeleteRoutineParams{ID: id, BusinessID: businessID, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrRoutineNotFound
	}
	return nil
}

func routineFromRow(row sqlc.CaretBuddyRoutine) domain.Routine {
	return domain.Routine{
		ID: row.ID, BuddyID: row.BuddyID, BusinessID: row.BusinessID, UserID: row.UserID, Name: row.Name,
		Instructions: row.Instructions, EveryMinutes: int(row.EveryMinutes), Enabled: row.Enabled,
		NextRunAt: timePointer(row.NextRunAt), LastRunAt: timePointer(row.LastRunAt),
	}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
