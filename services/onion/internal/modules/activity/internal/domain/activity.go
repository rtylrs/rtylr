package domain

import (
	"time"
)

type ActivityLog struct {
	ID          int64
	OwnerID     int64
	BusinessID  int64
	App         string
	EntityType  string
	EntityID    int64
	Action      string
	Description string
	UserID      int64
	UserName    string
	CreatedAt   time.Time
}

// ActivityPage is a single page of activity plus the total count for the same
// owner/app filter, so the client can render pagination controls.
type ActivityPage struct {
	Entries []ActivityLog
	Total   int64
}
