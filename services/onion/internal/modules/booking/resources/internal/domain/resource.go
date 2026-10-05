package domain

import "time"

type Resource struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	Type       string
	Capacity   int
	IsActive   bool
	CreatedAt  time.Time
}

type ResourceInput struct {
	Name     string
	Type     string
	Capacity int
	IsActive bool
}
