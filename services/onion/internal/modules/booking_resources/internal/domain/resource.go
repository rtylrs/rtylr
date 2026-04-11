package domain

import "time"

type Resource struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"ownerId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Capacity  int       `json:"capacity"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type ResourceInput struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Capacity int    `json:"capacity"`
	IsActive bool   `json:"isActive"`
}
