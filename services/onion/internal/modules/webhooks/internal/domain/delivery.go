package domain

import "time"

// Delivery is one queued attempt to send one event to one registration. The
// dispatcher (internal/platform/eventbus) creates these; the sender drains
// them.
type Delivery struct {
	ID         uint64
	BusinessID int64
	WebhookID  int64
	EventID    uint64
	// Attempts is the count after the claim, so a delivery on its last
	// permitted try already reads Attempts == MaxAttempts.
	Attempts    int
	MaxAttempts int
}

// Registration is the subscriber a delivery is addressed to.
type Registration struct {
	ID       int64
	URL      string
	Secret   string
	IsActive bool
}

// Event is the outbox row a delivery was created from.
type Event struct {
	ID          uint64
	Type        string
	SubjectType string
	SubjectID   int64
	Payload     []byte
	OccurredAt  time.Time
}
