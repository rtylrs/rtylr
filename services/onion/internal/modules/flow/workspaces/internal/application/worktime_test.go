package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// These exercise validation that returns before the repository is touched, so
// a Service with a nil repo is enough - hitting the repo would panic and fail
// the test, which is exactly the guarantee being checked: the invalid input
// never reaches storage.

func TestCreateWorkDayLine_RejectsMissingProjectID(t *testing.T) {
	svc := &Service{}
	_, err := svc.CreateWorkDayLine(context.Background(), domain.WorkDayLineInput{
		WorkDayID:       1,
		BusinessID:      1,
		DurationMinutes: 30,
	})
	if !errors.Is(err, ErrProjectRequired) {
		t.Fatalf("want ErrProjectRequired, got %v", err)
	}
}

func TestCreateWorkDayLine_RejectsInvalidWorkModeBeforeProjectCheck(t *testing.T) {
	svc := &Service{}
	_, err := svc.CreateWorkDayLine(context.Background(), domain.WorkDayLineInput{
		WorkDayID:  1,
		BusinessID: 1,
		WorkMode:   "not-a-real-mode",
	})
	if !errors.Is(err, ErrInvalidWorkMode) {
		t.Fatalf("want ErrInvalidWorkMode, got %v", err)
	}
}

func TestUpdateWorkDayLine_RejectsMissingProjectID(t *testing.T) {
	svc := &Service{}
	_, err := svc.UpdateWorkDayLine(context.Background(), domain.WorkDayLineInput{
		ID:              1,
		BusinessID:      1,
		DurationMinutes: 30,
	})
	if !errors.Is(err, ErrProjectRequired) {
		t.Fatalf("want ErrProjectRequired, got %v", err)
	}
}
