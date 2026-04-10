package domain

import "time"

type APIKey struct {
	ID         int64      `json:"id"`
	OwnerID    int64      `json:"ownerId"`
	Name       string     `json:"name"`
	KeyPrefix  string     `json:"keyPrefix"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	IsActive   bool       `json:"isActive"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type CreateAPIKeyInput struct {
	Name    string   `json:"name"`
	Scopes  []string `json:"scopes"`
}
