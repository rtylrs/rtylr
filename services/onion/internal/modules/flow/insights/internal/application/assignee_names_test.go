package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/insights/internal/domain"
)

func TestNameOverdueAssigneesNamesAssignedTasksInOneRead(t *testing.T) {
	var calls int
	var askedBusiness int64
	var askedIDs []int64
	names := func(_ context.Context, businessID int64, ids []int64) (map[int64]string, error) {
		calls++
		askedBusiness, askedIDs = businessID, ids
		return map[int64]string{7: "Abed Amouneh"}, nil
	}
	tasks := []domain.OverdueTask{
		{ID: 1, AssigneeID: 7},
		{ID: 2},
		{ID: 3, AssigneeID: 7},
		{ID: 4, AssigneeID: 9},
	}

	if err := NameOverdueAssignees(context.Background(), names, 42, tasks); err != nil {
		t.Fatalf("NameOverdueAssignees: %v", err)
	}
	if calls != 1 || askedBusiness != 42 || len(askedIDs) != 2 || askedIDs[0] != 7 || askedIDs[1] != 9 {
		t.Errorf("read %d time(s) for business %d ids %v, want once for 42 with [7 9]", calls, askedBusiness, askedIDs)
	}
	want := []string{"Abed Amouneh", "", "Abed Amouneh", ""}
	for i, task := range tasks {
		if task.AssigneeName != want[i] {
			t.Errorf("task %d name = %q, want %q", task.ID, task.AssigneeName, want[i])
		}
	}
}

func TestNameOverdueAssigneesSkipsTheReadWhenNothingIsAssigned(t *testing.T) {
	names := func(context.Context, int64, []int64) (map[int64]string, error) {
		t.Fatal("names read with no assigned task")
		return nil, nil
	}
	if err := NameOverdueAssignees(context.Background(), names, 42, []domain.OverdueTask{{ID: 1}}); err != nil {
		t.Fatalf("NameOverdueAssignees: %v", err)
	}
	if err := NameOverdueAssignees(context.Background(), nil, 42, []domain.OverdueTask{{ID: 1, AssigneeID: 7}}); err != nil {
		t.Fatalf("nil names: %v", err)
	}
}

func TestNameOverdueAssigneesPropagatesReadFailures(t *testing.T) {
	want := errors.New("identity unavailable")
	names := func(context.Context, int64, []int64) (map[int64]string, error) { return nil, want }
	if err := NameOverdueAssignees(context.Background(), names, 42, []domain.OverdueTask{{ID: 1, AssigneeID: 7}}); !errors.Is(err, want) {
		t.Errorf("err = %v, want the read failure", err)
	}
}
