package domain

import "time"

type Notification struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	UserID    int64     `json:"userId"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"isRead"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"createdAt"`
}
