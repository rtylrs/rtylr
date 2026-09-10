package domain

// Run is one queued execution of one automation against one event. The
// dispatcher (internal/platform/eventbus) creates these; the executor drains
// them.
type Run struct {
	ID           uint64
	BusinessID   int64
	AutomationID int64
	EventID      uint64
	// Attempts is the count after this run was claimed, so a run on its last
	// permitted try already reads Attempts == MaxAttempts.
	Attempts    int
	MaxAttempts int
}

// Event is the outbox row a run was created from, as much of it as the
// executor needs.
type Event struct {
	ID          uint64
	Type        string
	SubjectType string
	SubjectID   int64
	// Payload is the producer's JSON, read for whatever an action needs.
	Payload []byte
}

// CustomerContact is a send_email recipient.
type CustomerContact struct {
	ID    int64
	Email string
	Name  string
}
