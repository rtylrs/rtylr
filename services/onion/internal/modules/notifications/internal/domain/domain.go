package domain

import "time"

type Notification struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	UserID     int64
	Title      string
	Body       string
	Type       string
	IsRead     bool
	Link       string
	CreatedAt  time.Time
}
