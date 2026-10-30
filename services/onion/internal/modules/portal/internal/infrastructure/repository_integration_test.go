package infrastructure

import (
	"context"
	"slices"
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
	for range 2 { // the second save updates the same row
		if err := repo.SaveLayout(ctx, fixtureUserID, saved); err != nil {
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

	if err := repo.SaveLayout(ctx, fixtureUserID, layout.Layout{}); err != nil {
		t.Fatal(err)
	}
	var rows int
	if err := db.QueryRow("SELECT COUNT(*) FROM portal_preferences WHERE user_id = ?", fixtureUserID).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("restoring the default left %d rows (%v)", rows, err)
	}
}
