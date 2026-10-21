package authorization

import (
	"context"
	"errors"
	"strconv"
	"sync"
)

var ErrInvalidScope = errors.New("invalid assignment scope")

// ScopeLookup reads current committed references without participating in the caller's transaction.
type ScopeLookup func(context.Context, int64, string, int64) (bool, error)

var scopeLookup struct {
	sync.RWMutex
	reader ScopeLookup
}

func SetScopeLookup(reader ScopeLookup) func() {
	scopeLookup.Lock()
	previous := scopeLookup.reader
	scopeLookup.reader = reader
	scopeLookup.Unlock()
	return func() { scopeLookup.Lock(); scopeLookup.reader = previous; scopeLookup.Unlock() }
}

func ValidateAssignmentScope(ctx context.Context, businessID int64, scopeType, scopeID string) error {
	id, err := strconv.ParseInt(scopeID, 10, 64)
	if err != nil || id <= 0 || businessID <= 0 {
		return ErrInvalidScope
	}
	if scopeType != "pos_location" && scopeType != "flow_project" {
		return ErrInvalidScope
	}
	scopeLookup.RLock()
	reader := scopeLookup.reader
	scopeLookup.RUnlock()
	if reader == nil {
		return errors.New("assignment scope reader is not configured")
	}
	exists, err := reader(ctx, businessID, scopeType, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidScope
	}
	return nil
}
