package domain

import "time"

type APIKey struct {
	ID         int64
	OwnerID    int64
	BusinessID int64
	Name       string
	KeyPrefix  string
	Scopes     []string
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	IsActive   bool
	CreatedAt  time.Time
}

type CreateAPIKeyInput struct {
	Name   string
	Scopes []string
}
