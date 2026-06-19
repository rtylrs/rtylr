package domain

import "time"

type Resource struct {
	ID         int64     `json:"id"`
	OwnerID    int64     `json:"owner_id"`
	BusinessID int64     `json:"business_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Capacity   int       `json:"capacity"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type ResourceInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Capacity int    `json:"capacity"`
	IsActive bool   `json:"is_active"`
}
