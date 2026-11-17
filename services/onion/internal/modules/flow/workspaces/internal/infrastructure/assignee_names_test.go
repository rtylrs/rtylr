package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application/identity"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/infrastructure"
)

// AssigneeNames reads through the identity reader that names a listed task's
// assignee, scoped to one business, and keeps only the names it found.
func TestAssigneeNamesReadsFlowIdentityNames(t *testing.T) {
	var asked []identity.IdentityKey
	reader := func(_ context.Context, keys []identity.IdentityKey) (map[identity.IdentityKey]identity.Identity, error) {
		asked = append(asked, keys...)
		return map[identity.IdentityKey]identity.Identity{
			{BusinessID: 42, UserID: 7}: {Name: "Abed Amouneh", Email: "abed@example.com"},
			{BusinessID: 42, UserID: 8}: {Email: "no-name@example.com"},
			{BusinessID: 43, UserID: 9}: {Name: "Other business"},
		}, nil
	}
	repo := infrastructure.NewRepository(nil).WithIdentityReaders(reader, nil, nil)

	names, err := repo.AssigneeNames(context.Background(), 42, []int64{7, 8, 9, 0})
	if err != nil {
		t.Fatalf("AssigneeNames: %v", err)
	}
	if len(names) != 1 || names[7] != "Abed Amouneh" {
		t.Errorf("names = %v, want only 7 named, with no email fallback", names)
	}
	if len(asked) != 3 {
		t.Errorf("asked for %v, want the three real ids", asked)
	}
	for _, key := range asked {
		if key.BusinessID != 42 {
			t.Errorf("asked %+v outside business 42", key)
		}
	}
}

func TestAssigneeNamesPropagatesReadFailures(t *testing.T) {
	want := errors.New("identity unavailable")
	reader := func(context.Context, []identity.IdentityKey) (map[identity.IdentityKey]identity.Identity, error) {
		return nil, want
	}
	repo := infrastructure.NewRepository(nil).WithIdentityReaders(reader, nil, nil)
	if _, err := repo.AssigneeNames(context.Background(), 42, []int64{7}); !errors.Is(err, want) {
		t.Errorf("err = %v, want the reader's failure", err)
	}
}

func TestAssigneeNamesWithNoIDsReadsNothing(t *testing.T) {
	repo := infrastructure.NewRepository(nil)
	names, err := repo.AssigneeNames(context.Background(), 42, nil)
	if err != nil || len(names) != 0 {
		t.Errorf("names = %v, %v; want empty with no reader needed", names, err)
	}
}
