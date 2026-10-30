package echo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/caret/buddy"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

// buddyToolExecutor lets a person manage their buddy's routines from chat.
// The buddy module is attached after start-up, so it is read at call time.
type buddyToolExecutor struct{ module *Module }

func (e *buddyToolExecutor) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case "buddy.routines", "buddy.create_routine", "buddy.update_routine", "buddy.delete_routine":
	default:
		return nil, false, nil
	}
	if e.module == nil || e.module.buddies == nil {
		return nil, true, errors.New("Caret Buddy is not available right now")
	}
	user, ok := authPkg.GetAuthUser(ctx)
	if !ok || user == nil || user.UserID <= 0 || user.BusinessID <= 0 {
		return nil, true, errors.New("sign in to a business to use Caret Buddy")
	}
	buddies := e.module.buddies
	switch name {
	case "buddy.routines":
		routines, err := buddies.Routines(ctx, user.BusinessID, user.UserID)
		if err != nil {
			return nil, true, err
		}
		out := make([]map[string]any, 0, len(routines))
		for _, r := range routines {
			out = append(out, routineOutput(r))
		}
		return map[string]any{"routines": out, "count": len(out)}, true, nil
	case "buddy.create_routine":
		first, err := inputTime(input, "first_run_at", "next_run_at", "start_at")
		if err != nil {
			return nil, true, err
		}
		r, err := buddies.CreateRoutine(ctx, user.BusinessID, user.UserID,
			inputString(input, "name", "title"), inputString(input, "instructions", "task", "prompt"),
			inputInt(input, "every_minutes", "interval_minutes"), first)
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"routine": routineOutput(r), "message": "Routine created. Results will be posted in the Caret Buddy chat."}, true, nil
	case "buddy.update_routine":
		id := int64(inputInt(input, "id", "routine_id"))
		if id <= 0 {
			return nil, true, errors.New("say which routine to change by its id (from buddy.routines)")
		}
		change := buddy.RoutineChange{}
		if v, ok := input["enabled"].(bool); ok {
			change.Enabled = &v
		}
		if v := inputString(input, "name"); v != "" {
			change.Name = &v
		}
		if v := inputString(input, "instructions"); v != "" {
			change.Instructions = &v
		}
		if v := inputInt(input, "every_minutes", "interval_minutes"); v > 0 {
			change.EveryMinutes = &v
		}
		next, err := inputTime(input, "next_run_at", "first_run_at")
		if err != nil {
			return nil, true, err
		}
		change.NextRunAt = next
		r, err := buddies.UpdateRoutine(ctx, user.BusinessID, user.UserID, id, change)
		if errors.Is(err, buddy.ErrRoutineNotFound) {
			return nil, true, fmt.Errorf("no routine %d; list them with buddy.routines", id)
		}
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"routine": routineOutput(r)}, true, nil
	default:
		id := int64(inputInt(input, "id", "routine_id"))
		err := buddies.DeleteRoutine(ctx, user.BusinessID, user.UserID, id)
		if errors.Is(err, buddy.ErrRoutineNotFound) {
			return nil, true, fmt.Errorf("no routine %d; list them with buddy.routines", id)
		}
		if err != nil {
			return nil, true, err
		}
		return map[string]any{"deleted_id": id}, true, nil
	}
}

func routineOutput(r buddy.Routine) map[string]any {
	out := map[string]any{
		"id": r.ID, "name": r.Name, "instructions": r.Instructions,
		"every_minutes": r.EveryMinutes, "enabled": r.Enabled,
	}
	if r.NextRunAt != nil && r.Enabled {
		out["next_run_at"] = r.NextRunAt.UTC().Format(time.RFC3339)
	}
	if r.LastRunAt != nil {
		out["last_run_at"] = r.LastRunAt.UTC().Format(time.RFC3339)
	}
	return out
}

func inputString(input map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := input[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func inputInt(input map[string]any, keys ...string) int {
	for _, key := range keys {
		switch v := input[key].(type) {
		case float64:
			if v > 0 && v < math.MaxInt32 {
				return int(v)
			}
		case int:
			return v
		case int64:
			return int(v)
		}
	}
	return 0
}

func inputTime(input map[string]any, keys ...string) (*time.Time, error) {
	value := inputString(input, keys...)
	if value == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, fmt.Errorf("give the time as RFC3339 with its offset, e.g. 2026-10-31T20:30:00+03:00")
	}
	return &t, nil
}
