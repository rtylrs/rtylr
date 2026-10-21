package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/testsupport"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestViewCreatorCanMutateV2View(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := testsupport.NewRepository(db)
	id, err := repo.CreateView(ctx, domain.ViewInput{
		BusinessID: 902, WorkspaceID: 0, CreatedBy: 903,
		Name: "QC Private View", SortDir: "asc", Visibility: "private",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_views WHERE id = ? AND business_id = ?", id, 902); err != nil {
			t.Error(err)
		}
	})
	for _, actor := range []int64{901, 904} {
		if err := repo.DeleteView(ctx, id, actor, 902); err == nil {
			t.Fatalf("non-creator %d deleted the view", actor)
		}
	}
	if err := repo.DeleteView(ctx, id, 903, 999); err == nil {
		t.Fatal("creator deleted a view through another business")
	}
	if err := repo.UpdateView(ctx, domain.UpdateViewInput{
		ID: id, ActorID: 903, BusinessID: 902, Name: "Renamed view",
		SortDir: "asc", Visibility: "private",
	}); err != nil {
		t.Fatalf("creator update: %v", err)
	}
	var visibility string
	if err := db.QueryRowContext(ctx, "SELECT visibility FROM flow_views WHERE id = ? AND business_id = ?", id, 902).Scan(&visibility); err != nil {
		t.Fatal(err)
	}
	if visibility != "private" {
		t.Fatalf("visibility = %q, want private", visibility)
	}
	if err := repo.DeleteView(ctx, id, 903, 902); err != nil {
		t.Fatalf("creator delete: %v", err)
	}
}

func TestViewLegacyWorkspaceMembersRetainMutationAccess(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	repo := testsupport.NewRepository(db)
	result, err := db.ExecContext(ctx, "INSERT INTO flow_workspaces (owner_id, business_id, name, created_by) VALUES (?, ?, ?, ?)", 911, 912, "Legacy view test", 911)
	if err != nil {
		t.Fatal(err)
	}
	workspaceID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_workspace_members WHERE workspace_id = ?", workspaceID); err != nil {
			t.Error(err)
		}
		if _, err := db.ExecContext(ctx, "DELETE FROM flow_workspaces WHERE id = ? AND business_id = ?", workspaceID, 912); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.ExecContext(ctx, "INSERT INTO flow_workspace_members (workspace_id, user_id) VALUES (?, ?)", workspaceID, 913); err != nil {
		t.Fatal(err)
	}
	for _, actorID := range []int64{911, 913} {
		id, err := repo.CreateView(ctx, domain.ViewInput{
			BusinessID: 912, WorkspaceID: workspaceID, CreatedBy: 914,
			Name: "Legacy shared view", SortDir: "asc", Visibility: "shared",
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := db.ExecContext(ctx, "DELETE FROM flow_views WHERE id = ? AND business_id = ?", id, 912); err != nil {
				t.Error(err)
			}
		})
		if err := repo.DeleteView(ctx, id, actorID, 999); err == nil {
			t.Fatal("legacy actor deleted through another business")
		}
		if err := repo.DeleteView(ctx, id, actorID, 912); err != nil {
			t.Fatalf("legacy actor %d: %v", actorID, err)
		}
	}
}
