package echo

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
)

const toolsBusinessID = 920503

func TestBuddyToolsManageTheCallersRoutines(t *testing.T) {
	db := testutil.OpenDB(t)
	clean := func() {
		for _, q := range []string{
			"DELETE FROM caret_buddy_routines WHERE business_id = ?",
			"DELETE FROM caret_buddies WHERE business_id = ?",
		} {
			if _, err := db.Exec(q, toolsBusinessID); err != nil {
				t.Fatal(err)
			}
		}
	}
	clean()
	t.Cleanup(clean)

	exec := &buddyToolExecutor{module: &Module{buddies: buddy.NewModule(context.Background(), db)}}
	as := func(userID int64) context.Context {
		return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{UserID: userID, BusinessID: toolsBusinessID})
	}

	out, executed, err := exec.ExecuteCaretTool(as(5), "buddy.create_routine", map[string]any{
		"name": "Gym check-in", "instructions": "Ask if I went to the gym", "every_minutes": float64(1440),
		"first_run_at": time.Now().Add(48 * time.Hour).In(time.FixedZone("", 3*3600)).Format(time.RFC3339),
	})
	if err != nil || !executed {
		t.Fatalf("create: %v %v", executed, err)
	}
	created := out["routine"].(map[string]any)
	id := created["id"].(int64)
	if next, _ := time.Parse(time.RFC3339, created["next_run_at"].(string)); next.Before(time.Now().Add(47 * time.Hour)) {
		t.Fatalf("first run = %v", created["next_run_at"])
	}

	if out, _, _ := exec.ExecuteCaretTool(as(6), "buddy.routines", nil); out["count"] != 0 {
		t.Fatalf("another member sees routines: %v", out)
	}
	if _, _, err := exec.ExecuteCaretTool(as(6), "buddy.delete_routine", map[string]any{"id": float64(id)}); err == nil {
		t.Fatal("another member deleted the routine")
	}

	out, _, err = exec.ExecuteCaretTool(as(5), "buddy.update_routine", map[string]any{"id": float64(id), "enabled": false})
	if err != nil || out["routine"].(map[string]any)["enabled"] != false {
		t.Fatalf("pause: %v %v", out, err)
	}
	if _, _, err := exec.ExecuteCaretTool(as(5), "buddy.create_routine", map[string]any{
		"name": "Too often", "instructions": "x", "every_minutes": float64(5),
	}); err == nil {
		t.Fatal("a routine every 5 minutes was accepted")
	}
	if _, _, err := exec.ExecuteCaretTool(as(5), "buddy.update_routine", map[string]any{"id": float64(id), "next_run_at": "2099-01-01T00:00:00Z"}); err == nil {
		t.Fatal("a start years away was accepted")
	}
	if _, _, err := exec.ExecuteCaretTool(as(5), "buddy.delete_routine", map[string]any{"id": float64(id)}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if out, _, _ := exec.ExecuteCaretTool(as(5), "buddy.routines", nil); out["count"] != 0 {
		t.Fatalf("routine survived delete: %v", out)
	}
}
