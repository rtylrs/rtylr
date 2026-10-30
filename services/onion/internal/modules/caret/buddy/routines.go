package buddy

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy/internal/domain"
	"github.com/voxire/onion/internal/modules/caret/buddy/internal/infrastructure"
)

type Routine = domain.Routine

var ErrRoutineNotFound = infrastructure.ErrRoutineNotFound

const (
	MinRoutineMinutes = 15
	MaxRoutineMinutes = 43200
)

// RoutineChange is a partial update; nil fields keep their current value.
type RoutineChange struct {
	Name         *string
	Instructions *string
	EveryMinutes *int
	Enabled      *bool
	NextRunAt    *time.Time
}

// Routines lists the routines a person set up in a business.
func (m *Module) Routines(ctx context.Context, businessID, userID int64) ([]Routine, error) {
	return m.repo.Routines(ctx, businessID, userID)
}

// CreateRoutine schedules a routine. Its first run is at firstRun, or one
// interval from now when firstRun is nil or in the past.
func (m *Module) CreateRoutine(ctx context.Context, businessID, userID int64, name, instructions string, everyMinutes int, firstRun *time.Time) (Routine, error) {
	name, instructions = strings.TrimSpace(name), strings.TrimSpace(instructions)
	if name == "" || instructions == "" {
		return Routine{}, errors.New("a routine needs a name and instructions")
	}
	if err := validMinutes(everyMinutes); err != nil {
		return Routine{}, err
	}
	if err := validFirstRun(firstRun); err != nil {
		return Routine{}, err
	}
	next := firstRunAt(firstRun, everyMinutes)
	return m.repo.CreateRoutine(ctx, businessID, userID, Routine{
		Name: truncate(name, 120), Instructions: instructions, EveryMinutes: everyMinutes, NextRunAt: &next,
	})
}

func (m *Module) UpdateRoutine(ctx context.Context, businessID, userID, id int64, change RoutineChange) (Routine, error) {
	routine, err := m.repo.Routine(ctx, businessID, userID, id)
	if err != nil {
		return Routine{}, err
	}
	if change.Name != nil && strings.TrimSpace(*change.Name) != "" {
		routine.Name = truncate(strings.TrimSpace(*change.Name), 120)
	}
	if change.Instructions != nil && strings.TrimSpace(*change.Instructions) != "" {
		routine.Instructions = strings.TrimSpace(*change.Instructions)
	}
	if err := validFirstRun(change.NextRunAt); err != nil {
		return Routine{}, err
	}
	rescheduled := change.NextRunAt != nil
	if change.EveryMinutes != nil {
		if err := validMinutes(*change.EveryMinutes); err != nil {
			return Routine{}, err
		}
		rescheduled = rescheduled || routine.EveryMinutes != *change.EveryMinutes
		routine.EveryMinutes = *change.EveryMinutes
	}
	if change.Enabled != nil {
		rescheduled = rescheduled || (*change.Enabled && !routine.Enabled)
		routine.Enabled = *change.Enabled
	}
	if rescheduled || routine.NextRunAt == nil {
		next := firstRunAt(change.NextRunAt, routine.EveryMinutes)
		routine.NextRunAt = &next
	}
	if err := m.repo.UpdateRoutine(ctx, routine); err != nil {
		return Routine{}, err
	}
	return routine, nil
}

func (m *Module) DeleteRoutine(ctx context.Context, businessID, userID, id int64) error {
	return m.repo.DeleteRoutine(ctx, businessID, userID, id)
}

func validMinutes(minutes int) error {
	if minutes < MinRoutineMinutes || minutes > MaxRoutineMinutes {
		return errors.New("a routine can run at most every 15 minutes and at least once a month")
	}
	return nil
}

func validFirstRun(t *time.Time) error {
	if t != nil && t.After(time.Now().AddDate(1, 0, 0)) {
		return errors.New("a routine has to start within a year")
	}
	return nil
}

func firstRunAt(requested *time.Time, everyMinutes int) time.Time {
	now := time.Now()
	if requested != nil && requested.After(now) {
		return requested.UTC()
	}
	return now.Add(time.Duration(everyMinutes) * time.Minute).UTC()
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}
