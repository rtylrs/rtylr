package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/recruit/voxhire/internal/domain"
)

// Clearing a due date is a legitimate partial update, while an update that
// carries no field at all is still rejected.
func TestUpdateTaskAcceptsDueAtOnlyPartialUpdate(t *testing.T) {
	svc := &Service{repo: &mockRepo{}}

	if _, err := svc.UpdateTask(context.Background(), domain.TaskInput{ID: 5, DueAtSet: true}); err != nil {
		t.Fatalf("due_at-only update rejected: %v", err)
	}

	if _, err := svc.UpdateTask(context.Background(), domain.TaskInput{ID: 5}); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty update error = %v, want ErrInvalidInput", err)
	}
}
