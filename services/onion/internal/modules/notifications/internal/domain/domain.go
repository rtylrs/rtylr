package domain

import "time"

type Notification struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	UserID     int64     `json:"user_id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	Type       string    `json:"type"`
	IsRead     bool      `json:"is_read"`
	Link       string    `json:"link"`
	CreatedAt  time.Time `json:"created_at"`
}
