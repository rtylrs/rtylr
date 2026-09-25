package domain

import (
	"time"
)

type ActivityLog struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	BusinessID  int64     `json:"business_id"`
	App         string    `json:"app"`
	EntityType  string    `json:"entity_type"`
	EntityID    int64     `json:"entity_id"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActivityPage is a single page of activity plus the total count for the same
// owner/app filter, so the client can render pagination controls.
type ActivityPage struct {
	Entries []ActivityLog
	Total   int64
}
