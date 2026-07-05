package hubdeveloper

import (
	"context"
	"time"
)

type APIKey struct {
	ID         int64
	Name       string
	KeyPrefix  string
	Scopes     []string
	IsActive   bool
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

type Webhook struct {
	ID              int64
	URL             string
	Events          []string
	IsActive        bool
	LastTriggeredAt *time.Time
	CreatedAt       time.Time
}

type APIKeysProvider interface {
	List(ctx context.Context, businessID int64) ([]APIKey, error)
	Create(ctx context.Context, userID, businessID int64, name string, scopes []string) (id int64, rawKey string, err error)
	Revoke(ctx context.Context, businessID, keyID int64) error
}

type WebhooksProvider interface {
	List(ctx context.Context, businessID int64) ([]Webhook, error)
	Create(ctx context.Context, ownerID, businessID int64, url string, events []string) (int64, error)
	Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) (Webhook, error)
	Delete(ctx context.Context, businessID, webhookID int64) error
}
