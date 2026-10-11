package domain

import (
	"errors"
	"fmt"
)

// MaxOrderedTaskIDs caps how many card ids one move call may reorder. A board
// column is far smaller than this; the cap only bounds the CASE expression the
// repository builds.
const MaxOrderedTaskIDs = 500

// ErrInvalidTaskOrder marks a malformed ordered_task_ids list (client error).
var ErrInvalidTaskOrder = errors.New("invalid task order")

// ErrTaskOrderOutsideProject means a listed id is not a task of the moved
// task's project (or does not exist). Nothing is written.
var ErrTaskOrderOutsideProject = errors.New("ordered tasks must belong to the moved task's project")

// ValidateTaskOrder checks an ordered_task_ids list. An empty list is valid
// (legacy behaviour: no reordering). Otherwise it must hold at most
// MaxOrderedTaskIDs positive, unique ids and include the moved task.
func ValidateTaskOrder(movedTaskID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	if len(ids) > MaxOrderedTaskIDs {
		return fmt.Errorf("%w: at most %d ids allowed", ErrInvalidTaskOrder, MaxOrderedTaskIDs)
	}
	seen := make(map[int64]struct{}, len(ids))
	found := false
	for _, id := range ids {
		if id <= 0 {
			return fmt.Errorf("%w: ids must be positive", ErrInvalidTaskOrder)
		}
		if _, dup := seen[id]; dup {
			return fmt.Errorf("%w: duplicate id %d", ErrInvalidTaskOrder, id)
		}
		seen[id] = struct{}{}
		if id == movedTaskID {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("%w: the moved task must be in the list", ErrInvalidTaskOrder)
	}
	return nil
}
