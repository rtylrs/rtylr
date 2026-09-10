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

// DeliveryFilter narrows a delivery listing. A zero WebhookID and an empty
// Status mean no filter rather than match nothing.
type DeliveryFilter struct {
	WebhookID int64
	Status    string
	Limit     int32
	Offset    int64
}

// DeliveryRecord is a delivery attempt as the owner of the registration sees
// it, with the event type resolved so a row means something on its own.
type DeliveryRecord struct {
	// ID and EventID are uint64 because both queue tables key on
	// BIGINT UNSIGNED. They narrow to int64 at the contract boundary.
	ID             uint64
	WebhookID      int64
	EventID        uint64
	EventType      string
	Status         string
	Attempts       int
	MaxAttempts    int
	ResponseStatus int
	LastError      string
	NextAttemptAt  *time.Time
	DeliveredAt    *time.Time
	CreatedAt      time.Time
}
