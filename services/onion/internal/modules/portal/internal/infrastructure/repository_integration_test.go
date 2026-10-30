package infrastructure

import (
	"context"
	"slices"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/modules/portal/internal/layout"
	"github.com/voxire/onion/internal/testutil"
)

const fixtureUserID = 930117

func TestLayoutRoundTripsAndForgetsTheDefault(t *testing.T) {
	db := testutil.OpenDB(t)
	clean := func() {
		if _, err := db.Exec("DELETE FROM portal_preferences WHERE user_id IN (?, ?)", fixtureUserID, fixtureUserID+1); err != nil {
			t.Fatalf("clean fixtures: %v", err)
		}
	}
	clean()
	t.Cleanup(clean)
	repo := NewRepository(db)
	ctx := context.Background()

	if got, err := repo.Layout(ctx, fixtureUserID); err != nil || !got.Empty() {
		t.Fatalf("a person who never arranged anything: %+v, %v", got, err)
	}
	saved := layout.Layout{Rail: &layout.Arrangement{Top: []string{"caret", "home"}, Hidden: []string{"pos"}}}
	for range 2 { // the second change updates the same row
		if _, err := repo.UpdateLayout(ctx, fixtureUserID, func(layout.Layout) (layout.Layout, error) { return saved, nil }); err != nil {
			t.Fatal(err)
		}
	}
	got, err := repo.Layout(ctx, fixtureUserID)
	if err != nil || got.Rail == nil || !slices.Equal(got.Rail.Top, []string{"caret", "home"}) || got.Workspace != nil {
		t.Fatalf("read back %+v, %v", got, err)
	}
	if other, _ := repo.Layout(ctx, fixtureUserID+1); !other.Empty() {
		t.Fatal("another person sees this layout")
	}
	if _, err := repo.UpdateLayout(ctx, fixtureUserID, func(layout.Layout) (layout.Layout, error) { return layout.Layout{}, nil }); err != nil {
		t.Fatal(err)
	}
	if got, err := repo.Layout(ctx, fixtureUserID); err != nil || !got.Empty() {
		t.Fatalf("restoring the default left %+v (%v)", got, err)
	}
}

// Two changes at the same moment, as when Caret runs tool calls in parallel,
// both land: each edits the other's result, not a stale copy.
func TestParallelUpdatesDontLoseChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	clean := func() {
		if _, err := db.Exec("DELETE FROM portal_preferences WHERE user_id = ?", fixtureUserID+2); err != nil {
			t.Fatalf("clean fixtures: %v", err)
		}
	}
	clean()
	t.Cleanup(clean)
	repo := NewRepository(db)
	ctx := context.Background()
	hide := func(id string) func(layout.Layout) (layout.Layout, error) {
		return func(current layout.Layout) (layout.Layout, error) {
			next := layout.Arrangement{Top: []string{"home"}}
			if current.Rail != nil {
				next = *current.Rail
			}
			next.Hidden = append(slices.Clone(next.Hidden), id)
			current.Rail = &next
			return current, nil
		}
	}
	var wg sync.WaitGroup
	for _, id := range []string{"pos", "hr", "crm", "erp"} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := repo.UpdateLayout(ctx, fixtureUserID+2, hide(id)); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	got, err := repo.Layout(ctx, fixtureUserID+2)
	if err != nil || got.Rail == nil || len(got.Rail.Hidden) != 4 {
		t.Fatalf("lost a change: %+v, %v", got.Rail, err)
	}
}
