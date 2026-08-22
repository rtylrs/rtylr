package domain

import "time"

type APIKey struct {
	ID         int64      `json:"id"`
	OwnerID    int64      `json:"owner_id"`
	BusinessID int64      `json:"business_id"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateAPIKeyInput struct {
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}
