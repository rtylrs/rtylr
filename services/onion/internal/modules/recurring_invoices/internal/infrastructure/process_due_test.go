package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestProcessDueSchedulesContinuesAfterScheduleError(t *testing.T) {
	now := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	due := []dueSchedule{
		{id: 11, businessID: 101},
		{id: 12, businessID: 102},
		{id: 13, businessID: 103},
	}

	got := processDueSchedules(context.Background(), now, due, func(_ context.Context, id, _ int64, _ time.Time) (bool, error) {
		switch id {
		case 11:
			return false, errors.New("template missing")
		case 12:
			return true, nil
		default:
			return false, nil
		}
	})

	if len(got) != 1 || got[0] != 102 {
		t.Fatalf("created business IDs = %v, want [102]", got)
	}
}

func TestProcessDueSchedulesSkipsMissingRows(t *testing.T) {
	now := time.Date(2026, time.August, 22, 0, 0, 0, 0, time.UTC)
	due := []dueSchedule{{id: 11, businessID: 101}, {id: 12, businessID: 102}}

	got := processDueSchedules(context.Background(), now, due, func(_ context.Context, id, _ int64, _ time.Time) (bool, error) {
		if id == 11 {
			return false, sql.ErrNoRows
		}
		return true, nil
	})

	if len(got) != 1 || got[0] != 102 {
		t.Fatalf("created business IDs = %v, want [102]", got)
	}
}
