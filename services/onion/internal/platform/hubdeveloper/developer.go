package hubdeveloper

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrInvalidWebhook marks a registration the caller can correct: a bad
	// URL or an unsupported event.
	ErrInvalidWebhook  = errors.New("invalid webhook")
	ErrWebhookNotFound = errors.New("webhook not found")
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

// Delivery is one attempt to deliver one event to one registration, as the
// tenant who owns the registration sees it.
//
// The delivery engine dead-letters an endpoint that keeps failing, and without
// this the owner has no way to find out: the rows live only in the database
// and the first sign is somebody noticing events stopped arriving.
type Delivery struct {
	ID        int64
	WebhookID int64
	EventID   int64
	EventType string
	// Status is pending, delivering, delivered, failed or dead. A dead
	// delivery has spent its attempts and will not be retried.
	Status      string
	Attempts    int
	MaxAttempts int
	// ResponseStatus is the HTTP status the endpoint answered, zero when the
	// request never got a response at all.
	ResponseStatus int
	// LastError is why the most recent attempt failed, including a short
	// excerpt of the endpoint's own response when it sent one.
	LastError     string
	NextAttemptAt *time.Time
	DeliveredAt   *time.Time
	CreatedAt     time.Time
}

// DeliveryFilter narrows a listing. A zero WebhookID and an empty Status mean
// no filter rather than match nothing.
type DeliveryFilter struct {
	WebhookID int64
	Status    string
	Limit     int32
	Offset    int64
}

type APIKeysProvider interface {
	List(ctx context.Context, businessID int64) ([]APIKey, error)
	Create(ctx context.Context, userID, businessID int64, name string, scopes []string) (id int64, rawKey string, err error)
	Revoke(ctx context.Context, businessID, keyID int64) error
}

type WebhooksProvider interface {
	List(ctx context.Context, businessID int64) ([]Webhook, error)
	// Create registers a webhook and returns its signing secret, which is
	// never readable again.
	Create(ctx context.Context, ownerID, businessID int64, url string, events []string) (id int64, secret string, err error)
	Update(ctx context.Context, businessID, webhookID int64, url string, events []string, isActive bool) (Webhook, error)
	Delete(ctx context.Context, businessID, webhookID int64) error
	// ListDeliveries returns a page of delivery attempts for the business,
	// most recent first, with the total matching the filter.
	ListDeliveries(ctx context.Context, businessID int64, filter DeliveryFilter) ([]Delivery, int64, error)
}
